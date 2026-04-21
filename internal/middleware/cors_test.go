package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnableCORS(t *testing.T) {
	// 1. Setup a simple test handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 2. Wrap with EnableCORS
	corsHandler := EnableCORS(nextHandler)

	// 3. Test a standard POST request
	t.Run("Standard POST request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		expectedOrigin := "http://localhost:4200"
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != expectedOrigin {
			t.Errorf("expected Access-Control-Allow-Origin %s, got %s", expectedOrigin, origin)
		}

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	// 4. Test an OPTIONS preflight request
	t.Run("OPTIONS preflight request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()

		corsHandler.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status code %d for preflight, got %d", http.StatusNoContent, w.Code)
		}

		expectedMethods := "POST, GET, OPTIONS, PUT, DELETE"
		if methods := w.Header().Get("Access-Control-Allow-Methods"); methods != expectedMethods {
			t.Errorf("expected Access-Control-Allow-Methods %s, got %s", expectedMethods, methods)
		}
	})
}
