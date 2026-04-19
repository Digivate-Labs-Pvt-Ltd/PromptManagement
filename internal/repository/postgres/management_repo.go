package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"prompt-management/internal/domain"
)

type managementRepo struct {
	db *pgxpool.Pool
}

// NewManagementRepository creates a new PostgreSQL prompt management repository.
func NewManagementRepository(db *pgxpool.Pool) domain.ManagementStore {
	return &managementRepo{db: db}
}

func (r *managementRepo) Insert(ctx context.Context, pm *domain.PromptManagement) error {
	query := `
		INSERT INTO prompt_management (client, use_case, document_type, category, stage_name, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		pm.Client,
		pm.UseCase,
		pm.DocumentType,
		pm.Category,
		pm.StageName,
		pm.CreatedBy,
	).Scan(&pm.ID, &pm.CreatedAt, &pm.UpdatedAt)

	if err != nil {
		return fmt.Errorf("could not insert prompt group: %w", err)
	}

	return nil
}

func (r *managementRepo) Update(ctx context.Context, pm *domain.PromptManagement) error {
	query := `
		UPDATE prompt_management
		SET client = $1, use_case = $2, document_type = $3, category = $4, stage_name = $5, updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, query,
		pm.Client,
		pm.UseCase,
		pm.DocumentType,
		pm.Category,
		pm.StageName,
		pm.ID,
	).Scan(&pm.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("could not update prompt group: %w", err)
	}

	return nil
}

func (r *managementRepo) GetByID(ctx context.Context, id string) (*domain.PromptManagement, error) {
	query := `
		SELECT id, client, use_case, document_type, category, stage_name, active_item_id, created_by, created_at, updated_at
		FROM prompt_management
		WHERE id = $1 AND deleted_at IS NULL`

	var pm domain.PromptManagement
	err := r.db.QueryRow(ctx, query, id).Scan(
		&pm.ID,
		&pm.Client,
		&pm.UseCase,
		&pm.DocumentType,
		&pm.Category,
		&pm.StageName,
		&pm.ActiveItemID,
		&pm.CreatedBy,
		&pm.CreatedAt,
		&pm.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("could not get prompt group: %w", err)
	}

	return &pm, nil
}

func (r *managementRepo) List(ctx context.Context, f domain.ListFilters) ([]*domain.PromptManagement, int, error) {
	baseQuery := `
		FROM prompt_management
		WHERE deleted_at IS NULL`
	
	args := []interface{}{}
	argIdx := 1

	if f.Client != "" {
		baseQuery += fmt.Sprintf(" AND client = $%d", argIdx)
		args = append(args, f.Client)
		argIdx++
	}

	if f.UseCase != "" {
		baseQuery += fmt.Sprintf(" AND use_case = $%d", argIdx)
		args = append(args, f.UseCase)
		argIdx++
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("could not count prompt groups: %w", err)
	}

	// Fetch data
	if f.PerPage <= 0 {
		f.PerPage = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}

	dataQuery := `
		SELECT id, client, use_case, document_type, category, stage_name, active_item_id, created_by, created_at, updated_at ` +
		baseQuery +
		fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	
	args = append(args, f.PerPage, (f.Page-1)*f.PerPage)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("could not query prompt groups: %w", err)
	}
	defer rows.Close()

	var groups []*domain.PromptManagement
	for rows.Next() {
		var pm domain.PromptManagement
		err := rows.Scan(
			&pm.ID,
			&pm.Client,
			&pm.UseCase,
			&pm.DocumentType,
			&pm.Category,
			&pm.StageName,
			&pm.ActiveItemID,
			&pm.CreatedBy,
			&pm.CreatedAt,
			&pm.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("could not scan prompt group: %w", err)
		}
		groups = append(groups, &pm)
	}

	return groups, total, nil
}

func (r *managementRepo) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE prompt_management
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	res, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("could not delete prompt group: %w", err)
	}

	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
