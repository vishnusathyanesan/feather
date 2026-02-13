package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Webhook struct {
	ID        uuid.UUID `json:"id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Name      string    `json:"name"`
	Token     string    `json:"token,omitempty"`
	CreatorID uuid.UUID `json:"creator_id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateWebhookRequest struct {
	ChannelID uuid.UUID `json:"channel_id" validate:"required"`
	Name      string    `json:"name" validate:"required,min=2,max=100"`
}

type WebhookPayload struct {
	Channel  string          `json:"channel"`
	Title    string          `json:"title" validate:"required,max=200"`
	Severity string          `json:"severity" validate:"required,oneof=info warning critical"`
	Message  string          `json:"message" validate:"required,max=5000"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}
