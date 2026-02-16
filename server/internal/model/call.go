package model

import (
	"time"

	"github.com/google/uuid"
)

type CallStatus string
type CallType string

const (
	CallStatusRinging    CallStatus = "ringing"
	CallStatusInProgress CallStatus = "in_progress"
	CallStatusEnded      CallStatus = "ended"
	CallStatusMissed     CallStatus = "missed"
	CallStatusDeclined   CallStatus = "declined"
)

const (
	CallTypeAudio CallType = "audio"
	CallTypeVideo CallType = "video"
)

type Call struct {
	ID          uuid.UUID  `json:"id"`
	ChannelID   uuid.UUID  `json:"channel_id"`
	InitiatorID uuid.UUID  `json:"initiator_id"`
	CallType    CallType   `json:"call_type"`
	Status      CallStatus `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CallParticipant struct {
	CallID   uuid.UUID  `json:"call_id"`
	UserID   uuid.UUID  `json:"user_id"`
	JoinedAt *time.Time `json:"joined_at,omitempty"`
	LeftAt   *time.Time `json:"left_at,omitempty"`
}

type InitiateCallRequest struct {
	ChannelID uuid.UUID `json:"channel_id" validate:"required"`
	CallType  CallType  `json:"call_type" validate:"required,oneof=audio video"`
}
