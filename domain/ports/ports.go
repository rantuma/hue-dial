package ports

import (
	domainconfig "github.com/rantuma/hue-dial/domain/config"
	"github.com/rantuma/hue-dial/domain/device"
)

type (
	DeviceEvent struct {
		DeviceID    string
		ButtonEvent *DeviceButtonEvent
		DialEvent   *DeviceDialEvent
	}

	DeviceButtonEvent struct {
		EventType device.ButtonEventType
	}

	DeviceDialEvent struct {
		Direction device.Direction
		Steps     int
	}

	LightController interface {
		IdentifyLight(id string) error
		TurnOffLamp(id string) error
		SetTunedBrightnessOfLamp(id string, brightnessDelta int) error
	}

	EventSource interface {
		SubscribeToEvents(eventsChan chan DeviceEvent) error
	}

	ConfigStore interface {
		Exists() bool
		Load() (domainconfig.SetupConfig, error)
		Save(cfg domainconfig.SetupConfig) error
		Path() string
	}
)
