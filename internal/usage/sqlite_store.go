//go:build !js || !wasm

package usage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/env"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/logger"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// SQLiteStore implements the Store interface using an embedded SQLite database.
type SQLiteStore struct {
	db       *sql.DB
	dbPath   string
	mu       sync.RWMutex
	keyCache sync.Map
}

// NewSQLiteStore initializes a SQLite usage database at the specified path or default config directory.
func NewSQLiteStore(customPath string) (*SQLiteStore, error) {
	dbPath := customPath
	if dbPath == "" {
		if envPath, ok := env.Get("DATABASE_PATH"); ok && envPath != "" {
			dbPath = envPath
		} else if envPath, ok := env.Get("USAGE_DB_PATH"); ok && envPath != "" {
			dbPath = envPath
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				homeDir = "."
			}
			configDir := filepath.Join(homeDir, ".config", "antigravity-oauth-proxy")
			if err := os.MkdirAll(configDir, 0o755); err != nil {
				logger.Get().Warn().Err(err).Msg("Failed to create config dir for sqlite; falling back to current dir")
				configDir = "."
			}
			dbPath = filepath.Join(configDir, "data.sqlite")
		}
	} else if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}

	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database at %s: %w", dbPath, err)
	}

	// Optimize connection pool for SQLite WAL mode
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	store := &SQLiteStore{
		db:     db,
		dbPath: dbPath,
	}

	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite migration failed: %w", err)
	}

	if err := store.seedDefaultAdmin(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to seed default admin user: %w", err)
	}

	if err := store.seedDefaultAPIKey(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to seed default api key: %w", err)
	}

	logger.Get().Info().Str("path", dbPath).Msg("SQLite usage & telemetry store initialized")
	return store, nil
}

// DBPath returns the path to the sqlite file.
func (s *SQLiteStore) DBPath() string {
	return s.dbPath
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS requests (
		id TEXT PRIMARY KEY,
		timestamp DATETIME NOT NULL,
		endpoint TEXT NOT NULL,
		model TEXT NOT NULL,
		stream INTEGER NOT NULL,
		status_code INTEGER NOT NULL,
		duration_ms INTEGER NOT NULL,
		prompt_tokens INTEGER NOT NULL,
		completion_tokens INTEGER NOT NULL,
		total_tokens INTEGER NOT NULL,
		cost_savings REAL NOT NULL,
		error_message TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_requests_timestamp ON requests(timestamp);
	CREATE INDEX IF NOT EXISTS idx_requests_model ON requests(model);
	CREATE INDEX IF NOT EXISTS idx_requests_status ON requests(status_code);

	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		key TEXT NOT NULL,
		key_hash TEXT UNIQUE NOT NULL,
		key_prefix TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		last_used_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) seedDefaultAdmin() error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash default admin password: %w", err)
	}

	now := time.Now().UTC()
	_, err = s.db.Exec(
		"INSERT INTO users (username, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)",
		"admin", string(hash), now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert default admin: %w", err)
	}

	logger.Get().Info().Msg("Seeded default admin user with username 'admin'")
	return nil
}

func (s *SQLiteStore) seedDefaultAPIKey() error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM api_keys").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	adminKey, ok := env.Get("ADMIN_API_KEY")
	if !ok || strings.TrimSpace(adminKey) == "" {
		return nil
	}

	adminKey = strings.TrimSpace(adminKey)
	ctx := context.Background()
	_, err = s.CreateAPIKey(ctx, "Default (from ADMIN_API_KEY)", adminKey)
	if err != nil {
		return fmt.Errorf("failed to seed default API key from ADMIN_API_KEY: %w", err)
	}

	logger.Get().Info().Msg("Seeded default API key from ADMIN_API_KEY environment variable")
	return nil
}

