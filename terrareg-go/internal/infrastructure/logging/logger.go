package logging

import (
	"context"
	"time"
)

// Logger interface for structured logging
// This abstraction allows different implementations for production vs testing
type Logger interface {
	// Log levels return an Event for chaining
	Debug() Event
	Info() Event
	Warn() Event
	Error() Event

	// With creates a logger with additional context
	With() Logger

	// WithContext attaches the logger to a Go context
	WithContext(ctx context.Context) context.Context
}

// Event represents a log event with chained field setters
type Event interface {
	Str(key, val string) Event
	Strs(key string, vals []string) Event
	Int(key string, i int) Event
	Int64(key string, i int64) Event
	Bool(key string, b bool) Event
	Dur(key string, d time.Duration) Event
	Time(key string, t time.Time) Event
	Err(err error) Event
	Msg(msg string)
}
