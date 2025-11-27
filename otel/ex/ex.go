package main

import (
	"context"

	"github.com/nzlov/utils/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type App struct{}

func (a *App) Run(ctx context.Context) error {
	A(ctx)
	return nil
}

func A(ctx context.Context) {
	ctx, span := otel.Start(ctx, "A", trace.WithAttributes(attribute.String("name", "AA")))
	defer span.End()

	otel.Info(ctx, "aaa", "a", "A")
	B(ctx)
}

func B(ctx context.Context) {
	ctx, span := otel.Start(ctx, "B", trace.WithAttributes(attribute.String("name", "BB")))
	defer span.End()

	otel.With("b", "B").Info(ctx, "bbb")
	C(ctx)
}

func C(ctx context.Context) {
	ctx, span := otel.Start(ctx, "C", trace.WithAttributes(attribute.String("name", "CC")))
	defer span.End()

	log, ctx := otel.ForLog(ctx, "c", "C")

	log.Info(ctx, "ccc", "c1", "c")
	D(ctx)
}

func D(ctx context.Context) {
	ctx, span := otel.Start(ctx, "D", trace.WithAttributes(attribute.String("name", "DD")))
	defer span.End()

	log, ctx := otel.ForLog(ctx, "d", "D")

	log.Info(ctx, "ddd", "d1", "d")
}

func (a *App) Shutdown(ctx context.Context) error {
	return nil
}

func main() {
	cfg := otel.Config{
		Name:          "tot",
		Type:          "httpc",
		LogSource:     true,
		MetricDisable: true,
		TraceDisable:  true,
	}
	if err := cfg.Run(new(App)); err != nil {
		panic(err)
	}
}
