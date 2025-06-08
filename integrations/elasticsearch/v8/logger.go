package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"sync"
	"time"

	"github.com/ensarkovankaya/go-logging"
)

var (
	defaultIndexName    = "logs"
	defaultLevel        = logging.LevelDebug
	defaultFlushTimeout = time.Second * 5
	flushCheckInterval  = time.Millisecond * 100
)

type Option func(l *Logger)
type IndexBuilder func(ctx context.Context, logger *Logger, level logging.Level, msg string, fields []logging.Field) string

const Type = "elasticsearch"

var DefaultIndexBuilder IndexBuilder = func(_ context.Context, logger *Logger, _ logging.Level, _ string, _ []logging.Field) string {
	return fmt.Sprintf("%s-%s", defaultIndexName, time.Now().Format("2006-01-02-15-04-05"))
}

type Logger struct {
	Client       *elasticsearch8.Client
	Name         string
	NowFunc      func() time.Time
	Fields       []logging.Field
	Level        logging.Level
	IndexBuilder IndexBuilder
	FlushTimeout time.Duration
	Ctx          context.Context
	Cancel       context.CancelFunc
	DebugLogger  logging.Logger
	Transport    esapi.Transport
	lock         sync.Locker
	count        int
}

func New(options ...Option) *Logger {
	client, err := Initialize()
	if err != nil {
		panic(fmt.Sprintf("Elastic initialization failed: %v", err))
	}
	ctx, cancel := context.WithCancel(context.Background())
	logger := &Logger{
		Client:       client,
		NowFunc:      time.Now,
		Level:        defaultLevel,
		IndexBuilder: DefaultIndexBuilder,
		FlushTimeout: defaultFlushTimeout,
		Ctx:          ctx,
		Cancel:       cancel,
		DebugLogger:  &noopLogger{},
		Transport:    DefaultTransport,
		lock:         &sync.Mutex{},
		count:        0,
	}
	for _, opt := range options {
		opt(logger)
	}
	return logger
}

func (l *Logger) Type() string {
	return Type
}

func (l *Logger) Named(name string) logging.Logger {
	_l := l.clone()
	if name != "" && l.Name != "" {
		name = fmt.Sprintf("%s.%s", l.Name, name)
	}
	_l.Name = name
	return _l
}

func (l *Logger) Clone() logging.Logger {
	return l.clone()
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	return ctx
}

func (l *Logger) With(fields ...logging.Field) logging.Logger {
	_l := l.clone()
	_l.Fields = append(l.Fields, fields...)
	return l
}

func (l *Logger) Debug(_ context.Context, msg string, fields ...logging.Field) {
	if l.CanLog(logging.LevelDebug) {
		l.LogAsync(logging.LevelDebug, msg, fields...)
	}
}

func (l *Logger) Info(_ context.Context, msg string, fields ...logging.Field) {
	if l.CanLog(logging.LevelInfo) {
		l.LogAsync(logging.LevelInfo, msg, fields...)
	}
}

func (l *Logger) Warning(_ context.Context, msg string, fields ...logging.Field) {
	if l.CanLog(logging.LevelWarning) {
		l.LogAsync(logging.LevelWarning, msg, fields...)
	}
}

func (l *Logger) Error(_ context.Context, msg string, fields ...logging.Field) {
	if l.CanLog(logging.LevelError) {
		l.LogAsync(logging.LevelError, msg, fields...)
	}
}

