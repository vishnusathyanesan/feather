package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/feather-chat/feather/internal/model"
)

type Hub struct {
	instanceID string
	clients    map[uuid.UUID]*Client
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
	redis      *redis.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

type redisEnvelope struct {
	InstanceID string          `json:"instance_id"`
	Data       json.RawMessage `json:"data"`
}

func NewHub(redisClient *redis.Client) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		instanceID: uuid.New().String(),
		clients:    make(map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		redis:      redisClient,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (h *Hub) Run() {
	// Start Redis subscriber
	go h.subscribeRedis()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			slog.Info("client connected", "user_id", client.UserID, "client_id", client.ID)

			// Broadcast presence update
			h.broadcastPresence(client.UserID, true)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				close(client.send)
				delete(h.clients, client.ID)
			}
			h.mu.Unlock()
			slog.Info("client disconnected", "user_id", client.UserID, "client_id", client.ID)

			h.broadcastPresence(client.UserID, false)

		case <-h.ctx.Done():
			return
		}
	}
}

func (h *Hub) Stop() {
	h.cancel()
}

func (h *Hub) BroadcastToChannel(channelID uuid.UUID, data []byte, excludeClientID uuid.UUID) {
	// Always deliver locally first
	h.deliverToChannel(channelID, data, excludeClientID)

	// Also publish to Redis for cross-instance delivery
	if h.redis != nil {
		envelope, err := json.Marshal(redisEnvelope{
			InstanceID: h.instanceID,
			Data:       data,
		})
		if err != nil {
			slog.Error("failed to marshal redis envelope", "error", err)
			return
		}
		h.redis.Publish(h.ctx, "feather:channel:"+channelID.String(), envelope)
	}
}

func (h *Hub) BroadcastEvent(channelID uuid.UUID, event model.WebSocketEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal event", "error", err)
		return
	}
	h.BroadcastToChannel(channelID, data, uuid.Nil)
}

func (h *Hub) deliverToChannel(channelID uuid.UUID, data []byte, excludeClientID uuid.UUID) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.ID == excludeClientID {
			continue
		}
		if client.IsSubscribed(channelID) {
			select {
			case client.send <- data:
			default:
				// Client buffer full, skip
			}
		}
	}
}

func (h *Hub) subscribeRedis() {
	if h.redis == nil {
		return
	}

	pubsub := h.redis.PSubscribe(h.ctx, "feather:channel:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			// Unwrap envelope and skip messages from this instance
			var env redisEnvelope
			if err := json.Unmarshal([]byte(msg.Payload), &env); err != nil {
				continue
			}
			if env.InstanceID == h.instanceID {
				continue // Already delivered locally
			}

			// Extract channel ID from Redis channel name
			channelIDStr := msg.Channel[len("feather:channel:"):]
			channelID, err := uuid.Parse(channelIDStr)
			if err != nil {
				continue
			}
			h.deliverToChannel(channelID, env.Data, uuid.Nil)

		case <-h.ctx.Done():
			return
		}
	}
}

func (h *Hub) broadcastPresence(userID uuid.UUID, online bool) {
	payload := map[string]interface{}{
		"user_id": userID,
		"online":  online,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal presence payload", "error", err)
		return
	}
	event := model.WebSocketEvent{
		Type:    model.EventPresenceUpdate,
		Payload: data,
	}
	eventData, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal presence event", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.send <- eventData:
		default:
		}
	}
}

func (h *Hub) GetOnlineUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	seen := make(map[uuid.UUID]bool)
	var users []uuid.UUID
	for _, client := range h.clients {
		if !seen[client.UserID] {
			seen[client.UserID] = true
			users = append(users, client.UserID)
		}
	}
	return users
}
