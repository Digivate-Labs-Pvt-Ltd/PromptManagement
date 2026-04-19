package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"prompt-management/pkg/response"
)

// RecoverPanic returns a middleware that recovers from any panic during request processing.
func RecoverPanic(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.Header().Set("Connection", "close")
					logger.Error("panic recovered", "error", fmt.Errorf("%s", err))
					response.Error(w, http.StatusInternalServerError, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
