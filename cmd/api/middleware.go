package main

import (
	"fmt"
	"net/http"
	"time"

	"prompt-management/pkg/response"
)

// logRequest logs basic information about the incoming HTTP request.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture the status code
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(ww, r)

		app.logger.Info("request processed",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"remote_addr", r.RemoteAddr,
			"status", ww.status,
			"duration", time.Since(start).String(),
		)
	})
}

// recoverPanic recovers from any panic during request processing.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.logger.Error("panic recovered", "error", fmt.Errorf("%s", err))
				response.Error(w, http.StatusInternalServerError, "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// responseWriter is a wrapper for http.ResponseWriter to capture the HTTP status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
