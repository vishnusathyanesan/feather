package reaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feather-chat/feather/internal/model"
)

var ErrMessageNotFound = errors.New("message not found")

type BroadcastFunc func(channelID uuid.UUID, event model.WebSocketEvent)

type Service struct {
	db        *pgxpool.Pool
	broadcast BroadcastFunc
}

func NewService(db *pgxpool.Pool, broadcast BroadcastFunc) *Service {
	return &Service{db: db, broadcast: broadcast}
}

func (s *Service) AddReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	// Get message to verify it exists and get channel_id
	var channelID uuid.UUID
	err := s.db.QueryRow(ctx,
		"SELECT channel_id FROM messages WHERE id = $1 AND deleted_at IS NULL",
		messageID,
	).Scan(&channelID)
	if err != nil {
		return ErrMessageNotFound
	}

	query := `
		INSERT INTO reactions (id, message_id, user_id, emoji)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
	`
	_, err = s.db.Exec(ctx, query, uuid.New(), messageID, userID, emoji)
	if err != nil {
		return fmt.Errorf("add reaction: %w", err)
	}

	if s.broadcast != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"message_id": messageID,
			"user_id":    userID,
			"emoji":      emoji,
		})
		s.broadcast(channelID, model.WebSocketEvent{
			Type:      model.EventReactionAdded,
			ChannelID: channelID.String(),
			Payload:   payload,
		})
	}

	return nil
}

func (s *Service) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	var channelID uuid.UUID
	err := s.db.QueryRow(ctx,
		"SELECT channel_id FROM messages WHERE id = $1 AND deleted_at IS NULL",
		messageID,
	).Scan(&channelID)
	if err != nil {
		return ErrMessageNotFound
	}

	_, err = s.db.Exec(ctx,
		"DELETE FROM reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3",
		messageID, userID, emoji,
	)
	if err != nil {
		return fmt.Errorf("remove reaction: %w", err)
	}

	if s.broadcast != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"message_id": messageID,
			"user_id":    userID,
			"emoji":      emoji,
		})
		s.broadcast(channelID, model.WebSocketEvent{
			Type:      model.EventReactionRemoved,
			ChannelID: channelID.String(),
			Payload:   payload,
		})
	}

	return nil
}
