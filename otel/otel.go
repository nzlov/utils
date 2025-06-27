package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	_log  *slog.Logger
	Info  func(ctx context.Context, msg string, args ...any)
	Error func(ctx context.Context, msg string, args ...any)
)

var (
	Tracer trace.Tracer
	Start  func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
)

var Meter metric.Meter

type ctxLogKey struct{}

var _ctxLogKey = &ctxLogKey{}

func ForLog(ctx context.Context, args ...any) (*slog.Logger, context.Context) {
	contextLog := ctx.Value(_ctxLogKey)
	if contextLog != nil {
		if len(args) == 0 {
			return contextLog.(*slog.Logger), ctx
		}
		l := contextLog.(*slog.Logger).With(args...)
		return l, CtxLog(ctx, l)
	}
	l := _log.With(args...)
	return l, CtxLog(ctx, l)
}

func CtxLog(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, _ctxLogKey, log)
}
