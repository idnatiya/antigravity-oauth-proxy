package server

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/env"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/logger"
)

// adminMiddleware checks the API key in the supported Gemini and OpenAI locations,
// validating against both environment ADMIN_API_KEY and dashboard-managed API keys.
func (s *Server) adminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminKey, hasAdminKey := env.Get("ADMIN_API_KEY")
		if (!hasAdminKey || adminKey == "") && s.usageStore == nil {
			logger.Get().Error().Msg("Neither ADMIN_API_KEY nor API key store is configured")
			w.Header().Set("Cache-Control", "no-store")
			http.Error(w, "API keys not configured", http.StatusInternalServerError)
			return
		}

		var providedToken string
		authHeader := r.Header.Get("Authorization")
		googApiKey := r.Header.Get("X-Goog-Api-Key")
		xAPIKey := r.Header.Get("X-API-Key")
		keyParam := r.URL.Query().Get("key")

		if authHeader != "" {
			// Expect "Bearer <token>" format, case-insensitive
			parts := strings.Fields(authHeader)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				logger.Get().Warn().Msgf("Invalid Authorization header format for protected endpoint: %s %s from %s",
					r.Method, r.URL.Path, r.RemoteAddr)
				w.Header().Set("Cache-Control", "no-store")
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}
			providedToken = parts[1]
		} else if googApiKey != "" {
			// Use X-Goog-Api-Key header directly
			providedToken = googApiKey
		} else if xAPIKey != "" {
			providedToken = xAPIKey
		} else if keyParam != "" {
			// Gemini-compatible clients commonly send API keys in this parameter.
			providedToken = keyParam
		} else {
			logger.Get().Warn().Msgf("Missing API key for protected endpoint: %s %s from %s",
				r.Method, r.URL.Path, r.RemoteAddr)
			w.Header().Set("Cache-Control", "no-store")
			http.Error(w, "Unauthorized: Missing API key", http.StatusUnauthorized)
			return
		}

		// 1. Check against environment ADMIN_API_KEY fallback
		if hasAdminKey && adminKey != "" {
			providedHash := sha256.Sum256([]byte(providedToken))
			expectedHash := sha256.Sum256([]byte(adminKey))
			if subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) == 1 {
				logger.Get().Info().Msgf("Protected request authorized via ADMIN_API_KEY: %s %s from %s",
					r.Method, r.URL.Path, r.RemoteAddr)
				next(w, r)
				return
			}
		}

		// 2. Check against dashboard-managed API keys in usage store
		if s.usageStore != nil {
			apiKey, valid, err := s.usageStore.ValidateAPIKey(r.Context(), providedToken)
			if err == nil && valid && apiKey != nil {
				logger.Get().Info().Str("key_name", apiKey.Name).Int64("key_id", apiKey.ID).Msgf("Protected request authorized via API key: %s %s from %s",
					r.Method, r.URL.Path, r.RemoteAddr)
				go func(id int64) {
					_ = s.usageStore.TouchAPIKey(context.Background(), id)
				}(apiKey.ID)
				next(w, r)
				return
			}
		}

		// Neither matched
		logger.Get().Warn().Msgf("Invalid API key for protected endpoint: %s %s from %s",
			r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set("Cache-Control", "no-store")
		http.Error(w, "Unauthorized: Invalid API key", http.StatusUnauthorized)
	}
}
