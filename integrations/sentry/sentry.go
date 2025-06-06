package sentry

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"os"
	"strconv"
	"time"

	"github.com/ensarkovankaya/go-logging/core"
)

type ClientOption = func(*sentry.ClientOptions)

var FLushTimeout = time.Second * 5

func Initialize(opts ...ClientOption) *sentry.Hub {
	options := &sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		ServerName:  os.Getenv("APP_NAME"),
		Environment: core.ReadEnvironment(),
	}
	options.EnableLogs, _ = core.ParseBool("SENTRY_ENABLE_LOGS", true, false)
	options.Debug, _ = core.ParseBool("SENTRY_DEBUG", false, false)
	options.AttachStacktrace, _ = core.ParseBool("SENTRY_ATTACH_STACKTRACE", true, false)
	options.EnableTracing, _ = core.ParseBool("SENTRY_ENABLE_TRACING", true, false)
	if os.Getenv("SENTRY_SAMPLE_RATE") != "" {
		rate, err := strconv.ParseFloat(os.Getenv("SENTRY_SAMPLE_RATE"), 64)
		if err != nil {
			options.SampleRate = rate
		}
	}
	if os.Getenv("SENTRY_TRACE_SAMPLE_RATE") != "" {
		rate, err := strconv.ParseFloat(os.Getenv("SENTRY_TRACE_SAMPLE_RATE"), 64)
		if err != nil {
			options.TracesSampleRate = rate
		}
	}
	if os.Getenv("SENTRY_MAX_BREADCRUMBS") != "" {
		rate, err := strconv.ParseInt(os.Getenv("SENTRY_MAX_BREADCRUMBS"), 10, 64)
		if err != nil {
			options.MaxBreadcrumbs = int(rate)
		}
	}
	for _, opt := range opts {
		opt(options)
	}
	client, err := sentry.NewClient(*options)
	if err != nil {
		panic(err)
	}
	hub := sentry.NewHub(client, sentry.NewScope())
	return hub
}

func init() {
	if os.Getenv("SENTRY_FLUSH_TIMEOUT") != "" {
		timeout, err := time.ParseDuration(os.Getenv("SENTRY_FLUSH_TIMEOUT"))
		if err != nil {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Invalid SENTRY_FLUSH_TIMEOUT value: %s, using default %v\n",
				os.Getenv("SENTRY_FLUSH_TIMEOUT"),
				FLushTimeout,
			)
		} else {
			FLushTimeout = timeout
		}
	}
}
