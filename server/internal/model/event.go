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

	// DM events
	EventDMCreated EventType = "dm.created"

	// Mention events
	EventMentionNew EventType = "mention.new"

	// Call events
	EventCallInitiate     EventType = "call.initiate"
	EventCallRinging      EventType = "call.ringing"
	EventCallAccept       EventType = "call.accept"
	EventCallAccepted     EventType = "call.accepted"
	EventCallDecline      EventType = "call.decline"
	EventCallDeclined     EventType = "call.declined"
	EventCallOffer        EventType = "call.offer"
	EventCallAnswer       EventType = "call.answer"
	EventCallICECandidate EventType = "call.ice_candidate"
	EventCallHangup       EventType = "call.hangup"
	EventCallEnded        EventType = "call.ended"
	EventCallMissed       EventType = "call.missed"
)

type WebSocketEvent struct {
	Type      EventType       `json:"type"`
	ChannelID string          `json:"channel_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}
