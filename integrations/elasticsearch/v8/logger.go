package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ensarkovankaya/go-logging/core"
)

var (
	defaultIndexName   = "logs"
	defaultLevel       = core.LevelDebug
	flushCheckInterval = time.Second
)

type Option func(l *Logger)
type IndexBuilder func(ctx context.Context, logger *Logger, level core.Level, msg string, fields []core.Field) string

const Type = "elasticsearch"

var DefaultIndexBuilder IndexBuilder = func(_ context.Context, logger *Logger, _ core.Level, _ string, _ []core.Field) string {
	return defaultIndexName
}

type Logger struct {
	Client       *elasticsearch8.Client
	Name         string
	NowFunc      func() time.Time
	Extra        []core.Field
	Level        core.Level
	IndexBuilder IndexBuilder
	Ctx          context.Context
	DebugLogger  core.Interface
	lock         sync.Locker
	count        int
}

func New(options ...Option) *Logger {
	client, err := Initialize()
	if err != nil {
		panic(fmt.Sprintf("Elastic initialization failed: %v", err))
	}
	logger := &Logger{
		Client:       client,
		NowFunc:      time.Now,
		Level:        defaultLevel,
		IndexBuilder: DefaultIndexBuilder,
		Ctx:          context.Background(),
		DebugLogger:  &noopLogger{},
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

func (l *Logger) Named(name string) core.Interface {
	_l := l.clone()
	if name != "" && l.Name != "" {
		name = fmt.Sprintf("%s.%s", l.Name, name)
	}
	_l.Name = name
	return _l
}

func (l *Logger) Clone() core.Interface {
	return l.clone()
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	return ctx
}

func (l *Logger) With(fields ...core.Field) core.Interface {
	_l := l.clone()
	_l.Extra = append(l.Extra, fields...)
	return l
}

func (l *Logger) Debug(_ context.Context, msg string, fields ...core.Field) {
	if l.CanLog(core.LevelDebug) {
		l.LogAsync(core.LevelDebug, msg, fields...)
	}
}

func (l *Logger) Info(_ context.Context, msg string, fields ...core.Field) {
	if l.CanLog(core.LevelInfo) {
		l.LogAsync(core.LevelInfo, msg, fields...)
	}
}

func (l *Logger) Warning(_ context.Context, msg string, fields ...core.Field) {
	if l.CanLog(core.LevelWarning) {
		l.LogAsync(core.LevelWarning, msg, fields...)
	}
}

func (l *Logger) Error(_ context.Context, msg string, fields ...core.Field) {
	if l.CanLog(core.LevelError) {
		l.LogAsync(core.LevelError, msg, fields...)
	}
}

func (l *Logger) Flush(ctx context.Context) error {
	l.DebugLogger.Debug(ctx, "Flushing index")
	if l.getCount() <= 0 {
		return nil
	}
	intervalTimer := time.NewTimer(flushCheckInterval)
	defer intervalTimer.Stop()
	for {
		select {
		case <-ctx.Done():
			count := l.getCount()
			if count > 0 {
				l.DebugLogger.Error(ctx, "Flush cancelled before all logs were flushed", core.F("remaining", count))
				return fmt.Errorf("not all logs were flushed: %d remaining", count)
			}
			l.DebugLogger.Info(ctx, "Flush cancelled, no logs remaining")
			return nil
		case <-intervalTimer.C:
			count := l.getCount()
			if count <= 0 {
				l.DebugLogger.Info(ctx, "Flush completed, no logs remaining")
				return nil
			}
			l.DebugLogger.Debug(ctx, "Checking for remaining logs to flush", core.F("remaining", count))
			intervalTimer.Reset(flushCheckInterval)
		}
	}
}

func (l *Logger) LogAsync(level core.Level, msg string, fields ...core.Field) {
	l.increaseCount()
	go func() {
		defer l.decreaseCount()
		_, _ = l.Log(l.Ctx, level, msg, fields)
	}()
}

func (l *Logger) Log(ctx context.Context, level core.Level, msg string, fields []core.Field) (*esapi.Response, error) {
	body, err := l.buildDocument(level, msg, fields)
	if err != nil {
		l.DebugLogger.Error(ctx, "Failed to build body", core.E(err))
		return nil, err
	}
	index := l.IndexBuilder(ctx, l, level, msg, fields)
	resp, err := l.Client.API.Index(index, body)
	if err != nil {
		l.DebugLogger.Error(l.Ctx, "Failed to create index", core.E(err))
		return nil, errors.Join(fmt.Errorf("failed to create index"), err)
	}
	if resp.IsError() {
		l.DebugLogger.Error(l.Ctx, "Index request failed", core.F("status", resp.StatusCode), core.F("headers", resp.Header), core.F("response", resp.String()))
	} else {
		l.DebugLogger.Info(l.Ctx, "Index request succeeded", core.F("body", resp.String()))
	}
	return resp, nil
}

func (l *Logger) buildDocument(level core.Level, msg string, fields []core.Field) (io.Reader, error) {
	payload := map[string]any{
		"timestamp": l.NowFunc().Format(time.RFC3339),
		"level":     level.String(),
	}
	if msg != "" {
		payload["message"] = msg
	}
	if l.Name != "" {
		payload["name"] = l.Name
	}

	if len(l.Extra) > 0 {
		for _, field := range l.Extra {
			if _, ok := payload[field.Key]; ok {
				l.DebugLogger.Warning(l.Ctx, "Field already exists in payload, overwriting", core.F("key", field.Key))
			}
			payload[field.Key] = field.Value
		}
	}

	if len(fields) > 0 {
		data := make(map[string]any, len(fields))
		for _, field := range fields {
			data[field.Key] = field.Value
		}
		payload["data"] = data
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal log payload: %w", err)
	}

	return bytes.NewReader(body), nil
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

func (l *Logger) CanLog(level core.Level) bool {
	return l.Level <= level
}

func (l *Logger) clone() *Logger {
	_l := *l
	_l.lock = &sync.Mutex{}
	_l.count = 0
	_l.Extra = make([]core.Field, 0, len(l.Extra))
	for _, f := range l.Extra {
		_l.Extra = append(_l.Extra, f)
	}
	return &_l
}

func init() {
	if os.Getenv("ELASTIC_LOGGER_FLUSH_CHECK_INTERVAL") != "" {
		if interval, err := time.ParseDuration(os.Getenv("ELASTIC_LOGGER_FLUSH_CHECK_INTERVAL")); err == nil {
			flushCheckInterval = interval
		} else {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Invalid ELASTIC_LOGGER_FLUSH_CHECK_INTERVAL value: %s, using default %v\n",
				os.Getenv("ELASTIC_LOGGER_FLUSH_CHECK_INTERVAL"),
				flushCheckInterval,
			)
		}
	}
}
