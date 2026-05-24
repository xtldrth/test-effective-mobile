package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xtldrth/test-efmov-go/internal"
	"github.com/xtldrth/test-efmov-go/internal/config"
)

func main() {
	cfg := config.DefaultConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app, err := internal.NewApp(cfg, 10*time.Second, logger.WithGroup("App"))
	if err != nil {
		logger.Error("new app", slog.String("error", err.Error()))
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	if err := app.Run(ctx); err != nil {
		logger.Error("app run", slog.String("error", err.Error()))
	}
}
