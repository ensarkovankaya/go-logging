package logging

import (
	"context"
	"fmt"

	"github.com/ensarkovankaya/go-logging/core"
	"github.com/ensarkovankaya/go-logging/integrations/batch"
	"github.com/ensarkovankaya/go-logging/integrations/console"
	"github.com/ensarkovankaya/go-logging/integrations/sentry"
)

type Field = core.Field
type Interface = core.Interface
type Level = core.Level

const (
	LevelDebug    = core.LevelDebug
	LevelInfo     = core.LevelInfo
	LevelWarning  = core.LevelWarning
	LevelError    = core.LevelError
	LevelDisabled = core.LevelDisabled
)

var (
	globalLogger *batch.Logger
)

var (
	autoConfigure bool
	autoConfigEnv = "LOGGER_AUTO_CONFIG"
)

// G returns the global logger instance.
func G() *batch.Logger {
	return globalLogger
}

func ReplaceGlobal(logger *batch.Logger) {
	globalLogger = logger
}

// F exports core.F
func F(key string, value any) Field {
	return core.F(key, value)
}

// E exports core.E
func E(err error) Field {
	return core.E(err)
}

// L returns the logger from the context.
func L(ctx context.Context) core.Interface {
	logger := FromContext(ctx)
	if logger == nil {
		return globalLogger
	}
	return logger
}

// Named returns a function that creates a named logger from the context.
// Example usage:
//
//	package integration
//
//	logger := logging.Named("integration")
//	func someFunc(ctx context.Context) {
//	    logger(ctx).Info(ctx, "This is a log message")
//		// this will log messages with the name "integration"
//		// {"logger":"integration","level":"info","msg":"This is a log message"}
//	}
func Named(name string) func(ctx context.Context) core.Interface {
	return func(ctx context.Context) core.Interface {
		return L(ctx).Named(name)
	}
}

func init() {
	var err error
	globalLogger = batch.New()
	if autoConfigure, err = core.ParseBool(autoConfigEnv, false, false); err != nil {
		panic(fmt.Sprintf("Failed to parse %v: %v\n", autoConfigEnv, err))
	}
	if autoConfigure {
		AutoConfigure()
	}
}

func AutoConfigure() {
	if console.IsActive() {
		globalLogger.AddIntegration(console.New())
	}
	if sentry.IsActive() {
		globalLogger.AddIntegration(sentry.New())
	}
}
