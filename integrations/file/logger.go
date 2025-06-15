package file

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ensarkovankaya/go-logging/integrations/console"
)

type Option func(*Logger)

type Logger struct {
	*console.Logger
	File *os.File
	Path string
}

func New(filePath string, options ...Option) *Logger {
	logger := &Logger{
		Logger: console.New(),
		Path:   filePath,
	}
	if logger.Path != "" {
		file, err := os.OpenFile(logger.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic(fmt.Errorf("failed to open log file: %w", err))
		}
		logger.File = file
	}

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(cfg)
	writer := zapcore.AddSync(logger.File)
	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)
	logger.Transport = zap.New(core)

	for _, opt := range options {
		opt(logger)
	}
	return logger
}
