package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/logger"
)

type createAPIKeyRequest struct {
	Name string `json:"name"`
	Key  string `json:"key,omitempty"`
}

// apiKeysHandler manages API keys via GET, POST, and DELETE /api/keys.
func (s *Server) apiKeysHandler(w http.ResponseWriter, r *http.Request) {
	if s.usageStore == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Usage/API key store not configured"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleListAPIKeys(w, r)
	case http.MethodPost:
		s.handleCreateAPIKey(w, r)
	case http.MethodDelete:
		s.handleDeleteAPIKey(w, r)
	default:
		w.Header().Set("Allow", "GET, POST, DELETE")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := s.usageStore.ListAPIKeys(r.Context())
	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to list API keys")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to list API keys"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"keys": keys,
	})
}

func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Key name is required"})
		return
	}

	apiKey, err := s.usageStore.CreateAPIKey(r.Context(), req.Name, req.Key)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to create API key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	logger.Get().Info().Str("name", apiKey.Name).Int64("id", apiKey.ID).Msg("API key created via dashboard")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(apiKey)
}

func (s *Server) handleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Query parameter 'id' is required"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid API key ID"})
		return
	}

	if err := s.usageStore.DeleteAPIKey(r.Context(), id); err != nil {
		logger.Get().Error().Err(err).Int64("id", id).Msg("Failed to delete API key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	logger.Get().Info().Int64("id", id).Msg("API key deleted via dashboard")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
	})
}
