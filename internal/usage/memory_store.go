//go:build js && wasm

package usage

import (
	"context"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// MemoryStore provides an in-memory implementation of Store for Cloudflare Workers.
type MemoryStore struct {
	mu       sync.RWMutex
	records  []*RequestRecord
	user     *User
	capacity int
}

// NewMemoryStore initializes a memory store.
func NewMemoryStore() *MemoryStore {
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	now := time.Now().UTC()
	return &MemoryStore{
		records:  make([]*RequestRecord, 0, 1000),
		capacity: 1000,
		user: &User{
			ID:           1,
			Username:     "admin",
			PasswordHash: string(hash),
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
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
