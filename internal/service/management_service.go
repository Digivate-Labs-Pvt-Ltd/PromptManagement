package service

import (
	"context"

	"prompt-management/internal/domain"
)

// ManagementService handles business logic for prompt management.
type ManagementService struct {
	repo domain.ManagementStore
}

// NewManagementService creates a new ManagementService.
func NewManagementService(repo domain.ManagementStore) *ManagementService {
	return &ManagementService{repo: repo}
}

// CreateRequest defines the payload for creating a prompt group.
type CreateRequest struct {
	Client       string `json:"client"`
	UseCase      string `json:"use_case"`
	DocumentType string `json:"document_type"`
	Category     string `json:"category"`
	StageName    string `json:"stage_name"`
}

// UpdateRequest defines the payload for updating a prompt group.
type UpdateRequest struct {
	ID           string `json:"id"`
	Client       string `json:"client"`
	UseCase      string `json:"use_case"`
	DocumentType string `json:"document_type"`
	Category     string `json:"category"`
	StageName    string `json:"stage_name"`
}

func (s *ManagementService) Create(ctx context.Context, req CreateRequest, userID string) (*domain.PromptManagement, error) {
	// Basic validation
	if req.Client == "" || req.UseCase == "" || req.DocumentType == "" || req.Category == "" || req.StageName == "" {
		return nil, domain.ErrValidation
	}

	pm := &domain.PromptManagement{
		Client:       req.Client,
		UseCase:      req.UseCase,
		DocumentType: req.DocumentType,
		Category:     req.Category,
		StageName:    req.StageName,
		CreatedBy:    userID,
	}

	if err := s.repo.Insert(ctx, pm); err != nil {
		return nil, err
	}

	return pm, nil
}

func (s *ManagementService) Update(ctx context.Context, req UpdateRequest) (*domain.PromptManagement, error) {
	if req.ID == "" {
		return nil, domain.ErrValidation
	}

	pm, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Client != "" {
		pm.Client = req.Client
	}
	if req.UseCase != "" {
		pm.UseCase = req.UseCase
	}
	if req.DocumentType != "" {
		pm.DocumentType = req.DocumentType
	}
	if req.Category != "" {
		pm.Category = req.Category
	}
	if req.StageName != "" {
		pm.StageName = req.StageName
	}

	if err := s.repo.Update(ctx, pm); err != nil {
		return nil, err
	}

	return pm, nil
}

func (s *ManagementService) GetByID(ctx context.Context, id string) (*domain.PromptManagement, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ManagementService) List(ctx context.Context, filters domain.ListFilters) ([]*domain.PromptManagement, int, error) {
	return s.repo.List(ctx, filters)
}

func (s *ManagementService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
