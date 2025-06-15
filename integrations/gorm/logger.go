package gorm

import (
	"context"
	"time"

	gormLogger "gorm.io/gorm/logger"

	"github.com/ensarkovankaya/go-logging/core"
)

type Logger struct {
	Logger core.Interface
}

func New(logger core.Interface) gormLogger.Interface {
	return &Logger{
		Logger: logger,
	}
}

func (l *Logger) LogMode(_ gormLogger.LogLevel) gormLogger.Interface {
	return l
}

func (l *Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Info(ctx, msg, core.F("data", data))
}

func (l *Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Warning(ctx, msg, core.F("data", data))
}

func (l *Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Error(ctx, msg, core.F("data", data))
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	query, rowsAffected := fc()
	elapsed := time.Since(begin)
	l.Logger.Debug(ctx, "", core.F("elapsed", elapsed.Seconds()), core.F("query", query), core.F("rowsAffected", rowsAffected), core.E(err))
}
