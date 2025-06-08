package core

import (
	"context"
)

type Interface interface {
	Type() string
	Named(string) Interface
	Clone() Interface
	WithContext(context.Context) context.Context
	With(...Field) Interface
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warning(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Flush(ctx context.Context) error
}
