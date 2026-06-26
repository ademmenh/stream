package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"go-starter/internal/app"
	"go-starter/internal/config/infrastructure"
)

func main() {
	cfg := infrastructure.NewConfigAdapter()
	app := app.CreateApp(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := app.Start(":" + cfg.Port()); err != nil {
			slog.Error("server shutdown", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	app.Shutdown(shutdownCtx)
}
