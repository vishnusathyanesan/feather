package model

import (
	"time"

	"github.com/google/uuid"
)

type ChannelType string

const (
	ChannelPublic  ChannelType = "public"
	ChannelPrivate ChannelType = "private"
	ChannelSystem  ChannelType = "system"
)

type Channel struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Topic       string      `json:"topic"`
	Description string      `json:"description"`
	Type        ChannelType `json:"type"`
	IsReadonly  bool        `json:"is_readonly"`
	CreatorID   *uuid.UUID  `json:"creator_id,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	UnreadCount int         `json:"unread_count,omitempty"`
	MemberCount int         `json:"member_count,omitempty"`
}

type CreateChannelRequest struct {
	Name        string      `json:"name" validate:"required,min=2,max=100"`
	Topic       string      `json:"topic" validate:"max=500"`
	Description string      `json:"description" validate:"max=2000"`
	Type        ChannelType `json:"type" validate:"required,oneof=public private system"`
	IsReadonly  bool        `json:"is_readonly"`
}

type UpdateChannelRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Topic       *string `json:"topic" validate:"omitempty,max=500"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	IsReadonly  *bool   `json:"is_readonly"`
}

type ChannelMember struct {
	ChannelID  uuid.UUID `json:"channel_id"`
	UserID     uuid.UUID `json:"user_id"`
	Role       string    `json:"role"`
	LastReadAt time.Time `json:"last_read_at"`
	JoinedAt   time.Time `json:"joined_at"`
	User       *User     `json:"user,omitempty"`
}

type InviteMemberRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}
