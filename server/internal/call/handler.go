package call

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/feather-chat/feather/internal/config"
)

type Handler struct {
	service   *Service
	webrtcCfg config.WebRTCConfig
}

func NewHandler(service *Service, webrtcCfg config.WebRTCConfig) *Handler {
	return &Handler{service: service, webrtcCfg: webrtcCfg}
}

func (h *Handler) GetCallHistory(w http.ResponseWriter, r *http.Request) {
	channelID, err := uuid.Parse(chi.URLParam(r, "channelID"))
	if err != nil {
		writeError(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	calls, err := h.service.ListByChannel(r.Context(), channelID)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, calls, http.StatusOK)
}

func (h *Handler) GetActiveCall(w http.ResponseWriter, r *http.Request) {
	channelIDStr := r.URL.Query().Get("channel_id")
	if channelIDStr == "" {
		writeError(w, "channel_id query param required", http.StatusBadRequest)
		return
	}

	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		writeError(w, "invalid channel_id", http.StatusBadRequest)
		return
	}

	call, err := h.service.GetActiveCall(r.Context(), channelID)
	if err != nil {
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if call == nil {
		writeJSON(w, nil, http.StatusOK)
		return
	}

	writeJSON(w, call, http.StatusOK)
}

func (h *Handler) GetRTCConfig(w http.ResponseWriter, r *http.Request) {
	type iceServer struct {
		URLs       []string `json:"urls"`
		Username   string   `json:"username,omitempty"`
		Credential string   `json:"credential,omitempty"`
	}

	var servers []iceServer
	for _, s := range h.webrtcCfg.STUNServers {
		servers = append(servers, iceServer{URLs: []string{s}})
	}
	for _, t := range h.webrtcCfg.TURNServers {
		servers = append(servers, iceServer{
			URLs:       []string{t.URL},
			Username:   t.Username,
			Credential: t.Credential,
		})
	}

	// Default STUN if none configured
	if len(servers) == 0 {
		servers = []iceServer{{URLs: []string{"stun:stun.l.google.com:19302"}}}
	}

	writeJSON(w, map[string]interface{}{
		"ice_servers": servers,
		"enabled":     h.webrtcCfg.Enabled,
	}, http.StatusOK)
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
