// Package logger provides standardized logging interfaces and implementations
// for consistent logging across applications. It supports:
//   - Leveled logging (Debug, Info, Warn, Error)
//   - Structured logging with fields
//   - Context-aware logging
package logger

import "context"

// Logger is the fundamental logging interface used throughout the application.
// Implementations must be thread-safe.
type Logger interface {
	// Debug logs a debug-level message with optional key-value fields.
	Debug(msg string, args ...any)

	// Info logs an info-level message with optional key-value fields.
	Info(msg string, args ...any)

	// Warn logs a warning-level message with optional key-value fields.
	Warn(msg string, args ...any)

	// Error logs a message at Error level. Arguments are handled as key-value pairs.
	Error(msg string, args ...any)

	// WithContext returns a new Logger that inherits context for distributed tracing.
	WithContext(ctx context.Context) Logger

	// WithField returns a new Logger with a single additional field.
	WithField(key string, value any) Logger

	// WithFields returns a new Logger with additional structured fields.
	WithFields(fields map[string]any) Logger
}

// Level represents the severity of a log message.
type Level int

const (
	// DebugLevel logs verbose debugging information.
	DebugLevel Level = iota

	// InfoLevel logs routine operational messages (e.g., "service started").
	InfoLevel

	// WarnLevel indicates non-critical issues that may require investigation.
	WarnLevel

	// ErrorLevel indicates critical failures that require immediate action.
	ErrorLevel
)

// String returns the canonical name of the log level (e.g., "DEBUG").
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
