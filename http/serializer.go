package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Serializer struct {
	RequestJSON  bool
	ResponseJSON bool
}

func (t *Serializer) Request(request *http.Request) (map[string]any, error) {
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
		"headers":       t.Headers(request.Header),
		"proto":         request.Proto,
		"contentLength": request.ContentLength,
		"body":          nil,
	}
	if request.Body != nil {
		reader, err := request.GetBody()
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to get body: %v", err), err)
		}
		body, err := io.ReadAll(reader)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to read body: %v", err), err)
		}

		if t.RequestJSON && request.Header.Get("Content-Type") == "application/json" {
			var jsonBody map[string]any
			if err = json.Unmarshal(body, &jsonBody); err != nil {
				return nil, errors.Join(fmt.Errorf("failed to unmarshal JSON body"), err)
			}
			serialized["body"] = jsonBody
		} else {
			serialized["body"] = string(body)
		}
	}
	return serialized, nil
}

func (t *Serializer) Response(response *http.Response) (map[string]any, error) {
	if response == nil {
		return nil, nil
	}
	serialized := map[string]any{
		"statusCode":    response.StatusCode,
		"proto":         response.Proto,
		"headers":       t.Headers(response.Header),
		"contentLength": response.ContentLength,
		"body":          nil,
	}
	if response.Body != nil {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to read body"), err)
		}
		defer func() { _ = response.Body.Close() }()
		if t.ResponseJSON && response.Header.Get("Content-Type") == "application/json" {
			var jsonBody map[string]any
			if err = json.Unmarshal(body, &jsonBody); err != nil {
				return nil, errors.Join(fmt.Errorf("failed to unmarshal JSON body"), err)
			}
			serialized["body"] = jsonBody
		} else {
			serialized["body"] = string(body)
		}
		// Reset the body so it can be read again later
		response.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	return serialized, nil
}

func (t *Serializer) Headers(headers http.Header) map[string]any {
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
