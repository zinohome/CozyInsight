package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"cozy-insight/internal/model"
)

type ShareLinkRepository struct {
	db *sqlx.DB
}

func NewShareLinkRepository(db *sqlx.DB) *ShareLinkRepository {
	return &ShareLinkRepository{db: db}
}

func (r *ShareLinkRepository) Create(ctx context.Context, link *model.ShareLink) error {
	query := `INSERT INTO share_links (token, resource_type, resource_id, created_by, expire_at, status)
			  VALUES (:token, :resource_type, :resource_id, :created_by, :expire_at, :status)`
	result, err := r.db.NamedExecContext(ctx, query, link)
	if err != nil {
		return fmt.Errorf("create share link failed: %w", err)
	}
	id, _ := result.LastInsertId()
	link.ID = uint64(id)
	return nil
}

func (r *ShareLinkRepository) FindByToken(ctx context.Context, token string) (*model.ShareLink, error) {
	var link model.ShareLink
	query := `SELECT * FROM share_links WHERE token = ? AND status = 1`
	if err := r.db.GetContext(ctx, &link, query, token); err != nil {
		return nil, fmt.Errorf("find share link failed: %w", err)
	}
	return &link, nil
}

func (r *ShareLinkRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE share_links SET status = 0 WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("delete share link failed: %w", err)
	}
	return nil
}

func (r *ShareLinkRepository) ListByResource(ctx context.Context, resourceType string, resourceID uint64) ([]model.ShareLink, error) {
	var links []model.ShareLink
	query := `SELECT * FROM share_links WHERE resource_type = ? AND resource_id = ? AND status = 1`
	if err := r.db.SelectContext(ctx, &links, query, resourceType, resourceID); err != nil {
		return nil, fmt.Errorf("list share links failed: %w", err)
	}
	return links, nil
}
