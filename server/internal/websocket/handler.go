package websocket

import (
	"context"
	"log/slog"
	"net/http"

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

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid claims", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	userName, _ := claims["name"].(string)

	conn, err := ws.Accept(w, r, &ws.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Error("websocket accept error", "error", err)
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
