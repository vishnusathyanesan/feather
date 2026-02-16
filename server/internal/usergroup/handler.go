package usergroup

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
	var req model.CreateUserGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	g, err := h.service.Create(r.Context(), req, userID)
	if err != nil {
		if errors.Is(err, ErrGroupNameTaken) {
			writeError(w, "group name already taken", http.StatusConflict)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, g, http.StatusCreated)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	groups, err := h.service.List(r.Context(), search)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if groups == nil {
		groups = []model.UserGroup{}
	}

	writeJSON(w, groups, http.StatusOK)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, "invalid group id", http.StatusBadRequest)
		return
	}

	g, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			writeError(w, "group not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, g, http.StatusOK)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, "invalid group id", http.StatusBadRequest)
		return
	}

	var req model.UpdateUserGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	g, err := h.service.Update(r.Context(), id, req, userID, userRole)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			writeError(w, "group not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, ErrForbidden) {
			writeError(w, "forbidden", http.StatusForbidden)
			return
		}
		if errors.Is(err, ErrGroupNameTaken) {
			writeError(w, "group name already taken", http.StatusConflict)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, g, http.StatusOK)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, "invalid group id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	if err := h.service.Delete(r.Context(), id, userID, userRole); err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			writeError(w, "group not found", http.StatusNotFound)
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

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, "invalid group id", http.StatusBadRequest)
		return
	}

	var req model.GroupMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.AddMember(r.Context(), groupID, req.UserID); err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			writeError(w, "group not found", http.StatusNotFound)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, "invalid group id", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveMember(r.Context(), groupID, userID); err != nil {
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
