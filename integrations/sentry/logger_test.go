package sentry

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"sync"
	"testing"
	"time"

	"github.com/ensarkovankaya/go-logging/core"
)

type testCase struct {
	Name                string
	EventLevel          Level
	LogLevel            Level
	BreadcrumbLevel     Level
	ExpectedEvents      int
	ExpectedLogs        int
	ExpectedBreadcrumbs int
}

type MockTransport struct {
	T               *testing.T
	mu              *sync.Mutex
	EventCount      int
	LogCount        int
	BreadcrumbCount int
	events          []*sentry.Event
	lastEvent       *sentry.Event
}

func (t *MockTransport) Configure(options sentry.ClientOptions) {
	t.T.Logf("[MockTransport] Configure: %+v", options)
}

func (t *MockTransport) SendEvent(event *sentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.T.Logf("[MockTransport] SendEvent: %+v", event)
	t.events = append(t.events, event)
	if event.Type == "" {
		t.EventCount++
	}
	t.LogCount += len(event.Logs)
	t.BreadcrumbCount += len(event.Breadcrumbs)
	t.lastEvent = event
}

func (t *MockTransport) Flush(duration time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.T.Logf("[MockTransport] Flush: %+v", duration)
	return true
}

func (t *MockTransport) Events() []*sentry.Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.T.Logf("[MockTransport] Events: %v", len(t.events))
	return t.events
}

func (t *MockTransport) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.T.Logf("[MockTransport] Close")
}

func TestLogger_DefaultSettings(t *testing.T) {
	testLogger(t, testCase{
		Name:                "Debug",
		EventLevel:          defaultEventLevel,
		LogLevel:            defaultLogLevel,
		BreadcrumbLevel:     defaultBreadcrumbLevel,
		ExpectedEvents:      1,
		ExpectedLogs:        3,
		ExpectedBreadcrumbs: 3,
	})
}

func TestLogger_Debug(t *testing.T) {
	testLogger(t, testCase{
		Name:                "Debug",
		EventLevel:          LevelDebug,
		LogLevel:            LevelDebug,
		BreadcrumbLevel:     LevelDebug,
		ExpectedEvents:      4,
		ExpectedLogs:        4,
		ExpectedBreadcrumbs: 0,
	})
}

func testLogger(t *testing.T, _case testCase) {
	t.Run(_case.Name, func(t *testing.T) {
		ctx := context.Background()
		logger, transport := getLoggerForTest(t, func(l *Logger) {
			l.LogLevel = _case.LogLevel
			l.EventLevel = _case.EventLevel
			l.BreadcrumbLevel = _case.BreadcrumbLevel
		})
		logger.Debug(ctx, "Debug message", getFields()...)
		logger.Info(ctx, "Info message", getFields()...)
		logger.Warning(ctx, "Warning message", getFields()...)
		logger.Error(ctx, "Error message", getFields()...)
		if err := logger.Flush(ctx); err != nil {
			t.Errorf("Flush failed: %v", err)
		}
		if transport.EventCount != _case.ExpectedEvents {
			t.Errorf("Expected %d event, got %d", _case.ExpectedEvents, transport.EventCount)
		}
		if transport.LogCount != _case.ExpectedLogs {
			t.Errorf("Expected %d logs, got %d", _case.ExpectedLogs, transport.LogCount)
		}
		if transport.BreadcrumbCount != _case.ExpectedBreadcrumbs {
			t.Errorf("Expected %d breadcrumbs, got %d", _case.ExpectedBreadcrumbs, transport.BreadcrumbCount)
		}
	})
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

func getFields() []core.Field {
	type A struct {
		F1 string `json:"f_1"`
		F2 int
	}
	err := fmt.Errorf("some error")
	zone, _ := time.LoadLocation("Europe/Istanbul")
	timestamp := time.Date(2023, 10, 1, 12, 0, 0, 0, zone)
	twoMinutesTwentyThree := time.Minute*2 + time.Second*23

	return []core.Field{
		core.F("str", "value"),
		core.F("int", 1),
		core.F("float", 1.8),
		core.F("bool", true),
		core.F("nil", nil),
		core.F("struct", A{F1: "test", F2: 123}),
		core.F("struct_pointer", &A{F1: "test", F2: 123}),
		core.F("slice", []string{"a", "b", "c"}),
		core.F("map", map[string]int{"a": 1, "b": 2}),
		core.F("error", err),
		core.F("error_pointer", &err),
		core.F("context", context.Background()),
		core.F("context_with_value", context.WithValue(context.Background(), "key", "value")),
		core.F("timestamp", timestamp),
		core.F("duration", twoMinutesTwentyThree),
	}
}
