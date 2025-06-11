module github.com/ensarkovankaya/go-logging

go 1.24

require (
	github.com/cenkalti/backoff/v5 v5.0.2
	github.com/elastic/go-elasticsearch/v8 v8.18.0
	github.com/getsentry/sentry-go v0.33.0
	github.com/google/uuid v1.6.0
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.12.2
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.36.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.36.0
	go.opentelemetry.io/otel/log v0.12.2
	go.opentelemetry.io/otel/sdk v1.36.0
	go.opentelemetry.io/otel/sdk/log v0.12.2
	go.opentelemetry.io/otel/sdk/metric v1.36.0
	go.opentelemetry.io/otel/trace v1.36.0
	go.uber.org/zap v1.27.0
	gorm.io/gorm v1.30.0
)

require (
	github.com/elastic/elastic-transport-go/v8 v8.7.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)
