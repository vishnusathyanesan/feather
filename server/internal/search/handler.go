package search

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	q := r.URL.Query()

	params := SearchParams{
		Query:  q.Get("q"),
		UserID: userID,
		Limit:  20,
	}

	if channelIDStr := q.Get("channel_id"); channelIDStr != "" {
		cid, err := uuid.Parse(channelIDStr)
		if err == nil {
			params.ChannelID = &cid
		}
	}

	if authorIDStr := q.Get("user_id"); authorIDStr != "" {
		aid, err := uuid.Parse(authorIDStr)
		if err == nil {
			params.AuthorID = &aid
		}
	}

	if has := q.Get("has"); has == "link" {
		params.HasLink = true
	} else if has == "code" {
		params.HasCode = true
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	if offsetStr := q.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = offset
		}
	}

	result, err := h.service.Search(r.Context(), params)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, result, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
