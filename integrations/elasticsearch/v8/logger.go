package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esutil"

	"github.com/ensarkovankaya/go-logging/core"
)

const (
	envIndexName = "ELASTICSEARCH_INDEX_NAME"
	envLogLevel  = "ELASTICSEARCH_LOG_LEVEL"
	actionType   = "index"
)

var (
	defaultIndexName = "logs"
	defaultLevel     = core.LevelDebug
)

type Option func(l *Logger)
type IndexBuilder func(ctx context.Context, logger *Logger, level core.Level, msg string, fields []core.Field) (string, error)
type DocumentIDBuilder func() string

const Type = "elasticsearch"

var (
	DefaultIndexBuilder IndexBuilder = func(_ context.Context, _ *Logger, _ core.Level, _ string, _ []core.Field) (string, error) {
		return fmt.Sprintf("%v-%v", defaultIndexName, time.Now().Format("2006-01-02-15-01-05")), nil
	}
	DefaultDocumentIDBuilder DocumentIDBuilder = func() string {
		return ""
	}
)

type Logger struct {
	Name              string
	NowFunc           func() time.Time
	Extra             []core.Field
	Level             core.Level
	IndexBuilder      IndexBuilder
	DocumentIDBuilder DocumentIDBuilder
	DebugLogger       core.Interface
	Sink              esutil.BulkIndexer
	OnSuccess         func(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem)
	OnFailure         func(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem, err error)
}

func New(options ...Option) *Logger {
	debugLogger := &noopLogger{}
	logger := &Logger{
		NowFunc:           time.Now,
		Level:             defaultLevel,
		IndexBuilder:      DefaultIndexBuilder,
		DocumentIDBuilder: DefaultDocumentIDBuilder,
		DebugLogger:       debugLogger,
		Sink:              globalSink,
		OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem) {
			debugLogger.Info(ctx, "Document added to indexer", core.F("documentID", item.DocumentID), core.F("resp", resp))
		},
		OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, resp esutil.BulkIndexerResponseItem, err error) {
			debugLogger.Error(ctx, "Failed to add document to indexer", core.E(err), core.F("documentID", item.DocumentID), core.F("resp", resp))
		},
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
	_l.Extra = append(_l.Extra, fields...)
	return _l
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...core.Field) {
	if l.CanLog(core.LevelDebug) {
		l.Log(ctx, core.LevelDebug, msg, fields)
	}
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...core.Field) {
	if l.CanLog(core.LevelInfo) {
		l.Log(ctx, core.LevelInfo, msg, fields)
	}
}

func (l *Logger) Warning(ctx context.Context, msg string, fields ...core.Field) {
	if l.CanLog(core.LevelWarning) {
		l.Log(ctx, core.LevelWarning, msg, fields)
	}
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...core.Field) {
	if l.CanLog(core.LevelError) {
		l.Log(ctx, core.LevelError, msg, fields)
	}
}

func (l *Logger) Flush(ctx context.Context) error {
	if err := l.Sink.Close(ctx); err != nil {
		l.DebugLogger.Error(ctx, "Failed to close indexer", core.E(err))
		return err
	}
	l.DebugLogger.Debug(ctx, "Flushed indexer", core.F("stats", l.Sink.Stats()))
	return nil
}

func (l *Logger) Log(ctx context.Context, level core.Level, msg string, fields []core.Field) {
	body, err := l.buildDocument(ctx, level, msg, fields)
	if err != nil {
		l.DebugLogger.Error(ctx, "Failed to build body", core.E(err))
		return
	}
	index, err := l.IndexBuilder(ctx, l, level, msg, fields)
	if err != nil {
		l.DebugLogger.Error(ctx, "Failed to build index", core.E(err))
		return
	}
	if err = l.Sink.Add(ctx, esutil.BulkIndexerItem{
		DocumentID: l.DocumentIDBuilder(),
		Action:     actionType,
		Index:      index,
		Body:       body,
		OnSuccess:  l.OnSuccess,
		OnFailure:  l.OnFailure,
	}); err != nil {
		l.DebugLogger.Error(ctx, "Failed to add document to indexer", core.E(err), core.F("index", index))
	}
}

func (l *Logger) buildDocument(ctx context.Context, level core.Level, msg string, fields []core.Field) (io.ReadSeeker, error) {
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
				l.DebugLogger.Warning(ctx, "Field already exists in payload, overwriting", core.F("key", field.Key))
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

func (l *Logger) CanLog(level core.Level) bool {
	return l.Level <= level
}

func (l *Logger) clone() *Logger {
	_l := *l
	_l.Extra = make([]core.Field, 0)
	_l.Extra = append(_l.Extra, l.Extra...)
	return &_l
}

func init() {
	if os.Getenv(envIndexName) != "" {
		defaultIndexName = os.Getenv(envIndexName)
	}
	if os.Getenv(envLogLevel) != "" {
		level, err := core.ParseLevel(os.Getenv(envLogLevel))
		if err != nil {
			panic(fmt.Errorf("failed to parse environment %s: %w", envLogLevel, err))
		}
		defaultLevel = level
	}
}
