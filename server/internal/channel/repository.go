package channel

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

func (r *Repository) Create(ctx context.Context, ch *model.Channel) error {
	query := `
		INSERT INTO channels (id, name, topic, description, type, is_readonly, creator_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		ch.ID, ch.Name, ch.Topic, ch.Description, ch.Type, ch.IsReadonly,
		ch.CreatorID, ch.CreatedAt, ch.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create channel: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Channel, error) {
	query := `
		SELECT c.id, c.name, c.topic, c.description, c.type, c.is_readonly, c.creator_id, c.created_at, c.updated_at,
			   (SELECT COUNT(*) FROM channel_members cm WHERE cm.channel_id = c.id) as member_count
		FROM channels c WHERE c.id = $1
	`
	var ch model.Channel
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ch.ID, &ch.Name, &ch.Topic, &ch.Description, &ch.Type, &ch.IsReadonly,
		&ch.CreatorID, &ch.CreatedAt, &ch.UpdatedAt, &ch.MemberCount,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}
	return &ch, nil
}

func (r *Repository) GetByName(ctx context.Context, name string) (*model.Channel, error) {
	query := `
		SELECT c.id, c.name, c.topic, c.description, c.type, c.is_readonly, c.creator_id, c.created_at, c.updated_at,
			   (SELECT COUNT(*) FROM channel_members cm WHERE cm.channel_id = c.id) as member_count
		FROM channels c WHERE c.name = $1
	`
	var ch model.Channel
	err := r.db.QueryRow(ctx, query, name).Scan(
		&ch.ID, &ch.Name, &ch.Topic, &ch.Description, &ch.Type, &ch.IsReadonly,
		&ch.CreatorID, &ch.CreatedAt, &ch.UpdatedAt, &ch.MemberCount,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get channel by name: %w", err)
	}
	return &ch, nil
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID) ([]model.Channel, error) {
	query := `
		SELECT c.id, c.name, c.topic, c.description, c.type, c.is_readonly, c.creator_id, c.created_at, c.updated_at,
			   (SELECT COUNT(*) FROM channel_members cm2 WHERE cm2.channel_id = c.id) as member_count,
			   COALESCE(
				   (SELECT COUNT(*) FROM messages m
				    WHERE m.channel_id = c.id AND m.deleted_at IS NULL
				    AND m.created_at > COALESCE(
				        (SELECT cm3.last_read_at FROM channel_members cm3 WHERE cm3.channel_id = c.id AND cm3.user_id = $1),
				        '1970-01-01'::timestamptz
				    )
				   ), 0
			   ) as unread_count
		FROM channels c
		WHERE c.type = 'public'
		   OR EXISTS (SELECT 1 FROM channel_members cm WHERE cm.channel_id = c.id AND cm.user_id = $1)
		ORDER BY c.name ASC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	var channels []model.Channel
	for rows.Next() {
		var ch model.Channel
		if err := rows.Scan(
			&ch.ID, &ch.Name, &ch.Topic, &ch.Description, &ch.Type, &ch.IsReadonly,
			&ch.CreatorID, &ch.CreatedAt, &ch.UpdatedAt, &ch.MemberCount, &ch.UnreadCount,
		); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func (r *Repository) Update(ctx context.Context, ch *model.Channel) error {
	query := `
		UPDATE channels SET name = $1, topic = $2, description = $3, is_readonly = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, ch.Name, ch.Topic, ch.Description, ch.IsReadonly, ch.ID)
	if err != nil {
		return fmt.Errorf("update channel: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM channels WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	return nil
}

func (r *Repository) AddMember(ctx context.Context, channelID, userID uuid.UUID, role string) error {
	query := `
		INSERT INTO channel_members (channel_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (channel_id, user_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, channelID, userID, role)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (r *Repository) RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM channel_members WHERE channel_id = $1 AND user_id = $2", channelID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	return nil
}

func (r *Repository) IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM channel_members WHERE channel_id = $1 AND user_id = $2)",
		channelID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check membership: %w", err)
	}
	return exists, nil
}

func (r *Repository) GetMembers(ctx context.Context, channelID uuid.UUID) ([]model.ChannelMember, error) {
	query := `
		SELECT cm.channel_id, cm.user_id, cm.role, cm.last_read_at, cm.joined_at,
			   u.id, u.email, u.name, u.avatar_url, u.role, u.is_active, u.created_at, u.updated_at
		FROM channel_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.channel_id = $1
		ORDER BY cm.joined_at ASC
	`
	rows, err := r.db.Query(ctx, query, channelID)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	defer rows.Close()

	var members []model.ChannelMember
	for rows.Next() {
		var m model.ChannelMember
		var u model.User
		if err := rows.Scan(
			&m.ChannelID, &m.UserID, &m.Role, &m.LastReadAt, &m.JoinedAt,
			&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		m.User = &u
		members = append(members, m)
	}
	return members, nil
}

func (r *Repository) UpdateLastRead(ctx context.Context, channelID, userID uuid.UUID) error {
	query := `UPDATE channel_members SET last_read_at = NOW() WHERE channel_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, channelID, userID)
	if err != nil {
		return fmt.Errorf("update last read: %w", err)
	}
	return nil
}

func (r *Repository) GetMemberRole(ctx context.Context, channelID, userID uuid.UUID) (string, error) {
	var role string
	err := r.db.QueryRow(ctx,
		"SELECT role FROM channel_members WHERE channel_id = $1 AND user_id = $2",
		channelID, userID,
	).Scan(&role)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get member role: %w", err)
	}
	return role, nil
}
