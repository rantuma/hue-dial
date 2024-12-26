package logger

type (
	// Level represents a logging severity level.
	Level uint8

	// Logger is the logging port used by all layers.
	Logger interface {
		Level() Level
		SetLevel(level Level)

		Debug(args ...any)
		Debugf(format string, args ...any)

		Info(args ...any)
		Infof(format string, args ...any)

		Warn(args ...any)
		Warnf(format string, args ...any)

		Error(args ...any)
		Errorf(format string, args ...any)

		Fatal(args ...any)
		Fatalf(format string, args ...any)

		Panic(i ...any)
		Panicf(format string, args ...any)
	}
)

const (
	Panic Level = iota
	Fatal
	Error
	Warn
	Info
	Debug
	Trace
)

func (lv Level) String() string {
	switch lv {
	case Trace:
		return "TRACE"
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	case Panic:
		return "PANIC"
	default:
		return ""
	}
}
