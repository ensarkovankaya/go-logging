package batch

import (
	"context"
	"errors"

	"github.com/ensarkovankaya/go-logging/core"
)

const Type = "batch"

// Logger aggregates multiple core.Interface instances.
type Logger struct {
	integrations []core.Interface
}

func New() *Logger {
	return &Logger{
		integrations: make([]core.Interface, 0),
	}
}

func (l *Logger) Type() string {
	return Type
}

func (l *Logger) Named(name string) core.Interface {
	_l := &Logger{
		integrations: make([]core.Interface, 0, len(l.integrations)),
	}
	for _, integration := range l.integrations {
		_l.integrations = append(_l.integrations, integration.Named(name))
	}
	return _l
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	for _, integration := range l.integrations {
		ctx = integration.WithContext(ctx)
	}
	return ctx
}

func (l *Logger) With(fields ...core.Field) core.Interface {
	_t := &Logger{
		integrations: make([]core.Interface, 0, len(l.integrations)),
	}
	for _, integration := range l.integrations {
		_t.integrations = append(_t.integrations, integration.With(fields...))
	}
	return _t
}

func (l *Logger) Clone() core.Interface {
	_t := &Logger{
		integrations: make([]core.Interface, 0, len(l.integrations)),
	}
	for _, integration := range l.integrations {
		_t.integrations = append(_t.integrations, integration.Clone())
	}
	return _t
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...core.Field) {
	for _, integration := range l.integrations {
		integration.Debug(ctx, msg, fields...)
	}
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...core.Field) {
	for _, integration := range l.integrations {
		integration.Info(ctx, msg, fields...)
	}
}

func (l *Logger) Warning(ctx context.Context, msg string, fields ...core.Field) {
	for _, integration := range l.integrations {
		integration.Warning(ctx, msg, fields...)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...core.Field) {
	for _, integration := range l.integrations {
		integration.Error(ctx, msg, fields...)
	}
}

func (l *Logger) AddIntegration(integration core.Interface) {
	if integration != nil {
		l.integrations = append(l.integrations, integration)
	}
}

func (l *Logger) ReplaceIntegration(_type string, integration core.Interface) {
	for i, existing := range l.integrations {
		if existing.Type() == _type {
			l.integrations[i] = integration
			return
		}
	}
	l.AddIntegration(integration)
}

func (l *Logger) GetIntegration(_type string) core.Interface {
	for _, integration := range l.integrations {
		if integration.Type() == _type {
			return integration
		}
	}
	return nil
}

func (l *Logger) Flush(ctx context.Context) error {
	errs := make([]error, 0)
	for _, integration := range l.integrations {
		if err := integration.Flush(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
