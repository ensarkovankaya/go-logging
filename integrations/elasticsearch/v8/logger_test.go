package elastic

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esutil"

	"github.com/ensarkovankaya/go-logging/core"
)

func Test_Logger_Type(t *testing.T) {
	logger, _ := getTestLogger(t)
	_type := logger.Type()
	if _type != Type {
		t.Errorf("Expected logger type '%v', got '%s'", Type, logger.Type())
	}
}

func Test_Logger_Named(t *testing.T) {
	logger, _ := getTestLogger(t)
	name := "test"
	logger = logger.Named(name).(*Logger)
	if logger.Name != name {
		t.Errorf("Expected logger name '%s', got '%s'", name, logger.Name)
	}
	subName := "sub"
	logger = logger.Named(subName).(*Logger)
	if logger.Name != name+"."+subName {
		t.Errorf("Expected logger name '%s.%s', got '%s'", name, subName, logger.Name)
	}
}

func Test_Logger_CLone(t *testing.T) {
	logger, _ := getTestLogger(t)
	logger.Name = "test"
	logger.Extra = []core.Field{core.F("key", "value")}
	clone := logger.Clone().(*Logger)
	if clone.Name != logger.Name {
		t.Errorf("Expected cloned logger name '%s', got '%s'", logger.Name, clone.Name)
	}
	if clone.Level != logger.Level {
		t.Errorf("Expected cloned logger level '%v', got '%v'", logger.Level, clone.Level)
	}
	if len(clone.Extra) != len(logger.Extra) {
		t.Errorf("Expected cloned logger extra fields length '%d', got '%d'", len(logger.Extra), len(clone.Extra))
	}
	if clone.Extra[0].Key != "key" {
		t.Errorf("Expected cloned logger extra field key '%s', got '%s'", "key", clone.Extra[0].Key)
	}
	if clone.Extra[0].Value != "value" {
		t.Errorf("Expected cloned logger extra field value '%s', got '%s'", "value", clone.Extra[0].Value)
	}
}

func Test_Logger_WithContext(t *testing.T) {
	type ctxKeyType string
	const ctxKey ctxKeyType = "logger"
	ctx := context.WithValue(context.Background(), ctxKey, "value")
	logger, _ := getTestLogger(t)
	ctx = logger.WithContext(ctx)
	if ctx.Value(ctxKey) != "value" {
		t.Errorf("Expected context value 'value', got '%v'", ctx.Value("key"))
	}
}

func Test_Logger_With(t *testing.T) {
	logger, _ := getTestLogger(t)
	key := "testKey"
	value := "testValue"
	logger = logger.With(core.F(key, value)).(*Logger)
	if len(logger.Extra) != 1 {
		t.Errorf("Expected logger extra fields length '1', got '%d'", len(logger.Extra))
	}
	if logger.Extra[0].Key != key {
		t.Errorf("Expected logger extra field key '%s', got '%s'", key, logger.Extra[0].Key)
	}
	if logger.Extra[0].Value != value {
		t.Errorf("Expected logger extra field value '%s', got '%s'", value, logger.Extra[0].Value)
	}
}

func Test_Logger_CanLog_LevelDisabled(t *testing.T) {
	logger, _ := getTestLogger(t)
	logger.Level = core.LevelDisabled
	if logger.CanLog(core.LevelError) {
		t.Error("Expected logger to not log at LevelError when Level is set to LevelDisabled")
	}
	if logger.CanLog(core.LevelWarning) {
		t.Error("Expected logger to not log at LevelWarning when Level is set to LevelDisabled")
	}
	if logger.CanLog(core.LevelInfo) {
		t.Error("Expected logger to not log at LevelInfo when Level is set to LevelDisabled")
	}
	if logger.CanLog(core.LevelDebug) {
		t.Error("Expected logger to not log at LevelDebug when Level is set to LevelDisabled")
	}
}

