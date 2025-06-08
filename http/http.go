package http

import (
	"github.com/google/uuid"
	"net/http"
	"time"

	"github.com/ensarkovankaya/go-logging/core"
)

const DefaultRequestIDHeader = "X-Request-ID"

type Option func(*Transport)
type LogIDBuilder func(req *http.Request) string

var DefaultLogIDBuilder LogIDBuilder = func(req *http.Request) string {
	requestID := req.Header.Get(DefaultRequestIDHeader)
	if requestID != "" {
		return requestID
	}
	return uuid.NewString()
}

type Transport struct {
	Serializer
	Logger           core.Interface
	Transport        http.RoundTripper
	RequestIDBuilder LogIDBuilder
	NowFunc          func() time.Time
}

func NewTransport(logger core.Interface, options ...Option) *Transport {
	transport := &Transport{
		Serializer:       Serializer{},
		Logger:           logger,
		Transport:        http.DefaultTransport,
		RequestIDBuilder: DefaultLogIDBuilder,
		NowFunc:          time.Now,
	}
	for _, opt := range options {
		opt(transport)
	}
	return transport
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if GetLoggingDisabled(req.Context()) {
		return t.Transport.RoundTrip(req)
	}
	id := t.RequestIDBuilder(req)
	serializedRequest, err := t.Request(req)
	if err != nil {
		t.Logger.Error(req.Context(), "Failed to serialize request", core.F("id", id), core.E(err))
		return nil, err
	}
	t.Logger.Info(req.Context(), "Request", core.F("id", id), core.F("request", serializedRequest))
	startTime := t.NowFunc()
	response, err := t.Transport.RoundTrip(req)
	elapsed := t.NowFunc().Sub(startTime)
	if err != nil {
		t.Logger.Error(req.Context(), "HTTP error", core.F("id", id), core.E(err))
		return nil, err
	}
	serializedResponse, err := t.Response(response)
	if err != nil {
		t.Logger.Error(req.Context(), "Failed to serialize response", core.F("id", id), core.E(err))
		return nil, err
	}
	t.Logger.Info(req.Context(), "Response", core.F("id", id), core.F("response", serializedResponse), core.F("elapsed", elapsed))
	return response, nil
}
