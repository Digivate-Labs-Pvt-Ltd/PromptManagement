package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"prompt-management/internal/config"
	"prompt-management/internal/repository/postgres"
)

type application struct {
	config *config.Config
	logger *slog.Logger
	db     *pgxpool.Pool
}

func main() {
	// 1. Initialize Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. Load Config
	cfg := config.Load()

	// 3. Initialize DB Pool
	dbPool, err := postgres.NewPool(cfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	app := &application{
		config: cfg,
		logger: logger,
		db:     dbPool,
	}

	// 4. Setup Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 5. Graceful Shutdown Channel
	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	// 6. Start Server
	logger.Info("starting server", "addr", srv.Addr)
	err = srv.ListenAndServe()
	if err != http.ErrServerClosed {
		logger.Error("server failed to start", "error", err)
		os.Exit(1)
	}

	// 7. Wait for Shutdown Result
	err = <-shutdownError
	if err != nil {
		logger.Error("error during graceful shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("stopped server", "addr", srv.Addr)
}
