package http

import "context"

type contextKeyType string

var ContextDisableKey contextKeyType = "_http_logging_disable_"

// WithLoggingDisable sets a value in the context to disable the HTTP transport logging.
func WithLoggingDisable(ctx context.Context, disabled bool) context.Context {
	return context.WithValue(ctx, ContextDisableKey, disabled)
}

// GetLoggingDisabled retrieves the value from the context to determine if HTTP transport logging is disabled.
func GetLoggingDisabled(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if value, ok := ctx.Value(ContextDisableKey).(bool); ok {
		return value
	}
	return false
}
