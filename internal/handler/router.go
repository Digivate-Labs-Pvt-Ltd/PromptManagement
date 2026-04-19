package handler

import (
	"net/http"
)

// RouterConfig holds the handlers for the router.
type RouterConfig struct {
	Health *HealthHandler
	Auth   *AuthHandler
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

	return mux
}
