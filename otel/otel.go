package otel

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	log    *slog.Logger
	source bool
}

func (l *Logger) sources() []any {
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(3, pcs[:])

	fs := runtime.CallersFrames([]uintptr{pcs[0]})
	f, _ := fs.Next()
	return []any{
		"code_source", fmt.Sprintf("%s:%d %s", f.File, f.Line, f.Function),
	}
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	if l.source {
		args = append(l.sources(), args...)
	}
	l.log.InfoContext(ctx, msg, args...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	if l.source {
		args = append(l.sources(), args...)
	}
	l.log.ErrorContext(ctx, msg, args...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	if l.source {
		args = append(l.sources(), args...)
	}
	l.log.WarnContext(ctx, msg, args...)
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{log: l.log.With(args...), source: l.source}
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.log.Info(string(p))
	return len(p), nil
}

func (l *Logger) Printf(f string, args ...any) {
	l.log.Info(fmt.Sprintf(f, args...))
}

var (
	Log *Logger

	Info  func(ctx context.Context, msg string, args ...any)
	Error func(ctx context.Context, msg string, args ...any)
	Warn  func(ctx context.Context, msg string, args ...any)
	With  func(args ...any) *Logger
)

var (
	Tracer trace.Tracer
	Start  func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
)

var Meter metric.Meter

type ctxLogKey struct{}

var _ctxLogKey = &ctxLogKey{}

func ForLog(ctx context.Context, args ...any) (*Logger, context.Context) {
	contextLog := ctx.Value(_ctxLogKey)
	if contextLog != nil {
		if len(args) == 0 {
			return contextLog.(*Logger), ctx
		}
		l := contextLog.(*Logger).With(args...)
		return l, CtxLog(ctx, l)
	}
	l := With(args...)
	return l, CtxLog(ctx, l)
}

func CtxLog(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, _ctxLogKey, log)
}
