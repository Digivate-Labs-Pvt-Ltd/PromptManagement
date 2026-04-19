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
	CreatedBy    string     `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

// ManagementStore defines the interface for prompt management persistence.
type ManagementStore interface {
	Insert(ctx context.Context, pm *PromptManagement) error
	Update(ctx context.Context, pm *PromptManagement) error
	GetByID(ctx context.Context, id string) (*PromptManagement, error)
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
