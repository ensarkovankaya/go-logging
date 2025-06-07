package main

import (
	"context"
	"github.com/ensarkovankaya/go-logging/core"
)

type noopLogger struct {
}

func (l *noopLogger) Type() string {
	return "noop"
}

func (l *noopLogger) Named(_ string) core.Logger {
	return l
}

func (l *noopLogger) Clone() core.Logger {
	return l
}

func (l *noopLogger) WithContext(ctx context.Context) context.Context {
	return ctx
}

func (l *noopLogger) With(_ ...core.Field) core.Logger {
	return l
}

func (l *noopLogger) Debug(_ context.Context, _ string, _ ...core.Field) {
}

func (l *noopLogger) Info(_ context.Context, _ string, _ ...core.Field) {
}

func (l *noopLogger) Warning(_ context.Context, _ string, _ ...core.Field) {
}

func (l *noopLogger) Error(_ context.Context, _ string, _ ...core.Field) {
}

func (l *noopLogger) Flush(_ context.Context) error {
	return nil
}
