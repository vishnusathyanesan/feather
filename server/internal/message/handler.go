package message

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	var req model.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	msg, err := h.service.Create(r.Context(), channelID, req, userID, userRole)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, msg, http.StatusCreated)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	params := model.MessageListParams{
		ChannelID: channelID,
		Limit:     50,
	}

	if beforeStr := r.URL.Query().Get("before"); beforeStr != "" {
		beforeID, err := uuid.Parse(beforeStr)
		if err != nil {
			writeError(w, "invalid before id", http.StatusBadRequest)
			return
		}
		params.Before = &beforeID
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 && limit <= 100 {
			params.Limit = limit
		}
	}

	userID := middleware.GetUserID(r.Context())
	messages, err := h.service.List(r.Context(), params, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if messages == nil {
		messages = []model.Message{}
	}
	writeJSON(w, messages, http.StatusOK)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	messageID, err := uuid.Parse(chi.URLParam(r, "messageID"))
	if err != nil {
		writeError(w, "invalid message id", http.StatusBadRequest)
		return
	}

	var req model.UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	msg, err := h.service.Update(r.Context(), channelID, messageID, req, userID, userRole)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, msg, http.StatusOK)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	messageID, err := uuid.Parse(chi.URLParam(r, "messageID"))
	if err != nil {
		writeError(w, "invalid message id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.service.Delete(r.Context(), channelID, messageID, userID, userRole); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	messageID, err := uuid.Parse(chi.URLParam(r, "messageID"))
	if err != nil {
		writeError(w, "invalid message id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	messages, err := h.service.GetThread(r.Context(), messageID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if messages == nil {
		messages = []model.Message{}
	}
	writeJSON(w, messages, http.StatusOK)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrMessageNotFound):
		writeError(w, "message not found", http.StatusNotFound)
	case errors.Is(err, ErrForbidden):
		writeError(w, "forbidden", http.StatusForbidden)
	case errors.Is(err, ErrEditExpired):
		writeError(w, "edit window expired (24h)", http.StatusForbidden)
	case errors.Is(err, ErrReadonly):
		writeError(w, "channel is read-only", http.StatusForbidden)
	default:
		writeError(w, "internal server error", http.StatusInternalServerError)
	}
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
