package otel

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	Name          string `json:"name" yaml:"name" mapstructure:"name"`
	LogSource     bool   `json:"logSource" yaml:"logSource" mapstructure:"logSource"`
	Type          string `json:"type" yaml:"type" mapstructure:"type"`
	MetricDisable bool   `json:"metricDisable" yaml:"metricDisable" mapstructure:"metricDisable"`
	TraceDisable  bool   `json:"traceDisable" yaml:"traceDisable" mapstructure:"traceDisable"`
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func (cfg *Config) SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
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

	if !cfg.TraceDisable {
		// Set up propagator.
		prop := newPropagator(cfg)
		otel.SetTextMapPropagator(prop)

		// Set up trace provider.
		tracerProvider, nerr := newTraceProvider(cfg)
		if nerr != nil {
			handleErr(nerr)
			return
		}
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
	}

	if !cfg.TraceDisable {
		// Set up meter provider.
		meterProvider, nerr := newMeterProvider(cfg)
		if nerr != nil {
			handleErr(nerr)
			return
		}
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
	}

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(cfg)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	name := "default"
	if cfg.Name != "" {
		name = cfg.Name
	}

	Tracer = otel.Tracer(name)
	Start = Tracer.Start
	Meter = otel.Meter(name)

	_log = logger{
		log:    otelslog.NewLogger(name),
		source: cfg.LogSource,
	}

	Info = _log.Info
	Error = _log.Error
	Warn = _log.Warn
	With = _log.With

	return
}

func newPropagator(cfg *Config) propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(cfg *Config) (*trace.TracerProvider, error) {
	var exporter trace.SpanExporter
	var err error

	switch cfg.Type {
	case "http":
		exporter, err = otlptracehttp.New(
			context.Background(),
		)
		if err != nil {
			return nil, err
		}
	default:
		exporter, err = stdouttrace.New()
		if err != nil {
			return nil, err
		}
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
	)
	return traceProvider, nil
}

func newMeterProvider(cfg *Config) (*metric.MeterProvider, error) {
	var exporter metric.Exporter
	var err error

	switch cfg.Type {
	case "http":
		exporter, err = otlpmetrichttp.New(
			context.Background(),
		)
		if err != nil {
			return nil, err
		}
	default:
		exporter, err = stdoutmetric.New()
		if err != nil {
			return nil, err
		}
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(3*time.Second))),
		// Default is 1m. Set to 3s for demonstrative purposes.
	)
	return meterProvider, nil
}

func newLoggerProvider(cfg *Config) (*log.LoggerProvider, error) {
	var exporter log.Exporter
	var err error

	switch cfg.Type {
	case "http":
		exporter, err = otlploghttp.New(
			context.Background(),
		)
		if err != nil {
			return nil, err
		}
	default:
		exporter, err = stdoutlog.New()
		if err != nil {
			return nil, err
		}
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exporter)),
	)
	return loggerProvider, nil
}
