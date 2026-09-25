package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/credentials"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/usage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyCredsProvider struct {
	credentials.CredentialsProvider
}

func (d *dummyCredsProvider) Name() string {
	return "test_provider"
}

func (d *dummyCredsProvider) GetCredentials() (*credentials.OAuthCredentials, error) {
	return &credentials.OAuthCredentials{
		AccessToken:  "mock-token",
		RefreshToken: "mock-refresh",
		ExpiryDate:   time.Now().Add(1 * time.Hour).UnixMilli(),
	}, nil
}

func (d *dummyCredsProvider) SaveCredentials(*credentials.OAuthCredentials) error {
	return nil
}

func (d *dummyCredsProvider) RefreshToken() error {
	return nil
}

func setupTestServerWithUsage(t *testing.T) (*Server, *usage.SQLiteStore) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.sqlite")
	store, err := usage.NewSQLiteStore(dbPath)
	require.NoError(t, err)

	dummyProvider := &dummyCredsProvider{}
	srv := NewServer(dummyProvider, "test-project-123", WithUsageStore(store), WithAuthSecret([]byte("test-secret-key-32-bytes-long!")))
	return srv, store
}

func TestDashboardAuthAndUsageHandlers(t *testing.T) {
	srv, store := setupTestServerWithUsage(t)
	defer store.Close()

	ctx := context.Background()

	// Seed one request record
	require.NoError(t, store.RecordRequest(ctx, &usage.RequestRecord{
		ID:               "req-test-1",
		Timestamp:        time.Now().UTC(),
		Endpoint:         "/v1/chat/completions",
		Model:            "gemini-3.8-flash-high",
		Stream:           true,
		StatusCode:       200,
		DurationMs:       320,
		PromptTokens:     150,
		CompletionTokens: 80,
		TotalTokens:      230,
		CostSavings:      usage.CalculateCostSavings("gemini-3.8-flash-high", 150, 80),
	}))

	var sessionCookie *http.Cookie

	t.Run("LoginFailureWithWrongPassword", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": "admin",
			"password": "wrongpassword",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("LoginSuccessWithDefaultAdmin", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": "admin",
			"password": "admin",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		srv.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, "admin", resp["username"].(string))
		assert.NotEmpty(t, resp["token"])

		cookies := w.Result().Cookies()
		require.NotEmpty(t, cookies)
		for _, c := range cookies {
			if c.Name == sessionCookieName {
				sessionCookie = c
				break
			}
		}
		require.NotNil(t, sessionCookie)
	})

	t.Run("AuthMe", func(t *testing.T) {
		// Without cookie -> 401
		reqUnauth := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		wUnauth := httptest.NewRecorder()
		srv.ServeHTTP(wUnauth, reqUnauth)
		assert.Equal(t, http.StatusUnauthorized, wUnauth.Code)

		// With cookie -> 200
		reqAuth := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		reqAuth.AddCookie(sessionCookie)
		wAuth := httptest.NewRecorder()
		srv.ServeHTTP(wAuth, reqAuth)
		assert.Equal(t, http.StatusOK, wAuth.Code)

		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(wAuth.Body).Decode(&resp))
		assert.True(t, resp["authenticated"].(bool))
		assert.Equal(t, "admin", resp["username"].(string))
	})

	t.Run("UsageStatsRequiresAuth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/usage/stats?range=24h", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("UsageStatsSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/usage/stats?range=24h", nil)
		req.AddCookie(sessionCookie)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.NotNil(t, resp["stats"])
		assert.NotNil(t, resp["account"])

		statsMap := resp["stats"].(map[string]interface{})
		assert.Equal(t, float64(1), statsMap["total_requests"].(float64))
		assert.Equal(t, float64(150), statsMap["prompt_tokens"].(float64))
		assert.Equal(t, float64(80), statsMap["completion_tokens"].(float64))
	})

	t.Run("UsageRequestsSuccess", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/usage/requests?limit=10", nil)
		req.AddCookie(sessionCookie)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, float64(1), resp["total"].(float64))
		requests := resp["requests"].([]interface{})
		assert.Len(t, requests, 1)
	})

	t.Run("ChangePassword", func(t *testing.T) {
		// Wrong current password
		badBody, _ := json.Marshal(map[string]string{
			"current_password": "wrong",
			"new_password":     "newpassword123",
		})
		badReq := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(badBody))
		badReq.AddCookie(sessionCookie)
		badW := httptest.NewRecorder()
		srv.ServeHTTP(badW, badReq)
		assert.Equal(t, http.StatusBadRequest, badW.Code)

		// Correct current password
		goodBody, _ := json.Marshal(map[string]string{
			"current_password": "admin",
			"new_password":     "newpassword123",
		})
		goodReq := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", bytes.NewReader(goodBody))
		goodReq.AddCookie(sessionCookie)
		goodW := httptest.NewRecorder()
		srv.ServeHTTP(goodW, goodReq)
		assert.Equal(t, http.StatusOK, goodW.Code)

		// Verify login with old password fails
		oldLoginBody, _ := json.Marshal(map[string]string{
			"username": "admin",
			"password": "admin",
		})
		oldLoginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(oldLoginBody))
		oldLoginW := httptest.NewRecorder()
		srv.ServeHTTP(oldLoginW, oldLoginReq)
		assert.Equal(t, http.StatusUnauthorized, oldLoginW.Code)

		// Verify login with new password succeeds
		newLoginBody, _ := json.Marshal(map[string]string{
			"username": "admin",
			"password": "newpassword123",
		})
		newLoginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(newLoginBody))
		newLoginW := httptest.NewRecorder()
		srv.ServeHTTP(newLoginW, newLoginReq)
		assert.Equal(t, http.StatusOK, newLoginW.Code)
	})

	t.Run("Logout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("APIKeyManagement", func(t *testing.T) {
		// 1. Create API key without auth cookie -> 401
		createBody, _ := json.Marshal(map[string]string{
			"name": "Cursor Integration",
		})
		unauthReq := httptest.NewRequest(http.MethodPost, "/api/keys", bytes.NewReader(createBody))
		unauthW := httptest.NewRecorder()
		srv.ServeHTTP(unauthW, unauthReq)
		assert.Equal(t, http.StatusUnauthorized, unauthW.Code)

		// 2. Create API key with auth cookie -> 201 Created
		createReq := httptest.NewRequest(http.MethodPost, "/api/keys", bytes.NewReader(createBody))
		createReq.AddCookie(sessionCookie)
		createW := httptest.NewRecorder()
		srv.ServeHTTP(createW, createReq)
		assert.Equal(t, http.StatusCreated, createW.Code)

		var createdKey usage.APIKey
		require.NoError(t, json.NewDecoder(createW.Body).Decode(&createdKey))
		assert.Equal(t, "Cursor Integration", createdKey.Name)
		assert.NotEmpty(t, createdKey.Key)
		assert.Contains(t, createdKey.Key, "sk-agy-")

		// 3. List API keys -> 200
		listReq := httptest.NewRequest(http.MethodGet, "/api/keys", nil)
		listReq.AddCookie(sessionCookie)
		listW := httptest.NewRecorder()
		srv.ServeHTTP(listW, listReq)
		assert.Equal(t, http.StatusOK, listW.Code)

		var listResp struct {
			Keys []*usage.APIKey `json:"keys"`
		}
		require.NoError(t, json.NewDecoder(listW.Body).Decode(&listResp))
		assert.NotEmpty(t, listResp.Keys)

		// 4. Test adminMiddleware validation with the created API key
		// Sending request to a protected endpoint e.g. /admin/credentials/status with Bearer token
		apiReq := httptest.NewRequest(http.MethodGet, "/admin/credentials/status", nil)
		apiReq.Header.Set("Authorization", "Bearer "+createdKey.Key)
		apiW := httptest.NewRecorder()
		srv.ServeHTTP(apiW, apiReq)
		assert.Equal(t, http.StatusOK, apiW.Code)

		// Testing with wrong API key -> 401
		badReq := httptest.NewRequest(http.MethodGet, "/admin/credentials/status", nil)
		badReq.Header.Set("Authorization", "Bearer sk-agy-invalid-key-999")
		badW := httptest.NewRecorder()
		srv.ServeHTTP(badW, badReq)
		assert.Equal(t, http.StatusUnauthorized, badW.Code)

		// 5. Delete the key -> 200
		delReq := httptest.NewRequest(http.MethodDelete, "/api/keys?id=1", nil)
		delReq.AddCookie(sessionCookie)
		delW := httptest.NewRecorder()
		srv.ServeHTTP(delW, delReq)
		assert.Equal(t, http.StatusOK, delW.Code)

		// 6. Request with the deleted key now fails with 401
		deletedReq := httptest.NewRequest(http.MethodGet, "/admin/credentials/status", nil)
		deletedReq.Header.Set("Authorization", "Bearer "+createdKey.Key)
		deletedW := httptest.NewRecorder()
		srv.ServeHTTP(deletedW, deletedReq)
		assert.Equal(t, http.StatusUnauthorized, deletedW.Code)
	})

	t.Run("DashboardUIHTML", func(t *testing.T) {
		routes := []string{
			"/dashboard",
			"/dashboard/",
			"/dashboard/overview",
			"/dashboard/requests",
			"/dashboard/models",
			"/dashboard/security",
			"/dashboard/login",
		}

		for _, r := range routes {
			req := httptest.NewRequest(http.MethodGet, r, nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "route: "+r)
			assert.Contains(t, w.Header().Get("Content-Type"), "text/html", "route: "+r)
		}
	})
}
