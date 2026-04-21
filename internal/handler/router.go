package handler

import (
	"net/http"
	"prompt-management/internal/config"
	"prompt-management/internal/middleware"
)

// RouterConfig holds the handlers for the router.
type RouterConfig struct {
	Config     *config.Config
	Health     *HealthHandler
	Auth       *AuthHandler
	Management *ManagementHandler
	Item       *ItemHandler
}

// NewRouter registers all routes and returns the central mux.
func NewRouter(cfg RouterConfig) *http.ServeMux {
	mux := http.NewServeMux()

	// Health Endpoints (GET as requested in Issue #13)
	mux.HandleFunc("/healthz", cfg.Health.Liveness)
	mux.HandleFunc("/readyz", cfg.Health.Readiness)

	// Auth Endpoints (POST only as per interlock rules)
	mux.HandleFunc("/auth/register", cfg.Auth.Register)
	mux.HandleFunc("/auth/login", cfg.Auth.Login)

	// Prompt Management (Authenticated)
	authMW := func(next http.HandlerFunc) http.Handler {
		return middleware.Authenticate(cfg.Config, next)
	}

	mux.Handle("/auth/refresh", authMW(cfg.Auth.Refresh))

	// Group Endpoints
	mux.Handle("/prompts/create", authMW(cfg.Management.Create))
	mux.Handle("/prompts/create-full", authMW(cfg.Management.CreateFull))
	mux.Handle("/prompts/update", authMW(cfg.Management.Update))
	mux.Handle("/prompts/get", authMW(cfg.Management.Get))
	mux.Handle("/prompts/list", authMW(cfg.Management.List))
	mux.Handle("/prompts/delete", authMW(cfg.Management.Delete))

	// Item Endpoints
	mux.Handle("/prompts/items/add", authMW(cfg.Item.Add))
	mux.Handle("/prompts/items/list", authMW(cfg.Item.List))
	mux.Handle("/prompts/items/get", authMW(cfg.Item.Get))
	mux.Handle("/prompts/items/promote", authMW(cfg.Item.Promote))
	mux.Handle("/prompts/items/archive", authMW(cfg.Item.Archive))

	return mux
}
