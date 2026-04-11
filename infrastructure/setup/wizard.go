package setup

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	domainconfig "github.com/rantuma/hue-dial/domain/config"
	"github.com/rantuma/hue-dial/domain/ports"
	"github.com/rantuma/hue-dial/infrastructure/hue"
	"github.com/rantuma/hue-dial/pkg/logger"

	"charm.land/huh/v2"
)

// Wizard runs the interactive setup wizard and returns the resulting config.
func Wizard(
	log logger.Logger,
	store ports.ConfigStore,
) (domainconfig.SetupConfig, error) {
	log.Info("")
	log.Info("═══════════════════════════════════════════")
	log.Info("  Hue Tap Dial Switch — Setup Wizard")
	log.Info("═══════════════════════════════════════════")
	log.Info("")

	adapter := hue.NewUnauthenticated()

	bridgeIP, err := discoverBridge(log, adapter)
	if err != nil {
		return domainconfig.SetupConfig{},
			fmt.Errorf("bridge discovery failed: %w", err)
	}

	adapter.SetBridgeIP(bridgeIP)
	log.Infof("Using bridge at %s", bridgeIP)

	pairResult, err := pairWithBridge(log, adapter)
	if err != nil {
		return domainconfig.SetupConfig{},
			fmt.Errorf("pairing failed: %w", err)
	}

	adapter.SetKey(pairResult.Username)

	setupCfg, err := discoverAndConfigure(
		log, adapter, bridgeIP, pairResult.Username,
	)
	if err != nil {
		return domainconfig.SetupConfig{}, err
	}

	err = store.Save(setupCfg)
	if err != nil {
		return domainconfig.SetupConfig{},
			fmt.Errorf("failed to save config: %w", err)
	}

	log.Info("")
	log.Infof("Configuration saved to %s", store.Path())
	log.Info("Setup complete!")
	log.Info("")

	return setupCfg, nil
}

func pairWithBridge(
	log logger.Logger,
	adapter *hue.Adapter,
) (*hue.CreateUserSuccess, error) {
	log.Info("")
	log.Info("Press the link button on your Hue Bridge.")
	log.Info("Waiting up to 30 seconds...")
	log.Info("")

	result, err := PairBridge(adapter, "hue-dial#setup")
	if err != nil {
		return nil, err
	}

	log.Info("Paired successfully! API key obtained.")

	return result, nil
}

func discoverAndConfigure(
	log logger.Logger,
	adapter *hue.Adapter,
	bridgeIP string,
	apiKey string,
) (domainconfig.SetupConfig, error) {
	devices, err := adapter.GetDevices()
	if err != nil {
		return domainconfig.SetupConfig{},
			fmt.Errorf("failed to get devices: %w", err)
	}

	tapDial, err := findTapDialSwitch(devices)
	if err != nil {
		return domainconfig.SetupConfig{}, err
	}

	log.Infof(
		"Found Tap Dial Switch: %s (%s)",
		tapDial.Metadata.Name,
		tapDial.ProductData.ProductName,
	)

	buttons, err := adapter.GetButtons()
	if err != nil {
		return domainconfig.SetupConfig{},
			fmt.Errorf("failed to get buttons: %w", err)
	}

	buttonIDs := extractButtonIDs(tapDial, buttons)
	if len(buttonIDs) < 4 { //nolint:mnd // Tap Dial has exactly 4 buttons
		return domainconfig.SetupConfig{},
			fmt.Errorf(
				"expected 4 buttons on Tap Dial Switch, found %d",
				len(buttonIDs),
			)
	}

	dialID := extractDialID(tapDial)
	if dialID == "" {
		return domainconfig.SetupConfig{},
			errors.New("no dial service found on Tap Dial Switch")
	}

	lights, err := adapter.GetLights()
	if err != nil {
		return domainconfig.SetupConfig{},
			fmt.Errorf("failed to get lights: %w", err)
	}

	if len(lights) == 0 {
		return domainconfig.SetupConfig{},
			errors.New("no lights found on the bridge")
	}

	rooms, err := adapter.GetRooms()
	if err != nil {
		return domainconfig.SetupConfig{},
			fmt.Errorf("failed to get rooms: %w", err)
	}

	lampSelections, err := selectLampsForButtons(lights, rooms)
	if err != nil {
		return domainconfig.SetupConfig{},
			fmt.Errorf("lamp selection failed: %w", err)
	}

	invertDial, err := promptInvertDial()
	if err != nil {
		return domainconfig.SetupConfig{},
			fmt.Errorf("invert dial prompt failed: %w", err)
	}

	return buildSetupConfig(
		bridgeIP, apiKey, buttonIDs, dialID, lampSelections, invertDial,
	), nil
}

