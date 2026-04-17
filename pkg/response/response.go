package response

import (
	"encoding/json"
	"net/http"
)

// Envelope defines the standard JSON response structure.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrDetail  `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// ErrDetail defines the structure for error information.
type ErrDetail struct {
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// JSON sends a successful JSON response.
func JSON(w http.ResponseWriter, status int, data interface{}, meta ...interface{}) {
	var m interface{}
	if len(meta) > 0 {
		m = meta[0]
	}

	env := Envelope{
		Success: true,
		Data:    data,
		Meta:    m,
	}

	send(w, status, env)
}

// Error sends an error JSON response.
func Error(w http.ResponseWriter, status int, message string, details ...interface{}) {
	var d interface{}
	if len(details) > 0 {
		d = details[0]
	}

	env := Envelope{
		Success: false,
		Error: &ErrDetail{
			Message: message,
			Details: d,
		},
	}

	send(w, status, env)
}

func send(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}