func Test_Logger_CanLog_LevelError(t *testing.T) {
	logger, _ := getTestLogger(t)
	logger.Level = core.LevelError
	if !logger.CanLog(core.LevelError) {
		t.Error("Expected logger to log at LevelError when Level is set to LevelError")
	}
	if logger.CanLog(core.LevelWarning) {
		t.Error("Expected logger to not log at LevelWarning when Level is set to LevelError")
	}
	if logger.CanLog(core.LevelInfo) {
		t.Error("Expected logger to not log at LevelInfo when Level is set to LevelError")
	}
	if logger.CanLog(core.LevelDebug) {
		t.Error("Expected logger to not log at LevelDebug when Level is set to LevelError")
	}
}

func Test_Logger_CanLog_LevelWarning(t *testing.T) {
	logger, _ := getTestLogger(t)
	logger.Level = core.LevelWarning
	if !logger.CanLog(core.LevelError) {
		t.Error("Expected logger to log at LevelError when Level is set to LevelWarning")
	}
	if !logger.CanLog(core.LevelWarning) {
		t.Error("Expected logger to log at LevelWarning when Level is set to LevelWarning")
	}
	if logger.CanLog(core.LevelInfo) {
		t.Error("Expected logger to not log at LevelInfo when Level is set to LevelWarning")
	}
	if logger.CanLog(core.LevelDebug) {
		t.Error("Expected logger to not log at LevelDebug when Level is set to LevelWarning")
	}
}

func Test_Logger_CanLog_LevelInfo(t *testing.T) {
	logger, _ := getTestLogger(t)
	logger.Level = core.LevelInfo
	if !logger.CanLog(core.LevelError) {
		t.Error("Expected logger to log at LevelError when Level is set to LevelInfo")
	}
	if !logger.CanLog(core.LevelWarning) {
		t.Error("Expected logger to log at LevelWarning when Level is set to LevelInfo")
	}
	if !logger.CanLog(core.LevelInfo) {
		t.Error("Expected logger to log at LevelInfo when Level is set to LevelInfo")
	}
	if logger.CanLog(core.LevelDebug) {
		t.Error("Expected logger to not log at LevelDebug when Level is set to LevelInfo")
	}
}

func Test_Logger_CanLog_LevelDebug(t *testing.T) {
	logger, _ := getTestLogger(t)
	logger.Level = core.LevelDebug
	if !logger.CanLog(core.LevelError) {
		t.Error("Expected logger to log at LevelError when Level is set to LevelDebug")
	}
	if !logger.CanLog(core.LevelWarning) {
		t.Error("Expected logger to log at LevelWarning when Level is set to LevelDebug")
	}
	if !logger.CanLog(core.LevelInfo) {
		t.Error("Expected logger to log at LevelInfo when Level is set to LevelDebug")
	}
	if !logger.CanLog(core.LevelDebug) {
		t.Error("Expected logger to log at LevelDebug when Level is set to LevelDebug")
	}
}

func Test_Logger_Debug(t *testing.T) {
	loggerLevel := core.LevelDebug
	t.Run("Debug", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelDebug,
				Message:     "Debug message",
				Extra:       []core.Field{core.F("key1", "value1")},
				Fields:      []core.Field{core.F("field1", "value1")},
				ShouldIndex: true,
			},
		})
	})
	t.Run("Info", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelInfo,
				Message:     "Info message",
				Extra:       []core.Field{core.F("key2", "value2")},
				Fields:      []core.Field{core.F("field2", "value2")},
				ShouldIndex: true,
			},
		})
	})
	t.Run("Warning", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelWarning,
				Message:     "Warning message",
				Extra:       []core.Field{core.F("key3", "value3")},
				Fields:      []core.Field{core.F("field3", "value3")},
				ShouldIndex: true,
			},
		})
	})
	t.Run("Error", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelError,
				Message:     "Error message",
				Extra:       []core.Field{core.F("key4", "value4")},
				Fields:      []core.Field{core.F("field4", "value4")},
				ShouldIndex: true,
			},
		})
	})
}

