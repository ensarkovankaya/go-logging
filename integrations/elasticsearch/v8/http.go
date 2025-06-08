package elastic

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"

	"github.com/ensarkovankaya/go-logging"
)

const DefaultRequestIDHeader = "X-Request-ID"

type HttpTransportOption func(*HttpTransport)
type RequestIDSetter func(req *http.Request, header string) string

var DefaultRequestIDSetter RequestIDSetter = func(req *http.Request, header string) string {
	if req.Header.Get(header) == "" {
		req.Header.Set(header, uuid.NewString())
	}
	return req.Header.Get(header)
}

type httpTransportDisableKey string

var HttpTransportDisableKey httpTransportDisableKey = "_elastic_http_transport_disable"

// WithHttpTransportDisable sets a value in the context to disable the HTTP transport logging.
func WithHttpTransportDisable(ctx context.Context, disabled bool) context.Context {
	return context.WithValue(ctx, HttpTransportDisableKey, disabled)
}

// GetHttpTransportDisable retrieves the value from the context to determine if HTTP transport logging is disabled.
func GetHttpTransportDisable(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if value, ok := ctx.Value(HttpTransportDisableKey).(bool); ok {
		return value
	}
	return false
}

type HttpTransport struct {
	Transport       http.RoundTripper
	Logger          *Logger
	RequestIDSetter RequestIDSetter
	RequestIDHeader string
	NowFunc         func() time.Time
}

func NewHttpTransport(logger *Logger, options ...HttpTransportOption) *HttpTransport {
	t := &HttpTransport{
		Transport:       http.DefaultTransport,
		Logger:          logger,
		RequestIDSetter: DefaultRequestIDSetter,
		RequestIDHeader: DefaultRequestIDHeader,
		NowFunc:         time.Now,
	}
	for _, opt := range options {
		opt(t)
	}
	return t
}

func (t *HttpTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if GetHttpTransportDisable(request.Context()) {
		return t.Transport.RoundTrip(request)
	}
	requestID := t.RequestIDSetter(request, t.RequestIDHeader)
	serializedRequest, err := t.serializeRequest(request)
	if err != nil {
		t.Logger.Error(request.Context(), "Failed to serialize request", logging.F("requestID", requestID), logging.E(err))
		return nil, err
	}
	t.Logger.Info(request.Context(), "request", logging.F("requestID", requestID), logging.F("request", serializedRequest))
	startTime := t.NowFunc()
	response, err := t.Transport.RoundTrip(request)
	elapsed := t.NowFunc().Sub(startTime)
	if err != nil {
		t.Logger.Error(request.Context(), "HTTP error", logging.F("requestID", requestID), logging.E(err))
		return nil, err
	}
	serializedResponse, err := t.serializeResponse(response)
	if err != nil {
		t.Logger.Error(request.Context(), "Failed to serialize response", logging.F("requestID", requestID), logging.E(err))
		return nil, err
	}
	t.Logger.Info(request.Context(), "response", logging.F("requestID", requestID), logging.F("response", serializedResponse), logging.F("elapsed", elapsed.Seconds()))
	return response, nil
}

func (t *HttpTransport) serializeRequest(request *http.Request) (map[string]any, error) {
	if request == nil {
		return nil, nil
	}
	serialized := map[string]any{
		"method": request.Method,
		"url": map[string]any{
			"full":   request.URL.String(),
			"scheme": request.URL.Scheme,
			"host":   request.URL.Host,
			"path":   request.URL.Path,
			"query":  request.URL.RawQuery,
		},
		"headers":       t.serializeHeaders(request.Header),
		"proto":         request.Proto,
		"contentLength": request.ContentLength,
		"body":          nil,
	}
	if request.Body != nil {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to read body: %v", err), err)
		}
		serialized["body"] = string(body)
		// Reset the body so it can be read again later
		request.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	return serialized, nil
}

func (t *HttpTransport) serializeResponse(response *http.Response) (map[string]any, error) {
	if response == nil {
		return nil, nil
	}
	serialized := map[string]any{
		"statusCode":    response.StatusCode,
		"proto":         response.Proto,
		"headers":       t.serializeHeaders(response.Header),
		"contentLength": response.ContentLength,
		"body":          nil,
	}
	if response.Body != nil {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to read body: %v", err), err)
		}
		serialized["body"] = string(body)
		// Reset the body so it can be read again later
		response.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	return serialized, nil
}

func (t *HttpTransport) serializeHeaders(headers http.Header) map[string]any {
	serialized := make(map[string]any, len(headers))
	for key, values := range headers {
		if len(values) > 1 {
			// If there are multiple values for the same header, we store them as a slice.
			// This is useful for headers like "Set-Cookie" which can have multiple values.
			values = append([]string{}, values...) // Create a copy to avoid modifying the original slice
		} else if len(values) == 0 {
			// If there are no values, we set it to nil to indicate absence.
			serialized[key] = nil
		} else {
			// If there's only one value, we keep it as a single string.
			serialized[key] = values[0]
		}
	}
	return serialized
}
