package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feather-chat/feather/internal/model"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

type SearchParams struct {
	Query     string
	UserID    uuid.UUID // Requesting user (for permission scoping)
	ChannelID *uuid.UUID
	AuthorID  *uuid.UUID
	HasLink   bool
	HasCode   bool
	Limit     int
	Offset    int
}

type SearchResult struct {
	Messages   []model.Message `json:"messages"`
	TotalCount int             `json:"total_count"`
}

func (r *Repository) Search(ctx context.Context, params SearchParams) (*SearchResult, error) {
	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	// Base condition: not deleted and user has access
	conditions = append(conditions, "m.deleted_at IS NULL")

	// Permission scoping: only search channels the user is a member of
	conditions = append(conditions, fmt.Sprintf(`
		(m.channel_id IN (SELECT cm.channel_id FROM channel_members cm WHERE cm.user_id = $%d)
		 OR m.channel_id IN (SELECT c.id FROM channels c WHERE c.type = 'public'))
	`, argIdx))
	args = append(args, params.UserID)
	argIdx++

	// Full-text search
	if params.Query != "" {
		conditions = append(conditions, fmt.Sprintf("m.search_vector @@ plainto_tsquery('english', $%d)", argIdx))
		args = append(args, params.Query)
		argIdx++
	}

	// Channel filter
	if params.ChannelID != nil {
		conditions = append(conditions, fmt.Sprintf("m.channel_id = $%d", argIdx))
		args = append(args, *params.ChannelID)
		argIdx++
	}

	// Author filter
	if params.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("m.user_id = $%d", argIdx))
		args = append(args, *params.AuthorID)
		argIdx++
	}

	// has:link filter
	if params.HasLink {
		conditions = append(conditions, "m.content LIKE '%http%'")
	}

	// has:code filter
	if params.HasCode {
		conditions = append(conditions, "m.content LIKE '%```%'")
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM messages m WHERE %s", whereClause)
	var totalCount int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count search results: %w", err)
	}

	// Results query
	var orderClause string
	if params.Query != "" {
		orderClause = fmt.Sprintf("ORDER BY ts_rank(m.search_vector, plainto_tsquery('english', $%d)) DESC, m.created_at DESC", 2) // reuse query param
	} else {
		orderClause = "ORDER BY m.created_at DESC"
	}

	searchQuery := fmt.Sprintf(`
		SELECT m.id, m.channel_id, m.user_id, m.parent_id, m.content,
			   m.is_alert, m.alert_severity, m.alert_metadata,
			   m.edited_at, m.deleted_at, m.created_at,
			   u.id, u.email, u.name, u.avatar_url, u.role, u.is_active, u.created_at, u.updated_at,
			   0 as reply_count
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argIdx, argIdx+1)

	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.Query(ctx, searchQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		var user model.User
		var alertMetadata []byte
		if err := rows.Scan(
			&msg.ID, &msg.ChannelID, &msg.UserID, &msg.ParentID, &msg.Content,
			&msg.IsAlert, &msg.AlertSeverity, &alertMetadata,
			&msg.EditedAt, &msg.DeletedAt, &msg.CreatedAt,
			&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
			&msg.ReplyCount,
		); err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}
		if alertMetadata != nil {
			msg.AlertMetadata = alertMetadata
		}
		msg.User = &user
		messages = append(messages, msg)
	}

	return &SearchResult{
		Messages:   messages,
		TotalCount: totalCount,
	}, nil
}
