package model

import (
	"time"

	"github.com/google/uuid"
)

type FileAttachment struct {
	ID          uuid.UUID  `json:"id"`
	MessageID   *uuid.UUID `json:"message_id,omitempty"`
	ChannelID   uuid.UUID  `json:"channel_id"`
	UserID      uuid.UUID  `json:"user_id"`
	Filename    string     `json:"filename"`
	ContentType string     `json:"content_type"`
	SizeBytes   int64      `json:"size_bytes"`
	StorageKey  string     `json:"-"`
	DownloadURL string     `json:"download_url,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
