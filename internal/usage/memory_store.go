//go:build js && wasm

package usage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/env"
	"golang.org/x/crypto/bcrypt"
)

// MemoryStore provides an in-memory implementation of Store for Cloudflare Workers.
type MemoryStore struct {
	mu       sync.RWMutex
	records  []*RequestRecord
	user     *User
	keys     []*APIKey
	capacity int
}

// NewMemoryStore initializes a memory store.
func NewMemoryStore() *MemoryStore {
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	now := time.Now().UTC()
	store := &MemoryStore{
		records:  make([]*RequestRecord, 0, 1000),
		keys:     make([]*APIKey, 0),
		capacity: 1000,
		user: &User{
			ID:           1,
			Username:     "admin",
			PasswordHash: string(hash),
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	if adminKey, ok := env.Get("ADMIN_API_KEY"); ok && strings.TrimSpace(adminKey) != "" {
		adminKey = strings.TrimSpace(adminKey)
		keyHash := hashAPIKey(adminKey)
		store.keys = append(store.keys, &APIKey{
			ID:        1,
			Name:      "Default (from ADMIN_API_KEY)",
			Key:       adminKey,
			KeyHash:   keyHash,
			KeyPrefix: formatKeyPrefix(adminKey),
			CreatedAt: now,
		})
	}

	return store
}

func (m *MemoryStore) RecordRequest(ctx context.Context, req *RequestRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.records) >= m.capacity {
		// Evict oldest
		m.records = m.records[1:]
	}
	if req.LatencyMs == 0 && req.DurationMs > 0 {
		req.LatencyMs = req.DurationMs
	}
	req.IsStream = req.Stream
	m.records = append(m.records, req)
	return nil
}

func (m *MemoryStore) GetStats(ctx context.Context, timeRange string) (*StatsSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := &StatsSummary{
		ModelBreakdown: make([]ModelUsage, 0),
		Timeline:       make([]TimelinePoint, 0),
	}

	modelMap := make(map[string]*ModelUsage)
	timelineMap := make(map[string]*TimelinePoint)
	var totalDuration int64

	for _, r := range m.records {
		summary.TotalRequests++
		if r.StatusCode < 400 {
			summary.SuccessRequests++
		} else {
			summary.FailedRequests++
		}
		summary.PromptTokens += int64(r.PromptTokens)
		summary.CompletionTokens += int64(r.CompletionTokens)
		summary.TotalTokens += int64(r.TotalTokens)
		summary.EstimatedSavings += r.CostSavings
		totalDuration += r.DurationMs

		// Model breakdown
		mu, ok := modelMap[r.Model]
		if !ok {
			mu = &ModelUsage{Model: r.Model}
			modelMap[r.Model] = mu
		}
		mu.Requests++
		mu.RequestCount++
		mu.PromptTokens += int64(r.PromptTokens)
		mu.CompletionTokens += int64(r.CompletionTokens)
		mu.TotalTokens += int64(r.TotalTokens)
		mu.CostSavings += r.CostSavings

		// Timeline
		bucket := r.Timestamp.Format("2006-01-02 15:00")
		tp, ok := timelineMap[bucket]
		if !ok {
			tp = &TimelinePoint{TimeBucket: bucket}
			timelineMap[bucket] = tp
		}
		tp.Requests++
		tp.PromptTokens += int64(r.PromptTokens)
		tp.CompletionTokens += int64(r.CompletionTokens)
		tp.TotalTokens += int64(r.TotalTokens)
	}

	if summary.TotalRequests > 0 {
		summary.AvgDurationMs = float64(totalDuration) / float64(summary.TotalRequests)
	}

	for _, mu := range modelMap {
		summary.ModelBreakdown = append(summary.ModelBreakdown, *mu)
	}
	for _, tp := range timelineMap {
		summary.Timeline = append(summary.Timeline, *tp)
	}

	return summary, nil
}

func (m *MemoryStore) GetRequests(ctx context.Context, filter RequestFilter) ([]*RequestRecord, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*RequestRecord
	for i := len(m.records) - 1; i >= 0; i-- {
		r := m.records[i]
		if filter.Model != "" && r.Model != filter.Model {
			continue
		}
		if filter.Status > 0 && r.StatusCode != filter.Status {
			continue
		}
		if filter.Search != "" {
			term := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(r.Endpoint), term) &&
				!strings.Contains(strings.ToLower(r.Model), term) &&
				!strings.Contains(strings.ToLower(r.ErrorMessage), term) {
				continue
			}
		}
		matched = append(matched, r)
	}

	total := int64(len(matched))
	start := filter.Offset
	if start > len(matched) {
		start = len(matched)
	}
	end := start + filter.Limit
	if filter.Limit <= 0 {
		end = start + 50
	}
	if end > len(matched) {
		end = len(matched)
	}

	return matched[start:end], total, nil
}

func (m *MemoryStore) GetUser(ctx context.Context, username string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.user != nil && m.user.Username == username {
		return m.user, nil
	}
	return nil, nil
}

func (m *MemoryStore) UpdatePassword(ctx context.Context, username, newPasswordHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.user != nil && m.user.Username == username {
		m.user.PasswordHash = newPasswordHash
		m.user.UpdatedAt = time.Now().UTC()
		return nil
	}
	return nil
}

func (m *MemoryStore) Close() error {
	return nil
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sk-agy-" + hex.EncodeToString(bytes), nil
}

func hashAPIKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func formatKeyPrefix(key string) string {
	if len(key) <= 12 {
		return key
	}
	return key[:8] + "..." + key[len(key)-4:]
}

func (m *MemoryStore) ListAPIKeys(ctx context.Context) ([]*APIKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*APIKey, 0, len(m.keys))
	for i := len(m.keys) - 1; i >= 0; i-- {
		result = append(result, m.keys[i])
	}
	return result, nil
}

func (m *MemoryStore) CreateAPIKey(ctx context.Context, name string, customKey string) (*APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("api key name cannot be empty")
	}

	keyStr := strings.TrimSpace(customKey)
	if keyStr == "" {
		generated, err := generateAPIKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate secure key: %w", err)
		}
		keyStr = generated
	}

	keyHash := hashAPIKey(keyStr)
	keyPrefix := formatKeyPrefix(keyStr)
	now := time.Now().UTC()

	var nextID int64 = 1
	for _, k := range m.keys {
		if k.ID >= nextID {
			nextID = k.ID + 1
		}
	}

	apiKey := &APIKey{
		ID:        nextID,
		Name:      name,
		Key:       keyStr,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		CreatedAt: now,
	}

	m.keys = append(m.keys, apiKey)
	return apiKey, nil
}

func (m *MemoryStore) DeleteAPIKey(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, k := range m.keys {
		if k.ID == id {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("api key with id %d not found", id)
}

func (m *MemoryStore) ValidateAPIKey(ctx context.Context, rawKey string) (*APIKey, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		return nil, false, nil
	}

	keyHash := hashAPIKey(rawKey)
	for _, k := range m.keys {
		if k.KeyHash == keyHash {
			return k, true, nil
		}
	}
	return nil, false, nil
}

func (m *MemoryStore) TouchAPIKey(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	for _, k := range m.keys {
		if k.ID == id {
			k.LastUsedAt = &now
			return nil
		}
	}
	return nil
}
