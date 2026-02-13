package webhook

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
	var req model.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	wh, err := h.service.Create(r.Context(), req, userID)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, wh, http.StatusCreated)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	webhooks, err := h.service.List(r.Context(), userID)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if webhooks == nil {
		webhooks = []model.Webhook{}
	}
	writeJSON(w, webhooks, http.StatusOK)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	webhookID, err := uuid.Parse(chi.URLParam(r, "webhookID"))
	if err != nil {
		writeError(w, "invalid webhook id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())

	if err := h.service.Delete(r.Context(), webhookID, userID, userRole); err != nil {
		if errors.Is(err, ErrWebhookNotFound) {
			writeError(w, "webhook not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleIncoming(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, "missing token", http.StatusBadRequest)
		return
	}

	var payload model.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(payload); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.HandleIncoming(r.Context(), token, payload); err != nil {
		if errors.Is(err, ErrInvalidToken) {
			writeError(w, "invalid webhook token", http.StatusUnauthorized)
			return
		}
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
