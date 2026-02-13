package websocket

import (
	"encoding/json"

	"github.com/feather-chat/feather/internal/model"
	"github.com/google/uuid"
)

func NewEvent(eventType model.EventType, channelID string, payload interface{}) model.WebSocketEvent {
	data, _ := json.Marshal(payload)
	return model.WebSocketEvent{
		Type:      eventType,
		ChannelID: channelID,
		Payload:   data,
	}
}

func MarshalEvent(event model.WebSocketEvent) ([]byte, error) {
	return json.Marshal(event)
}

func UnmarshalEvent(data []byte) (*model.WebSocketEvent, error) {
	var event model.WebSocketEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

type TypingPayload struct {
	UserID    uuid.UUID `json:"user_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	UserName  string    `json:"user_name"`
}
