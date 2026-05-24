package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xtldrth/test-efmov-go/internal/config"
)

type App interface {
	Run(ctx context.Context) error
}

type app struct {
	server *http.Server
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func setupDB(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("pool new: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pool ping: %w", err)
	}
	return pool, nil
}

func NewApp(cfg config.Config, initTimeout time.Duration, logger *slog.Logger) (App, error) {
	ctx, cancel := context.WithTimeout(context.Background(), initTimeout)
	defer cancel()
	pool, err := setupDB(ctx, cfg.DB.ConnStr)
	if err != nil {
		return nil, fmt.Errorf("setup db: %w", err)
	}
	subscriptionsRepository := NewPSQLSubscriptionsRepository(pool)
	subscriptionsService := NewSubscriptionsService(
		subscriptionsRepository,
		cfg.App.SubscriptionsServiceTimeout,
		logger.WithGroup("Subscriptions Service"),
	)
	subscriptionsHandler := NewSubscriptionsHandler(subscriptionsService, logger.WithGroup("Subscriptions Handler"))
	mux := http.NewServeMux()
	RegisterSubscriptionsRoutes(mux, subscriptionsHandler)
	return app{
		pool: pool,
		server: &http.Server{
			Addr:    cfg.Server.Addr,
			Handler: mux,
		},
		logger: logger,
	}, nil
}

func (a app) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		a.logger.Info("shutting down application, to force it stop press CTRL^C")
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var wg sync.WaitGroup
		wg.Go(
			func() {
				if err := a.server.Shutdown(c); err != nil {
					a.logger.Error("server shutdown", slog.String("error", err.Error()))
				}
			},
		)
		wg.Go(func() {
			done := make(chan struct{})
			go func() {
				a.pool.Close()
				done <- struct{}{}
			}()
			select {
			case <-c.Done():
				a.logger.Error("pool close", slog.String("error", ctx.Err().Error()))
			case <-done:
			}
		})
		wg.Wait()
		return
	}()
	a.logger.Info("app started", slog.String("addr", a.server.Addr))
	if err := a.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
	return nil
}
