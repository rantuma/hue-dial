package device

type (
	ButtonEventType string

	Direction string

	ButtonMapping struct {
		ID       string
		Label    string
		LampID   string
		LampName string
	}

	DialMapping struct {
		ID    string
		Label string
	}
)

const (
	ButtonEventInitialPress ButtonEventType = "initial_press"
	ButtonEventShortRelease ButtonEventType = "short_release"
	ButtonEventLongPress    ButtonEventType = "long_press"
	ButtonEventRepeat       ButtonEventType = "repeat"
	ButtonEventLongRelease  ButtonEventType = "long_release"

	DirectionClockwise        Direction = "clock_wise"
	DirectionCounterClockwise Direction = "counter_clock_wise"
)
