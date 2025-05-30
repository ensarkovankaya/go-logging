package go_logging

import (
	"context"
	"errors"

	"github.com/ensarkovankaya/go-logging/core"
)

type Logger struct {
	Transports []core.Logger
}

func (l *Logger) Named(name string) *Logger {
	_l := &Logger{
		Transports: make([]core.Logger, 0, len(l.Transports)),
	}
	for _, transport := range l.Transports {
		_l.Transports = append(_l.Transports, transport.Named(name))
	}
	return _l
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	for _, transport := range l.Transports {
		ctx = transport.WithContext(ctx)
	}
	return ctx
}

func (l *Logger) With(fields ...Field) *Logger {
	_t := &Logger{
		Transports: make([]core.Logger, 0, len(l.Transports)),
	}
	for _, transport := range l.Transports {
		_t.Transports = append(_t.Transports, transport.With(fields...))
	}
	return _t
}

func (l *Logger) Clone() *Logger {
	_t := &Logger{
		Transports: make([]core.Logger, 0, len(l.Transports)),
	}
	for _, transport := range l.Transports {
		_t.Transports = append(_t.Transports, transport.Clone())
	}
	return _t
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...Field) {
	for _, transport := range l.Transports {
		transport.Debug(ctx, msg, fields...)
	}
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...Field) {
	for _, transport := range l.Transports {
		transport.Info(ctx, msg, fields...)
	}
}

func (l *Logger) Warning(ctx context.Context, msg string, fields ...Field) {
	for _, transport := range l.Transports {
		transport.Warning(ctx, msg, fields...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...Field) {
	for _, transport := range l.Transports {
		transport.Error(ctx, msg, fields...)
	}
}

func (l *Logger) AddTransport(transport core.Logger) {
	if transport != nil {
		l.Transports = append(l.Transports, transport)
	}
}

func (l *Logger) GetTransport(_type string) core.Logger {
	for _, transport := range l.Transports {
		if transport.Type() == _type {
			return transport
		}
	}
	return nil

}

func (l *Logger) Flush(ctx context.Context) error {
	errs := make([]error, 0)
	for _, transport := range l.Transports {
		if err := transport.Flush(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
