package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger interface for WinRM plugin logging
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// pluginLogger wraps logrus.Logger and implements Logger interface
type pluginLogger struct {
	*logrus.Logger
}

// Implement Logger interface methods
func (l *pluginLogger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
}

func (l *pluginLogger) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

func (l *pluginLogger) Info(args ...interface{}) {
	l.Logger.Info(args...)
}

func (l *pluginLogger) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
}

func (l *pluginLogger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
}

func (l *pluginLogger) Warnf(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
}

func (l *pluginLogger) Error(args ...interface{}) {
	l.Logger.Error(args...)
}

func (l *pluginLogger) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

// New creates a new logger with specified level and format
func New(level, format string) Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	// Set log level
	switch level {
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "verbose":
		log.SetLevel(logrus.TraceLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// Set log format
	switch format {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
		})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z",
		})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z",
		})
	}

	return &pluginLogger{Logger: log}
}
