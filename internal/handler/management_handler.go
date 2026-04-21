package handler

import (
	"errors"
	"net/http"

	"prompt-management/internal/domain"
	"prompt-management/internal/middleware"
	"prompt-management/internal/service"
	"prompt-management/pkg/response"
	"prompt-management/pkg/validator"
)

// ManagementHandler handles HTTP requests for prompt management.
type ManagementHandler struct {
	service *service.ManagementService
}

// NewManagementHandler creates a new ManagementHandler.
func NewManagementHandler(s *service.ManagementService) *ManagementHandler {
	return &ManagementHandler{service: s}
}

// Create handles the POST /prompts/create action.
func (h *ManagementHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input service.CreateRequest
	if err := validator.DecodeAndValidate(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	pm, err := h.service.Create(r.Context(), input, userID)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "invalid input")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create prompt group")
		return
	}

	response.JSON(w, http.StatusCreated, pm)
}

// Update handles the POST /prompts/update action.
func (h *ManagementHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input service.UpdateRequest
	if err := validator.DecodeAndValidate(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	pm, err := h.service.Update(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "prompt group not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update prompt group")
		return
	}

	response.JSON(w, http.StatusOK, pm)
}

// Get handles the POST /prompts/get action.
func (h *ManagementHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input struct {
		ID string `json:"id"`
	}
	if err := validator.DecodeAndValidate(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	pm, err := h.service.GetByID(r.Context(), input.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "prompt group not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get prompt group")
		return
	}

	response.JSON(w, http.StatusOK, pm)
}

// List handles the POST /prompts/list action.
func (h *ManagementHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input domain.ListFilters
	if err := validator.DecodeAndValidate(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	groups, total, err := h.service.List(r.Context(), input)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list prompt groups")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"groups": groups,
		"total":  total,
	})
}

// Delete handles the POST /prompts/delete action.
func (h *ManagementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input struct {
		ID string `json:"id"`
	}
	if err := validator.DecodeAndValidate(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Delete(r.Context(), input.ID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "prompt group not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete prompt group")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "prompt group deleted successfully"})
}

// CreateFull handles the POST /prompts/create-full action.
func (h *ManagementHandler) CreateFull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input service.BulkCreateRequest
	if err := validator.DecodeAndValidate(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	pm, err := h.service.CreateFull(r.Context(), input, userID)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create full prompt group")
		return
	}

	response.JSON(w, http.StatusCreated, pm)
}

// ListFull handles the POST /prompts/list-full action.
func (h *ManagementHandler) ListFull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input domain.ListFilters
	if err := validator.DecodeAndValidate(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	groups, total, err := h.service.ListFull(r.Context(), input)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list full prompt groups")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"groups": groups,
		"total":  total,
	})
}

