package slog

import "log/slog"

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
