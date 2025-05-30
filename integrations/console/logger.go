package console

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"strings"

	"go-logging/core"
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
	l.Transport = l.Transport.With(fields...)
	return l
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	if l == nil || l.Transport == nil {
		return ctx
	}
	return context.WithValue(ctx, CtxKey, l.Transport)
}

func (l *Logger) Clone() core.Logger {
	if l == nil {
		return nil
	}
	_l := *l
	_l.Transport = l.Transport.WithOptions(zap.AddCallerSkip(1))
	return &_l
}

func (l *Logger) Named(name string) core.Logger {
	if l == nil {
		return nil
	}
	name = strings.ReplaceAll(name, " ", "_")
	transport := l.Transport.WithOptions(zap.AddCallerSkip(1)).Named(name)
	_l := *l
	_l.Transport = transport
	_l.Name = transport.Name()
	return &_l
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...core.Field) {
	l.fromContext(ctx).Debug(msg, fields...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...core.Field) {
	l.fromContext(ctx).Info(msg, fields...)
}

func (l *Logger) Warning(ctx context.Context, msg string, fields ...core.Field) {
	l.fromContext(ctx).Warn(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...core.Field) {
	l.fromContext(ctx).Error(msg, fields...)
}

func (l *Logger) Flush(ctx context.Context) error {
	return l.fromContext(ctx).Sync()
}

func (l *Logger) fromContext(ctx context.Context) *zap.Logger {
	if l == nil {
		return nil
	}
	transport, ok := ctx.Value(CtxKey).(*zap.Logger)
	if !ok || transport == nil {
		return l.Transport
	}
	return transport
}
