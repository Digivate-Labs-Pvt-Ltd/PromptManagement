package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter is a wrapper for http.ResponseWriter to capture the HTTP status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// LogRequest returns a middleware that logs basic information about the incoming HTTP request.
func LogRequest(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap ResponseWriter to capture the status code
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(ww, r)

			logger.Info("request processed",
				"method", r.Method,
				"uri", r.URL.RequestURI(),
				"remote_addr", r.RemoteAddr,
				"status", ww.status,
				"duration", time.Since(start).String(),
			)
		})
	}
}
