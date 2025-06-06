package console

import (
	"context"
	"fmt"
	"go.uber.org/zap"

	"github.com/ensarkovankaya/go-logging/core"
)

type ctxKeyType string

const CtxKey ctxKeyType = "__console_logger__"

const Type = "console"

type Option func(l *Logger)

type Logger struct {
	Transport *zap.Logger
	Name      string
}

func New(opts ...Option) *Logger {
	zapLogger, err := Initialize()
	if err != nil {
		panic(fmt.Sprintf("Logger initialization failed: %v", err))
	}
	logger := &Logger{
		Transport: zapLogger,
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
	l.Transport = l.Transport.With(l.serialize(fields...)...)
	return l
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, CtxKey, l.Transport)
}

func (l *Logger) Clone() core.Logger {
	_l := *l
	_l.Transport = l.Transport.WithOptions()
	return &_l
}

func (l *Logger) Named(name string) core.Logger {
	_l := *l
	_l.Transport = l.Transport.WithOptions().Named(name)
	_l.Name = l.Transport.Name()
	return &_l
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...core.Field) {
	l.getLogger(ctx).Debug(msg, l.serialize(fields...)...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...core.Field) {
	l.getLogger(ctx).Info(msg, l.serialize(fields...)...)
}

func (l *Logger) Warning(ctx context.Context, msg string, fields ...core.Field) {
	l.getLogger(ctx).Warn(msg, l.serialize(fields...)...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...core.Field) {
	l.getLogger(ctx).Error(msg, l.serialize(fields...)...)
}

func (l *Logger) Flush(ctx context.Context) error {
	return l.getLogger(ctx).Sync()
}

func (l *Logger) getLogger(ctx context.Context) *zap.Logger {
	transport, ok := ctx.Value(CtxKey).(*zap.Logger)
	if !ok || transport == nil {
		return l.Transport
	}
	return transport
}

func (l *Logger) serialize(fields ...core.Field) []zap.Field {
	serialized := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		serialized = append(serialized, zap.Any(field.Key, field.Value))
	}
	return serialized
}
