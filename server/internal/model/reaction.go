package model

import (
	"time"

	"github.com/google/uuid"
)

type Reaction struct {
	ID        uuid.UUID `json:"id"`
	MessageID uuid.UUID `json:"message_id"`
	UserID    uuid.UUID `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type ReactionGroup struct {
	Emoji string      `json:"emoji"`
	Count int         `json:"count"`
	Users []uuid.UUID `json:"users"`
}

type AddReactionRequest struct {
	Emoji string `json:"emoji" validate:"required,min=1,max=50"`
}
