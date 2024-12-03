package gin

import "log/slog"

type GinSlog struct {
	log *slog.Logger
}

func NewGinSlog(log *slog.Logger) *GinSlog {
	return &GinSlog{log}
}

func (g *GinSlog) Write(p []byte) (n int, err error) {
	g.log.Info(string(p), "source", "gin")
	return len(p), nil
}