func Test_Logger_Info(t *testing.T) {
	loggerLevel := core.LevelInfo
	t.Run("Debug", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelDebug,
				Message:     "Debug message",
				Extra:       []core.Field{core.F("key1", "value1")},
				Fields:      []core.Field{core.F("field1", "value1")},
				ShouldIndex: false,
			},
		})
	})
	t.Run("Info", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelInfo,
				Message:     "Info message",
				Extra:       []core.Field{core.F("key2", "value2")},
				Fields:      []core.Field{core.F("field2", "value2")},
				ShouldIndex: true,
			},
		})
	})
	t.Run("Warning", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelWarning,
				Message:     "Warning message",
				Extra:       []core.Field{core.F("key3", "value3")},
				Fields:      []core.Field{core.F("field3", "value3")},
				ShouldIndex: true,
			},
		})
	})
	t.Run("Error", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelError,
				Message:     "Error message",
				Extra:       []core.Field{core.F("key4", "value4")},
				Fields:      []core.Field{core.F("field4", "value4")},
				ShouldIndex: true,
			},
		})
	})
}

func Test_Logger_Warning(t *testing.T) {
	loggerLevel := core.LevelWarning
	t.Run("Debug", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelDebug,
				Message:     "Debug message",
				Extra:       []core.Field{core.F("key1", "value1")},
				Fields:      []core.Field{core.F("field1", "value1")},
				ShouldIndex: false,
			},
		})
	})
	t.Run("Info", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelInfo,
				Message:     "Info message",
				Extra:       []core.Field{core.F("key2", "value2")},
				Fields:      []core.Field{core.F("field2", "value2")},
				ShouldIndex: false,
			},
		})
	})
	t.Run("Warning", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelWarning,
				Message:     "Warning message",
				Extra:       []core.Field{core.F("key3", "value3")},
				Fields:      []core.Field{core.F("field3", "value3")},
				ShouldIndex: true,
			},
		})
	})
	t.Run("Error", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelError,
				Message:     "Error message",
				Extra:       []core.Field{core.F("key4", "value4")},
				Fields:      []core.Field{core.F("field4", "value4")},
				ShouldIndex: true,
			},
		})
	})
}

func Test_Logger_Error(t *testing.T) {
	loggerLevel := core.LevelError
	t.Run("Debug", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelDebug,
				Message:     "Debug message",
				Extra:       []core.Field{core.F("key1", "value1")},
				Fields:      []core.Field{core.F("field1", "value1")},
				ShouldIndex: false,
			},
		})
	})
	t.Run("Info", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelInfo,
				Message:     "Info message",
				Extra:       []core.Field{core.F("key2", "value2")},
				Fields:      []core.Field{core.F("field2", "value2")},
				ShouldIndex: false,
			},
		})
	})
	t.Run("Warning", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelWarning,
				Message:     "Warning message",
				Extra:       []core.Field{core.F("key3", "value3")},
				Fields:      []core.Field{core.F("field3", "value3")},
				ShouldIndex: false,
			},
		})
	})
	t.Run("Error", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelError,
				Message:     "Error message",
				Extra:       []core.Field{core.F("key4", "value4")},
				Fields:      []core.Field{core.F("field4", "value4")},
				ShouldIndex: true,
			},
		})
	})
}

func Test_Logger_Disabled(t *testing.T) {
	loggerLevel := core.LevelDisabled
	t.Run("Debug", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelDebug,
				Message:     "Debug message",
				Extra:       []core.Field{core.F("key1", "value1")},
				Fields:      []core.Field{core.F("field1", "value1")},
				ShouldIndex: false,
			},
		})
	})
	t.Run("Info", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelInfo,
				Message:     "Info message",
				Extra:       []core.Field{core.F("key2", "value2")},
				Fields:      []core.Field{core.F("field2", "value2")},
				ShouldIndex: false,
			},
		})
	})
	t.Run("Warning", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelWarning,
				Message:     "Warning message",
				Extra:       []core.Field{core.F("key3", "value3")},
				Fields:      []core.Field{core.F("field3", "value3")},
				ShouldIndex: false,
			},
		})
	})
	t.Run("Error", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelError,
				Message:     "Error message",
				Extra:       []core.Field{core.F("key4", "value4")},
				Fields:      []core.Field{core.F("field4", "value4")},
				ShouldIndex: false,
			},
		})
	})
}

