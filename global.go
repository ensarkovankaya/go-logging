package go_logging

import (
	"context"
	"go.uber.org/zap"
	"reflect"

	"github.com/ensarkovankaya/go-logging/core"
	"github.com/ensarkovankaya/go-logging/integrations/console"
	"github.com/ensarkovankaya/go-logging/integrations/sentry"
)

type Field = core.Field
type ctxKeyType string

const CtxKey ctxKeyType = "__logger__"

var (
	globalLogger *Logger
)

// G returns the global logger instance.
func G() *Logger {
	return globalLogger
}

func ReplaceGlobal(logger *Logger) {
	*globalLogger = *logger
}

// F is a helper function to create a Field.
// If the value is a pointer, it dereferences it to get the underlying value.
// If the value is an error, it uses zap.NamedError to log it with the provided key.
// If the value is not an error or a pointer, it uses zap.Any to log the value.
func F(key string, value any) Field {
	switch v := value.(type) {
	case error:
		return zap.NamedError(key, v)
	}
	// If value is a pointer, dereference it to get the underlying value
	vf := reflect.ValueOf(value)
	if vf.Kind() == reflect.Ptr {
		value = vf.Elem().Interface()
	}
	return zap.Any(key, value)
}

// E is a helper function to create a Field for errors
func E(err error) Field {
	return zap.NamedError("error", err)
}

func L(ctx context.Context) *Logger {
	logger := FromContext(ctx)
	if logger == nil {
		return globalLogger
	}
	return logger
}

// Named returns a function that creates a named logger.
func Named(names ...string) func(ctx context.Context) *Logger {
	return func(ctx context.Context) *Logger {
		logger := L(ctx)
		for _, name := range names {
			logger = logger.Named(name)
		}
		return logger
	}
}

func FromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(CtxKey).(*Logger)
	if !ok || logger == nil {
		return globalLogger
	}
	return logger
}

func WithContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, CtxKey, logger)
}

func Flush() error {
	if globalLogger == nil {
		return nil
	}
	return globalLogger.Flush(context.Background())
}

func init() {
	logger := &Logger{}
	if console.IsActive() {
		logger.AddTransport(console.New())
	}
	if sentry.IsActive() {
		logger.AddTransport(sentry.New())
	}
	globalLogger = logger
}
