package core

import (
	"context"
	"go.uber.org/zap"
)

type Field = zap.Field

type Logger interface {
	Type() string
	Named(string) Logger
	Clone() Logger
	WithContext(context.Context) context.Context
	With(...zap.Field) Logger
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warning(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Flush(ctx context.Context) error
}
