package slog

import "log/slog"

type Writer struct {
	log *slog.Logger

	source string
}

func NewWriter(log *slog.Logger) *Writer {
	return NewWriterSource(log, "writer")
}

func NewWriterSource(log *slog.Logger, source string) *Writer {
	return &Writer{source: source, log: log}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.log.Info(string(p), "source", w.source)
	return len(p), nil
}
