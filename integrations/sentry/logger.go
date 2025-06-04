package sentry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/getsentry/sentry-go/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"time"

	"github.com/ensarkovankaya/go-logging/core"
)

var (
	DefaultLevel           = zapcore.ErrorLevel
	DefaultBreadcrumbLevel = zapcore.InfoLevel
)

var (
	envSentryLogLevel        = "SENTRY_LOG_LEVEL"
	envSentryBreadcrumbLevel = "SENTRY_BREADCRUMB_LEVEL"
)

const Type = "sentry"

type Option func(l *Logger)

type Logger struct {
	Level           zapcore.Level
	BreadcrumbLevel zapcore.Level
	Hub             *sentry.Hub
	FlushTimeout    time.Duration
	Name            string
	NowFunc         func() time.Time
}

func New(opts ...Option) *Logger {
	logger := &Logger{
		Level:           DefaultLevel,
		BreadcrumbLevel: DefaultBreadcrumbLevel,
		Hub:             globalHub,
		FlushTimeout:    FLushTimeout,
		NowFunc:         time.Now,
	}
	for _, opt := range opts {
		opt(logger)
	}
	return logger
}

func (l *Logger) Type() string {
	return Type
}

func (l *Logger) With(fields ...core.Field) core.Logger {
	if len(fields) == 0 {
		return l.copy()
	}
	_l := l.copy()
	_l.Hub.WithScope(func(scope *sentry.Scope) {
		scope.SetExtras(l.encode(fields...))
	})
	return _l
}

// WithContext returns a new context with the Sentry hub attached.
func (l *Logger) WithContext(ctx context.Context) context.Context {
	if l.Hub == nil {
		return ctx
	}
	return sentry.SetHubOnContext(ctx, l.Hub)
}

func (l *Logger) Clone() core.Logger {
	return l.copy()
}

func (l *Logger) Named(name string) core.Logger {
	name = strings.ReplaceAll(name, " ", "_")
	if l.Name != "" {
		name = l.Name + "." + name
	}
	_l := l.copy()
	_l.Name = name
	_l.Hub.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("logger", name)
	})
	return _l
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	if l.canLog(zapcore.DebugLevel) {
		l.captureEvent(ctx, zapcore.DebugLevel, msg, fields...)
	} else if l.canBreadcrumb(zapcore.DebugLevel) {
		l.addBreadcrumb(ctx, zapcore.DebugLevel, msg, fields...)
	}
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if l.canLog(zapcore.InfoLevel) {
		l.captureEvent(ctx, zapcore.InfoLevel, msg, fields...)
	} else if l.canBreadcrumb(zapcore.InfoLevel) {
		l.addBreadcrumb(ctx, zapcore.InfoLevel, msg, fields...)
	}
}

func (l *Logger) Warning(ctx context.Context, msg string, fields ...zap.Field) {
	if l.canLog(zapcore.WarnLevel) {
		l.captureEvent(ctx, zapcore.WarnLevel, msg, fields...)
	} else if l.canBreadcrumb(zapcore.WarnLevel) {
		l.addBreadcrumb(ctx, zapcore.WarnLevel, msg, fields...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	if l.canLog(zapcore.ErrorLevel) {
		l.captureEvent(ctx, zapcore.ErrorLevel, msg, fields...)
	} else if l.canBreadcrumb(zapcore.ErrorLevel) {
		l.addBreadcrumb(ctx, zapcore.ErrorLevel, msg, fields...)
	}
}

func (l *Logger) Flush(_ context.Context) error {
	if ok := l.Hub.Flush(l.FlushTimeout); !ok {
		return fmt.Errorf("failed to flush Sentry hub within %s", l.FlushTimeout.String())
	}
	return nil
}

func (l *Logger) SetLevel(level zapcore.Level) {
	l.Level = level
}

func (l *Logger) SetBreadcrumbLevel(level zapcore.Level) {
	l.BreadcrumbLevel = level
}

func (l *Logger) copy() *Logger {
	_l := *l
	_l.Hub = l.Hub.Clone()
	return &_l
}

func (l *Logger) canBreadcrumb(level zapcore.Level) bool {
	return level >= l.BreadcrumbLevel
}

func (l *Logger) canLog(level zapcore.Level) bool {
	return l.Level <= level
}

func (l *Logger) addBreadcrumb(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	l.getHub(ctx).AddBreadcrumb(&sentry.Breadcrumb{
		Level:     l.getSentryLevel(level),
		Message:   msg,
		Data:      l.encode(fields...),
		Timestamp: l.NowFunc(),
	}, nil)
}

func (l *Logger) getSentryLevel(level zapcore.Level) sentry.Level {
	switch level {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel:
		return sentry.LevelError
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return sentry.LevelFatal
	default:
		return ""
	}
}

func (l *Logger) captureEvent(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	l.getHub(ctx).CaptureEvent(&sentry.Event{
		Level:     l.getSentryLevel(level),
		Message:   msg,
		Extra:     l.encode(fields...),
		Timestamp: l.NowFunc(),
	})
}

func (l *Logger) attachAttributes(logger sentry.Logger, fields ...zap.Field) {
	if len(fields) == 0 {
		return
	}
	attributeMap := l.encode(fields...)
	var err error
	var decoded []byte
	for key, value := range attributeMap {
		if decoded, err = json.Marshal(value); err != nil {
			decoded = []byte(fmt.Sprintf("[decode error]: %v", err))
		}
		logger.SetAttributes(attribute.String(key, string(decoded)))
	}
}

func (l *Logger) encode(fields ...zap.Field) map[string]interface{} {
	encoder := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(encoder)
	}
	return encoder.Fields
}

func (l *Logger) getHub(ctx context.Context) *sentry.Hub {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		return hub
	}
	return l.Hub
}

func init() {
	if os.Getenv(envSentryLogLevel) != "" {
		if level, err := zapcore.ParseLevel(os.Getenv(envSentryLogLevel)); err == nil {
			DefaultLevel = level
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid SENTRY_LOG_LEVEL: %s, using default level: %s\n", os.Getenv(envSentryLogLevel), DefaultLevel)
		}
	}
	if os.Getenv(envSentryBreadcrumbLevel) != "" {
		if level, err := zapcore.ParseLevel(os.Getenv(envSentryBreadcrumbLevel)); err == nil {
			DefaultBreadcrumbLevel = level
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid SENTRY_BREADCRUMB_LEVEL: %s, using default level: %s\n", os.Getenv(envSentryBreadcrumbLevel), DefaultBreadcrumbLevel)
		}
	}
}
