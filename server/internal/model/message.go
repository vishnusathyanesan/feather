package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID            uuid.UUID        `json:"id"`
	ChannelID     uuid.UUID        `json:"channel_id"`
	UserID        uuid.UUID        `json:"user_id"`
	ParentID      *uuid.UUID       `json:"parent_id,omitempty"`
	Content       string           `json:"content"`
	IsAlert       bool             `json:"is_alert"`
	AlertSeverity *string          `json:"alert_severity,omitempty"`
	AlertMetadata json.RawMessage  `json:"alert_metadata,omitempty"`
	EditedAt      *time.Time       `json:"edited_at,omitempty"`
	DeletedAt     *time.Time       `json:"deleted_at,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	User          *User            `json:"user,omitempty"`
	Reactions     []ReactionGroup  `json:"reactions,omitempty"`
	Attachments   []FileAttachment `json:"attachments,omitempty"`
	ReplyCount    int              `json:"reply_count"`
}

type CreateMessageRequest struct {
	Content       string      `json:"content" validate:"required,min=1,max=10000"`
	ParentID      *uuid.UUID  `json:"parent_id"`
	AttachmentIDs []uuid.UUID `json:"attachment_ids"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=10000"`
}

type MessageListParams struct {
	ChannelID uuid.UUID
	Before    *uuid.UUID
	Limit     int
}
