package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

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

	B(ctx)

	otel.Info(ctx, "aaa", "a", "A")
	C(ctx)
}

func B(ctx context.Context) {
	var pcs [10]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(0, pcs[:])

	fs := runtime.CallersFrames(pcs[:])
	for {
		f, more := fs.Next()
		fmt.Printf("code_source:%s:%d %s\n", f.File, f.Line, f.Function)
		time.Sleep(time.Second)
		if !more {
			break
		}
	}
}

func C(ctx context.Context) {
	ctx, span := otel.Start(ctx, "C", trace.WithAttributes(attribute.String("name", "CC")))
	defer span.End()

	ctx = otel.With(ctx, "c", "C")

	otel.Info(ctx, "ccc", "c1", "c")
	D(ctx)
}

func D(ctx context.Context) {
	ctx, span := otel.Start(ctx, "D", trace.WithAttributes(attribute.String("name", "DD")))
	defer span.End()

	ctx = otel.With(ctx, "d", "D")

	otel.Info(ctx, "ddd", "d1", "d")
}

func (a *App) Shutdown(ctx context.Context) error {
	return nil
}

func main() {
	cfg := otel.Config{
		Name:          "tot",
		Type:          "httpc",
		LogSource:     true,
		MetricDisable: false,
		TraceDisable:  false,
	}
	if err := cfg.Run(new(App)); err != nil {
		panic(err)
	}
}
