package elastic

import (
	"fmt"
	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"os"
	"strings"
)

type ClientOption func(cfg *elasticsearch8.Config)

func Initialize(options ...ClientOption) (*elasticsearch8.Client, error) {
	cfg := &elasticsearch8.Config{}
	cert := os.Getenv("ELASTICSEARCH_CA_CERT")
	if cert != "" {
		if strings.HasPrefix(cert, "-----BEGIN CERTIFICATE-----") {
			cfg.CACert = []byte(cert)
		} else if strings.HasSuffix(cert, ".crt") {
			certData, err := os.ReadFile(cert)
			if err != nil {
				return nil, err
			}
			cfg.CACert = certData
		} else {
			return nil, fmt.Errorf("environment ELASTICSEARCH_CA_CERT should be a file or a PEM encoded certificate")
		}
	}
	for _, opt := range options {
		opt(cfg)
	}
	return elasticsearch8.NewClient(*cfg)
}
