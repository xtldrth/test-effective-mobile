package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xtldrth/test-efmov-go/internal"
	"github.com/xtldrth/test-efmov-go/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	configPathPtr := flag.String("config", "", "config path")
	flag.Parse()
	var cfg config.Config
	if *configPathPtr == "" {
		logger.Warn("config weren't provided, using default config")
		cfg = config.DefaultConfig()
	} else {
		cfg = config.MustParseConfig(*configPathPtr)
	}
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
