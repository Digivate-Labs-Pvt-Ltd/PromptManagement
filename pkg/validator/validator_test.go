package validator

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestDecodeAndValidate(t *testing.T) {
	t.Run("Valid body", func(t *testing.T) {
		body := `{"name": "test", "age": 25}`
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
		var dst mockRequest

		err := DecodeAndValidate(r, &dst)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if dst.Name != "test" || dst.Age != 25 {
			t.Errorf("wrong data: %+v", dst)
		}
	})

	t.Run("Empty body", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(""))
		var dst mockRequest
		err := DecodeAndValidate(r, &dst)
		if err == nil || err.Error() != "body must not be empty" {
			t.Errorf("expected 'body must not be empty', got %v", err)
		}
	})

	t.Run("Bad JSON", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{invalid}"))
		var dst mockRequest
		err := DecodeAndValidate(r, &dst)
		if err == nil {
			t.Error("expected error for bad JSON")
		}
	})

	t.Run("Unknown field", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name": "test", "extra": true}`))
		var dst mockRequest
		err := DecodeAndValidate(r, &dst)
		if err == nil {
			t.Error("expected error for unknown field")
		}
	})
}
