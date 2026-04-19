package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"prompt-management/internal/handler"
)

func TestHealthCheck(t *testing.T) {
	// 1. Setup minimal application
	
	// We pass nil for DB since we are testing routing
	healthHandler := handler.NewHealthHandler(nil)
	
	mux := handler.NewRouter(handler.RouterConfig{
		Health: healthHandler,
		Auth:   nil, // Not testing auth here
	})

	t.Run("GET /healthz - Liveness OK", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("POST /healthz - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
		rr := httptest.NewRecorder()
		
		mux.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", rr.Code)
		}
	})

	t.Run("GET /readyz - Database Ping (Panic expected)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rr := httptest.NewRecorder()
		
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic due to nil db pool")
			}
		}()

		mux.ServeHTTP(rr, req)
	})
}
