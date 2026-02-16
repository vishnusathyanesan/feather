package invitation

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
	var req model.CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.MaxUses == 0 {
		req.MaxUses = 1
	}
	if req.TTLDays == 0 {
		req.TTLDays = 7
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	inv, err := h.service.Create(r.Context(), req, userID)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, inv, http.StatusCreated)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	invitations, err := h.service.List(r.Context(), userID)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, invitations, http.StatusOK)
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, "invalid invitation id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	if err := h.service.Revoke(r.Context(), id, userID, userRole); err != nil {
		if errors.Is(err, ErrInvitationNotFound) {
			writeError(w, "invitation not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrForbidden) {
			writeError(w, "forbidden", http.StatusForbidden)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	inv, err := h.service.Validate(r.Context(), token)
	if err != nil {
		if errors.Is(err, ErrInvitationNotFound) {
			writeError(w, "invitation not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrInvitationExpired) {
			writeError(w, "invitation has expired", http.StatusGone)
			return
		}
		if errors.Is(err, ErrInvitationMaxUsed) {
			writeError(w, "invitation has reached max uses", http.StatusGone)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, inv, http.StatusOK)
}

func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	userID := middleware.GetUserID(r.Context())

	if err := h.service.Accept(r.Context(), token, userID); err != nil {
		if errors.Is(err, ErrInvitationNotFound) {
			writeError(w, "invitation not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrInvitationExpired) {
			writeError(w, "invitation has expired", http.StatusGone)
			return
		}
		if errors.Is(err, ErrInvitationMaxUsed) {
			writeError(w, "invitation has reached max uses", http.StatusGone)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "accepted"}, http.StatusOK)
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
