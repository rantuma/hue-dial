package hue

import "strings"

type (
	Bridges []Bridge

	Bridge struct {
		ID                string `json:"id"`
		Internalipaddress string `json:"internalipaddress"`
		Port              int    `json:"port"`
	}

	LightsResponse struct {
		Errors Errors `json:"errors"`
		Data   Lights `json:"data"`
	}
)

type (
	// Errors is a list of API-level error responses. Exported for tests.
	Errors []Error

	// Error is an API-level error response. Exported for tests.
	Error struct {
		Description string `json:"description"`
	}
)

func (errs Errors) String() string {
	if len(errs) == 0 {
		return "No errors"
	}

	descriptions := make([]string, len(errs))
	for i, err := range errs {
		descriptions[i] = err.Description
	}

	return strings.Join(descriptions, "; ")
}

type (
	Owner struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	}

	Metadata struct {
		Name       string `json:"name"`
		Archetype  string `json:"archetype"`
		FixedMired int    `json:"fixed_mired"`
		Function   string `json:"function"`
	}

	ProductData struct {
		Function string `json:"function"`
	}

	Dimming struct {
		Brightness  float64 `json:"brightness"`
		MinDimLevel float64 `json:"min_dim_level"`
		Mode        string  `json:"mode"`
	}

	Dynamics struct {
		Status       string   `json:"status"`
		StatusValues []string `json:"status_values"`
		Speed        float64  `json:"speed"`
		SpeedValid   bool     `json:"speed_valid"`
	}

	Alert struct {
		ActionValues []string `json:"action_values"`
	}

	Signaling struct {
		SignalValues []string `json:"signal_values"`
	}

	Effects struct {
		StatusValues []string `json:"status_values"`
		Status       string   `json:"status"`
		EffectValues []string `json:"effect_values"`
	}

	action struct {
		EffectValues []string `json:"effect_values"`
	}

	status struct {
		Effect       string   `json:"effect"`
		EffectValues []string `json:"effect_values"`
	}

	EffectsV2 struct {
		Action action `json:"action"`
		Status status `json:"status"`
	}

	typeOnPowerup struct {
		Mode string `json:"mode"`
		On   typeOn `json:"on"`
	}

	typeOn struct {
		On bool `json:"on"`
	}

	Powerup struct {
		Preset     string        `json:"preset"`
		Configured bool          `json:"configured"`
		On         typeOnPowerup `json:"on"`
		Dimming    Dimming       `json:"dimming"`
	}

	Lights []Light

	Light struct {
		ID          string      `json:"id"`
		IDV1        string      `json:"id_v1"`
		Owner       Owner       `json:"owner"`
		Metadata    Metadata    `json:"metadata"`
		ProductData ProductData `json:"product_data"`
		ServiceID   int         `json:"service_id"`
		On          typeOn      `json:"on"`
		Dimming     Dimming     `json:"dimming"`
		Dynamics    Dynamics    `json:"dynamics"`
		Alert       Alert       `json:"alert"`
		Signaling   Signaling   `json:"signaling"`
		Mode        string      `json:"mode"`
		Effects     Effects     `json:"effects"`
		EffectsV2   EffectsV2   `json:"effects_v2"`
		Powerup     Powerup     `json:"powerup"`
		Type        string      `json:"type"`
	}
)

type (
	identifyBody struct {
		Identify identify `json:"identify"`
	}

	identify struct {
		Action string `json:"action"`
	}

	lightsResponseBody struct {
		Errors Errors `json:"errors"`
	}

	setLightStateBody struct {
		On           typeOn        `json:"on"`
		DimmingDelta *dimmingDelta `json:"dimming_delta,omitempty"`
	}

	dimmingDelta struct {
		Action          dimmingDeltaAction `json:"action,omitempty"`
		BrightnessDelta int                `json:"brightness_delta,omitempty"`
	}

	dimmingDeltaAction string
)

const (
	dimmingDeltaActionUp   dimmingDeltaAction = "up"
	dimmingDeltaActionDown dimmingDeltaAction = "down"
)

// --- Event stream types ---

