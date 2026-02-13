package file

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/feather-chat/feather/internal/middleware"
)

type FileAttachment struct {
	ID          uuid.UUID  `json:"id"`
	MessageID   *uuid.UUID `json:"message_id,omitempty"`
	ChannelID   uuid.UUID  `json:"channel_id"`
	UserID      uuid.UUID  `json:"user_id"`
	Filename    string     `json:"filename"`
	ContentType string     `json:"content_type"`
	SizeBytes   int64      `json:"size_bytes"`
	StorageKey  string     `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Handler struct {
	storage *Storage
	db      *pgxpool.Pool
	maxSize int64
}

func NewHandler(storage *Storage, db *pgxpool.Pool, maxSize int64) *Handler {
	return &Handler{storage: storage, db: db, maxSize: maxSize}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)
	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		writeError(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileID := uuid.New()
	ext := filepath.Ext(header.Filename)
	storageKey := fmt.Sprintf("%s/%s%s", channelID.String(), fileID.String(), ext)

	if err := h.storage.Upload(r.Context(), storageKey, file, header.Size, header.Header.Get("Content-Type")); err != nil {
		writeError(w, "upload failed", http.StatusInternalServerError)
		return
	}

	attachment := FileAttachment{
		ID:          fileID,
		ChannelID:   channelID,
		UserID:      userID,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		SizeBytes:   header.Size,
		StorageKey:  storageKey,
		CreatedAt:   time.Now(),
	}

	query := `
		INSERT INTO file_attachments (id, channel_id, user_id, filename, content_type, size_bytes, storage_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = h.db.Exec(r.Context(), query,
		attachment.ID, attachment.ChannelID, attachment.UserID,
		attachment.Filename, attachment.ContentType, attachment.SizeBytes, attachment.StorageKey,
	)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, attachment, http.StatusCreated)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	fileID, err := uuid.Parse(chi.URLParam(r, "fileID"))
	if err != nil {
		writeError(w, "invalid file id", http.StatusBadRequest)
		return
	}

	var storageKey string
	err = h.db.QueryRow(r.Context(),
		"SELECT storage_key FROM file_attachments WHERE id = $1",
		fileID,
	).Scan(&storageKey)
	if err == pgx.ErrNoRows {
		writeError(w, "file not found", http.StatusNotFound)
		return
	}
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	url, err := h.storage.GetPresignedURL(r.Context(), storageKey)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
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
