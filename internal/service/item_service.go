package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"prompt-management/internal/domain"
)

// ItemService handles business logic for prompt items.
type ItemService struct {
	repo domain.ItemStore
}

// NewItemService creates a new ItemService.
func NewItemService(repo domain.ItemStore) *ItemService {
	return &ItemService{repo: repo}
}

// AddItemRequest defines the payload for creating a new prompt version.
type AddItemRequest struct {
	ManagementID     string                 `json:"management_id"`
	QuestionKey      string                 `json:"question_key"`
	PromptText       string                 `json:"prompt_text"`
	VectorPrompt     *string                `json:"vector_prompt"`
	GenerationConfig map[string]interface{} `json:"generation_config"`
	ResponseSchema   map[string]interface{} `json:"response_schema"`
	TopK             *float64               `json:"top_k"`
	RankingMethod    *string                `json:"ranking_method"`
	ChangeLog        *string                `json:"change_log"`
}

func (s *ItemService) Add(ctx context.Context, req AddItemRequest, userID string) (*domain.PromptItem, error) {
	// Basic validation
	if req.ManagementID == "" || req.QuestionKey == "" || req.PromptText == "" {
		return nil, domain.ErrValidation
	}

	// Figure out the version number
	latestVersion, err := s.repo.GetLatestVersionByGroupAndKey(ctx, req.ManagementID, req.QuestionKey)
	if err != nil {
		return nil, err
	}

	nextVersion := bumpVersion(latestVersion)

	item := &domain.PromptItem{
		ManagementID:     req.ManagementID,
		QuestionKey:      req.QuestionKey,
		PromptText:       req.PromptText,
		VectorPrompt:     req.VectorPrompt,
		GenerationConfig: req.GenerationConfig,
		ResponseSchema:   req.ResponseSchema,
		TopK:             req.TopK,
		RankingMethod:    req.RankingMethod,
		VersionNo:        nextVersion,
		Status:           "draft",
		ChangeLog:        req.ChangeLog,
		CreatedByID:      userID,
	}

	if err := s.repo.Insert(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *ItemService) List(ctx context.Context, filters domain.ItemListFilters) ([]*domain.PromptItem, int, error) {
	if filters.ManagementID == "" {
		return nil, 0, domain.ErrValidation
	}
	return s.repo.List(ctx, filters)
}

func (s *ItemService) GetByID(ctx context.Context, id string) (*domain.PromptItem, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ItemService) Promote(ctx context.Context, managementID string, itemID string) error {
	if managementID == "" || itemID == "" {
		return domain.ErrValidation
	}
	return s.repo.Promote(ctx, managementID, itemID)
}

func (s *ItemService) Archive(ctx context.Context, id string) error {
	return s.repo.Archive(ctx, id)
}

// bumpVersion automatically increments a semantic version string.
// If nil is passed, implies starting version: "v1.0.0".
func bumpVersion(v *string) string {
	if v == nil || *v == "" {
		return "v1.0.0"
	}

	// Very simple semver bump logic on the patch version
	// "v1.0.5" -> "v1.0.6"
	str := *v
	if strings.HasPrefix(str, "v") {
		str = str[1:]
	}
	parts := strings.Split(str, ".")
	if len(parts) != 3 {
		return *v + "-1" // Fallback if it's not strictly "X.Y.Z"
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return *v + "-1"
	}

	return fmt.Sprintf("v%s.%s.%d", parts[0], parts[1], patch+1)
}
