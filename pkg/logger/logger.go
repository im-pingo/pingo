package logger

// Level represents the severity level of a log message
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// Logger defines the interface that any logging implementation must satisfy
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithFields(fields map[string]interface{}) Logger
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger
	SetLevel(level Level)
}

var defaultLogger Logger

// SetDefaultLogger sets the default logger implementation
func SetDefaultLogger(l Logger) {
	defaultLogger = l
}

// GetDefaultLogger returns the current default logger
func GetDefaultLogger() Logger {
	return defaultLogger
}

// func init() {
// 	// Initialize the default logger with default configuration
// 	defaultLogger = NewLogrusLoggerWithConfig(DefaultConfig())
// }

// Debug logs a message at DebugLevel using the default logger
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Debugf logs a formatted message at DebugLevel using the default logger
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Info logs a message at InfoLevel using the default logger
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

// Infof logs a formatted message at InfoLevel using the default logger
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Warn logs a message at WarnLevel using the default logger
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warnf logs a formatted message at WarnLevel using the default logger
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Error logs a message at ErrorLevel using the default logger
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf logs a formatted message at ErrorLevel using the default logger
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatal logs a message at FatalLevel using the default logger
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Fatalf logs a formatted message at FatalLevel using the default logger
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// WithFields returns a new logger with the specified fields
func WithFields(fields map[string]interface{}) Logger {
	return defaultLogger.WithFields(fields)
}

// WithField returns a new logger with the specified field
func WithField(key string, value interface{}) Logger {
	return defaultLogger.WithField(key, value)
}

// WithError returns a new logger with the specified error
func WithError(err error) Logger {
	return defaultLogger.WithError(err)
}

// SetLevel sets the log level of the default logger
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}
