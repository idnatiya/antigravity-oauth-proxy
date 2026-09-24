package usage

import (
	"context"
	"strings"
	"time"
)

// RequestRecord captures telemetry data for an individual API request.
type RequestRecord struct {
	ID               string    `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	Endpoint         string    `json:"endpoint"`
	Model            string    `json:"model"`
	Stream           bool      `json:"stream"`
	StatusCode       int       `json:"status_code"`
	DurationMs       int64     `json:"duration_ms"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	CostSavings      float64   `json:"cost_savings"`
	ErrorMessage     string    `json:"error_message,omitempty"`
}

// ModelUsage contains token and request totals for a specific model.
type ModelUsage struct {
	Model            string  `json:"model"`
	Requests         int64   `json:"requests"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	CostSavings      float64 `json:"cost_savings"`
}

// TimelinePoint represents an aggregated bucket of telemetry over time.
type TimelinePoint struct {
	TimeBucket       string `json:"time_bucket"`
	Time             string `json:"time"`
	Requests         int64  `json:"requests"`
	RequestCount     int64  `json:"request_count"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	TotalTokens      int64  `json:"total_tokens"`
}

// StatsSummary provides aggregated KPIs and charts for the usage dashboard.
type StatsSummary struct {
	TotalRequests    int64           `json:"total_requests"`
	SuccessRequests  int64           `json:"success_requests"`
	FailedRequests   int64           `json:"failed_requests"`
	ErrorRequests    int64           `json:"error_requests"`
	PromptTokens     int64           `json:"prompt_tokens"`
	CompletionTokens int64           `json:"completion_tokens"`
	TotalTokens      int64           `json:"total_tokens"`
	EstimatedSavings float64         `json:"estimated_savings"`
	AvgDurationMs    float64         `json:"avg_duration_ms"`
	AvgLatencyMs     float64         `json:"avg_latency_ms"`
	ModelBreakdown   []ModelUsage    `json:"model_breakdown"`
	Timeline         []TimelinePoint `json:"timeline"`
	TimeSeries       []TimelinePoint `json:"time_series"`
}

// User represents a dashboard admin user account.
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RequestFilter defines search, filter, and pagination criteria.
type RequestFilter struct {
	Limit     int
	Offset    int
	Model     string
	Status    int
	Search    string
	StartTime *time.Time
	EndTime   *time.Time
}

// Store defines persistent storage operations for usage metrics and authentication.
type Store interface {
	RecordRequest(ctx context.Context, req *RequestRecord) error
	GetStats(ctx context.Context, timeRange string) (*StatsSummary, error)
	GetRequests(ctx context.Context, filter RequestFilter) ([]*RequestRecord, int64, error)
	GetUser(ctx context.Context, username string) (*User, error)
	UpdatePassword(ctx context.Context, username, newPasswordHash string) error
	Close() error
}

// CalculateCostSavings calculates the estimated commercial API savings based on model and token counts.
// Pricing benchmarks based on commercial Gemini API tiers:
// - Gemini Pro: ~$1.25 / 1M prompt tokens, ~$5.00 / 1M completion tokens.
// - Gemini Flash: ~$0.15 / 1M prompt tokens, ~$0.60 / 1M completion tokens.
func CalculateCostSavings(model string, promptTokens, completionTokens int) float64 {
	lower := strings.ToLower(model)
	var promptRate, completionRate float64

	if strings.Contains(lower, "pro") {
		promptRate = 1.25 / 1_000_000.0
		completionRate = 5.00 / 1_000_000.0
	} else {
		// Flash default
		promptRate = 0.15 / 1_000_000.0
		completionRate = 0.60 / 1_000_000.0
	}

	return (float64(promptTokens) * promptRate) + (float64(completionTokens) * completionRate)
}
