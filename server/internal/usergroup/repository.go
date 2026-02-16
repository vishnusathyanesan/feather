package usergroup

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

func (r *Repository) Create(ctx context.Context, g *model.UserGroup) error {
	query := `
		INSERT INTO user_groups (id, name, description, creator_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, g.ID, g.Name, g.Description, g.CreatorID, g.CreatedAt, g.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create user group: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.UserGroup, error) {
	query := `SELECT id, name, description, creator_id, created_at, updated_at FROM user_groups WHERE id = $1`
	var g model.UserGroup
	err := r.db.QueryRow(ctx, query, id).Scan(&g.ID, &g.Name, &g.Description, &g.CreatorID, &g.CreatedAt, &g.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user group: %w", err)
	}
	return &g, nil
}

func (r *Repository) GetByName(ctx context.Context, name string) (*model.UserGroup, error) {
	query := `SELECT id, name, description, creator_id, created_at, updated_at FROM user_groups WHERE LOWER(name) = LOWER($1)`
	var g model.UserGroup
	err := r.db.QueryRow(ctx, query, name).Scan(&g.ID, &g.Name, &g.Description, &g.CreatorID, &g.CreatedAt, &g.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user group by name: %w", err)
	}
	return &g, nil
}

func (r *Repository) List(ctx context.Context, search string) ([]model.UserGroup, error) {
	query := `SELECT id, name, description, creator_id, created_at, updated_at FROM user_groups`
	var args []interface{}
	if search != "" {
		query += ` WHERE LOWER(name) LIKE LOWER($1)`
		args = append(args, "%"+search+"%")
	}
	query += ` ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list user groups: %w", err)
	}
	defer rows.Close()

	var groups []model.UserGroup
	for rows.Next() {
		var g model.UserGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatorID, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user group: %w", err)
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *Repository) Update(ctx context.Context, g *model.UserGroup) error {
	query := `UPDATE user_groups SET name = $1, description = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, g.Name, g.Description, g.ID)
	if err != nil {
		return fmt.Errorf("update user group: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM user_groups WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete user group: %w", err)
	}
	return nil
}

func (r *Repository) AddMember(ctx context.Context, groupID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO user_group_members (group_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		groupID, userID,
	)
	if err != nil {
		return fmt.Errorf("add group member: %w", err)
	}
	return nil
}

func (r *Repository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM user_group_members WHERE group_id = $1 AND user_id = $2", groupID, userID)
	if err != nil {
		return fmt.Errorf("remove group member: %w", err)
	}
	return nil
}

func (r *Repository) GetMembers(ctx context.Context, groupID uuid.UUID) ([]model.User, error) {
	query := `
		SELECT u.id, u.email, u.name, u.avatar_url, u.role, u.is_active, u.created_at, u.updated_at
		FROM users u
		JOIN user_group_members gm ON gm.user_id = u.id
		WHERE gm.group_id = $1
		ORDER BY u.name ASC
	`
	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("get group members: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
