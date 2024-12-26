package logging

import (
	"math"

	pkglogger "github.com/rantuma/hue-dial/pkg/logger"
	"github.com/sirupsen/logrus"
)

var logrusLogger = logrus.New() //nolint:gochecknoglobals // package-level logger instance

type (
	logger struct {
		level pkglogger.Level
	}
)

// New creates a Logger that writes to logrus and Azure Application Insights.
func New(
	level pkglogger.Level,
) pkglogger.Logger {
	log := &logger{
		level: level,
	}
	logrusLogger.SetLevel(logrus.Level(level))
	return log
}

func (l *logger) Level() pkglogger.Level {
	if logrusLogger.Level > math.MaxUint8 {
		l.Warnf(
			"log level %d exceeds maximum allowed value %d\n",
			logrusLogger.Level, math.MaxUint8,
		)
		return pkglogger.Level(math.MaxUint8)
	}
	return pkglogger.Level(logrusLogger.Level) //#nosec G115
}

func (l *logger) SetLevel(level pkglogger.Level) {
	logrusLogger.SetLevel(logrus.Level(level))
}

func (l *logger) Debug(args ...any) {
	logrusLogger.Debug(args...)
}

func (l *logger) Debugf(format string, args ...any) {
	logrusLogger.Debugf(format, args...)
}

func (l *logger) Info(args ...any) {
	logrusLogger.Info(args...)
}

func (l *logger) Infof(format string, args ...any) {
	logrusLogger.Infof(format, args...)
}

func (l *logger) Warn(args ...any) {
	logrusLogger.Warn(args...)
}

func (l *logger) Warnf(format string, args ...any) {
	logrusLogger.Warnf(format, args...)
}

func (l *logger) Error(args ...any) {
	logrusLogger.Error(args...)
}

func (l *logger) Errorf(format string, args ...any) {
	logrusLogger.Errorf(format, args...)
}

func (l *logger) Fatal(args ...any) {
	logrusLogger.Fatal(args...)
}

func (l *logger) Fatalf(format string, args ...any) {
	logrusLogger.Fatalf(format, args...)
}

func (l *logger) Panic(args ...any) {
	logrusLogger.Panic(args...)
}

func (l *logger) Panicf(format string, args ...any) {
	logrusLogger.Panicf(format, args...)
}