// RecordRequest writes a completed request to SQLite.
func (s *SQLiteStore) RecordRequest(ctx context.Context, req *RequestRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `
	INSERT INTO requests (
		id, timestamp, endpoint, model, stream, status_code,
		duration_ms, prompt_tokens, completion_tokens, total_tokens,
		cost_savings, error_message
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	streamInt := 0
	if req.Stream {
		streamInt = 1
	}

	_, err := s.db.ExecContext(ctx, query,
		req.ID,
		req.Timestamp.UTC().Format(time.RFC3339Nano),
		req.Endpoint,
		req.Model,
		streamInt,
		req.StatusCode,
		req.DurationMs,
		req.PromptTokens,
		req.CompletionTokens,
		req.TotalTokens,
		req.CostSavings,
		req.ErrorMessage,
	)
	return err
}

// GetStats returns aggregated telemetry KPIs, model breakdowns, and timeline points.
func (s *SQLiteStore) GetStats(ctx context.Context, timeRange string) (*StatsSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var since time.Time
	now := time.Now().UTC()
	bucketFormat := "%Y-%m-%d %H:00"

	switch strings.ToLower(timeRange) {
	case "1h":
		since = now.Add(-1 * time.Hour)
		bucketFormat = "%Y-%m-%d %H:%M"
	case "7d":
		since = now.Add(-7 * 24 * time.Hour)
		bucketFormat = "%Y-%m-%d"
	case "30d":
		since = now.Add(-30 * 24 * time.Hour)
		bucketFormat = "%Y-%m-%d"
	case "all":
		since = time.Time{}
		bucketFormat = "%Y-%m-%d"
	case "24h":
		fallthrough
	default:
		since = now.Add(-24 * time.Hour)
		bucketFormat = "%Y-%m-%d %H:00"
	}

	summary := &StatsSummary{
		ModelBreakdown: make([]ModelUsage, 0),
		Timeline:       make([]TimelinePoint, 0),
	}

	// 1. Overall Aggregates
	aggregateQuery := `
	SELECT 
		COUNT(*),
		COALESCE(SUM(CASE WHEN status_code < 400 THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(prompt_tokens), 0),
		COALESCE(SUM(completion_tokens), 0),
		COALESCE(SUM(total_tokens), 0),
		COALESCE(SUM(cost_savings), 0.0),
		COALESCE(AVG(duration_ms), 0.0)
	FROM requests
	`
	var args []interface{}
	if !since.IsZero() {
		aggregateQuery += " WHERE timestamp >= ?"
		args = append(args, since.Format(time.RFC3339Nano))
	}

	err := s.db.QueryRowContext(ctx, aggregateQuery, args...).Scan(
		&summary.TotalRequests,
		&summary.SuccessRequests,
		&summary.FailedRequests,
		&summary.PromptTokens,
		&summary.CompletionTokens,
		&summary.TotalTokens,
		&summary.EstimatedSavings,
		&summary.AvgDurationMs,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to query stats summary: %w", err)
	}

	// 2. Model Breakdown
	modelQuery := `
	SELECT 
		model,
		COUNT(*),
		COALESCE(SUM(prompt_tokens), 0),
		COALESCE(SUM(completion_tokens), 0),
		COALESCE(SUM(total_tokens), 0),
		COALESCE(SUM(cost_savings), 0.0)
	FROM requests
	`
	if !since.IsZero() {
		modelQuery += " WHERE timestamp >= ?"
	}
	modelQuery += " GROUP BY model ORDER BY SUM(total_tokens) DESC"

	mRows, err := s.db.QueryContext(ctx, modelQuery, args...)
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var mu ModelUsage
			if err := mRows.Scan(&mu.Model, &mu.Requests, &mu.PromptTokens, &mu.CompletionTokens, &mu.TotalTokens, &mu.CostSavings); err == nil {
				mu.RequestCount = mu.Requests
				summary.ModelBreakdown = append(summary.ModelBreakdown, mu)
			}
		}
	}

	// 3. Timeline
	timelineQuery := fmt.Sprintf(`
	SELECT 
		strftime('%s', timestamp) AS bucket,
		COUNT(*),
		COALESCE(SUM(prompt_tokens), 0),
		COALESCE(SUM(completion_tokens), 0),
		COALESCE(SUM(total_tokens), 0)
	FROM requests
	`, bucketFormat)
	var tArgs []interface{}
	if !since.IsZero() {
		timelineQuery += " WHERE timestamp >= ?"
		tArgs = append(tArgs, since.Format(time.RFC3339Nano))
	}
	timelineQuery += fmt.Sprintf(" GROUP BY bucket ORDER BY bucket ASC")

	tRows, err := s.db.QueryContext(ctx, timelineQuery, tArgs...)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var tp TimelinePoint
			if err := tRows.Scan(&tp.TimeBucket, &tp.Requests, &tp.PromptTokens, &tp.CompletionTokens, &tp.TotalTokens); err == nil {
				tp.Time = tp.TimeBucket
				tp.RequestCount = tp.Requests
				summary.Timeline = append(summary.Timeline, tp)
			}
		}
	}

	summary.ErrorRequests = summary.FailedRequests
	summary.AvgLatencyMs = summary.AvgDurationMs
	summary.TimeSeries = summary.Timeline

	return summary, nil
}

// GetRequests fetches a paginated, filterable list of requests.
func (s *SQLiteStore) GetRequests(ctx context.Context, filter RequestFilter) ([]*RequestRecord, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var whereClauses []string
	var args []interface{}

	if filter.Model != "" {
		whereClauses = append(whereClauses, "model = ?")
		args = append(args, filter.Model)
	}
	if filter.Status > 0 {
		whereClauses = append(whereClauses, "status_code = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		whereClauses = append(whereClauses, "(endpoint LIKE ? OR model LIKE ? OR error_message LIKE ?)")
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern, pattern)
	}
	if filter.StartTime != nil {
		whereClauses = append(whereClauses, "timestamp >= ?")
		args = append(args, filter.StartTime.UTC())
	}
	if filter.EndTime != nil {
		whereClauses = append(whereClauses, "timestamp <= ?")
		args = append(args, filter.EndTime.UTC())
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total matching
	countQuery := "SELECT COUNT(*) FROM requests" + whereSQL
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Select records
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	dataQuery := `
	SELECT 
		id, timestamp, endpoint, model, stream, status_code,
		duration_ms, prompt_tokens, completion_tokens, total_tokens,
		cost_savings, COALESCE(error_message, '')
	FROM requests
	` + whereSQL + " ORDER BY timestamp DESC LIMIT ? OFFSET ?"

	queryArgs := append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, dataQuery, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]*RequestRecord, 0)
	for rows.Next() {
		var r RequestRecord
		var streamInt int
		if err := rows.Scan(
			&r.ID,
			&r.Timestamp,
			&r.Endpoint,
			&r.Model,
			&streamInt,
			&r.StatusCode,
			&r.DurationMs,
			&r.PromptTokens,
			&r.CompletionTokens,
			&r.TotalTokens,
			&r.CostSavings,
			&r.ErrorMessage,
		); err != nil {
			return nil, 0, err
		}
		r.Stream = streamInt == 1
		r.IsStream = r.Stream
		r.LatencyMs = r.DurationMs
		records = append(records, &r)
	}

	return records, total, nil
}

// GetUser retrieves a user by username.
func (s *SQLiteStore) GetUser(ctx context.Context, username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := "SELECT id, username, password_hash, created_at, updated_at FROM users WHERE username = ?"
	var u User
	err := s.db.QueryRowContext(ctx, query, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// UpdatePassword updates the bcrypt password hash for a user.
func (s *SQLiteStore) UpdatePassword(ctx context.Context, username, newPasswordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, "UPDATE users SET password_hash = ?, updated_at = ? WHERE username = ?", newPasswordHash, now, username)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("user '%s' not found", username)
	}
	return nil
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
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

// ListAPIKeys retrieves all registered API keys ordered newest first.
func (s *SQLiteStore) ListAPIKeys(ctx context.Context) ([]*APIKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, "SELECT id, name, key, key_hash, key_prefix, created_at, last_used_at FROM api_keys ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var k APIKey
		var lastUsed sql.NullTime
		if err := rows.Scan(&k.ID, &k.Name, &k.Key, &k.KeyHash, &k.KeyPrefix, &k.CreatedAt, &lastUsed); err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}
		if lastUsed.Valid {
			t := lastUsed.Time
			k.LastUsedAt = &t
		}
		keys = append(keys, &k)
	}
	if keys == nil {
		keys = make([]*APIKey, 0)
	}
	return keys, nil
}

// CreateAPIKey generates or registers a new API key.
func (s *SQLiteStore) CreateAPIKey(ctx context.Context, name string, customKey string) (*APIKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	res, err := s.db.ExecContext(ctx,
		"INSERT INTO api_keys (name, key, key_hash, key_prefix, created_at) VALUES (?, ?, ?, ?, ?)",
		name, keyStr, keyHash, keyPrefix, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert api key: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	apiKey := &APIKey{
		ID:        id,
		Name:      name,
		Key:       keyStr,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		CreatedAt: now,
	}

	s.keyCache.Store(keyHash, apiKey)
	return apiKey, nil
}

// DeleteAPIKey removes an API key by ID.
func (s *SQLiteStore) DeleteAPIKey(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var keyHash string
	err := s.db.QueryRowContext(ctx, "SELECT key_hash FROM api_keys WHERE id = ?", id).Scan(&keyHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("api key with id %d not found", id)
		}
		return err
	}

	_, err = s.db.ExecContext(ctx, "DELETE FROM api_keys WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}

	s.keyCache.Delete(keyHash)
	return nil
}

// ValidateAPIKey verifies an incoming raw API key against stored keys.
func (s *SQLiteStore) ValidateAPIKey(ctx context.Context, rawKey string) (*APIKey, bool, error) {
	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		return nil, false, nil
	}

	keyHash := hashAPIKey(rawKey)
	if val, ok := s.keyCache.Load(keyHash); ok {
		return val.(*APIKey), true, nil
	}

	s.mu.RLock()
	var k APIKey
	var lastUsed sql.NullTime
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, key, key_hash, key_prefix, created_at, last_used_at FROM api_keys WHERE key_hash = ?",
		keyHash,
	).Scan(&k.ID, &k.Name, &k.Key, &k.KeyHash, &k.KeyPrefix, &k.CreatedAt, &lastUsed)
	s.mu.RUnlock()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if lastUsed.Valid {
		t := lastUsed.Time
		k.LastUsedAt = &t
	}

	s.keyCache.Store(keyHash, &k)
	return &k, true, nil
}

// TouchAPIKey updates the last_used_at timestamp for a key.
func (s *SQLiteStore) TouchAPIKey(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, "UPDATE api_keys SET last_used_at = ? WHERE id = ?", now, id)
	return err
}