func Test_Logger_Debug_Multiple(t *testing.T) {
	loggerLevel := core.LevelDebug
	t.Run("Debug", func(t *testing.T) {
		testLogger(t, loggerLevel, []testCase{
			{
				Index:       testIndex,
				Level:       core.LevelDebug,
				Message:     "Debug message",
				Extra:       []core.Field{core.F("key1", "value1")},
				Fields:      []core.Field{core.F("field1", "value1")},
				ShouldIndex: true,
			},
			{
				Index:       testIndex,
				Level:       core.LevelInfo,
				Message:     "Info message",
				Extra:       []core.Field{core.F("key2", "value2")},
				Fields:      []core.Field{core.F("field2", "value2")},
				ShouldIndex: true,
			},
			{
				Index:       testIndex,
				Level:       core.LevelWarning,
				Message:     "Warning message",
				Extra:       []core.Field{core.F("key3", "value3")},
				Fields:      []core.Field{core.F("field3", "value3")},
				ShouldIndex: true,
			},
			{
				Index:       testIndex,
				Level:       core.LevelError,
				Message:     "Error message",
				Extra:       []core.Field{core.F("key4", "value4")},
				Fields:      []core.Field{core.F("field4", "value4")},
				ShouldIndex: true,
			},
		})
	})
}