func buildSetupConfig(
	bridgeIP string,
	apiKey string,
	buttonIDs []string,
	dialID string,
	lamps []lampChoice,
	invertDial bool,
) domainconfig.SetupConfig {
	labels := [4]string{
		"top left", "top right", "bottom left", "bottom right",
	}

	cfg := domainconfig.SetupConfig{
		BridgeIP: bridgeIP,
		APIKey:   apiKey,
		Dial: domainconfig.SetupDialMapping{
			DialID:     dialID,
			Label:      "brightness control",
			InvertDial: invertDial,
		},
	}

	for i, label := range labels {
		cfg.Buttons = append(cfg.Buttons, domainconfig.SetupButtonMapping{
			ButtonID: buttonIDs[i],
			Label:    label,
			LampID:   lamps[i].id,
			LampName: lamps[i].name,
		})
	}

	return cfg
}

// discoverBridge finds the bridge IP via auto-discovery or manual input.
func discoverBridge(
	log logger.Logger,
	adapter *hue.Adapter,
) (string, error) {
	log.Info("Discovering Hue bridges on your network...")

	bridges, err := adapter.GetBridges()
	if err != nil || len(bridges) == 0 {
		log.Info("Could not auto-discover a bridge.")

		return promptBridgeIP()
	}

	if len(bridges) == 1 {
		log.Infof("Found bridge: %s", bridges[0].Internalipaddress)
		return bridges[0].Internalipaddress, nil
	}

	return selectBridge(bridges)
}

func promptBridgeIP() (string, error) {
	var ip string

	err := huh.NewInput().
		Title("Bridge IP address").
		Placeholder("192.168.1.x").
		Value(&ip).
		Run()
	if err != nil {
		return "", fmt.Errorf("bridge IP input failed: %w", err)
	}

	if ip == "" {
		return "", errors.New("no bridge IP provided")
	}

	return ip, nil
}

func selectBridge(bridges hue.Bridges) (string, error) {
	opts := make([]huh.Option[string], len(bridges)+1)
	for i, b := range bridges {
		label := b.Internalipaddress + " (ID: " + b.ID + ")"
		opts[i] = huh.NewOption(label, b.Internalipaddress)
	}

	opts[len(bridges)] = huh.NewOption(
		"Enter IP manually...", "__manual__",
	)

	var selected string

	err := huh.NewSelect[string]().
		Title("Multiple bridges found — select one").
		Options(opts...).
		Value(&selected).
		Run()
	if err != nil {
		return "", fmt.Errorf("bridge selection failed: %w", err)
	}

	if selected == "__manual__" {
		return promptBridgeIP()
	}

	return selected, nil
}

type (
	lampChoice struct {
		id   string
		name string
	}
)

func promptInvertDial() (bool, error) {
	var inverted bool

	err := huh.NewConfirm().
		Title("Should the dial rotation direction be inverted?").
		Description("If yes, the rotation direction will be inverted.").
		Affirmative("Yes").
		Negative("No").
		Value(&inverted).
		Run()
	if err != nil {
		return false, fmt.Errorf("invert dial confirm failed: %w", err)
	}

	return inverted, nil
}

