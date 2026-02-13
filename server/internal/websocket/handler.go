package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	ws "nhooyr.io/websocket"
)

type ChannelLister interface {
	GetUserChannelIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type WSHandler struct {
	hub       *Hub
	jwtSecret string
	channels  ChannelLister
}

func NewHandler(hub *Hub, jwtSecret string, channels ChannelLister) *WSHandler {
	return &WSHandler{
		hub:       hub,
		jwtSecret: jwtSecret,
		channels:  channels,
	}
}

// authMessage is the expected first message from the client.
type authMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type authPayload struct {
	Token string `json:"token"`
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.Accept(w, r, &ws.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Error("websocket accept error", "error", err)
		return
	}

	// Wait for auth message (first message must be auth within 10s)
	authCtx, authCancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer authCancel()

	_, data, err := conn.Read(authCtx)
	if err != nil {
		conn.Close(ws.StatusPolicyViolation, "auth timeout")
		return
	}

	var msg authMessage
	if err := json.Unmarshal(data, &msg); err != nil || msg.Type != "auth" {
		conn.Close(ws.StatusPolicyViolation, "first message must be auth")
		return
	}

	var payload authPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil || payload.Token == "" {
		conn.Close(ws.StatusPolicyViolation, "missing token")
		return
	}

	// Validate JWT
	userID, userName, err := h.validateToken(payload.Token)
	if err != nil {
		conn.Close(ws.StatusPolicyViolation, "invalid token")
		return
	}

	client := NewClient(conn, h.hub, userID, userName)

	// Subscribe to user's channels
	if h.channels != nil {
		channelIDs, err := h.channels.GetUserChannelIDs(r.Context(), userID)
		if err == nil {
			for _, chID := range channelIDs {
				client.SubscribeChannel(chID)
			}
		}
	}

	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (h *WSHandler) validateToken(tokenStr string) (uuid.UUID, string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, "", jwt.ErrTokenInvalidClaims
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return uuid.Nil, "", jwt.ErrTokenInvalidClaims
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, "", err
	}

	userName, _ := claims["name"].(string)
	return userID, userName, nil
}
