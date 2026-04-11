package config

import "github.com/rantuma/hue-dial/domain/device"

type (
	Config struct {
		Hue        HueConfig
		Brightness BrightnessConfig
		Devices    DevicesConfig
	}

	HueConfig struct {
		Key      string
		BridgeIP string
	}

	BrightnessConfig struct {
		DeltaFactor float64
	}

	DevicesConfig struct {
		Buttons []device.ButtonMapping
		Dials   []device.DialMapping
	}

	SetupConfig struct {
		BridgeIP string               `json:"bridge_ip"`
		APIKey   string               `json:"api_key"`
		Buttons  []SetupButtonMapping `json:"buttons"`
		Dial     SetupDialMapping     `json:"dial"`
	}

	SetupButtonMapping struct {
		ButtonID string `json:"button_id"`
		Label    string `json:"label"`
		LampID   string `json:"lamp_id"`
		LampName string `json:"lamp_name"`
	}

	SetupDialMapping struct {
		DialID     string `json:"dial_id"`
		Label      string `json:"label"`
		InvertDial bool   `json:"invert_dial"`
	}
)

const defaultDeltaFactor = 0.3

func (sc SetupConfig) ToConfig() Config {
	buttons := make([]device.ButtonMapping, len(sc.Buttons))
	for i, b := range sc.Buttons {
		buttons[i] = device.ButtonMapping{
			ID:       b.ButtonID,
			Label:    b.Label,
			LampID:   b.LampID,
			LampName: b.LampName,
		}
	}

	return Config{
		Hue: HueConfig{
			Key:      sc.APIKey,
			BridgeIP: sc.BridgeIP,
		},
		Brightness: BrightnessConfig{
			DeltaFactor: defaultDeltaFactor,
		},
		Devices: DevicesConfig{
			Buttons: buttons,
			Dials: []device.DialMapping{
				{ID: sc.Dial.DialID, Label: sc.Dial.Label, Inverted: sc.Dial.InvertDial},
			},
		},
	}
}
