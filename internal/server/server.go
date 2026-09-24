package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/antigravity"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/credentials"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/env"
	serverhttp "github.com/dvcrn/antigravity-oauth-proxy/internal/http"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/logger"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/usage"
	"github.com/google/uuid"
)

// Server represents the proxy server with its dependencies
type Server struct {
	httpClient        serverhttp.HTTPClient
	provider          credentials.CredentialsProvider
	projectID         string
	mux               *http.ServeMux
	antigravityClient *antigravity.Client
	googleAuth        *GoogleAuth
	accounts          *AccountPool
	usageStore        usage.Store
	authSecret        []byte
}

type Option func(*Server)

func WithGoogleAuth(store GoogleAuthStore) Option {
	return func(s *Server) {
		s.googleAuth = newGoogleAuth(store, s.httpClient)
	}
}

// WithAccountPool serves requests from multiple Google accounts with quota fallback.
func WithAccountPool(pool *AccountPool) Option {
	return func(s *Server) {
		s.accounts = pool
	}
}

func WithUsageStore(store usage.Store) Option {
	return func(s *Server) {
		s.usageStore = store
	}
}

func WithAuthSecret(secret []byte) Option {
	return func(s *Server) {
		s.authSecret = secret
	}
}

// NewServer creates a new server instance with the given credentials provider
func NewServer(provider credentials.CredentialsProvider, projectID string, options ...Option) *Server {
	s := &Server{
		httpClient:        serverhttp.NewHTTPClient(),
		provider:          provider,
		projectID:         projectID,
		mux:               http.NewServeMux(),
		antigravityClient: antigravity.NewClient(provider),
	}
	for _, option := range options {
		option(s)
	}
	if s.accounts == nil {
		s.accounts = &AccountPool{now: time.Now}
		s.accounts.AddDefault(provider, projectID)
	}

	if len(s.authSecret) == 0 {
		if secretStr, ok := env.Get("DASHBOARD_SECRET"); ok && secretStr != "" {
			s.authSecret = []byte(secretStr)
		} else if secretStr, ok := env.Get("SESSION_SECRET"); ok && secretStr != "" {
			s.authSecret = []byte(secretStr)
		} else {
			randomSecret := make([]byte, 32)
			_, _ = rand.Read(randomSecret)
			s.authSecret = randomSecret
		}
	}

	s.setupRoutes()

	return s
}

// Start launches the proxy server with the configured provider
func (s *Server) Start(addr string) error {
	// Load OAuth credentials on startup
	if err := s.LoadCredentials(false); errors.Is(err, fs.ErrNotExist) {
		logger.Get().Info().Msg("No default OAuth credentials; add Google accounts from the dashboard")
	} else if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to load OAuth credentials")
		logger.Get().Warn().Msg("The proxy will run but authentication will fail without valid credentials")
	}

	// Start periodic token refresh
	s.startTokenRefreshLoop()

	logger.Get().Info().Msgf("Starting proxy server on %s", addr)
	return http.ListenAndServe(addr, loggingMiddleware(s.mux))
}

// LoadCredentials loads OAuth credentials using the configured provider
func (s *Server) LoadCredentials(isPeriodicRefresh bool) error {
	creds, err := s.provider.GetCredentials()
	if err != nil {
		return err
	}

	// Check if token is expired (with a 5-minute buffer)
	if creds.ExpiryDate > 0 {
		expiryTime := time.Unix(creds.ExpiryDate/1000, 0)
		if time.Now().After(expiryTime.Add(-5 * time.Minute)) {
			logger.Get().Info().Msg("OAuth token is expired or expiring soon, attempting to refresh...")
			if err := s.provider.RefreshToken(); err != nil {
				logger.Get().Error().Err(err).Msg("Failed to refresh OAuth token")
				// Continue with the expired token, the API call might still work or will fail with 401
			}
		} else {
			if !isPeriodicRefresh {
				timeUntilExpiry := time.Until(expiryTime)
				logger.Get().Info().Dur("valid_for", timeUntilExpiry.Round(time.Second)).Msg("OAuth token valid")
			}
		}
	}

	if !isPeriodicRefresh {
		logger.Get().Info().Str("provider", s.provider.Name()).Msg("Loaded OAuth credentials")
	}
	return nil
}

