package console

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestLogger_LevelDisabled(t *testing.T) {
	logger := getLogger(t, LevelDisabled)
	ctx := context.Background()
	logger.Debug(ctx, "debug message", getFields()...)
	logger.Info(ctx, "info message", getFields()...)
	logger.Warning(ctx, "warning message", getFields()...)
	logger.Error(ctx, "error message", getFields()...)
}

func TestLogger_LevelDebug(t *testing.T) {
	logger := getLogger(t, LevelDebug)
	ctx := context.Background()
	logger.Debug(ctx, "debug message", getFields()...)
	logger.Info(ctx, "info message", getFields()...)
	logger.Warning(ctx, "warning message", getFields()...)
	logger.Error(ctx, "error message", getFields()...)
}

func TestLogger_LevelInfo(t *testing.T) {
	logger := getLogger(t, LevelInfo)
	ctx := context.Background()
	logger.Debug(ctx, "debug message", getFields()...)
	logger.Info(ctx, "info message", getFields()...)
	logger.Warning(ctx, "warning message", getFields()...)
	logger.Error(ctx, "error message", getFields()...)
}

func TestLogger_LevelWarning(t *testing.T) {
	logger := getLogger(t, LevelWarning)
	ctx := context.Background()
	logger.Debug(ctx, "debug message", getFields()...)
	logger.Info(ctx, "info message", getFields()...)
	logger.Warning(ctx, "warning message", getFields()...)
	logger.Error(ctx, "error message", getFields()...)
}

func TestLogger_LevelError(t *testing.T) {
	logger := getLogger(t, LevelError)
	ctx := context.Background()
	logger.Debug(ctx, "debug message", getFields()...)
	logger.Info(ctx, "info message", getFields()...)
	logger.Warning(ctx, "warning message", getFields()...)
	logger.Error(ctx, "error message", getFields()...)
}

func getLogger(t *testing.T, level Level) *Logger {
	zapLogger, err := Initialize(func(cfg *zap.Config) {
		cfg.Level = zap.NewAtomicLevelAt(level)
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
