package dm

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

// FindExistingDM finds an existing 1:1 DM channel between two users.
func (r *Repository) FindExistingDM(ctx context.Context, userID1, userID2 uuid.UUID) (*model.Channel, error) {
	query := `
		SELECT c.id, c.name, c.topic, c.description, c.type, c.is_readonly, c.creator_id, c.created_at, c.updated_at
		FROM channels c
		WHERE c.type = 'dm'
		  AND (SELECT COUNT(*) FROM channel_members cm WHERE cm.channel_id = c.id) = 2
		  AND EXISTS (SELECT 1 FROM channel_members cm WHERE cm.channel_id = c.id AND cm.user_id = $1)
		  AND EXISTS (SELECT 1 FROM channel_members cm WHERE cm.channel_id = c.id AND cm.user_id = $2)
		LIMIT 1
	`
	var ch model.Channel
	err := r.db.QueryRow(ctx, query, userID1, userID2).Scan(
		&ch.ID, &ch.Name, &ch.Topic, &ch.Description, &ch.Type, &ch.IsReadonly,
		&ch.CreatorID, &ch.CreatedAt, &ch.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find existing dm: %w", err)
	}
	return &ch, nil
}

// FindExistingGroupDM finds an existing group DM with exactly the given set of members.
func (r *Repository) FindExistingGroupDM(ctx context.Context, memberIDs []uuid.UUID) (*model.Channel, error) {
	if len(memberIDs) < 2 {
		return nil, nil
	}

	// Find group_dm channels where the member count matches and all specified users are members.
	query := `
		SELECT c.id, c.name, c.topic, c.description, c.type, c.is_readonly, c.creator_id, c.created_at, c.updated_at
		FROM channels c
		WHERE c.type = 'group_dm'
		  AND (SELECT COUNT(*) FROM channel_members cm WHERE cm.channel_id = c.id) = $1
	`
	// Add an EXISTS clause for each member
	args := []interface{}{len(memberIDs)}
	for i, id := range memberIDs {
		query += fmt.Sprintf(
			" AND EXISTS (SELECT 1 FROM channel_members cm WHERE cm.channel_id = c.id AND cm.user_id = $%d)",
			i+2,
		)
		args = append(args, id)
	}
	query += " LIMIT 1"

	var ch model.Channel
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&ch.ID, &ch.Name, &ch.Topic, &ch.Description, &ch.Type, &ch.IsReadonly,
		&ch.CreatorID, &ch.CreatedAt, &ch.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find existing group dm: %w", err)
	}
	return &ch, nil
}

// CreateDMChannel creates a DM or group_dm channel and adds all members.
func (r *Repository) CreateDMChannel(ctx context.Context, ch *model.Channel, memberIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO channels (id, name, topic, description, type, is_readonly, creator_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		ch.ID, ch.Name, ch.Topic, ch.Description, ch.Type, ch.IsReadonly,
		ch.CreatorID, ch.CreatedAt, ch.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create dm channel: %w", err)
	}

	for _, memberID := range memberIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO channel_members (channel_id, user_id, role) VALUES ($1, $2, 'member')
			 ON CONFLICT (channel_id, user_id) DO NOTHING`,
			ch.ID, memberID,
		)
		if err != nil {
			return fmt.Errorf("add dm member: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// ListDMs returns all DM/group_dm channels for a user, with member info.
func (r *Repository) ListDMs(ctx context.Context, userID uuid.UUID) ([]model.Channel, error) {
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
		JOIN channel_members cm ON cm.channel_id = c.id AND cm.user_id = $1
		WHERE c.type IN ('dm', 'group_dm')
		ORDER BY c.updated_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list dms: %w", err)
	}
	defer rows.Close()

	var channels []model.Channel
	for rows.Next() {
		var ch model.Channel
		if err := rows.Scan(
			&ch.ID, &ch.Name, &ch.Topic, &ch.Description, &ch.Type, &ch.IsReadonly,
			&ch.CreatorID, &ch.CreatedAt, &ch.UpdatedAt, &ch.MemberCount, &ch.UnreadCount,
		); err != nil {
			return nil, fmt.Errorf("scan dm: %w", err)
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

// GetDMMembers returns members of a DM channel with user info.
func (r *Repository) GetDMMembers(ctx context.Context, channelID uuid.UUID) ([]model.User, error) {
	query := `
		SELECT u.id, u.email, u.name, u.avatar_url, u.role, u.is_active, u.created_at, u.updated_at
		FROM users u
		JOIN channel_members cm ON cm.user_id = u.id
		WHERE cm.channel_id = $1
		ORDER BY u.name ASC
	`
	rows, err := r.db.Query(ctx, query, channelID)
	if err != nil {
		return nil, fmt.Errorf("get dm members: %w", err)
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
