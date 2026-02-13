package model

import "encoding/json"

type EventType string

const (
	EventMessageNew      EventType = "message.new"
	EventMessageUpdated  EventType = "message.updated"
	EventMessageDeleted  EventType = "message.deleted"
	EventReactionAdded   EventType = "reaction.added"
	EventReactionRemoved EventType = "reaction.removed"
	EventTyping          EventType = "typing"
	EventPresenceUpdate  EventType = "presence.update"
	EventChannelCreated  EventType = "channel.created"
	EventChannelUpdated  EventType = "channel.updated"
	EventChannelDeleted  EventType = "channel.deleted"
	EventMemberJoined    EventType = "member.joined"
	EventMemberLeft      EventType = "member.left"
)

type WebSocketEvent struct {
	Type      EventType       `json:"type"`
	ChannelID string          `json:"channel_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}
