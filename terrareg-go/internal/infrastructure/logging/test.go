package logging

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestLogger uses testing.T.Log() which only shows output on test failure
// or when running with go test -v
type TestLogger struct {
	t *testing.T
}

// NewTestLogger creates a new TestLogger for the given test
func NewTestLogger(t *testing.T) Logger {
	return &TestLogger{t: t}
}

// Debug returns a debug level event
func (t *TestLogger) Debug() Event {
	return &testEvent{t: t.t, level: "DEBUG"}
}

// Info returns an info level event
func (t *TestLogger) Info() Event {
	return &testEvent{t: t.t, level: "INFO"}
}

// Warn returns a warn level event
func (t *TestLogger) Warn() Event {
	return &testEvent{t: t.t, level: "WARN"}
}

// Error returns an error level event
func (t *TestLogger) Error() Event {
	return &testEvent{t: t.t, level: "ERROR"}
}

// With creates a logger with additional context
func (t *TestLogger) With() Logger {
	return t
}

// WithContext attaches the logger to a Go context
func (t *TestLogger) WithContext(ctx context.Context) context.Context {
	// For tests, we don't need to attach the logger to context
	return ctx
}

// testEvent represents a log event for testing
type testEvent struct {
	t      *testing.T
	level  string
	fields []string
	err    error
}

// Str adds a string field to the event
func (e *testEvent) Str(key, val string) Event {
	e.fields = append(e.fields, fmt.Sprintf("%s=%s", key, val))
	return e
}

// Strs adds a string slice field to the event
func (e *testEvent) Strs(key string, vals []string) Event {
	e.fields = append(e.fields, fmt.Sprintf("%s=%s", key, strings.Join(vals, ",")))
	return e
}

// Int adds an integer field to the event
func (e *testEvent) Int(key string, i int) Event {
	e.fields = append(e.fields, fmt.Sprintf("%s=%d", key, i))
	return e
}

// Int64 adds an int64 field to the event
func (e *testEvent) Int64(key string, i int64) Event {
	e.fields = append(e.fields, fmt.Sprintf("%s=%d", key, i))
	return e
}

// Bool adds a boolean field to the event
func (e *testEvent) Bool(key string, b bool) Event {
	e.fields = append(e.fields, fmt.Sprintf("%s=%t", key, b))
	return e
}

// Dur adds a duration field to the event
func (e *testEvent) Dur(key string, d time.Duration) Event {
	e.fields = append(e.fields, fmt.Sprintf("%s=%s", key, d))
	return e
}

// Time adds a time field to the event
func (e *testEvent) Time(key string, t time.Time) Event {
	e.fields = append(e.fields, fmt.Sprintf("%s=%s", key, t.Format(time.RFC3339)))
	return e
}

// Err adds an error field to the event
func (e *testEvent) Err(err error) Event {
	if err != nil {
		e.err = err
	}
	return e
}

// Msg sends the event with the given message
func (e *testEvent) Msg(msg string) {
	// Build the log message
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", e.level))
	parts = append(parts, msg)

	// Add fields
	if len(e.fields) > 0 {
		parts = append(parts, strings.Join(e.fields, " "))
	}

	// Add error if present
	if e.err != nil {
		parts = append(parts, fmt.Sprintf("error=%s", e.err))
	}

	// Log using t.Log() - only shown on failure or with -v flag
	e.t.Log(strings.Join(parts, " "))
}
