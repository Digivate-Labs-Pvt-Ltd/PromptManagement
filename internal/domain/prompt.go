package domain

import (
	"context"
	"time"
)

// PromptManagement represents a prompt metadata group.
type PromptManagement struct {
	ID           string     `json:"id"`
	Client       string     `json:"client"`
	UseCase      string     `json:"use_case"`
	DocumentType string     `json:"document_type"`
	Category     string     `json:"category"`
	StageName    string     `json:"stage_name"`
	ActiveItemID *string    `json:"active_item_id"`
	CreatedByID  string     `json:"created_by_id"`
	CreatedBy    string     `json:"created_by"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Prompts      []*PromptItem `json:"prompts,omitempty"`
	DeletedAt    *time.Time    `json:"-"`
}

// ManagementStore defines the interface for prompt management persistence.
type ManagementStore interface {
	Insert(ctx context.Context, pm *PromptManagement) error
	Update(ctx context.Context, pm *PromptManagement) error
	GetByID(ctx context.Context, id string) (*PromptManagement, error)
	GetByBusinessKey(ctx context.Context, client, useCase, docType string) (*PromptManagement, error)
	List(ctx context.Context, filters ListFilters) ([]*PromptManagement, int, error)
	Delete(ctx context.Context, id string) error
}

// ListFilters represents the filtering parameters for listing prompt groups.
type ListFilters struct {
	Client  string `json:"client"`
	UseCase string `json:"use_case"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
}

// PromptItem represents a versioned prompt item.
type PromptItem struct {
	ID               string                 `json:"id"`
	ManagementID     string                 `json:"management_id"`
	QuestionKey      string                 `json:"question_key"`
	PromptText       string                 `json:"prompt_text"`
	VectorPrompt     *string                `json:"vector_prompt"`
	GenerationConfig map[string]interface{} `json:"generation_config"`
	ResponseSchema   map[string]interface{} `json:"response_schema"`
	TopK             *float64               `json:"top_k"`
	RankingMethod    *string                `json:"ranking_method"`
	VersionNo        string                 `json:"version_no"`
	Status           string                 `json:"status"`
	IsActive         bool                   `json:"is_active"`
	ChangeLog        *string                `json:"change_log"`
	CreatedByID      string                 `json:"created_by_id"`
	CreatedBy        string                 `json:"created_by"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	DeletedAt        *time.Time             `json:"-"`
}

// ItemStore defines the interface for prompt item persistence.
type ItemStore interface {
	Insert(ctx context.Context, item *PromptItem) error
	GetByID(ctx context.Context, id string) (*PromptItem, error)
	GetLatestVersionByGroupAndKey(ctx context.Context, managementID string, questionKey string) (*string, error)
	List(ctx context.Context, filters ItemListFilters) ([]*PromptItem, int, error)
	Promote(ctx context.Context, managementID string, itemID string) error
	GetActiveItemsByManagementIDs(ctx context.Context, ids []string) ([]*PromptItem, error)
	Archive(ctx context.Context, id string) error
}

// ItemListFilters represents the filtering parameters for listing prompt items.
type ItemListFilters struct {
	ManagementID string `json:"management_id"`
	Status       *string `json:"status"`
	Page         int    `json:"page"`
	PerPage      int    `json:"per_page"`
}
