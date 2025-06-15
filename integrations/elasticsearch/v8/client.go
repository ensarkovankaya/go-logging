// Package elastic provides a client for interacting with Elasticsearch.
// References:
//   - https://github.com/elastic/go-elasticsearch/blob/main/_examples/bulk/indexer.go
package elastic

import (
	"fmt"
	"github.com/cenkalti/backoff/v5"
	"github.com/elastic/go-elasticsearch/v8"
	"os"
	"strings"
	"time"
)

var (
	globalClient *elasticsearch.Client
)

type ClientOption func(cfg *elasticsearch.Config)

func ReplaceClient(client *elasticsearch.Client) {
	globalClient = client
}

func NewClient(options ...ClientOption) (*elasticsearch.Client, error) {
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
	if IsActive() {
		var err error
		if globalClient, err = NewClient(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Elasticsearch client: %v\n", err)
		}
		if globalSink, err = NewSink(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to initialize Elasticsearch bulk indexer: %v\n", err)
		}
	}
}
