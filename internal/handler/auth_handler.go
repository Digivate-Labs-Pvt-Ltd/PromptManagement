package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"prompt-management/internal/domain"
	"prompt-management/internal/middleware"
	"prompt-management/internal/service"
	"prompt-management/pkg/response"
	"prompt-management/pkg/validator"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input registerRequest
	err := validator.DecodeAndValidate(r, &input)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Register(r.Context(), input.Email, input.Username, input.FullName, input.Password)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			response.Error(w, http.StatusConflict, "user with this email or username already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	response.JSON(w, http.StatusCreated, user)
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type loginResponse struct {
	Value string `json:"Value"`
	Error string `json:"Error"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendLoginResponse(w, http.StatusMethodNotAllowed, "", "method not allowed")
		return
	}

	var input loginRequest
	err := validator.DecodeAndValidate(r, &input)
	if err != nil {
		h.sendLoginResponse(w, http.StatusBadRequest, "", err.Error())
		return
	}

	// 3. Generate JWT
	token, err := h.service.Login(r.Context(), input.Identifier, input.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			h.sendLoginResponse(w, http.StatusUnauthorized, "", "invalid credentials")
			return
		}
		h.sendLoginResponse(w, http.StatusInternalServerError, "", "login failed")
		return
	}

	h.sendLoginResponse(w, http.StatusOK, token, "")
}

func (h *AuthHandler) sendLoginResponse(w http.ResponseWriter, status int, value, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(loginResponse{
		Value: value,
		Error: err,
	})
}

// Refresh handles stateless token extensions.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// The UserID is inserted directly by the active authenticate middleware.
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok || userID == "" {
		response.Error(w, http.StatusUnauthorized, "invalid session context identity")
		return
	}

	token, err := h.service.RefreshToken(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "token refresh generation failed")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"token": token})
}
