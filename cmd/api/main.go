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
	"prompt-management/internal/handler"
	"prompt-management/internal/middleware"
	"prompt-management/internal/repository/postgres"
	"prompt-management/internal/service"
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

	// 4. Initialize Handlers/Services
	userRepo := postgres.NewUserRepository(dbPool)
	authService := service.NewAuthService(cfg, userRepo)
	authHandler := handler.NewAuthHandler(authService)

	mgmtRepo := postgres.NewManagementRepository(dbPool)
	mgmtService := service.NewManagementService(mgmtRepo)
	mgmtHandler := handler.NewManagementHandler(mgmtService)

	itemRepo := postgres.NewItemRepository(dbPool)
	itemService := service.NewItemService(itemRepo)
	itemHandler := handler.NewItemHandler(itemService)

	healthHandler := handler.NewHealthHandler(dbPool)

	// 5. Setup Router
	mux := handler.NewRouter(handler.RouterConfig{
		Config:     cfg,
		Health:     healthHandler,
		Auth:       authHandler,
		Management: mgmtHandler,
		Item:       itemHandler,
	})

	// 6. Apply Global Middleware
	// Order: Recovery -> Logger -> CORS -> Mux
	var handler http.Handler = mux
	handler = middleware.EnableCORS(handler)
	handler = middleware.LogRequest(logger)(handler)
	handler = middleware.RecoverPanic(logger)(handler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 7. Graceful Shutdown Channel
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

	// 8. Start Server
	logger.Info("starting server", "addr", srv.Addr)
	err = srv.ListenAndServe()
	if err != http.ErrServerClosed {
		logger.Error("server failed to start", "error", err)
		os.Exit(1)
	}

	// 9. Wait for Shutdown Result
	err = <-shutdownError
	if err != nil {
		logger.Error("error during graceful shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("stopped server", "addr", srv.Addr)
}
