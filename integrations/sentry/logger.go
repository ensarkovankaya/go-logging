package sentry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/getsentry/sentry-go/attribute"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"time"

	"github.com/ensarkovankaya/go-logging/core"
)

var (
	defaultLogLevel        = zapcore.InfoLevel
	defaultEventLevel      = zapcore.ErrorLevel
	defaultBreadcrumbLevel = zapcore.DebugLevel
)

var (
	envSentryLogLevel        = "SENTRY_LOG_LEVEL"
	envSentryEventLevel      = "SENTRY_EVENT_LEVEL"
	envSentryBreadcrumbLevel = "SENTRY_BREADCRUMB_LEVEL"
)

const Type = "sentry"

type Option func(l *Logger)

type Logger struct {
	LogLevel        zapcore.Level
	EventLevel      zapcore.Level
	BreadcrumbLevel zapcore.Level
	Hub             *sentry.Hub
	FlushTimeout    time.Duration
	Name            string
	NowFunc         func() time.Time
}

func New(opts ...Option) *Logger {
	logger := &Logger{
		LogLevel:        defaultLogLevel,
		EventLevel:      defaultEventLevel,
		BreadcrumbLevel: defaultBreadcrumbLevel,
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
		scope.SetExtras(l.transform(fields...))
	})
	return _l
}

