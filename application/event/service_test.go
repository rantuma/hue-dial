package event_test

import (
	"testing"

	"github.com/rantuma/hue-dial/application/event"
	domainconfig "github.com/rantuma/hue-dial/domain/config"
	"github.com/rantuma/hue-dial/domain/device"
	"github.com/rantuma/hue-dial/domain/ports"
	"github.com/rantuma/hue-dial/pkg/logger"

	"github.com/stretchr/testify/assert"
)

type (
	mockLogger struct{}

	mockLightController struct {
		brightnessCalls []brightnessCall
	}

	brightnessCall struct {
		id    string
		delta int
	}

	mockEventSource struct{}
)

func (m *mockLogger) Level() logger.Level   { return 0 }
func (m *mockLogger) SetLevel(logger.Level) {}
func (m *mockLogger) Debug(...any)          {}
func (m *mockLogger) Debugf(string, ...any) {}
func (m *mockLogger) Info(...any)           {}
func (m *mockLogger) Infof(string, ...any)  {}
func (m *mockLogger) Warn(...any)           {}
func (m *mockLogger) Warnf(string, ...any)  {}
func (m *mockLogger) Error(...any)          {}
func (m *mockLogger) Errorf(string, ...any) {}
func (m *mockLogger) Fatal(...any)          {}
func (m *mockLogger) Fatalf(string, ...any) {}
func (m *mockLogger) Panic(...any)          {}
func (m *mockLogger) Panicf(string, ...any) {}

func (m *mockLightController) IdentifyLight(string) error { return nil }
func (m *mockLightController) TurnOffLamp(string) error   { return nil }

func (m *mockLightController) SetTunedBrightnessOfLamp(id string, delta int) error {
	m.brightnessCalls = append(m.brightnessCalls, brightnessCall{id: id, delta: delta})
	return nil
}

func (m *mockEventSource) SubscribeToEvents(chan ports.DeviceEvent) error {
	return nil
}

func newTestService(lc *mockLightController, inverted bool) event.Service {
	cfg := domainconfig.Config{
		Brightness: domainconfig.BrightnessConfig{DeltaFactor: 0.5},
		Devices: domainconfig.DevicesConfig{
			Buttons: []device.ButtonMapping{
				{ID: "btn-1", Label: "test button", LampID: "lamp-1"},
			},
			Dials: []device.DialMapping{
				{ID: "dial-1", Label: "test dial", Inverted: inverted},
			},
		},
	}

	return event.New(&mockLogger{}, lc, &mockEventSource{}, cfg)
}

func TestService_DialBrightness(t *testing.T) {
	tests := []struct {
		name          string
		direction     device.Direction
		steps         int
		inverted      bool
		expectedDelta int
	}{
		{
			name:          "clockwise normal",
			direction:     device.DirectionClockwise,
			steps:         10,
			inverted:      false,
			expectedDelta: 5,
		},
		{
			name:          "counter-clockwise normal",
			direction:     device.DirectionCounterClockwise,
			steps:         10,
			inverted:      false,
			expectedDelta: -5,
		},
		{
			name:          "clockwise inverted",
			direction:     device.DirectionClockwise,
			steps:         10,
			inverted:      true,
			expectedDelta: -5,
		},
		{
			name:          "counter-clockwise inverted",
			direction:     device.DirectionCounterClockwise,
			steps:         10,
			inverted:      true,
			expectedDelta: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := &mockLightController{}
			svc := newTestService(lc, tt.inverted)

			err := svc.Start()
			assert.NoError(t, err) //nolint:testifylint // project uses assert

			svc.HandleEvent(ports.DeviceEvent{
				DeviceID: "btn-1",
				ButtonEvent: &ports.DeviceButtonEvent{
					EventType: device.ButtonEventShortRelease,
				},
			})

			svc.HandleEvent(ports.DeviceEvent{
				DeviceID: "dial-1",
				DialEvent: &ports.DeviceDialEvent{
					Direction: tt.direction,
					Steps:     tt.steps,
				},
			})

			assert.Len(t, lc.brightnessCalls, 1)
			assert.Equal(t, "lamp-1", lc.brightnessCalls[0].id)
			assert.Equal(t, tt.expectedDelta, lc.brightnessCalls[0].delta)
		})
	}
}
