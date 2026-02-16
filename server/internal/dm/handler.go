package dm

import (
	"encoding/json"
	"errors"
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

func (h *Handler) CreateDM(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	ch, err := h.service.GetOrCreateDM(r.Context(), userID, req.UserID)
	if err != nil {
		if errors.Is(err, ErrSelfDM) {
			writeError(w, "cannot create DM with yourself", http.StatusBadRequest)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ch, http.StatusOK)
}

func (h *Handler) CreateGroupDM(w http.ResponseWriter, r *http.Request) {
	var req model.CreateGroupDMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	ch, err := h.service.CreateGroupDM(r.Context(), userID, req.UserIDs)
	if err != nil {
		if errors.Is(err, ErrDMTooFew) || errors.Is(err, ErrDMTooMany) {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ch, http.StatusCreated)
}

func (h *Handler) ListDMs(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	dms, err := h.service.ListDMs(r.Context(), userID)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if dms == nil {
		dms = []model.Channel{}
	}

	writeJSON(w, dms, http.StatusOK)
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
