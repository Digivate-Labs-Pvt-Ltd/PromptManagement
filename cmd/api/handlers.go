package main

import (
	"context"
	"net/http"
	"time"

	"prompt-management/pkg/response"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Validate Method (Allow GET for infra, POST per project rules)
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 2. Check Database Connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := app.db.Ping(ctx)
	status := "available"
	if err != nil {
		app.logger.Error("database health check failed", "error", err)
		status = "unavailable"
	}

	// 3. Prepare response data
	data := map[string]interface{}{
		"status":      status,
		"environment": "development", // TODO: pull from config if needed
		"version":     "1.0.0",
	}

	// 4. Send response
	if status == "unavailable" {
		response.Error(w, http.StatusServiceUnavailable, "system is partially unavailable", data)
		return
	}

	response.JSON(w, http.StatusOK, data)
}
