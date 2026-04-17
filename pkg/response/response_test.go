package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"foo": "bar"}
	meta := map[string]int{"total": 1}

	JSON(w, http.StatusOK, data, meta)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var env Envelope
	err := json.NewDecoder(w.Body).Decode(&env)
	if err != nil {
		t.Fatal(err)
	}

	if !env.Success {
		t.Error("expected success to be true")
	}

	resData := env.Data.(map[string]interface{})
	if resData["foo"] != "bar" {
		t.Errorf("expected data foo=bar, got %v", resData["foo"])
	}

	resMeta := env.Meta.(map[string]interface{})
	if resMeta["total"] != 1.0 { // json decodes numbers as float64
		t.Errorf("expected meta total=1, got %v", resMeta["total"])
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusBadRequest, "bad request", map[string]string{"field": "required"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var env Envelope
	err := json.NewDecoder(w.Body).Decode(&env)
	if err != nil {
		t.Fatal(err)
	}

	if env.Success {
		t.Error("expected success to be false")
	}

	if env.Error.Message != "bad request" {
		t.Errorf("expected message 'bad request', got %s", env.Error.Message)
	}

	details := env.Error.Details.(map[string]interface{})
	if details["field"] != "required" {
		t.Errorf("expected detail field=required, got %v", details["field"])
	}
}
