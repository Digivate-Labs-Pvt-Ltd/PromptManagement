package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"prompt-management/internal/domain"
)

type itemRepo struct {
	db *pgxpool.Pool
}

// NewItemRepository creates a new PostgreSQL prompt item repository.
func NewItemRepository(db *pgxpool.Pool) domain.ItemStore {
	return &itemRepo{db: db}
}

func (r *itemRepo) Insert(ctx context.Context, item *domain.PromptItem) error {
	query := `
		WITH inserted AS (
			INSERT INTO prompt_item (
				management_id, question_key, prompt_text, vector_prompt, 
				generation_config, response_schema, top_k, ranking_method, 
				version_no, status, change_log, created_by
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id, created_by, created_at, updated_at
		)
		SELECT i.id, i.created_by, u.username, i.created_at, i.updated_at
		FROM inserted i
		JOIN users u ON i.created_by = u.id`

	err := r.db.QueryRow(ctx, query,
		item.ManagementID,
		item.QuestionKey,
		item.PromptText,
		item.VectorPrompt,
		item.GenerationConfig,
		item.ResponseSchema,
		item.TopK,
		item.RankingMethod,
		item.VersionNo,
		item.Status,
		item.ChangeLog,
		item.CreatedByID,
	).Scan(&item.ID, &item.CreatedByID, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("could not insert prompt item: %w", err)
	}

	return nil
}

func (r *itemRepo) GetByID(ctx context.Context, id string) (*domain.PromptItem, error) {
	query := `
		SELECT 
			i.id, i.management_id, i.question_key, i.prompt_text, i.vector_prompt,
			i.generation_config, i.response_schema, i.top_k, i.ranking_method,
			i.version_no, i.status, i.change_log, i.created_by, u.username,
			i.created_at, i.updated_at
		FROM prompt_item i
		JOIN users u ON i.created_by = u.id
		WHERE i.id = $1 AND i.deleted_at IS NULL`

	var item domain.PromptItem
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.ManagementID,
		&item.QuestionKey,
		&item.PromptText,
		&item.VectorPrompt,
		&item.GenerationConfig,
		&item.ResponseSchema,
		&item.TopK,
		&item.RankingMethod,
		&item.VersionNo,
		&item.Status,
		&item.ChangeLog,
		&item.CreatedByID,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("could not get prompt item: %w", err)
	}

	return &item, nil
}

func (r *itemRepo) GetLatestVersionByGroupAndKey(ctx context.Context, managementID string, questionKey string) (*string, error) {
	query := `
		SELECT version_no
		FROM prompt_item
		WHERE management_id = $1 AND question_key = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`

	var version string
	err := r.db.QueryRow(ctx, query, managementID, questionKey).Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found is acceptable here, it implies it's the very first version
		}
		return nil, fmt.Errorf("could not get latest version: %w", err)
	}

	return &version, nil
}

func (r *itemRepo) List(ctx context.Context, f domain.ItemListFilters) ([]*domain.PromptItem, int, error) {
	baseQuery := `
		FROM prompt_item i
		JOIN users u ON i.created_by = u.id
		WHERE i.management_id = $1 AND i.deleted_at IS NULL`

	args := []interface{}{f.ManagementID}
	argIdx := 2

	if f.Status != nil && *f.Status != "" {
		baseQuery += fmt.Sprintf(" AND i.status = $%d", argIdx)
		args = append(args, *f.Status)
		argIdx++
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("could not count prompt items: %w", err)
	}

	// Fetch data
	if f.PerPage <= 0 {
		f.PerPage = 20
	}
	if f.Page <= 0 {
		f.Page = 1
	}

	dataQuery := `
		SELECT 
			i.id, i.management_id, i.question_key, i.prompt_text, i.vector_prompt,
			i.generation_config, i.response_schema, i.top_k, i.ranking_method,
			i.version_no, i.status, i.change_log, i.created_by, u.username,
			i.created_at, i.updated_at ` +
		baseQuery +
		fmt.Sprintf(" ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)

	args = append(args, f.PerPage, (f.Page-1)*f.PerPage)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("could not query prompt items: %w", err)
	}
	defer rows.Close()

	var items []*domain.PromptItem
	for rows.Next() {
		var item domain.PromptItem
		err := rows.Scan(
			&item.ID,
			&item.ManagementID,
			&item.QuestionKey,
			&item.PromptText,
			&item.VectorPrompt,
			&item.GenerationConfig,
			&item.ResponseSchema,
			&item.TopK,
			&item.RankingMethod,
			&item.VersionNo,
			&item.Status,
			&item.ChangeLog,
			&item.CreatedByID,
			&item.CreatedBy,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("could not scan prompt item: %w", err)
		}
		items = append(items, &item)
	}

	return items, total, nil
}

func (r *itemRepo) Promote(ctx context.Context, managementID string, itemID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Archive current active item for the management group
	archiveOldQuery := `
		UPDATE prompt_item 
		SET status = 'archived', updated_at = NOW() 
		WHERE id = (SELECT active_item_id FROM prompt_management WHERE id = $1)
		AND id != $2 AND status = 'active'`
	
	_, err = tx.Exec(ctx, archiveOldQuery, managementID, itemID)
	if err != nil {
		return fmt.Errorf("could not archive previous active item: %w", err)
	}

	// 2. Set new active item
	activateNewQuery := `
		UPDATE prompt_item 
		SET status = 'active', updated_at = NOW() 
		WHERE id = $1 AND management_id = $2`

	res, err := tx.Exec(ctx, activateNewQuery, itemID, managementID)
	if err != nil {
		return fmt.Errorf("could not activate new item: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("item not found or invalid management id: %w", domain.ErrNotFound)
	}

	// 3. Update pointer in management table
	updatePointerQuery := `
		UPDATE prompt_management 
		SET active_item_id = $1, updated_at = NOW() 
		WHERE id = $2`

	_, err = tx.Exec(ctx, updatePointerQuery, itemID, managementID)
	if err != nil {
		return fmt.Errorf("could not update management active pointer: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *itemRepo) Archive(ctx context.Context, id string) error {
	query := `
		UPDATE prompt_item
		SET status = 'archived', updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	res, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("could not archive prompt item: %w", err)
	}

	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
