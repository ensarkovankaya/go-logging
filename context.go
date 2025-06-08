package logging

import (
	"context"

	"github.com/ensarkovankaya/go-logging/core"
	"github.com/ensarkovankaya/go-logging/integrations/batch"
)

type ctxKeyType string

const CtxKey ctxKeyType = "__logger__"

// FromContext retrieves the logger from the context.
func FromContext(ctx context.Context) core.Interface {
	logger, ok := ctx.Value(CtxKey).(*batch.Logger)
	if ok {
		return logger
	}
	return nil
}

// WithContext adds a logger to the context.
func WithContext(ctx context.Context, logger core.Interface) context.Context {
	return context.WithValue(ctx, CtxKey, logger)
}
