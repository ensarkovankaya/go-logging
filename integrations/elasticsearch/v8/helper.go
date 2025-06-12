package elastic

import (
	"context"
	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"net/http"
	"time"

	"github.com/ensarkovankaya/go-logging/core"
)

type clientLogger struct {
	ctx                 context.Context
	logger              core.Interface
	requestBodyEnabled  bool
	responseBodyEnabled bool
}

func NewClientLogger(logger core.Interface, options ...func(l *clientLogger)) elastictransport.Logger {
	l := &clientLogger{ctx: context.Background(), logger: logger}
	for _, opt := range options {
		opt(l)
	}
	return l
}

func (l *clientLogger) LogRoundTrip(req *http.Request, res *http.Response, err error, start time.Time, dur time.Duration) error {
	l.logger.Debug(l.ctx, "Elasticsearch round trip", core.F("req", req), core.F("res", res), core.F("err", err), core.F("start", start), core.F("duration", dur))
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
