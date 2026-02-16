package model

import (
	"time"

	"github.com/google/uuid"
)

type UserGroup struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatorID   uuid.UUID `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Members     []User    `json:"members,omitempty"`
}

type CreateUserGroupRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
}

type UpdateUserGroupRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

type GroupMemberRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

type Mention struct {
	ID               uuid.UUID  `json:"id"`
	MessageID        uuid.UUID  `json:"message_id"`
	ChannelID        uuid.UUID  `json:"channel_id"`
	MentionedUserID  *uuid.UUID `json:"mentioned_user_id,omitempty"`
	MentionedGroupID *uuid.UUID `json:"mentioned_group_id,omitempty"`
	MentionType      string     `json:"mention_type"`
	IsRead           bool       `json:"is_read"`
	CreatedAt        time.Time  `json:"created_at"`
	Message          *Message   `json:"message,omitempty"`
}

type MarkMentionsReadRequest struct {
	ChannelID uuid.UUID `json:"channel_id" validate:"required"`
}
