package slog

import (
	"context"
	"fmt"
	"log/slog"
)

type loggerKey struct{}

var _loggerKey = loggerKey{}

type Writer struct {
	log *slog.Logger
}

func NewWriter(log *slog.Logger) *Writer {
	return &Writer{log: log}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.log.Info(string(p))
	return len(p), nil
}

func (w *Writer) Printf(f string, args ...any) {
	w.log.Info(fmt.Sprintf(f, args...))
}

func For(c context.Context) *slog.Logger {
	cfg := c.Value(_loggerKey)
	if cfg != nil {
		return cfg.(*slog.Logger)
	}
	return slog.Default()
}

func Set(c context.Context, l *slog.Logger) context.Context {
	return context.WithValue(c, _loggerKey, l)
}

func SetDef(c context.Context) context.Context {
	return context.WithValue(c, _loggerKey, slog.Default())
}
