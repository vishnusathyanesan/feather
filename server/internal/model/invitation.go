package model

import (
	"time"

	"github.com/google/uuid"
)

type WorkspaceInvitation struct {
	ID         uuid.UUID  `json:"id"`
	InviterID  uuid.UUID  `json:"inviter_id"`
	Email      *string    `json:"email,omitempty"`
	Token      string     `json:"token"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	AcceptedBy *uuid.UUID `json:"accepted_by,omitempty"`
	MaxUses    int        `json:"max_uses"`
	UseCount   int        `json:"use_count"`
	CreatedAt  time.Time  `json:"created_at"`
	InviteURL  string     `json:"invite_url,omitempty"`
}

type CreateInvitationRequest struct {
	Email   *string `json:"email" validate:"omitempty,email,max=255"`
	MaxUses int     `json:"max_uses" validate:"min=1,max=1000"`
	TTLDays int     `json:"ttl_days" validate:"min=1,max=30"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required"`
}
