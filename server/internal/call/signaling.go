package call

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/model"
)

// SignalingMessage represents SDP/ICE messages forwarded between peers.
type SignalingMessage struct {
	CallID   uuid.UUID       `json:"call_id"`
	FromUser uuid.UUID       `json:"from_user"`
	ToUser   uuid.UUID       `json:"to_user"`
	Data     json.RawMessage `json:"data"`
}

// RelaySignaling forwards an SDP offer/answer/ICE candidate to the target user.
func (s *Service) RelaySignaling(eventType model.EventType, msg SignalingMessage) {
	if s.sendToUser == nil {
		return
	}

	payload, _ := json.Marshal(msg)
	event := model.WebSocketEvent{
		Type:    eventType,
		Payload: payload,
	}
	data, _ := json.Marshal(event)
	s.sendToUser(msg.ToUser, data)
}
