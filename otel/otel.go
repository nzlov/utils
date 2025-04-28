package otel

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	SSL         bool              `json:"ssl" yaml:"ssl" mapstructure:"ssl"`
	Endpoint    string            `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`
	EndpointURL string            `json:"endpoint_url" yaml:"endpoint_url" mapstructure:"endpoint_url"`
	Headers     map[string]string `json:"headers" yaml:"headers" mapstructure:"headers"`
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func (cfg Config) SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator(cfg)
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(cfg)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(cfg)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(cfg)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

func newPropagator(cfg Config) propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(cfg Config) (*trace.TracerProvider, error) {
	opts := []otlptracehttp.Option{}

	if cfg.Endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(cfg.Endpoint))
	}

	if cfg.EndpointURL != "" {
		opts = append(opts, otlptracehttp.WithEndpointURL(cfg.EndpointURL))
	}

	if !cfg.SSL {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if cfg.Headers != nil {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Headers))
	}

	traceExporter, err := otlptracehttp.New(
		context.Background(),
		opts...,
	)
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
	)
	return traceProvider, nil
}

func newMeterProvider(cfg Config) (*metric.MeterProvider, error) {
	opts := []otlpmetrichttp.Option{}

	if cfg.Endpoint != "" {
		opts = append(opts, otlpmetrichttp.WithEndpoint(cfg.Endpoint))
	}

	if cfg.EndpointURL != "" {
		opts = append(opts, otlpmetrichttp.WithEndpointURL(cfg.EndpointURL))
	}

	if !cfg.SSL {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	if cfg.Headers != nil {
		opts = append(opts, otlpmetrichttp.WithHeaders(cfg.Headers))
	}

	metricExporter, err := otlpmetrichttp.New(
		context.Background(),
		opts...,
	)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)), // Default is 1m. Set to 3s for demonstrative purposes.

	)
	return meterProvider, nil
}

func newLoggerProvider(cfg Config) (*log.LoggerProvider, error) {
	opts := []otlploghttp.Option{}

	if cfg.Endpoint != "" {
		opts = append(opts, otlploghttp.WithEndpoint(cfg.Endpoint))
	}

	if cfg.EndpointURL != "" {
		opts = append(opts, otlploghttp.WithEndpointURL(cfg.EndpointURL))
	}

	if !cfg.SSL {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	if cfg.Headers != nil {
		opts = append(opts, otlploghttp.WithHeaders(cfg.Headers))
	}

	logExporter, err := otlploghttp.New(
		context.Background(),
		opts...,
	)
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}
