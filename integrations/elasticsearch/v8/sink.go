package elastic

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	envFlushBytes    = "ELASTICSEARCH_SINK_FLUSH_BYTES"
	envFlushInterval = "ELASTICSEARCH_SINK_FLUSH_INTERVAL"
	envNumWorkers    = "ELASTICSEARCH_SINK_NUM_WORKERS"
)

var (
	globalSink    esutil.BulkIndexer
	flushBytes    = int(5e+6) // 5 MB
	flushInterval = 30 * time.Second
	numWorkers    = runtime.NumCPU() // Default is the number of CPU cores
)

type SinkOption func(cfg *esutil.BulkIndexerConfig)

func ReplaceGlobalSink(indexer esutil.BulkIndexer) {
	globalSink = indexer
}

func InitializeSink(options ...SinkOption) (esutil.BulkIndexer, error) {
	cfg := &esutil.BulkIndexerConfig{
		Client:        globalClient,
		NumWorkers:    numWorkers,
		FlushBytes:    flushBytes,
		FlushInterval: flushInterval,
	}
	for _, opt := range options {
		opt(cfg)
	}
	return esutil.NewBulkIndexer(*cfg)
}

func init() {
	if os.Getenv(envFlushInterval) != "" {
		if interval, err := time.ParseDuration(os.Getenv(envFlushInterval)); err == nil {
			flushInterval = interval
		} else {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Invalid %v value, using default %v\n",
				envFlushInterval,
				flushInterval,
			)
		}
	}
	if os.Getenv(envFlushBytes) != "" {
		if bytes, err := strconv.ParseInt(os.Getenv(envFlushBytes), 10, 64); err == nil {
			flushBytes = int(bytes)
		} else {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Invalid %v value, using default %v\n",
				envFlushBytes,
				flushBytes,
			)
		}
	}
	if os.Getenv(envNumWorkers) != "" {
		if workerCount, err := strconv.Atoi(os.Getenv(envNumWorkers)); err == nil && workerCount > 0 {
			numWorkers = workerCount
		} else {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Invalid %v value, using default %v\n",
				envNumWorkers,
				numWorkers,
			)
		}
	}
}
