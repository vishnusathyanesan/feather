package invitation

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

func (r *Repository) Create(ctx context.Context, inv *model.WorkspaceInvitation) error {
	query := `
		INSERT INTO workspace_invitations (id, inviter_id, email, token, expires_at, max_uses, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		inv.ID, inv.InviterID, inv.Email, inv.Token, inv.ExpiresAt, inv.MaxUses, inv.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create invitation: %w", err)
	}
	return nil
}

func (r *Repository) GetByToken(ctx context.Context, token string) (*model.WorkspaceInvitation, error) {
	query := `
		SELECT id, inviter_id, email, token, expires_at, accepted_at, accepted_by, max_uses, use_count, created_at
		FROM workspace_invitations WHERE token = $1
	`
	var inv model.WorkspaceInvitation
	err := r.db.QueryRow(ctx, query, token).Scan(
		&inv.ID, &inv.InviterID, &inv.Email, &inv.Token, &inv.ExpiresAt,
		&inv.AcceptedAt, &inv.AcceptedBy, &inv.MaxUses, &inv.UseCount, &inv.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get invitation by token: %w", err)
	}
	return &inv, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.WorkspaceInvitation, error) {
	query := `
		SELECT id, inviter_id, email, token, expires_at, accepted_at, accepted_by, max_uses, use_count, created_at
		FROM workspace_invitations WHERE id = $1
	`
	var inv model.WorkspaceInvitation
	err := r.db.QueryRow(ctx, query, id).Scan(
		&inv.ID, &inv.InviterID, &inv.Email, &inv.Token, &inv.ExpiresAt,
		&inv.AcceptedAt, &inv.AcceptedBy, &inv.MaxUses, &inv.UseCount, &inv.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get invitation by id: %w", err)
	}
	return &inv, nil
}

func (r *Repository) IncrementUseCount(ctx context.Context, id uuid.UUID, acceptedBy uuid.UUID) error {
	query := `
		UPDATE workspace_invitations
		SET use_count = use_count + 1,
		    accepted_at = COALESCE(accepted_at, NOW()),
		    accepted_by = COALESCE(accepted_by, $2)
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, acceptedBy)
	if err != nil {
		return fmt.Errorf("increment use count: %w", err)
	}
	return nil
}

func (r *Repository) ListByInviter(ctx context.Context, inviterID uuid.UUID) ([]model.WorkspaceInvitation, error) {
	query := `
		SELECT id, inviter_id, email, token, expires_at, accepted_at, accepted_by, max_uses, use_count, created_at
		FROM workspace_invitations WHERE inviter_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, inviterID)
	if err != nil {
		return nil, fmt.Errorf("list invitations: %w", err)
	}
	defer rows.Close()

	var invitations []model.WorkspaceInvitation
	for rows.Next() {
		var inv model.WorkspaceInvitation
		if err := rows.Scan(
			&inv.ID, &inv.InviterID, &inv.Email, &inv.Token, &inv.ExpiresAt,
			&inv.AcceptedAt, &inv.AcceptedBy, &inv.MaxUses, &inv.UseCount, &inv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan invitation: %w", err)
		}
		invitations = append(invitations, inv)
	}
	return invitations, rows.Err()
}

func (r *Repository) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM workspace_invitations WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("revoke invitation: %w", err)
	}
	return nil
}
