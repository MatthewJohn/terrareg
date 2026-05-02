package logging

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// ZeroLogger wraps zerolog.Logger to implement the logging interface
type ZeroLogger struct {
	logger zerolog.Logger
}

// NewZeroLogger creates a new ZeroLogger from a zerolog.Logger
func NewZeroLogger(zl zerolog.Logger) Logger {
	return &ZeroLogger{logger: zl}
}

// Debug returns a debug level event
func (z *ZeroLogger) Debug() Event {
	return &zeroEvent{e: z.logger.Debug()}
}

// Info returns an info level event
func (z *ZeroLogger) Info() Event {
	return &zeroEvent{e: z.logger.Info()}
}

// Warn returns a warn level event
func (z *ZeroLogger) Warn() Event {
	return &zeroEvent{e: z.logger.Warn()}
}

// Error returns an error level event
func (z *ZeroLogger) Error() Event {
	return &zeroEvent{e: z.logger.Error()}
}

// With creates a logger with additional context
func (z *ZeroLogger) With() Logger {
	// For zerolog, With() returns a logger that can add context
	// We'll just return the same logger since zerolog handles context differently
	return z
}

// WithContext attaches the logger to a Go context
func (z *ZeroLogger) WithContext(ctx context.Context) context.Context {
	return z.logger.WithContext(ctx)
}

// zeroEvent wraps zerolog.Event to implement the Event interface
type zeroEvent struct {
	e *zerolog.Event
}

// Str adds a string field to the event
func (z *zeroEvent) Str(key, val string) Event {
	return &zeroEvent{e: z.e.Str(key, val)}
}

// Strs adds a string slice field to the event
func (z *zeroEvent) Strs(key string, vals []string) Event {
	return &zeroEvent{e: z.e.Strs(key, vals)}
}

// Int adds an integer field to the event
func (z *zeroEvent) Int(key string, i int) Event {
	return &zeroEvent{e: z.e.Int(key, i)}
}

// Int64 adds an int64 field to the event
func (z *zeroEvent) Int64(key string, i int64) Event {
	return &zeroEvent{e: z.e.Int64(key, i)}
}

// Bool adds a boolean field to the event
func (z *zeroEvent) Bool(key string, b bool) Event {
	return &zeroEvent{e: z.e.Bool(key, b)}
}

// Dur adds a duration field to the event
func (z *zeroEvent) Dur(key string, d time.Duration) Event {
	return &zeroEvent{e: z.e.Dur(key, d)}
}

// Time adds a time field to the event
func (z *zeroEvent) Time(key string, t time.Time) Event {
	return &zeroEvent{e: z.e.Time(key, t)}
}

// Err adds an error field to the event
func (z *zeroEvent) Err(err error) Event {
	if err == nil {
		return z
	}
	return &zeroEvent{e: z.e.Err(err)}
}

// Msg sends the event with the given message
func (z *zeroEvent) Msg(msg string) {
	z.e.Msg(msg)
}
