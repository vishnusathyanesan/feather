package audit

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Logger struct {
	db *pgxpool.Pool
}

func NewLogger(db *pgxpool.Pool) *Logger {
	return &Logger{db: db}
}

type Entry struct {
	UserID     uuid.UUID
	Action     string
	EntityType string
	EntityID   uuid.UUID
	Metadata   map[string]interface{}
	IPAddress  string
}

func (l *Logger) Log(ctx context.Context, entry Entry) {
	query := `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, metadata, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := l.db.Exec(ctx, query,
		entry.UserID, entry.Action, entry.EntityType, entry.EntityID, entry.Metadata, entry.IPAddress,
	)
	if err != nil {
		slog.Error("failed to write audit log", "error", err, "action", entry.Action)
	}
}
