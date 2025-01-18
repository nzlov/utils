package main

import (
	"log/slog"

	"github.com/nzlov/utils/zinc"
)

func main() {
	cfg := zinc.Config{
		Host:     "http://localhost:4080",
		User:     "admin",
		Password: "",
		Index:    "test",
	}

	z := cfg.With(zinc.WithCache(1))
	slog.SetDefault(
		slog.New(slog.NewJSONHandler(z, &slog.HandlerOptions{
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.MessageKey {
					a.Key = "message"
				}
				return a
			},
		})),
	)

	slog.Info("test1")
	slog.Info("test2")
	slog.Info("test3")
	slog.Info("test4")
	slog.Info("test5")
	select {}
}
