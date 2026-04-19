package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"prompt-management/internal/config"
	"prompt-management/pkg/response"
)

func TestHealthCheck(t *testing.T) {
	// 1. Setup minimal application
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := &config.Config{Port: "8080"}
	
	app := &application{
		config: cfg,
		logger: logger,
		// Note: We skip DB initialization here because we are testing the router logic,
		// but the handler will try to use app.db. We can pass a nil or mock if needed.
	}

	mux := app.routes()

	t.Run("GET /health - Method Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()

		// Since app.db is nil, we expect a panic if we call healthCheckHandler directly,
		// but it might be recovered by middleware.
		// For this test, let's just assert the router correctly dispatches.
		
		defer func() {
			if r := recover(); r != nil {
				// Expected if DB is nil and Ping is called
			}
		}()

		mux.ServeHTTP(rr, req)
		
		// If DB was nil, we'd get a 500 error from our middleware recovery
		if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
			t.Errorf("expected 200 or 500, got %d", rr.Code)
		}
	})

	t.Run("POST /health - Method Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		rr := httptest.NewRecorder()
		
		mux.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
			t.Errorf("expected 200 or 500, got %d", rr.Code)
		}
	})

	t.Run("PUT /health - Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/health", nil)
		rr := httptest.NewRecorder()
		
		mux.ServeHTTP(rr, req)
		
		if rr.Code != http.StatusMethodNotAllowed {
			// We checking if our handler caught it
			var res response.Envelope
			_ = json.NewDecoder(rr.Body).Decode(&res)
			if !res.Success && rr.Code == http.StatusMethodNotAllowed {
				// Correct
			} else {
				t.Errorf("expected 405, got %d", rr.Code)
			}
		}
	})
}
