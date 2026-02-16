package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/feather-chat/feather/internal/model"
	"github.com/google/uuid"
	ws "nhooyr.io/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
	maxMsgSize = 16384
)

type Client struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	UserName  string
	conn      *ws.Conn
	hub       *Hub
	send      chan []byte
	channels  map[uuid.UUID]bool
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewClient(conn *ws.Conn, hub *Hub, userID uuid.UUID, userName string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		ID:       uuid.New(),
		UserID:   userID,
		UserName: userName,
		conn:     conn,
		hub:      hub,
		send:     make(chan []byte, 256),
		channels: make(map[uuid.UUID]bool),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (c *Client) SubscribeChannel(channelID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.channels[channelID] = true
}

func (c *Client) UnsubscribeChannel(channelID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.channels, channelID)
}

func (c *Client) IsSubscribed(channelID uuid.UUID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.channels[channelID]
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(ws.StatusNormalClosure, "")
		c.cancel()
	}()

	c.conn.SetReadLimit(maxMsgSize)

	for {
		_, data, err := c.conn.Read(c.ctx)
		if err != nil {
			if ws.CloseStatus(err) != ws.StatusNormalClosure {
				slog.Debug("websocket read error", "user_id", c.UserID, "error", err)
			}
			return
		}

		var event model.WebSocketEvent
		if err := json.Unmarshal(data, &event); err != nil {
			continue
		}

		// Handle client-sent events
		switch event.Type {
		case model.EventTyping:
			c.handleTyping(event)
		case model.EventCallInitiate, model.EventCallAccept, model.EventCallDecline,
			model.EventCallOffer, model.EventCallAnswer, model.EventCallICECandidate,
			model.EventCallHangup:
			if c.hub.callHandler != nil {
				c.hub.callHandler(c.UserID, event)
			}
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close(ws.StatusNormalClosure, "")
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			ctx, cancel := context.WithTimeout(c.ctx, writeWait)
			err := c.conn.Write(ctx, ws.MessageText, message)
			cancel()
			if err != nil {
				return
			}
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(c.ctx, writeWait)
			err := c.conn.Ping(ctx)
			cancel()
			if err != nil {
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Client) handleTyping(event model.WebSocketEvent) {
	channelIDStr := event.ChannelID
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		return
	}

	if !c.IsSubscribed(channelID) {
		return
	}

	payload := TypingPayload{
		UserID:    c.UserID,
		ChannelID: channelID,
		UserName:  c.UserName,
	}

	data, _ := json.Marshal(NewEvent(model.EventTyping, channelIDStr, payload))
	c.hub.BroadcastToChannel(channelID, data, c.ID)
}
