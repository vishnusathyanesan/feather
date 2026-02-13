package channel

import (
	"encoding/json"
	"errors"
	"net/http"

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
	var req model.CreateChannelRequest
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

	ch, err := h.service.Create(r.Context(), req, userID, userRole)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, ch, http.StatusCreated)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	channels, err := h.service.List(r.Context(), userID)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if channels == nil {
		channels = []model.Channel{}
	}
	writeJSON(w, channels, http.StatusOK)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	ch, err := h.service.GetByID(r.Context(), channelID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, ch, http.StatusOK)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	var req model.UpdateChannelRequest
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

	ch, err := h.service.Update(r.Context(), channelID, req, userID, userRole)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, ch, http.StatusOK)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	userRole := middleware.GetUserRole(r.Context())
	if err := h.service.Delete(r.Context(), channelID, userRole); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if err := h.service.Join(r.Context(), channelID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if err := h.service.Leave(r.Context(), channelID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) InviteMember(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	var req model.InviteMemberRequest
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

	if err := h.service.InviteMember(r.Context(), channelID, userID, req.UserID, userRole); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMembers(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	members, err := h.service.GetMembers(r.Context(), channelID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if members == nil {
		members = []model.ChannelMember{}
	}
	writeJSON(w, members, http.StatusOK)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	if err := h.service.MarkRead(r.Context(), channelID, userID); err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrChannelNotFound):
		writeError(w, "channel not found", http.StatusNotFound)
	case errors.Is(err, ErrNameTaken):
		writeError(w, "channel name already taken", http.StatusConflict)
	case errors.Is(err, ErrForbidden):
		writeError(w, "forbidden", http.StatusForbidden)
	case errors.Is(err, ErrNotMember):
		writeError(w, "not a channel member", http.StatusForbidden)
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
