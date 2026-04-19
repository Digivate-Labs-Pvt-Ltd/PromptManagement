package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"prompt-management/pkg/response"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	db *pgxpool.Pool
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// Liveness handles the /healthz endpoint.
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

// Readiness handles the /readyz endpoint.
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		response.Error(w, http.StatusServiceUnavailable, "database not reachable")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "READY"})
}
