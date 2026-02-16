package mention

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

func (r *Repository) Create(ctx context.Context, m *model.Mention) error {
	query := `
		INSERT INTO mentions (id, message_id, channel_id, mentioned_user_id, mentioned_group_id, mention_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		m.ID, m.MessageID, m.ChannelID, m.MentionedUserID, m.MentionedGroupID, m.MentionType, m.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create mention: %w", err)
	}
	return nil
}

func (r *Repository) GetUnreadByUser(ctx context.Context, userID uuid.UUID) ([]model.Mention, error) {
	query := `
		SELECT mn.id, mn.message_id, mn.channel_id, mn.mentioned_user_id, mn.mentioned_group_id,
			   mn.mention_type, mn.is_read, mn.created_at,
			   m.id, m.channel_id, m.user_id, m.content, m.created_at,
			   u.id, u.name, u.avatar_url
		FROM mentions mn
		JOIN messages m ON m.id = mn.message_id
		JOIN users u ON u.id = m.user_id
		WHERE mn.mentioned_user_id = $1 AND mn.is_read = false
		ORDER BY mn.created_at DESC
		LIMIT 50
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get unread mentions: %w", err)
	}
	defer rows.Close()

	var mentions []model.Mention
	for rows.Next() {
		var mn model.Mention
		var msg model.Message
		var u model.User
		if err := rows.Scan(
			&mn.ID, &mn.MessageID, &mn.ChannelID, &mn.MentionedUserID, &mn.MentionedGroupID,
			&mn.MentionType, &mn.IsRead, &mn.CreatedAt,
			&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content, &msg.CreatedAt,
			&u.ID, &u.Name, &u.AvatarURL,
		); err != nil {
			return nil, fmt.Errorf("scan mention: %w", err)
		}
		msg.User = &u
		mn.Message = &msg
		mentions = append(mentions, mn)
	}
	return mentions, rows.Err()
}

func (r *Repository) MarkReadByChannel(ctx context.Context, userID, channelID uuid.UUID) error {
	query := `UPDATE mentions SET is_read = true WHERE mentioned_user_id = $1 AND channel_id = $2 AND is_read = false`
	_, err := r.db.Exec(ctx, query, userID, channelID)
	if err != nil {
		return fmt.Errorf("mark mentions read: %w", err)
	}
	return nil
}

// GetUserByName finds a user by their name (case-insensitive).
func (r *Repository) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	var u model.User
	err := r.db.QueryRow(ctx,
		`SELECT id, name FROM users WHERE LOWER(name) = LOWER($1) LIMIT 1`, name,
	).Scan(&u.ID, &u.Name)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by name: %w", err)
	}
	return &u, nil
}

// GetGroupByName finds a user group by name (case-insensitive).
func (r *Repository) GetGroupByName(ctx context.Context, name string) (*model.UserGroup, error) {
	var g model.UserGroup
	err := r.db.QueryRow(ctx,
		`SELECT id, name FROM user_groups WHERE LOWER(name) = LOWER($1) LIMIT 1`, name,
	).Scan(&g.ID, &g.Name)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get group by name: %w", err)
	}
	return &g, nil
}

// GetGroupMemberIDs returns user IDs for a group.
func (r *Repository) GetGroupMemberIDs(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `SELECT user_id FROM user_group_members WHERE group_id = $1`, groupID)
	if err != nil {
		return nil, fmt.Errorf("get group member ids: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetChannelMemberIDs returns all member user IDs for a channel.
func (r *Repository) GetChannelMemberIDs(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `SELECT user_id FROM channel_members WHERE channel_id = $1`, channelID)
	if err != nil {
		return nil, fmt.Errorf("get channel member ids: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