//nolint:gocyclo
func testLogger(t *testing.T, loggerLevel core.Level, logs []testCase) {
	logger, transport := getTestLogger(t)
	logger.Level = loggerLevel

	expectedStats := esutil.BulkIndexerStats{}

	indexedLogs := make([]testCase, 0)
	for _, log := range logs {
		if log.ShouldIndex {
			indexedLogs = append(indexedLogs, log)
			expectedStats.NumIndexed++
			expectedStats.NumAdded++
			expectedStats.NumFlushed++
		}
	}

	for _, log := range logs {
		switch log.Level {
		case core.LevelDebug:
			logger.Extra = log.Extra
			logger.Debug(context.Background(), log.Message, log.Fields...)
		case core.LevelInfo:
			logger.Extra = log.Extra
			logger.Info(context.Background(), log.Message, log.Fields...)
		case core.LevelWarning:
			logger.Extra = log.Extra
			logger.Warning(context.Background(), log.Message, log.Fields...)
		case core.LevelError:
			logger.Extra = log.Extra
			logger.Error(context.Background(), log.Message, log.Fields...)
		case core.LevelDisabled:
			continue
		default:
			t.Fatalf("Unsupported log loggerLevel: %v", loggerLevel)
		}
	}

	if err := logger.Sink.Close(context.Background()); err != nil {
		t.Fatalf("Failed to close sink: %v", err)
	}

	// Check the transport for the expected request
	stats := logger.Sink.Stats()
	if stats.NumAdded != expectedStats.NumAdded {
		t.Errorf("Expected %d logs to be added, got %d", expectedStats.NumAdded, stats.NumAdded)
	}
	if stats.NumFlushed != expectedStats.NumFlushed {
		t.Errorf("Expected %d logs to be flushed, got %d", expectedStats.NumFlushed, stats.NumFlushed)
	}
	if stats.NumFailed != expectedStats.NumFailed {
		t.Errorf("Expected %d logs to fail, got %d", expectedStats.NumFailed, stats.NumFailed)
	}
	if stats.NumIndexed != expectedStats.NumIndexed {
		t.Errorf("Expected %d logs to be indexed, got %d", expectedStats.NumIndexed, stats.NumIndexed)
	}
	if stats.NumCreated != expectedStats.NumCreated {
		t.Errorf("Expected %d logs to be created, got %d", expectedStats.NumCreated, stats.NumCreated)
	}
	if stats.NumUpdated != expectedStats.NumUpdated {
		t.Errorf("Expected %d logs to be updated, got %d", expectedStats.NumUpdated, stats.NumUpdated)
	}
	if stats.NumDeleted != expectedStats.NumDeleted {
		t.Errorf("Expected %d logs to be deleted, got %d", expectedStats.NumDeleted, stats.NumDeleted)
	}

	if len(indexedLogs) != len(transport.IndexRequests) {
		t.Errorf("Expected %d indexed logs, got %d", len(indexedLogs), len(transport.IndexRequests))
	}

	indexedRequestMap := map[int]map[string]any{}
	for i, indexRequest := range transport.IndexRequests {
		indexedRequestMap[i] = indexRequest
	}

	for _, log := range indexedLogs {
		var indexRequest map[string]any
		var found bool
		var key int
		for key, indexRequest = range indexedRequestMap {
			message, ok := indexRequest["message"].(string)
			if !ok {
				t.Errorf("Expected indexed log message to be a string, got %T", indexRequest["message"])
				continue
			}
			if message == log.Message {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected indexed log with message '%s' not found in requests", log.Message)
			continue
		}
		delete(indexedRequestMap, key) // remove the found request to avoid duplicates

		// Validate indexed log has correct level
		level, ok := indexRequest["level"].(string)
		if !ok {
			t.Errorf("Expected indexed log level to be a string, got %T", indexRequest["level"])
			continue
		}
		if level != log.Level.String() {
			t.Errorf("Expected indexed log level '%s', got '%s'", log.Level.String(), level)
		}

		// Validate indexed log has timestamp
		timestamp, ok := indexRequest["timestamp"].(string)
		if !ok {
			t.Errorf("Expected indexed log timestamp to be a string, got %T", indexRequest["timestamp"])
			continue
		}
		if timestamp != testTimestamp.Format(time.RFC3339) {
			t.Errorf("Expected indexed log timestamp '%s', got '%s'", testTimestamp.Format(time.RFC3339), timestamp)
		}

		// Validate extra fields add to indexed
		for _, field := range log.Extra {
			if _, ok = indexRequest[field.Key]; !ok {
				t.Errorf("Expected indexed log extra field '%s' not found in request", field.Key)
			}
		}

		// Validate fields add to indexed log data
		data, ok := indexRequest["data"].(map[string]any)
		if !ok {
			t.Errorf("Expected indexed log data to be a map[string]any, got %T", indexRequest["data"])
			continue
		}

		for _, field := range log.Fields {
			if _, ok = data[field.Key]; !ok {
				t.Errorf("Expected indexed log field '%s' not found in request", field.Key)
			}
		}
	}
}

func getTestLogger(t *testing.T, options ...Option) (*Logger, *mockRoundTrip) {
	transport := newMockTransport(t)
	sink := getTestSink(t, transport)
	options = append([]Option{func(l *Logger) {
		l.Sink = sink
		l.IndexBuilder = func(_ context.Context, _ *Logger, _ core.Level, _ string, _ []core.Field) (string, error) {
			return testIndex, nil
		}
		l.NowFunc = testNowFunc
	}}, options...)
	logger := New(options...)
	return logger, transport
}

func getTestSink(t *testing.T, transport http.RoundTripper, options ...SinkOption) esutil.BulkIndexer {
	t.Helper()
	client := getTestClient(t, transport)
	options = append([]SinkOption{func(cfg *esutil.BulkIndexerConfig) {
		cfg.Client = client
	}}, options...)
	sink, err := NewSink(options...)
	if err != nil {
		t.Fatalf("Failed to create sink: %v", err)
	}
	return sink
}