func selectLampsForButtons(
	lights hue.Lights,
	rooms hue.Rooms,
) ([]lampChoice, error) {
	deviceToRoom := make(map[string]string)
	for _, r := range rooms {
		for _, c := range r.Children {
			if c.Rtype == "device" {
				deviceToRoom[c.Rid] = r.Metadata.Name
			}
		}
	}

	labelFor := func(l hue.Light) string {
		if room, ok := deviceToRoom[l.Owner.Rid]; ok {
			return strings.ToLower(room + " - " + l.Metadata.Name)
		}
		return strings.ToLower(l.Metadata.Name)
	}

	sort.Slice(lights, func(i, j int) bool {
		return labelFor(lights[i]) < labelFor(lights[j])
	})

	opts := make([]huh.Option[string], len(lights))
	for idx, lt := range lights {
		label := lt.Metadata.Name
		if room, ok := deviceToRoom[lt.Owner.Rid]; ok {
			label = room + " - " + lt.Metadata.Name
		}

		if lt.On.On {
			label += " (on)"
		} else {
			label += " (off)"
		}

		opts[idx] = huh.NewOption(label, lt.ID)
	}

	nameByID := make(map[string]string, len(lights))
	for _, lt := range lights {
		nameByID[lt.ID] = lt.Metadata.Name
	}

	labels := [4]string{
		"top left", "top right", "bottom left", "bottom right",
	}

	selections := make([]string, 4) //nolint:mnd // 4 buttons

	groups := make([]*huh.Group, 4) //nolint:mnd // 4 buttons
	for i, label := range labels {
		groups[i] = huh.NewGroup(
			huh.NewSelect[string]().
				Title("Button \"" + label + "\" → select lamp").
				Options(opts...).
				Value(&selections[i]),
		)
	}

	form := huh.NewForm(groups...)

	if err := form.Run(); err != nil {
		return nil, err
	}

	choices := make([]lampChoice, 4) //nolint:mnd // 4 buttons
	for i, sel := range selections {
		choices[i] = lampChoice{
			id:   sel,
			name: nameByID[sel],
		}
	}

	return choices, nil
}

// findTapDialSwitch locates a device that has both button and relative_rotary services.
func findTapDialSwitch(devices hue.Devices) (hue.Device, error) {
	var candidates []hue.Device

	for _, dev := range devices {
		hasButton := false
		hasRotary := false

		for _, s := range dev.Services {
			switch s.Rtype {
			case "button":
				hasButton = true
			case "relative_rotary":
				hasRotary = true
			}
		}

		if hasButton && hasRotary {
			candidates = append(candidates, dev)
		}
	}

	if len(candidates) == 0 {
		return hue.Device{}, errors.New(
			"no Tap Dial Switch found (device with button + dial services)",
		)
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	// Multiple Tap Dials — let user select.
	opts := make([]huh.Option[string], len(candidates))
	for i, c := range candidates {
		opts[i] = huh.NewOption(
			fmt.Sprintf("%s (%s)", c.Metadata.Name, c.ProductData.ProductName),
			c.ID,
		)
	}

	var selectedID string
	err := huh.NewSelect[string]().
		Title("Multiple Tap Dial Switches found — select one").
		Options(opts...).
		Value(&selectedID).
		Run()
	if err != nil {
		return hue.Device{}, err
	}

	for _, c := range candidates {
		if c.ID == selectedID {
			return c, nil
		}
	}

	return hue.Device{}, errors.New("selected device not found")
}

// extractButtonIDs returns the button resource IDs for the given device, sorted
// by control_id (1=top-left, 2=top-right, 3=bottom-left, 4=bottom-right).
func extractButtonIDs(dev hue.Device, allButtons hue.ButtonResources) []string {
	deviceButtonRIDs := make(map[string]bool)
	for _, s := range dev.Services {
		if s.Rtype == "button" {
			deviceButtonRIDs[s.Rid] = true
		}
	}

	var matched []hue.ButtonResource
	for _, b := range allButtons {
		if deviceButtonRIDs[b.ID] {
			matched = append(matched, b)
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Metadata.ControlID < matched[j].Metadata.ControlID
	})

	ids := make([]string, len(matched))
	for i, b := range matched {
		ids[i] = b.ID
	}

	return ids
}

// extractDialID returns the relative_rotary service RID for the device.
func extractDialID(dev hue.Device) string {
	for _, s := range dev.Services {
		if s.Rtype == "relative_rotary" {
			return s.Rid
		}
	}

	return ""
}
