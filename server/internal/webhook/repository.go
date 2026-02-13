package webhook

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feather-chat/feather/internal/model"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, wh *model.Webhook) error {
	query := `
		INSERT INTO webhooks (id, channel_id, name, token, creator_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, wh.ID, wh.ChannelID, wh.Name, wh.Token, wh.CreatorID, wh.IsActive)
	if err != nil {
		return fmt.Errorf("create webhook: %w", err)
	}
	return nil
}

func (r *Repository) GetByToken(ctx context.Context, token string) (*model.Webhook, error) {
	query := `SELECT id, channel_id, name, token, creator_id, is_active, created_at FROM webhooks WHERE token = $1 AND is_active = true`
	var wh model.Webhook
	err := r.db.QueryRow(ctx, query, token).Scan(&wh.ID, &wh.ChannelID, &wh.Name, &wh.Token, &wh.CreatorID, &wh.IsActive, &wh.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get webhook by token: %w", err)
	}
	return &wh, nil
}

func (r *Repository) List(ctx context.Context, creatorID uuid.UUID) ([]model.Webhook, error) {
	query := `SELECT id, channel_id, name, '', creator_id, is_active, created_at FROM webhooks WHERE creator_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, creatorID)
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []model.Webhook
	for rows.Next() {
		var wh model.Webhook
		if err := rows.Scan(&wh.ID, &wh.ChannelID, &wh.Name, &wh.Token, &wh.CreatorID, &wh.IsActive, &wh.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan webhook: %w", err)
		}
		webhooks = append(webhooks, wh)
	}
	return webhooks, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM webhooks WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete webhook: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Webhook, error) {
	query := `SELECT id, channel_id, name, token, creator_id, is_active, created_at FROM webhooks WHERE id = $1`
	var wh model.Webhook
	err := r.db.QueryRow(ctx, query, id).Scan(&wh.ID, &wh.ChannelID, &wh.Name, &wh.Token, &wh.CreatorID, &wh.IsActive, &wh.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get webhook: %w", err)
	}
	return &wh, nil
}
