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

// ItemHandler handles HTTP requests for prompt items (versions).
type ItemHandler struct {
	service *service.ItemService
}

// NewItemHandler creates a new ItemHandler.
func NewItemHandler(s *service.ItemService) *ItemHandler {
	return &ItemHandler{service: s}
}

// Add handles the POST /prompts/items/add action.
func (h *ItemHandler) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input service.AddItemRequest
	if err := validator.DecodeAndValidate(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.service.Add(r.Context(), input, userID)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "invalid input")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to add prompt item version")
		return
	}

	response.JSON(w, http.StatusCreated, item)
}

// Get handles the POST /prompts/items/get action.
func (h *ItemHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	item, err := h.service.GetByID(r.Context(), input.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "prompt item not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get prompt item")
		return
	}

	response.JSON(w, http.StatusOK, item)
}

// List handles the POST /prompts/items/list action.
func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input domain.ItemListFilters
	if err := validator.DecodeAndValidate(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	items, total, err := h.service.List(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			response.Error(w, http.StatusBadRequest, "invalid input missing management id")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to list prompt items")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
	})
}

// Promote handles the POST /prompts/items/promote action.
func (h *ItemHandler) Promote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var input struct {
		ManagementID string `json:"management_id"`
		ItemID       string `json:"item_id"`
	}
	if err := validator.DecodeAndValidate(r, &input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Promote(r.Context(), input.ManagementID, input.ItemID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "prompt item or group not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "prompt item promoted successfully"})
}

// Archive handles the POST /prompts/items/archive action.
func (h *ItemHandler) Archive(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.Archive(r.Context(), input.ID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "prompt item not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to archive prompt item")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "prompt item archived successfully"})
}
