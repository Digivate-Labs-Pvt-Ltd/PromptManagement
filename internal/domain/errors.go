package domain

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound   = errors.New("resource not found")
	ErrConflict   = errors.New("resource already exists")
	ErrForbidden  = errors.New("permission denied")
	ErrValidation = errors.New("validation failed")
	ErrInternal   = errors.New("internal server error")
	ErrUnauthorized = errors.New("unauthorized access")
)

// MapToHTTPStatus translates a domain error to an HTTP status code.
func MapToHTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