// WithContext returns a new context with the Sentry hub attached.
func (l *Logger) WithContext(ctx context.Context) context.Context {
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

func (l *Logger) Debug(ctx context.Context, msg string, fields ...core.Field) {
	if l.CanCaptureEvent(zapcore.DebugLevel) {
		l.CaptureEvent(ctx, zapcore.DebugLevel, msg, fields...)
	} else if l.CanBreadcrumb(zapcore.DebugLevel) {
		l.AddBreadcrumb(ctx, zapcore.DebugLevel, msg, fields...)
	}
	if l.CanLog(zapcore.DebugLevel) {
		l.Log(ctx, zapcore.DebugLevel, msg, fields...)
	}
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...core.Field) {
	if l.CanCaptureEvent(zapcore.InfoLevel) {
		l.CaptureEvent(ctx, zapcore.InfoLevel, msg, fields...)
	} else if l.CanBreadcrumb(zapcore.InfoLevel) {
		l.AddBreadcrumb(ctx, zapcore.InfoLevel, msg, fields...)
	}
	if l.CanLog(zapcore.InfoLevel) {
		l.Log(ctx, zapcore.InfoLevel, msg, fields...)
	}
}

func (l *Logger) Warning(ctx context.Context, msg string, fields ...core.Field) {
	if l.CanCaptureEvent(zapcore.WarnLevel) {
		l.CaptureEvent(ctx, zapcore.WarnLevel, msg, fields...)
	} else if l.CanBreadcrumb(zapcore.WarnLevel) {
		l.AddBreadcrumb(ctx, zapcore.WarnLevel, msg, fields...)
	}
	if l.CanLog(zapcore.WarnLevel) {
		l.Log(ctx, zapcore.WarnLevel, msg, fields...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...core.Field) {
	if l.CanCaptureEvent(zapcore.ErrorLevel) {
		l.CaptureEvent(ctx, zapcore.ErrorLevel, msg, fields...)
	} else if l.CanBreadcrumb(zapcore.ErrorLevel) {
		l.AddBreadcrumb(ctx, zapcore.ErrorLevel, msg, fields...)
	}
	if l.CanLog(zapcore.ErrorLevel) {
		l.Log(ctx, zapcore.ErrorLevel, msg, fields...)
	}
}

func (l *Logger) Flush(_ context.Context) error {
	if ok := l.Hub.Flush(l.FlushTimeout); !ok {
		return fmt.Errorf("failed to flush Sentry hub within %s", l.FlushTimeout.String())
	}
	return nil
}

func (l *Logger) SetEventLevel(level zapcore.Level) {
	l.EventLevel = level
}

func (l *Logger) SetLogLevel(level zapcore.Level) {
	l.LogLevel = level
}

func (l *Logger) SetBreadcrumbLevel(level zapcore.Level) {
	l.BreadcrumbLevel = level
}

func (l *Logger) Log(ctx context.Context, level zapcore.Level, msg string, fields ...core.Field) {
	logger := sentry.NewLogger(sentry.SetHubOnContext(ctx, l.getHub(ctx)))
	l.attachAttributes(logger, fields...)
	switch level {
	case zapcore.DebugLevel:
		logger.Debug(ctx, msg)
	case zapcore.InfoLevel:
		logger.Info(ctx, msg)
	case zapcore.WarnLevel:
		logger.Warn(ctx, msg)
	case zapcore.ErrorLevel:
		logger.Error(ctx, msg)
	default:
		return
	}
}

func (l *Logger) CanBreadcrumb(level zapcore.Level) bool {
	return l.BreadcrumbLevel <= level
}

func (l *Logger) CanCaptureEvent(level zapcore.Level) bool {
	return l.EventLevel <= level
}

func (l *Logger) CanLog(level zapcore.Level) bool {
	return l.LogLevel <= level
}

func (l *Logger) AddBreadcrumb(ctx context.Context, level zapcore.Level, msg string, fields ...core.Field) {
	l.getHub(ctx).AddBreadcrumb(&sentry.Breadcrumb{
		Level:     l.getSentryLevel(level),
		Message:   msg,
		Data:      l.transform(fields...),
		Timestamp: l.NowFunc(),
	}, nil)
}

func (l *Logger) CaptureEvent(ctx context.Context, level zapcore.Level, msg string, fields ...core.Field) {
	l.getHub(ctx).CaptureEvent(&sentry.Event{
		Level:     l.getSentryLevel(level),
		Message:   msg,
		Extra:     l.transform(fields...),
		Timestamp: l.NowFunc(),
	})
}

func (l *Logger) copy() *Logger {
	_l := *l
	_l.Hub = l.Hub.Clone()
	return &_l
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

func (l *Logger) attachAttributes(logger sentry.Logger, fields ...core.Field) {
	if len(fields) == 0 {
		return
	}
	if l.Name != "" {
		logger.SetAttributes(attribute.String("logger", l.Name))
	}
	attributeMap := l.transform(fields...)
	var err error
	var decoded []byte
	for key, value := range attributeMap {
		if decoded, err = json.Marshal(value); err != nil {
			decoded = []byte(fmt.Sprintf("[decode error]: %v", err))
		}
		logger.SetAttributes(attribute.String(key, string(decoded)))
	}
}

func (l *Logger) transform(fields ...core.Field) map[string]interface{} {
	transformed := make(map[string]interface{}, len(fields))
	for _, field := range fields {
		transformed[field.Key] = field.Value
	}
	return transformed
}

func (l *Logger) getHub(ctx context.Context) *sentry.Hub {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		return hub
	}
	return l.Hub
}

func init() {
	if os.Getenv(envSentryEventLevel) != "" {
		if level, err := zapcore.ParseLevel(os.Getenv(envSentryEventLevel)); err == nil {
			defaultEventLevel = level
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid SENTRY_EVENT_LEVEL: %s, using default: %s\n", os.Getenv(envSentryEventLevel), defaultEventLevel.String())
		}
	}
	if os.Getenv(envSentryLogLevel) != "" {
		if level, err := zapcore.ParseLevel(os.Getenv(envSentryLogLevel)); err == nil {
			defaultLogLevel = level
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid SENTRY_LOG_LEVEL: %s, using default: %s\n", os.Getenv(envSentryLogLevel), defaultLogLevel.String())
		}
	}
	if os.Getenv(envSentryBreadcrumbLevel) != "" {
		if level, err := zapcore.ParseLevel(os.Getenv(envSentryBreadcrumbLevel)); err == nil {
			defaultBreadcrumbLevel = level
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid SENTRY_BREADCRUMB_LEVEL: %s, using default: %s\n", os.Getenv(envSentryBreadcrumbLevel), defaultBreadcrumbLevel.String())
		}
	}
}
