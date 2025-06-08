package file

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"

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

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	writer := zapcore.AddSync(logger.File)
	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)
	logger.Transport = zap.New(core)

	for _, opt := range options {
		opt(logger)
	}
	return logger
}