func (l *Logger) Flush(ctx context.Context) error {
	l.DebugLogger.Debug(ctx, "Flushing index")
	defer l.Cancel()
	if l.getCount() <= 0 {
		return nil
	}
	maxTimer := time.NewTimer(l.FlushTimeout)
	defer maxTimer.Stop()
	intervalTimer := time.NewTimer(flushCheckInterval)
	defer intervalTimer.Stop()
	for {
		select {
		case <-ctx.Done():
			count := l.getCount()
			if count > 0 {
				l.DebugLogger.Error(ctx, "Flush cancelled before all logs were flushed", logging.F("remaining", count))
				return fmt.Errorf("not all logs were flushed: %d remaining", l.count)
			}
			l.DebugLogger.Info(ctx, "Flush cancelled, no logs remaining")
			return nil
		case <-intervalTimer.C:
			count := l.getCount()
			if count <= 0 {
				l.DebugLogger.Info(ctx, "Flush completed, no logs remaining")
				return nil
			}
			l.DebugLogger.Debug(ctx, "Checking for remaining logs to flush", logging.F("remaining", count))
			intervalTimer.Reset(flushCheckInterval)
		case <-maxTimer.C:
			count := l.getCount()
			if count <= 0 {
				l.DebugLogger.Info(ctx, "Flush completed, no logs remaining")
				return nil
			}
			l.DebugLogger.Error(ctx, "Flush timed out", logging.F("remaining", count))
			return fmt.Errorf("timed out waiting for logs to be flushed: %d remaining", l.count)
		}
	}
}

func (l *Logger) LogAsync(level logging.Level, msg string, fields ...logging.Field) {
	l.increaseCount()
	go func() {
		defer l.decreaseCount()
		_, _ = l.Log(l.Ctx, level, msg, fields)
	}()
}

func (l *Logger) Log(ctx context.Context, level logging.Level, msg string, fields []logging.Field) (*esapi.Response, error) {
	req, err := l.BuildRequest(ctx, level, msg, fields)
	if err != nil {
		l.DebugLogger.Error(ctx, "Failed to build request", logging.E(err))
		return nil, err
	}
	resp, err := req.Do(ctx, l.Transport)
	if err != nil {
		l.DebugLogger.Error(l.Ctx, "Failed to create index request", logging.E(err))
		return nil, errors.Join(fmt.Errorf("failed to create index request"), err)
	}
	if resp.IsError() {
		l.DebugLogger.Error(l.Ctx, "Index request failed", logging.F("status", resp.StatusCode), logging.F("headers", resp.Header), logging.F("response", resp.String()))
	} else {
		l.DebugLogger.Info(l.Ctx, "Index request succeeded", logging.F("body", resp.String()))
	}
	return resp, nil
}

func (l *Logger) BuildRequest(ctx context.Context, level logging.Level, msg string, fields []logging.Field) (*esapi.IndexRequest, error) {
	index := l.IndexBuilder(ctx, l, level, msg, fields)

	payload := map[string]any{
		"timestamp": l.NowFunc().Format(time.RFC3339),
		"message":   msg,
		"logger":    l.Name,
		"level":     level.String(),
	}

	if len(l.Fields) > 0 {
		payload["context"] = make(map[string]any, len(l.Fields))
		for _, field := range l.Fields {
			payload[field.Key] = field.Value
		}
	}

	if len(fields) > 0 {
		payload["fields"] = make(map[string]any, len(fields))
		for _, field := range fields {
			payload[field.Key] = field.Value
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal log payload: %w", err)
	}

	return &esapi.IndexRequest{
		Index: index,
		Body:  bytes.NewReader(body),
	}, nil
}

func (l *Logger) decreaseCount() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.count--
}

func (l *Logger) increaseCount() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.count++
}

func (l *Logger) getCount() int {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.count
}

func (l *Logger) CanLog(level logging.Level) bool {
	return l.Level <= level
}

func (l *Logger) clone() *Logger {
	ctx, cancel := context.WithCancel(context.WithoutCancel(l.Ctx))
	_l := *l
	_l.Ctx = ctx
	_l.Cancel = cancel
	_l.lock = &sync.Mutex{}
	_l.count = 0
	_l.Fields = make([]logging.Field, 0, len(l.Fields))
	for _, f := range l.Fields {
		_l.Fields = append(_l.Fields, f)
	}
	return &_l
}

func init() {
}
