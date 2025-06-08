package elastic

import (
	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
)

type ClientOption func(cfg *elasticsearch8.Config)

func Initialize(options ...ClientOption) (*elasticsearch8.Client, error) {
	cfg := elasticsearch8.Config{}
	for _, opt := range options {
		opt(&cfg)
	}
	return elasticsearch8.NewClient(cfg)
}
