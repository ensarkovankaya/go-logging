package elastic

import (
	"fmt"
	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"os"
	"strings"
)

type ClientOption func(cfg *elasticsearch8.Config)

func Initialize(options ...ClientOption) (*elasticsearch8.Client, error) {
	var err error
	cfg := &elasticsearch8.Config{}
	if cfg.CACert, err = getCACert(); err != nil {
		return nil, err
	}

	for _, opt := range options {
		opt(cfg)
	}
	return elasticsearch8.NewClient(*cfg)
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
