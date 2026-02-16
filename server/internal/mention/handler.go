package mention

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/feather-chat/feather/internal/middleware"
	"github.com/feather-chat/feather/internal/model"
)

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service, validate *validator.Validate) *Handler {
	return &Handler{service: service, validate: validate}
}

func (h *Handler) GetUnread(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	mentions, err := h.service.GetUnreadMentions(r.Context(), userID)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if mentions == nil {
		mentions = []model.Mention{}
	}

	writeJSON(w, mentions, http.StatusOK)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	var req model.MarkMentionsReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if err := h.service.MarkRead(r.Context(), userID, req.ChannelID); err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
