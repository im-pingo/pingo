package logger

import (
	"os"
	"testing"
)

func init() {
	// Initialize the default logger with default configuration
	logg := NewLogrusLogger(WithLogrusConfig(Config{
		Level:        DebugLevel,
		OutputFormat: "json",
	}))

	logg.Entry.Logger.SetOutput(os.Stdout)
	SetDefaultLogger(logg)
}

func TestLogrusLogger(t *testing.T) {
	// Create a new logrus logger
	// log := NewLogrusLogger()

	// // Set it as the default logger
	// SetDefaultLogger(log)

	// // Get the default logger
	// logger := GetDefaultLogger()

	// Set log level
	SetLevel(DebugLevel)

	// Basic logging
	Debug("This is a debug message")
	Info("This is an info message")
	Warn("This is a warning message")

	// Formatted logging
	Debugf("Debug message with %s", "formatting")
	Infof("Info message with %d", 123)

	// Logging with fields
	logger := WithFields(map[string]interface{}{
		"component": "test",
		"id":        123,
	})
	logger.Info("This message includes fields")
}
