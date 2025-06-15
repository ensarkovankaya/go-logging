package console

import (
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ensarkovankaya/go-logging/core"
)

type ZapConfigOption = func(cfg *zap.Config)
type Level = zapcore.Level

var (
	logLevel   = core.LevelDebug
	callerSkip = 2
	debug      = false
)

var (
	envLogLevel      = "CONSOLE_LOG_LEVEL"
	envLogCallerSkip = "CONSOLE_LOG_CALLER_SKIP"
	envDebug         = "CONSOLE_DEBUG"
)

func IsActive() bool {
	return logLevel != core.LevelDisabled
}

func Initialize(cfgOptions ...ZapConfigOption) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(getZapLevel(logLevel))
	cfg.Development = debug
	options := []zap.Option{
		zap.AddStacktrace(zapcore.ErrorLevel),
	}
	if callerSkip != 0 {
		options = append(options, zap.AddCallerSkip(callerSkip))
	}
	for _, opt := range cfgOptions {
		opt(&cfg)
	}
	return cfg.Build(options...)
}

func init() {
	if os.Getenv(envLogLevel) != "" {
		if value, err := core.ParseLevel(os.Getenv(envLogLevel)); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid CONSOLE_LOG_LEVEL value: %s, using default %d\n", os.Getenv(envLogLevel), logLevel)
		} else {
			logLevel = value
		}
	}
	if os.Getenv(envLogCallerSkip) != "" {
		if value, err := strconv.Atoi(os.Getenv(envLogCallerSkip)); err == nil && callerSkip >= 0 {
			callerSkip = value
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid CONSOLE_LOG_CALLER_SKIP value: %s, using default %d\n", os.Getenv(envLogCallerSkip), callerSkip)
		}
	}
	if os.Getenv(envDebug) != "" {
		if value, err := strconv.ParseBool(os.Getenv(envDebug)); err == nil {
			debug = value
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid CONSOLE_DEBUG value: %s, using default %t\n", os.Getenv(envDebug), debug)
		}
	}
}

func getZapLevel(level core.Level) zapcore.Level {
	switch level {
	case core.LevelDebug:
		return zapcore.DebugLevel
	case core.LevelInfo:
		return zapcore.InfoLevel
	case core.LevelWarning:
		return zapcore.WarnLevel
	case core.LevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InvalidLevel
	}
}
