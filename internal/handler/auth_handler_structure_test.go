package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginResponseStructure(t *testing.T) {
	h := &AuthHandler{} // Mock service is not needed just for checking structure if we mock the handler call or just test the helper

	t.Run("sendLoginResponse Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.sendLoginResponse(w, http.StatusOK, "test-token", "")

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp loginResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Value != "test-token" {
			t.Errorf("expected Value 'test-token', got '%s'", resp.Value)
		}
		if resp.Error != "" {
			t.Errorf("expected Error empty, got '%s'", resp.Error)
		}
	})

	t.Run("sendLoginResponse Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.sendLoginResponse(w, http.StatusUnauthorized, "", "invalid credentials")

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}

		var resp loginResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Value != "" {
			t.Errorf("expected Value empty, got '%s'", resp.Value)
		}
		if resp.Error != "invalid credentials" {
			t.Errorf("expected Error 'invalid credentials', got '%s'", resp.Error)
		}
	})
}
