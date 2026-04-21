package service

import (
	"context"
	"fmt"

	"prompt-management/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ManagementService handles business logic for prompt management.
type ManagementService struct {
	db       *pgxpool.Pool
	repo     domain.ManagementStore
	itemRepo domain.ItemStore
}

// NewManagementService creates a new ManagementService.
func NewManagementService(db *pgxpool.Pool, repo domain.ManagementStore, itemRepo domain.ItemStore) *ManagementService {
	return &ManagementService{
		db:       db,
		repo:     repo,
		itemRepo: itemRepo,
	}
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

// BulkCreateRequest defines the payload for creating a group and multiple items.
type BulkCreateRequest struct {
	CreateRequest
	Items []AddItemRequest `json:"items"`
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
		CreatedByID:  userID,
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
	pm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Fetch associated prompts (all status/versions as requested for history)
	items, _, err := s.itemRepo.List(ctx, domain.ItemListFilters{
		ManagementID: id,
	})
	if err != nil {
		return nil, err
	}

	pm.Prompts = items
	return pm, nil
}

func (s *ManagementService) List(ctx context.Context, filters domain.ListFilters) ([]*domain.PromptManagement, int, error) {
	return s.repo.List(ctx, filters)
}

func (s *ManagementService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ManagementService) CreateFull(ctx context.Context, req BulkCreateRequest, userID string) (*domain.PromptManagement, error) {
	// 1. Validation
	if req.Client == "" || req.UseCase == "" || req.DocumentType == "" {
		return nil, domain.ErrValidation
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: at least one prompt item is required", domain.ErrValidation)
	}

	// 2. Start Transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 3. Create Group
	pm := &domain.PromptManagement{
		Client:       req.Client,
		UseCase:      req.UseCase,
		DocumentType: req.DocumentType,
		Category:     req.Category,
		StageName:    req.StageName,
		CreatedByID:  userID,
	}

	// Note: We need a way to pass the transaction down to the repo,
	// or perform the SQL here. Given the current repo structure,
	// we'll assume the repos are not yet transaction-aware for external Tx.
	// For now, I will implement the SQL directly in the service's transaction context
	// to ensure atomicity for this bulk operation.

	groupQuery := `
		INSERT INTO prompt_management (client, use_case, document_type, category, stage_name, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(ctx, groupQuery, pm.Client, pm.UseCase, pm.DocumentType, pm.Category, pm.StageName, userID).
		Scan(&pm.ID, &pm.CreatedAt, &pm.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert group: %w", err)
	}

	// 4. Create Items
	for _, itemReq := range req.Items {
		if itemReq.QuestionKey == "" || itemReq.PromptText == "" {
			return nil, fmt.Errorf("%w: item missing required fields", domain.ErrValidation)
		}

		itemQuery := `
			INSERT INTO prompt_item (
				management_id, question_key, prompt_text, vector_prompt, 
				generation_config, response_schema, top_k, ranking_method, 
				version_no, status, change_log, created_by
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id`

		var itemID string
		err = tx.QueryRow(ctx, itemQuery,
			pm.ID,
			itemReq.QuestionKey,
			itemReq.PromptText,
			itemReq.VectorPrompt,
			itemReq.GenerationConfig,
			itemReq.ResponseSchema,
			itemReq.TopK,
			itemReq.RankingMethod,
			"v1.0.0",
			"active", // Auto-promote to active
			itemReq.ChangeLog,
			userID,
		).Scan(&itemID)

		if err != nil {
			return nil, fmt.Errorf("failed to insert item %s: %w", itemReq.QuestionKey, err)
		}

		// Update the active pointer in the group to the LAST created item
		// (Common pattern for initial creation)
		_, err = tx.Exec(ctx, "UPDATE prompt_management SET active_item_id = $1 WHERE id = $2", itemID, pm.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update active pointer: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Set populated prompts for the response
	return s.GetByID(ctx, pm.ID)
}