// startTokenRefreshLoop starts a goroutine to periodically refresh the OAuth token.
func (s *Server) startTokenRefreshLoop() {
	// Get refresh interval from environment, default to 5 minutes
	refreshIntervalStr := env.GetOrDefault("TOKEN_REFRESH_INTERVAL", "5m")
	refreshInterval, err := time.ParseDuration(refreshIntervalStr)
	if err != nil {
		logger.Get().Warn().Err(err).Str("value", refreshIntervalStr).Msg("Invalid token refresh interval, defaulting to 5 minutes")
		refreshInterval = 5 * time.Minute
	}

	logger.Get().Info().Dur("refresh_interval", refreshInterval).Msg("Starting periodic token refresh")

	// Run the refresh loop in a separate goroutine
	go func() {
		// Create a ticker that fires at the specified interval
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for range ticker.C {
			logger.Get().Debug().Msg("Running periodic token refresh check...")
			if err := s.LoadCredentials(true); err != nil && !errors.Is(err, fs.ErrNotExist) {
				logger.Get().Error().Err(err).Msg("Error during periodic token refresh")
			}
		}
	}()
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/admin/credentials", s.adminMiddleware(s.credentialsHandler))
	s.mux.HandleFunc("/admin/credentials/status", s.adminMiddleware(s.credentialsStatusHandler))
	if s.googleAuth != nil {
		s.mux.HandleFunc("/admin/auth/start", s.adminMiddleware(s.googleAuthStartHandler))
		s.mux.HandleFunc("/admin/auth/status", s.adminMiddleware(s.googleAuthStatusHandler))
		s.mux.HandleFunc("/admin/tokens", s.adminMiddleware(s.tokensHandler))
		s.mux.HandleFunc("/admin/status", s.adminMiddleware(s.tokenStatusHandler))
	}
	s.mux.HandleFunc("/v1beta/models/", s.adminMiddleware(s.streamGenerateContentHandler))
	s.mux.HandleFunc("/v1/models/", s.modelsHandler)
	s.mux.HandleFunc("/v1/models", s.modelsHandler)
	s.mux.HandleFunc("/v1/chat/completions", s.adminMiddleware(s.openAIChatCompletionsHandler))

	// MCP endpoint. The handler is built once so the tool set is shared across
	// requests; the session itself is stateless.
	mcpHandler := s.adminMiddleware(s.mcpHandler())
	s.mux.HandleFunc("/mcp", mcpHandler)
	s.mux.HandleFunc("/mcp/", mcpHandler)

	// Dashboard Auth API
	s.mux.HandleFunc("/api/auth/login", s.handleLogin)
	s.mux.HandleFunc("/api/auth/logout", s.handleLogout)
	s.mux.HandleFunc("/api/auth/me", s.handleMe)
	s.mux.HandleFunc("/api/auth/change-password", s.dashboardAuthMiddleware(s.handleChangePassword))

	// Dashboard Usage Telemetry API
	s.mux.HandleFunc("/api/accounts", s.dashboardAuthMiddleware(s.accountsHandler))
	if s.googleAuth != nil {
		s.mux.HandleFunc("/api/accounts/auth/start", s.dashboardAuthMiddleware(s.googleAuthStartHandler))
		s.mux.HandleFunc("/api/accounts/auth/status", s.dashboardAuthMiddleware(s.googleAuthStatusHandler))
	}
	s.mux.HandleFunc("/api/usage/stats", s.dashboardAuthMiddleware(s.handleUsageStats))
	s.mux.HandleFunc("/api/usage/requests", s.dashboardAuthMiddleware(s.handleUsageRequests))

	// Dashboard UI Frontend
	s.mux.HandleFunc("/dashboard/", s.dashboardUIHandler)
	s.mux.HandleFunc("/dashboard", s.dashboardUIHandler)

	// Redirect root to /dashboard
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
}

// recordUsage logs a completed or failed request into usage store asynchronously.
func (s *Server) recordUsage(endpoint, model string, stream bool, statusCode int, duration time.Duration, promptTokens, completionTokens int, errMsg string) {
	if s.usageStore == nil {
		return
	}
	totalTokens := promptTokens + completionTokens
	record := &usage.RequestRecord{
		ID:               uuid.New().String(),
		Timestamp:        time.Now().UTC(),
		Endpoint:         endpoint,
		Model:            model,
		Stream:           stream,
		StatusCode:       statusCode,
		DurationMs:       duration.Milliseconds(),
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		CostSavings:      usage.CalculateCostSavings(model, promptTokens, completionTokens),
		ErrorMessage:     errMsg,
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.usageStore.RecordRequest(ctx, record); err != nil {
			logger.Get().Warn().Err(err).Msg("Failed to record usage telemetry")
		}
	}()
}

// ServeHTTP implements http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// credentialsHandler handles POST /admin/credentials for setting OAuth credentials
func (s *Server) credentialsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse request body - using the exact same format as oauth_creds.json
	var creds credentials.OAuthCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		logger.Get().Error().Err(err).Msg("Failed to decode credentials request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Save credentials
	if err := s.provider.SaveCredentials(&creds); err != nil {
		logger.Get().Error().Err(err).Msg("Failed to save credentials")
		http.Error(w, "Failed to save credentials", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"message": "Credentials saved successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// credentialsStatusHandler handles GET /admin/credentials/status
func (s *Server) credentialsStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Try to get current credentials
	creds, err := s.provider.GetCredentials()

	response := map[string]interface{}{
		"type":           "oauth",
		"hasCredentials": err == nil && creds != nil,
		"provider":       s.provider.Name(),
	}

	if err == nil && creds != nil {
		// Check expiry
		isExpired := false
		var expiresAt time.Time
		if creds.ExpiryDate > 0 {
			expiresAt = time.Unix(creds.ExpiryDate/1000, 0)
			isExpired = time.Now().After(expiresAt)
		}

		response["is_expired"] = isExpired
		if creds.ExpiryDate > 0 {
			response["expiry_date"] = creds.ExpiryDate
			response["expiry_date_formatted"] = expiresAt.Format(time.RFC3339)
		}
		response["has_refresh_token"] = creds.RefreshToken != ""
	} else if err != nil {
		response["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
