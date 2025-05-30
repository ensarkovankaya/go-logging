package sentry

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
	"testing"
	"time"
)

type MockTransport struct {
	T         *testing.T
	mu        *sync.Mutex
	logCount  int
	events    []*sentry.Event
	lastEvent *sentry.Event
}

func (t *MockTransport) Configure(options sentry.ClientOptions) {
	t.T.Logf("Configure: %+v", options)
}
func (t *MockTransport) SendEvent(event *sentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.T.Logf("SendEvent: %+v", event)
	t.events = append(t.events, event)
	t.logCount += len(event.Logs)
	t.lastEvent = event
}
func (t *MockTransport) Flush(duration time.Duration) bool {
	t.T.Logf("Flush: %+v", duration)
	return true
}
func (t *MockTransport) Events() []*sentry.Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.T.Logf("Events: %+v", t.events)
	return t.events
}
func (t *MockTransport) Close() {
	t.T.Logf("Close")
}

func TestLogger_Debug(t *testing.T) {
	testLogger(t, zap.DebugLevel, 4)
}

func TestLogger_Info(t *testing.T) {
	testLogger(t, zap.InfoLevel, 3)
}

func TestLogger_Warning(t *testing.T) {
	testLogger(t, zap.WarnLevel, 2)
}

func TestLogger_Error(t *testing.T) {
	testLogger(t, zap.ErrorLevel, 1)
}

func testLogger(t *testing.T, level zapcore.Level, expectedLogs int) {
	ctx := context.Background()
	logger, transport := getLoggerForTest(t, func(l *Logger) {
		l.Level = level
	})
	logger.Debug(ctx, "Debug message", getFields()...)
	logger.Info(ctx, "Info message", getFields()...)
	logger.Warning(ctx, "Warning message", getFields()...)
	logger.Error(ctx, "Error message", getFields()...)
	if err := logger.Flush(ctx); err != nil {
		t.Errorf("Flush failed: %v", err)
	}
	if transport.logCount != expectedLogs {
		t.Errorf("Expected %d logs, got %d", expectedLogs, transport.logCount)
	}
}

func getLoggerForTest(t *testing.T, opts ...Option) (*Logger, *MockTransport) {
	transport := &MockTransport{
		T:  t,
		mu: &sync.Mutex{},
	}
	hub := Initialize(func(opt *sentry.ClientOptions) {
		opt.Transport = transport
		opt.Debug = true
	})
	opts = append(opts, func(l *Logger) {
		l.Hub = hub
	})
	return New(opts...), transport
}

func getFields() []zap.Field {
	type A struct {
		F1 string `json:"f_1"`
		F2 int
	}
	err := fmt.Errorf("some error")
	zone, _ := time.LoadLocation("Europe/Istanbul")
	timestamp := time.Date(2023, 10, 1, 12, 0, 0, 0, zone)
	twoMinutesTwentyThree := time.Minute*2 + time.Second*23

	return []zap.Field{
		zap.Any("str", "value"),
		zap.Any("int", 1),
		zap.Any("float", 1.8),
		zap.Any("bool", true),
		zap.Any("nil", nil),
		zap.Any("struct", A{F1: "test", F2: 123}),
		zap.Any("struct_pointer", &A{F1: "test", F2: 123}),
		zap.Any("slice", []string{"a", "b", "c"}),
		zap.Any("map", map[string]int{"a": 1, "b": 2}),
		zap.Any("error", err),
		zap.Any("error_pointer", &err),
		zap.Any("context", context.Background()),
		zap.Any("context_with_value", context.WithValue(context.Background(), "key", "value")),
		zap.Any("timestamp", timestamp),
		zap.Any("duration", twoMinutesTwentyThree),
	}
}
