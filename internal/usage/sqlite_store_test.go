//go:build !js || !wasm

package usage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestSQLiteStore(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_usage.sqlite")

	store, err := NewSQLiteStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	t.Run("DefaultAdminSeeded", func(t *testing.T) {
		user, err := store.GetUser(ctx, "admin")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "admin", user.Username)

		// Verify default password is "admin"
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("admin"))
		assert.NoError(t, err, "default password should match 'admin'")
	})

	t.Run("UpdatePassword", func(t *testing.T) {
		newHash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
		require.NoError(t, err)

		err = store.UpdatePassword(ctx, "admin", string(newHash))
		require.NoError(t, err)

		user, err := store.GetUser(ctx, "admin")
		require.NoError(t, err)
		require.NotNil(t, user)

		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("secret123"))
		assert.NoError(t, err)
	})

	t.Run("RecordAndQueryRequests", func(t *testing.T) {
		now := time.Now().UTC()
		req1 := &RequestRecord{
			ID:               "req-1",
			Timestamp:        now.Add(-10 * time.Minute),
			Endpoint:         "/v1/chat/completions",
			Model:            "gemini-3.8-flash-high",
			Stream:           true,
			StatusCode:       200,
			DurationMs:       450,
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
			CostSavings:      CalculateCostSavings("gemini-3.8-flash-high", 100, 50),
		}

		req2 := &RequestRecord{
			ID:               "req-2",
			Timestamp:        now.Add(-5 * time.Minute),
			Endpoint:         "/v1beta/models/gemini-2.5-pro:generateContent",
			Model:            "gemini-2.5-pro",
			Stream:           false,
			StatusCode:       200,
			DurationMs:       1200,
			PromptTokens:     500,
			CompletionTokens: 200,
			TotalTokens:      700,
			CostSavings:      CalculateCostSavings("gemini-2.5-pro", 500, 200),
		}

		req3 := &RequestRecord{
			ID:           "req-3",
			Timestamp:    now,
			Endpoint:     "/v1/chat/completions",
			Model:        "gemini-3.8-flash-high",
			Stream:       false,
			StatusCode:   429,
			DurationMs:   100,
			ErrorMessage: "rate limit exceeded",
		}

		require.NoError(t, store.RecordRequest(ctx, req1))
		require.NoError(t, store.RecordRequest(ctx, req2))
		require.NoError(t, store.RecordRequest(ctx, req3))

		// Query Stats
		stats, err := store.GetStats(ctx, "24h")
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Equal(t, int64(3), stats.TotalRequests)
		assert.Equal(t, int64(2), stats.SuccessRequests)
		assert.Equal(t, int64(1), stats.FailedRequests)
		assert.Equal(t, int64(600), stats.PromptTokens)
		assert.Equal(t, int64(250), stats.CompletionTokens)
		assert.Equal(t, int64(850), stats.TotalTokens)
		assert.Greater(t, stats.EstimatedSavings, 0.0)

		assert.Len(t, stats.ModelBreakdown, 2)

		// Query Requests with filter
		list, total, err := store.GetRequests(ctx, RequestFilter{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, list, 3)
		assert.Equal(t, "req-3", list[0].ID) // ordered DESC

		// Filter by model
		listPro, totalPro, err := store.GetRequests(ctx, RequestFilter{Model: "gemini-2.5-pro"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), totalPro)
		assert.Len(t, listPro, 1)
		assert.Equal(t, "req-2", listPro[0].ID)

		// Filter by search
		listSearch, totalSearch, err := store.GetRequests(ctx, RequestFilter{Search: "rate limit"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), totalSearch)
		assert.Len(t, listSearch, 1)
		assert.Equal(t, "req-3", listSearch[0].ID)
	})
}
