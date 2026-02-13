package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/middleware"
	"github.com/feather-chat/feather/internal/model"
)

type AutoJoiner interface {
	AutoJoinUser(ctx context.Context, userID uuid.UUID) error
}

type Handler struct {
	service    *Service
	validate   *validator.Validate
	autoJoiner AutoJoiner
}

func NewHandler(service *Service, validate *validator.Validate, autoJoiner AutoJoiner) *Handler {
	return &Handler{service: service, validate: validate, autoJoiner: autoJoiner}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			writeError(w, "email already registered", http.StatusConflict)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Auto-join default channel
	if h.autoJoiner != nil {
		if err := h.autoJoiner.AutoJoinUser(r.Context(), resp.User.ID); err != nil {
			slog.Warn("failed to auto-join user to default channel", "error", err, "user_id", resp.User.ID)
		}
	}

	writeJSON(w, resp, http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCreds) || errors.Is(err, ErrUserDeactivated) {
			writeError(w, "invalid email or password", http.StatusUnauthorized)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp, http.StatusOK)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			writeError(w, "invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp, http.StatusOK)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			writeError(w, "user not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, user, http.StatusOK)
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
