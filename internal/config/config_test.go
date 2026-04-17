package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set an env var
	os.Setenv("PORT", "9999")
	defer os.Unsetenv("PORT")

	cfg := Load()

	if cfg.Port != "9999" {
		t.Errorf("expected PORT 9999, got %s", cfg.Port)
	}

	// Check default
	if cfg.JWTSecret != "super-secret-jwt-key" {
		t.Errorf("expected default JWTSecret, got %s", cfg.JWTSecret)
	}
}
