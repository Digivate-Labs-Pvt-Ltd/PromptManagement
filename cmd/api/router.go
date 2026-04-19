package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// 1. Register handlers
	mux.HandleFunc("/health", app.healthCheckHandler)

	// 2. Wrap with middleware
	return app.recoverPanic(app.logRequest(mux))
}
