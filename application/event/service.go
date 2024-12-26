package event

import (
	"fmt"
	"sync"

	domainconfig "github.com/rantuma/hue-dial/domain/config"
	"github.com/rantuma/hue-dial/domain/device"
	"github.com/rantuma/hue-dial/domain/ports"
	"github.com/rantuma/hue-dial/pkg/logger"
)

type (
	Service interface {
		Start() error
	}

	service struct {
		logger             logger.Logger
		lightController    ports.LightController
		eventSource        ports.EventSource
		config             domainconfig.Config
		buttonMappingsByID map[string]device.ButtonMapping
		dialMappingsByID   map[string]device.DialMapping
		state              state
	}

	state struct {
		mu            sync.RWMutex
		selectedLamps []string
	}
)

const (
	eventChannelBufferSize = 1_000_000
)

func New(
	log logger.Logger,
	lightController ports.LightController,
	eventSource ports.EventSource,
	config domainconfig.Config,
) Service {
	buttonMappingsByID := make(map[string]device.ButtonMapping)
	for _, btn := range config.Devices.Buttons {
		buttonMappingsByID[btn.ID] = btn
	}

	dialMappingsByID := make(map[string]device.DialMapping)
	for _, dial := range config.Devices.Dials {
		dialMappingsByID[dial.ID] = dial
	}

	return &service{
		logger:             log,
		lightController:    lightController,
		eventSource:        eventSource,
		config:             config,
		buttonMappingsByID: buttonMappingsByID,
		dialMappingsByID:   dialMappingsByID,
	}
}

func (svc *service) Start() error {
	eventChan := make(chan ports.DeviceEvent, eventChannelBufferSize)

	go func() {
		for evt := range eventChan {
			svc.dispatchEvent(evt)
		}
	}()

	err := svc.eventSource.SubscribeToEvents(eventChan)
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	return nil
}

func (svc *service) dispatchEvent(evt ports.DeviceEvent) {
	buttonMapping, buttonExists := svc.buttonMappingsByID[evt.DeviceID]
	if buttonExists && evt.ButtonEvent != nil {
		svc.handleButtonEvent(buttonMapping.Label, evt.ButtonEvent.EventType, buttonMapping.LampID)
		return
	}

	if dialMapping, exists := svc.dialMappingsByID[evt.DeviceID]; exists && evt.DialEvent != nil {
		err := svc.handleDial(evt.DialEvent.Direction, evt.DialEvent.Steps)
		if err != nil {
			svc.logger.Errorf("failed to handle dial %q: %q", dialMapping.Label, err.Error())
		}
		return
	}

	svc.logger.Debugf("unknown event: %#v", evt)
}

func (svc *service) handleButtonEvent(
	label string,
	eventType device.ButtonEventType,
	lampID string,
) {
	err := svc.handleClick(eventType, lampID)
	if err != nil {
		svc.logger.Errorf(
			"failed to handle %q %q: %v",
			label, eventType, err,
		)
	}
}

func (svc *service) handleDial(direction device.Direction, steps int) error {
	brightnessDelta := float64(steps) * svc.config.Brightness.DeltaFactor
	if direction == device.DirectionCounterClockwise {
		brightnessDelta = -brightnessDelta
	}

	for _, selectedLamp := range svc.state.selectedLamps {
		svc.logger.Debugf("set brightness of lamp %q to %d", selectedLamp, int(brightnessDelta))

		err := svc.lightController.SetTunedBrightnessOfLamp(selectedLamp, int(brightnessDelta))
		if err != nil {
			return fmt.Errorf(
				"failed to set brightness of lamp %q: %w",
				selectedLamp,
				err,
			)
		}
	}

	return nil
}

func (svc *service) handleClick(eventType device.ButtonEventType, id string) error {
	svc.state.mu.Lock()
	defer svc.state.mu.Unlock()

	switch eventType {
	case device.ButtonEventShortRelease:
		svc.state.selectedLamps = []string{id}

		err := svc.lightController.IdentifyLight(id)
		if err != nil {
			return fmt.Errorf("failed to identify lamp %q: %w", id, err)
		}

		svc.logger.Debugf("event %q handled", eventType)

	case device.ButtonEventLongPress:
		svc.state.selectedLamps = []string{id}

		err := svc.lightController.TurnOffLamp(id)
		if err != nil {
			return fmt.Errorf("failed to turn lamp %q off: %w", id, err)
		}

		svc.logger.Debugf("event %q handled", eventType)

	case device.ButtonEventInitialPress, device.ButtonEventRepeat, device.ButtonEventLongRelease:
		svc.logger.Debugf("event %q not handled", eventType)
	}

	return nil
}
