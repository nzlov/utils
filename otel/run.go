package otel

import (
	"context"
	"errors"
	"os"
	"os/signal"
)

type Run interface {
	Run() error
	Shutdown(context.Context) error
}

type run struct {
	f func() error
}

func (r run) Run() error {
	return r.f()
}

func (r run) Shutdown(ctx context.Context) error {
	return nil
}

func (cfg Config) RunFunc(f func() error) (err error) {
	return cfg.Run(run{f})
}

func (cfg Config) Run(r Run) (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := cfg.SetupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- r.Run()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, returns ErrServerClosed.
	err = r.Shutdown(context.Background())
	return
}
