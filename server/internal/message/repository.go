package message

import (
	"context"
	"encoding/json"
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

func (r *Repository) Create(ctx context.Context, msg *model.Message) error {
	query := `
		INSERT INTO messages (id, channel_id, user_id, parent_id, content, is_alert, alert_severity, alert_metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		msg.ID, msg.ChannelID, msg.UserID, msg.ParentID,
		msg.Content, msg.IsAlert, msg.AlertSeverity, msg.AlertMetadata, msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create message: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Message, error) {
	query := `
		SELECT m.id, m.channel_id, m.user_id, m.parent_id, m.content,
			   m.is_alert, m.alert_severity, m.alert_metadata,
			   m.edited_at, m.deleted_at, m.created_at,
			   u.id, u.email, u.name, u.avatar_url, u.role, u.is_active, u.created_at, u.updated_at,
			   (SELECT COUNT(*) FROM messages r WHERE r.parent_id = m.id AND r.deleted_at IS NULL) as reply_count
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.id = $1 AND m.deleted_at IS NULL
	`
	msg, err := r.scanMessage(r.db.QueryRow(ctx, query, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get message: %w", err)
	}
	return msg, nil
}

func (r *Repository) List(ctx context.Context, params model.MessageListParams) ([]model.Message, error) {
	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 50
	}

	var query string
	var args []interface{}

	if params.Before != nil {
		query = `
			SELECT m.id, m.channel_id, m.user_id, m.parent_id, m.content,
				   m.is_alert, m.alert_severity, m.alert_metadata,
				   m.edited_at, m.deleted_at, m.created_at,
				   u.id, u.email, u.name, u.avatar_url, u.role, u.is_active, u.created_at, u.updated_at,
				   (SELECT COUNT(*) FROM messages r WHERE r.parent_id = m.id AND r.deleted_at IS NULL) as reply_count
			FROM messages m
			JOIN users u ON u.id = m.user_id
			WHERE m.channel_id = $1 AND m.deleted_at IS NULL AND m.parent_id IS NULL
			  AND m.created_at < (SELECT created_at FROM messages WHERE id = $2)
			ORDER BY m.created_at DESC
			LIMIT $3
		`
		args = []interface{}{params.ChannelID, *params.Before, params.Limit}
	} else {
		query = `
			SELECT m.id, m.channel_id, m.user_id, m.parent_id, m.content,
				   m.is_alert, m.alert_severity, m.alert_metadata,
				   m.edited_at, m.deleted_at, m.created_at,
				   u.id, u.email, u.name, u.avatar_url, u.role, u.is_active, u.created_at, u.updated_at,
				   (SELECT COUNT(*) FROM messages r WHERE r.parent_id = m.id AND r.deleted_at IS NULL) as reply_count
			FROM messages m
			JOIN users u ON u.id = m.user_id
			WHERE m.channel_id = $1 AND m.deleted_at IS NULL AND m.parent_id IS NULL
			ORDER BY m.created_at DESC
			LIMIT $2
		`
		args = []interface{}{params.ChannelID, params.Limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		msg, err := r.scanMessageRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, *msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	return messages, nil
}

func (r *Repository) GetThread(ctx context.Context, parentID uuid.UUID) ([]model.Message, error) {
	query := `
		SELECT m.id, m.channel_id, m.user_id, m.parent_id, m.content,
			   m.is_alert, m.alert_severity, m.alert_metadata,
			   m.edited_at, m.deleted_at, m.created_at,
			   u.id, u.email, u.name, u.avatar_url, u.role, u.is_active, u.created_at, u.updated_at,
			   0 as reply_count
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.parent_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("get thread: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		msg, err := r.scanMessageRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan thread message: %w", err)
		}
		messages = append(messages, *msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate thread messages: %w", err)
	}

	return messages, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, content string) error {
	query := `UPDATE messages SET content = $1, edited_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, content, id)
	if err != nil {
		return fmt.Errorf("update message: %w", err)
	}
	return nil
}

func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE messages SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft delete message: %w", err)
	}
	return nil
}

func (r *Repository) GetReactions(ctx context.Context, messageID uuid.UUID) ([]model.ReactionGroup, error) {
	query := `
		SELECT emoji, COUNT(*) as count, array_agg(user_id) as users
		FROM reactions
		WHERE message_id = $1
		GROUP BY emoji
		ORDER BY MIN(created_at)
	`
	rows, err := r.db.Query(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("get reactions: %w", err)
	}
	defer rows.Close()

	var groups []model.ReactionGroup
	for rows.Next() {
		var g model.ReactionGroup
		if err := rows.Scan(&g.Emoji, &g.Count, &g.Users); err != nil {
			return nil, fmt.Errorf("scan reaction group: %w", err)
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reaction groups: %w", err)
	}
	return groups, nil
}

func (r *Repository) GetReactionsForMessages(ctx context.Context, messageIDs []uuid.UUID) (map[uuid.UUID][]model.ReactionGroup, error) {
	if len(messageIDs) == 0 {
		return make(map[uuid.UUID][]model.ReactionGroup), nil
	}

	query := `
		SELECT message_id, emoji, COUNT(*) as count, array_agg(user_id) as users
		FROM reactions
		WHERE message_id = ANY($1)
		GROUP BY message_id, emoji
		ORDER BY message_id, MIN(created_at)
	`
	rows, err := r.db.Query(ctx, query, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("get reactions for messages: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]model.ReactionGroup)
	for rows.Next() {
		var msgID uuid.UUID
		var g model.ReactionGroup
		if err := rows.Scan(&msgID, &g.Emoji, &g.Count, &g.Users); err != nil {
			return nil, fmt.Errorf("scan reaction: %w", err)
		}
		result[msgID] = append(result[msgID], g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reactions for messages: %w", err)
	}
	return result, nil
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func (r *Repository) scanMessage(row scannable) (*model.Message, error) {
	var msg model.Message
	var user model.User
	var alertMetadata []byte

	err := row.Scan(
		&msg.ID, &msg.ChannelID, &msg.UserID, &msg.ParentID, &msg.Content,
		&msg.IsAlert, &msg.AlertSeverity, &alertMetadata,
		&msg.EditedAt, &msg.DeletedAt, &msg.CreatedAt,
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&msg.ReplyCount,
	)
	if err != nil {
		return nil, err
	}

	if alertMetadata != nil {
		msg.AlertMetadata = json.RawMessage(alertMetadata)
	}
	msg.User = &user
	return &msg, nil
}

func (r *Repository) scanMessageRow(rows pgx.Rows) (*model.Message, error) {
	var msg model.Message
	var user model.User
	var alertMetadata []byte

	err := rows.Scan(
		&msg.ID, &msg.ChannelID, &msg.UserID, &msg.ParentID, &msg.Content,
		&msg.IsAlert, &msg.AlertSeverity, &alertMetadata,
		&msg.EditedAt, &msg.DeletedAt, &msg.CreatedAt,
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&msg.ReplyCount,
	)
	if err != nil {
		return nil, err
	}

	if alertMetadata != nil {
		msg.AlertMetadata = json.RawMessage(alertMetadata)
	}
	msg.User = &user
	return &msg, nil
}
