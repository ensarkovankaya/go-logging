package console

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"testing"
	"time"

	"github.com/ensarkovankaya/go-logging/core"
)

func TestLogger_LevelDisabled(t *testing.T) {
	logger := getLogger(t, core.LevelDisabled)
	ctx := context.Background()
	logger.Debug(ctx, "debug message", getFields()...)
	logger.Info(ctx, "info message", getFields()...)
	logger.Warning(ctx, "warning message", getFields()...)
	logger.Error(ctx, "error message", getFields()...)
}

func TestLogger_LevelDebug(t *testing.T) {
	logger := getLogger(t, core.LevelDebug)
	ctx := context.Background()
	logger.Debug(ctx, "debug message", getFields()...)
	logger.Info(ctx, "info message", getFields()...)
	logger.Warning(ctx, "warning message", getFields()...)
	logger.Error(ctx, "error message", getFields()...)
}

func TestLogger_LevelInfo(t *testing.T) {
	logger := getLogger(t, core.LevelInfo)
	ctx := context.Background()
	logger.Debug(ctx, "debug message", getFields()...)
	logger.Info(ctx, "info message", getFields()...)
	logger.Warning(ctx, "warning message", getFields()...)
	logger.Error(ctx, "error message", getFields()...)
}

func TestLogger_LevelWarning(t *testing.T) {
	logger := getLogger(t, core.LevelWarning)
	ctx := context.Background()
	logger.Debug(ctx, "debug message", getFields()...)
	logger.Info(ctx, "info message", getFields()...)
	logger.Warning(ctx, "warning message", getFields()...)
	logger.Error(ctx, "error message", getFields()...)
}

func TestLogger_LevelError(t *testing.T) {
	logger := getLogger(t, core.LevelError)
	ctx := context.Background()
	logger.Debug(ctx, "debug message", getFields()...)
	logger.Info(ctx, "info message", getFields()...)
	logger.Warning(ctx, "warning message", getFields()...)
	logger.Error(ctx, "error message", getFields()...)
}

func getLogger(t *testing.T, level core.Level) *Logger {
	zapLogger, err := Initialize(func(cfg *zap.Config) {
		cfg.Level = zap.NewAtomicLevelAt(getZapLevel(level))
	})
	if err != nil {
		t.Fatalf("failed to initialize zap: %v", err)
	}
	logger := New(func(l *Logger) {
		l.Transport = zapLogger
	})
	if logger == nil {
		t.Fatal("logger is nil")
	}
	return logger
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
