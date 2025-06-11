// Package elastic provides a client for interacting with Elasticsearch.
// References:
//   - https://github.com/elastic/go-elasticsearch/blob/main/_examples/bulk/indexer.go
package elastic

import (
	"fmt"
	"github.com/cenkalti/backoff/v5"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	envFlushBytes    = "ELASTICSEARCH_FLUSH_BYTES"
	envFlushInterval = "ELASTICSEARCH_FLUSH_INTERVAL"
)

var (
	globalClient      *elasticsearch.Client
	globalBulkIndexer esutil.BulkIndexer
	flushBytes        = int(5e+6) // 5 MB
	flushInterval     = 30 * time.Second
)

type (
	ClientOption      func(cfg *elasticsearch.Config)
	BulkIndexerOption func(cfg *esutil.BulkIndexerConfig)
)

func ReplaceClient(client *elasticsearch.Client) {
	globalClient = client
}

func ReplaceIndexer(indexer esutil.BulkIndexer) {
	globalBulkIndexer = indexer
}

func InitializeClient(options ...ClientOption) (*elasticsearch.Client, error) {
	var err error

	retryBackOff := backoff.NewExponentialBackOff()
	cfg := &elasticsearch.Config{
		// Retry on 429 TooManyRequests statuses
		RetryOnStatus: []int{502, 503, 504, 429},
		// Configure the backoff function
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackOff.Reset()
			}
			return retryBackOff.NextBackOff()
		},
		MaxRetries: 5,
	}
	if cfg.CACert, err = getCACert(); err != nil {
		return nil, err
	}

	for _, opt := range options {
		opt(cfg)
	}
	return elasticsearch.NewClient(*cfg)
}

func InitializeBulkIndexer(options ...BulkIndexerOption) (esutil.BulkIndexer, error) {
	cfg := &esutil.BulkIndexerConfig{
		Client:        globalClient,
		NumWorkers:    runtime.NumCPU(),
		FlushBytes:    flushBytes,
		FlushInterval: flushInterval,
	}
	for _, opt := range options {
		opt(cfg)
	}
	return esutil.NewBulkIndexer(*cfg)
}

// getCACert retrieves the CA certificate from the environment variable ELASTICSEARCH_CA_CERT.
func getCACert() ([]byte, error) {
	cert := os.Getenv("ELASTICSEARCH_CA_CERT")
	if cert == "" {
		return nil, nil
	}
	// Check if the cert is a PEM encoded certificate
	if strings.HasPrefix(cert, "-----BEGIN CERTIFICATE-----") {
		return []byte(cert), nil
	}
	// Check if the cert is a file path
	if strings.HasSuffix(cert, ".crt") || strings.HasSuffix(cert, ".pem") {
		certData, err := os.ReadFile(cert)
		if err != nil {
			return nil, err
		}
		return certData, nil
	}
	return nil, fmt.Errorf("environment ELASTICSEARCH_CA_CERT should be a file or a PEM encoded certificate")
}

func IsActive() bool {
	return os.Getenv("ELASTICSEARCH_URL") != ""
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
	if IsActive() {
		var err error
		if globalClient, err = InitializeClient(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Elasticsearch client: %v\n", err)
		}
		if globalBulkIndexer, err = InitializeBulkIndexer(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Elasticsearch bulk indexer: %v\n", err)
		}
	}
}
