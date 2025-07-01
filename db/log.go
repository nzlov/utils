package db

import (
	"fmt"
)

type Writer struct {
	log func(msg string, args ...any)
}

func NewWriter(log func(msg string, args ...any)) *Writer {
	return &Writer{log: log}
}

func (w *Writer) Printf(f string, args ...any) {
	w.log(fmt.Sprintf(f, args...))
}
