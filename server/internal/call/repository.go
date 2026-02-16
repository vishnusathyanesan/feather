package call

import (
	"context"
	"fmt"
	"time"

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

func (r *Repository) Create(ctx context.Context, c *model.Call) error {
	query := `
		INSERT INTO calls (id, channel_id, initiator_id, call_type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, c.ID, c.ChannelID, c.InitiatorID, c.CallType, c.Status, c.CreatedAt)
	if err != nil {
		return fmt.Errorf("create call: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Call, error) {
	query := `
		SELECT id, channel_id, initiator_id, call_type, status, started_at, ended_at, created_at
		FROM calls WHERE id = $1
	`
	var c model.Call
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.ChannelID, &c.InitiatorID, &c.CallType, &c.Status,
		&c.StartedAt, &c.EndedAt, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get call: %w", err)
	}
	return &c, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.CallStatus) error {
	query := `UPDATE calls SET status = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("update call status: %w", err)
	}
	return nil
}

func (r *Repository) SetStarted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE calls SET status = 'in_progress', started_at = $2 WHERE id = $1`,
		id, now,
	)
	if err != nil {
		return fmt.Errorf("set call started: %w", err)
	}
	return nil
}

func (r *Repository) SetEnded(ctx context.Context, id uuid.UUID, status model.CallStatus) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE calls SET status = $2, ended_at = $3 WHERE id = $1`,
		id, status, now,
	)
	if err != nil {
		return fmt.Errorf("set call ended: %w", err)
	}
	return nil
}

func (r *Repository) AddParticipant(ctx context.Context, callID, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		`INSERT INTO call_participants (call_id, user_id, joined_at) VALUES ($1, $2, $3)
		 ON CONFLICT (call_id, user_id) DO UPDATE SET joined_at = $3, left_at = NULL`,
		callID, userID, now,
	)
	if err != nil {
		return fmt.Errorf("add call participant: %w", err)
	}
	return nil
}

func (r *Repository) RemoveParticipant(ctx context.Context, callID, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(ctx,
		`UPDATE call_participants SET left_at = $3 WHERE call_id = $1 AND user_id = $2`,
		callID, userID, now,
	)
	if err != nil {
		return fmt.Errorf("remove call participant: %w", err)
	}
	return nil
}

// CountActiveParticipants returns the number of participants who haven't left the call.
func (r *Repository) CountActiveParticipants(ctx context.Context, callID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM call_participants WHERE call_id = $1 AND left_at IS NULL`,
		callID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active participants: %w", err)
	}
	return count, nil
}

func (r *Repository) GetActiveCallForChannel(ctx context.Context, channelID uuid.UUID) (*model.Call, error) {
	query := `
		SELECT id, channel_id, initiator_id, call_type, status, started_at, ended_at, created_at
		FROM calls WHERE channel_id = $1 AND status IN ('ringing', 'in_progress')
		ORDER BY created_at DESC LIMIT 1
	`
	var c model.Call
	err := r.db.QueryRow(ctx, query, channelID).Scan(
		&c.ID, &c.ChannelID, &c.InitiatorID, &c.CallType, &c.Status,
		&c.StartedAt, &c.EndedAt, &c.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active call: %w", err)
	}
	return &c, nil
}

func (r *Repository) ListByChannel(ctx context.Context, channelID uuid.UUID, limit int) ([]model.Call, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `
		SELECT id, channel_id, initiator_id, call_type, status, started_at, ended_at, created_at
		FROM calls WHERE channel_id = $1
		ORDER BY created_at DESC LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, channelID, limit)
	if err != nil {
		return nil, fmt.Errorf("list calls: %w", err)
	}
	defer rows.Close()

	var calls []model.Call
	for rows.Next() {
		var c model.Call
		if err := rows.Scan(&c.ID, &c.ChannelID, &c.InitiatorID, &c.CallType, &c.Status,
			&c.StartedAt, &c.EndedAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan call: %w", err)
		}
		calls = append(calls, c)
	}
	return calls, rows.Err()
}
