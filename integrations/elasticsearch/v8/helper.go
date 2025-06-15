package elastic

import (
	"context"
	"net/http"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8/esutil"

	"github.com/ensarkovankaya/go-logging/core"
	_http "github.com/ensarkovankaya/go-logging/http"
)

type clientLogger struct {
	_http.Serializer
	ctx                 context.Context
	logger              core.Interface
	requestBodyEnabled  bool
	responseBodyEnabled bool
}

func NewClientLogger(logger core.Interface, options ...func(l *clientLogger)) elastictransport.Logger {
	l := &clientLogger{
		ctx:                 context.Background(),
		logger:              logger,
		responseBodyEnabled: true,
		requestBodyEnabled:  true,
	}
	for _, opt := range options {
		opt(l)
	}
	return l
}

func (l *clientLogger) LogRoundTrip(req *http.Request, res *http.Response, clientErr error, start time.Time, dur time.Duration) error {
	request, err := l.Serializer.Request(req)
	if err != nil {
		return err
	}
	response, err := l.Serializer.Response(res)
	if err != nil {
		return err
	}
	l.logger.Debug(l.ctx, "Elasticsearch round trip", core.F("req", request), core.F("res", response), core.F("err", clientErr), core.F("start", start), core.F("duration", dur))
	return nil
}

func (l *clientLogger) RequestBodyEnabled() bool {
	return l.requestBodyEnabled
}

func (l *clientLogger) ResponseBodyEnabled() bool {
	return l.responseBodyEnabled
}

type bulkIndexerDebugLogger struct {
	ctx    context.Context
	logger core.Interface
}

func NewBulkIndexerDebugLogger(logger core.Interface, options ...func(l *bulkIndexerDebugLogger)) esutil.BulkIndexerDebugLogger {
	l := &bulkIndexerDebugLogger{ctx: context.Background(), logger: logger}
	for _, opt := range options {
		opt(l)
	}
	return l
}

func (l *bulkIndexerDebugLogger) Printf(msg string, args ...interface{}) {
	l.logger.Debug(l.ctx, msg, core.F("args", args))
}
