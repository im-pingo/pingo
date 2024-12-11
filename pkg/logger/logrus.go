package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	*logrus.Entry
}

// Config holds configuration options for the logger
// This can include log level, output format, etc.
type Config struct {
	Level           Level
	OutputFormat    string
	Output          io.Writer
	TimestampFormat string
	// Add more configuration options as needed
}

// DefaultConfig returns a Config with default settings
func DefaultConfig() Config {
	return Config{
		Level:           InfoLevel,
		OutputFormat:    "text",
		TimestampFormat: "2006-01-02 15:04:05.999",
		Output:          os.Stdout,
		// Set other default options here
	}
}

type LogrusLoggerOption func(l *logrusLogger)

func WithLogrusEntry(entry *logrus.Entry) LogrusLoggerOption {
	return func(l *logrusLogger) {
		l.Entry = entry
	}
}

func WithLogrusConfig(config Config) LogrusLoggerOption {
	return func(l *logrusLogger) {
		ll := logrus.New()
		l.Entry = ll.WithFields(logrus.Fields{})
		l.Entry.Logger.SetLevel(mapLevel(config.Level))
		timestampFormat := config.TimestampFormat
		if config.OutputFormat == "json" {
			l.Entry.Logger.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: timestampFormat,
			})
		} else {
			l.Entry.Logger.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: timestampFormat,
			})
		}

		if config.Output == nil {
			l.Entry.Logger.SetOutput(os.Stdout)
		} else {
			l.Entry.Logger.SetOutput(config.Output)
		}
	}
}

// NewLogrusLoggerWithConfig creates a new Logger implementation using logrus with custom configuration
func NewLogrusLogger(options ...LogrusLoggerOption) *logrusLogger {
	ll := &logrusLogger{}

	for _, opt := range options {
		opt(ll)
	}

	return ll
}

func (l *logrusLogger) SetLevel(level Level) {
	l.Entry.Logger.SetLevel(mapLevel(level))
}

func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{l.Entry.WithField(key, value)}
}

func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &logrusLogger{l.Entry.WithFields(fields)}
}

func (l *logrusLogger) WithError(err error) Logger {
	return &logrusLogger{l.Entry.WithError(err)}
}

// mapLevel maps the custom logger Level to logrus.Level
func mapLevel(level Level) logrus.Level {
	switch level {
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	case FatalLevel:
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}

func init() {
	// Ensure the default logger outputs to console
	defaultLogger := NewLogrusLogger(WithLogrusConfig(DefaultConfig()))
	SetDefaultLogger(defaultLogger)
}