type (
	events []event

	event struct {
		Data []eventData `json:"data"`
	}

	buttonReport struct {
		Event buttonEvent `json:"event"`
	}

	buttonEvent string

	button struct {
		ButtonReport buttonReport `json:"button_report"`
	}

	rotation struct {
		Direction string `json:"direction"`
		Steps     int    `json:"steps"`
	}

	rotaryReport struct {
		Rotation rotation `json:"rotation"`
	}

	relativeRotary struct {
		RotaryReport rotaryReport `json:"rotary_report"`
	}

	temperatureReport struct {
		Temperature float64 `json:"temperature"`
	}

	temperature struct {
		Temperature       float64           `json:"temperature"`
		TemperatureReport temperatureReport `json:"temperature_report"`
		TemperatureValid  bool              `json:"temperature_valid"`
	}

	motionReport struct {
		Motion bool `json:"motion"`
	}

	motion struct {
		Motion       bool         `json:"motion"`
		MotionReport motionReport `json:"motion_report"`
		MotionValid  bool         `json:"motion_valid"`
	}

	lightLevelReport struct {
		LightLevel int `json:"light_level"`
	}

	lightData struct {
		LightLevel       int              `json:"light_level"`
		LightLevelReport lightLevelReport `json:"light_level_report"`
		LightLevelValid  bool             `json:"light_level_valid"`
	}

	eventData struct {
		ID             string         `json:"id"`
		Button         button         `json:"button"`
		RelativeRotary relativeRotary `json:"relative_rotary"`
		Temperature    temperature    `json:"temperature"`
		Motion         motion         `json:"motion"`
		Dimming        Dimming        `json:"dimming"`
		Light          lightData      `json:"light"`
	}
)

// --- Device discovery types (GET /resource/device) ---

type (
	// ServiceRef is a reference to a service offered by a device.
	ServiceRef struct {
		Rid   string `json:"rid"`
		Rtype string `json:"rtype"`
	}

	DeviceProductData struct {
		ManufacturerName string `json:"manufacturer_name"`
		ModelID          string `json:"model_id"`
		ProductArchetype string `json:"product_archetype"`
		ProductName      string `json:"product_name"`
		SoftwareVersion  string `json:"software_version"`
		Certified        bool   `json:"certified"`
	}

	Device struct {
		ID          string            `json:"id"`
		Type        string            `json:"type"`
		Metadata    Metadata          `json:"metadata"`
		ProductData DeviceProductData `json:"product_data"`
		Services    []ServiceRef      `json:"services"`
	}

	Devices []Device

	DevicesResponse struct {
		Errors Errors  `json:"errors"`
		Data   Devices `json:"data"`
	}
)

// --- Button resource types (GET /resource/button) ---

type (
	ButtonMetadata struct {
		ControlID int `json:"control_id"`
	}

	ButtonResource struct {
		ID       string         `json:"id"`
		Type     string         `json:"type"`
		Owner    Owner          `json:"owner"`
		Metadata ButtonMetadata `json:"metadata"`
	}

	ButtonResources []ButtonResource

	ButtonsResponse struct {
		Errors Errors          `json:"errors"`
		Data   ButtonResources `json:"data"`
	}
)

// --- Room types (GET /resource/room) ---

type (
	Room struct {
		ID       string       `json:"id"`
		Type     string       `json:"type"`
		Metadata Metadata     `json:"metadata"`
		Children []ServiceRef `json:"children"`
	}

	Rooms []Room

	RoomsResponse struct {
		Errors Errors `json:"errors"`
		Data   Rooms  `json:"data"`
	}
)

// --- Pairing types (POST /api) ---

type (
	createUserRequest struct {
		DeviceType        string `json:"devicetype"`
		GenerateClientKey bool   `json:"generateclientkey"`
	}

	// CreateUserSuccess holds the username (API key) returned after pairing.
	CreateUserSuccess struct {
		Username  string `json:"username"`
		ClientKey string `json:"clientkey"`
	}

	// CreateUserResult represents a single item in the pairing response array.
	CreateUserResult struct {
		Success *CreateUserSuccess `json:"success,omitempty"`
		Error   *CreateUserError   `json:"error,omitempty"`
	}

	// CreateUserError represents a pairing error (e.g. link button not pressed).
	CreateUserError struct {
		Type        int    `json:"type"`
		Address     string `json:"address"`
		Description string `json:"description"`
	}

	// CreateUserResponse is the array returned by POST /api.
	CreateUserResponse []CreateUserResult
)
