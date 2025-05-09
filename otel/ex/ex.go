package main

import (
	"context"
	"log"
	"net/http"

	"github.com/nzlov/utils/otel"
	exotel "github.com/nzlov/utils/otel/ex/otel"
)

func main() {
	cfg := &otel.Config{}

	if err := cfg.Run(&App{}); err != nil {
		log.Fatal(err)
	}
}

type App struct {
	srv *http.Server
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run() error {
	a.srv = &http.Server{
		Addr: ":9999",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := exotel.Tracer.Start(r.Context(), "handler")
			defer span.End()

			exotel.Info(ctx, r.URL.String())
		}),
	}

	return a.srv.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.srv.Shutdown(ctx)
}
