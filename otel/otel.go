package otel

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
	Write([]byte) (int, error)
	Printf(string, ...any)
}

type logger struct {
	log    *slog.Logger
	source bool
}

func (l *logger) sources() []any {
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(4, pcs[:])

	fs := runtime.CallersFrames([]uintptr{pcs[0]})
	f, _ := fs.Next()
	return []any{
		"code_source", fmt.Sprintf("%s:%d %s", f.File, f.Line, f.Function),
	}
}

func (l *logger) Info(ctx context.Context, msg string, args ...any) {
	if l.source {
		args = append(l.sources(), args...)
	}
	l.log.InfoContext(ctx, msg, args...)
}

func (l *logger) Error(ctx context.Context, msg string, args ...any) {
	if l.source {
		args = append(l.sources(), args...)
	}
	l.log.ErrorContext(ctx, msg, args...)
}

func (l *logger) Warn(ctx context.Context, msg string, args ...any) {
	if l.source {
		args = append(l.sources(), args...)
	}
	l.log.WarnContext(ctx, msg, args...)
}

func (l *logger) With(args ...any) Logger {
	return &logger{log: l.log.With(args...), source: l.source}
}

func (l *logger) Write(p []byte) (n int, err error) {
	l.log.Info(string(p))
	return len(p), nil
}

func (l *logger) Printf(f string, args ...any) {
	l.log.Info(fmt.Sprintf(f, args...))
}

var Log Logger

var (
	Tracer trace.Tracer
	Start  func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
)

var Meter metric.Meter

type ctxLogKey struct{}

var _ctxLogKey = &ctxLogKey{}

func forLog(ctx context.Context) Logger {
	contextLog := ctx.Value(_ctxLogKey)
	if contextLog != nil {
		return contextLog.(Logger)
	}
	return Log
}

func Ctx(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, _ctxLogKey, log)
}

func Info(ctx context.Context, msg string, args ...any) {
	forLog(ctx).Info(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	forLog(ctx).Error(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	forLog(ctx).Warn(ctx, msg, args...)
}

func With(ctx context.Context, args ...any) context.Context {
	contextLog := ctx.Value(_ctxLogKey)
	if contextLog != nil {
		l := contextLog.(Logger).With(args...)
		return Ctx(ctx, l)
	}
	l := Log.With(args...)
	return Ctx(ctx, l)
}
