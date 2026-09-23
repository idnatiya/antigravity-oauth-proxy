package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/logger"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "ag_dashboard_session"
	sessionDuration   = 7 * 24 * time.Hour
)

type contextKey string

const userContextKey contextKey = "dashboard_user"

// createSessionToken generates a signed session token: base64(username).expiryUnix.signature
func createSessionToken(username string, secret []byte) string {
	expiry := time.Now().Add(sessionDuration).Unix()
	userB64 := base64.RawURLEncoding.EncodeToString([]byte(username))
	payload := fmt.Sprintf("%s.%d", userB64, expiry)

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s.%s", payload, sig)
}

// verifySessionToken parses and validates a signed session token.
func verifySessionToken(tokenStr string, secret []byte) (string, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid token format")
	}

	payload := fmt.Sprintf("%s.%s", parts[0], parts[1])
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", errors.New("invalid signature encoding")
	}

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(payload))
	expectedSig := h.Sum(nil)

	if !hmac.Equal(sig, expectedSig) {
		return "", errors.New("invalid token signature")
	}

	expiryUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().Unix() > expiryUnix {
		return "", errors.New("session expired")
	}

	userBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", errors.New("invalid username encoding")
	}

	return string(userBytes), nil
}

// dashboardAuthMiddleware protects dashboard API endpoints.
func (s *Server) dashboardAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var token string

		// 1. Check Cookie
		if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
			token = cookie.Value
		}

		// 2. Check Authorization header
		if token == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if token == "" {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		username, err := verifySessionToken(token, s.authSecret)
		if err != nil {
			http.Error(w, `{"error":"Invalid or expired session"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, username)
		next(w, r.WithContext(ctx))
	}
}

// handleLogin handles POST /api/auth/login.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"Username and password required"}`, http.StatusBadRequest)
		return
	}

	if s.usageStore == nil {
		http.Error(w, `{"error":"Usage store not configured"}`, http.StatusInternalServerError)
		return
	}

	user, err := s.usageStore.GetUser(r.Context(), req.Username)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to query user for login")
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, `{"error":"Invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"Invalid username or password"}`, http.StatusUnauthorized)
		return
	}

	token := createSessionToken(user.Username, s.authSecret)

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(sessionDuration),
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"username": user.Username,
		"token":    token,
	})
}

// handleLogout handles POST /api/auth/logout.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

// handleMe handles GET /api/auth/me.
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	var token string
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		token = cookie.Value
	}
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if token == "" {
		http.Error(w, `{"authenticated":false}`, http.StatusUnauthorized)
		return
	}

	username, err := verifySessionToken(token, s.authSecret)
	if err != nil {
		http.Error(w, `{"authenticated":false}`, http.StatusUnauthorized)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"username":      username,
	})
}

// handleChangePassword handles POST /api/auth/change-password.
func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	username, ok := r.Context().Value(userContextKey).(string)
	if !ok || username == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 4 {
		http.Error(w, `{"error":"New password must be at least 4 characters"}`, http.StatusBadRequest)
		return
	}

	if s.usageStore == nil {
		http.Error(w, `{"error":"Usage store not configured"}`, http.StatusInternalServerError)
		return
	}

	user, err := s.usageStore.GetUser(r.Context(), username)
	if err != nil || user == nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, `{"error":"Current password is incorrect"}`, http.StatusBadRequest)
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to hash new password")
		http.Error(w, `{"error":"Failed to hash new password"}`, http.StatusInternalServerError)
		return
	}

	if err := s.usageStore.UpdatePassword(r.Context(), username, string(newHash)); err != nil {
		logger.Get().Error().Err(err).Msg("Failed to update password in store")
		http.Error(w, `{"error":"Failed to update password"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password changed successfully",
	})
}
