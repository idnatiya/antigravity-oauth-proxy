package server

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/antigravity"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/credentials"
)

func TestAccountPoolFallsBackOnQuota(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	first := &account{id: "a", projectID: "p-a", client: &antigravity.Client{}}
	second := &account{id: "b", projectID: "p-b", client: &antigravity.Client{}}
	pool := &AccountPool{accounts: []*account{first, second}, now: func() time.Time { return now }}

	quota := &antigravity.UpstreamError{StatusCode: 429, Body: []byte(`{"error":{"details":[{"retryDelay":"30s"}]}}`)}
	var projects []string
	call := func(_ *antigravity.Client, r *antigravity.GenerateContentRequest) error {
		projects = append(projects, r.Project)
		if r.Project == "p-a" {
			return quota
		}
		return nil
	}

	if err := pool.try(&antigravity.GenerateContentRequest{Project: "orig"}, call); err != nil {
		t.Fatalf("expected fallback success, got %v", err)
	}
	if len(projects) != 2 || projects[0] != "p-a" || projects[1] != "p-b" {
		t.Fatalf("unexpected call order %v", projects)
	}
	if want := now.Add(30 * time.Second); !first.coolUntil.Equal(want) {
		t.Fatalf("coolUntil = %v, want %v", first.coolUntil, want)
	}

	// Cooling account is skipped on the next request.
	projects = nil
	_ = pool.try(&antigravity.GenerateContentRequest{}, call)
	if len(projects) != 1 || projects[0] != "p-b" {
		t.Fatalf("expected only p-b, got %v", projects)
	}

	// All accounts exhausted: the soonest-recovering one is tried and its error returned.
	projects = nil
	second.coolUntil = now.Add(time.Hour)
	err := pool.try(&antigravity.GenerateContentRequest{}, call)
	if !errors.Is(err, quota) || len(projects) != 1 || projects[0] != "p-a" {
		t.Fatalf("expected quota error from p-a, got %v %v", err, projects)
	}

	// Non-account errors are returned without trying other accounts.
	first.coolUntil, second.coolUntil = time.Time{}, time.Time{}
	projects = nil
	boom := &antigravity.UpstreamError{StatusCode: 400}
	err = pool.try(&antigravity.GenerateContentRequest{}, func(_ *antigravity.Client, r *antigravity.GenerateContentRequest) error {
		projects = append(projects, r.Project)
		return boom
	})
	if !errors.Is(err, boom) || len(projects) != 1 {
		t.Fatalf("expected single 400 attempt, got %v %v", err, projects)
	}

	if err := (&AccountPool{now: time.Now}).try(&antigravity.GenerateContentRequest{}, call); !errors.Is(err, errNoAccounts) {
		t.Fatalf("expected errNoAccounts, got %v", err)
	}
}

func TestParseRetryDelay(t *testing.T) {
	cases := map[string]time.Duration{
		`{"retryDelay": "3723.5s"}`:                   3723500 * time.Millisecond,
		`[{"metadata":{"quotaResetDelay":"1h2m3s"}}]`: time.Hour + 2*time.Minute + 3*time.Second,
		`{"retryDelay":"garbage"}`:                    defaultQuotaCooldown,
		``:                                            defaultQuotaCooldown,
	}
	for body, want := range cases {
		if got := parseRetryDelay([]byte(body)); got != want {
			t.Errorf("parseRetryDelay(%q) = %v, want %v", body, got, want)
		}
	}
}

func TestAccountPoolLogoutDefaultAccount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oauth_creds.json")
	if err := os.WriteFile(path, []byte(`{"access_token":"x"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	pool := &AccountPool{now: time.Now}
	pool.AddDefault(credentials.NewFileProviderAt(path), "p")
	if views := pool.list(); len(views) != 1 || !views[0].Removable {
		t.Fatalf("default file account should be removable: %+v", views)
	}
	if err := pool.remove("default"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("credentials file still exists: %v", err)
	}

	var upstreamErr *antigravity.UpstreamError
	err := pool.try(&antigravity.GenerateContentRequest{}, nil)
	if !errors.As(err, &upstreamErr) || upstreamErr.StatusCode != 503 {
		t.Fatalf("expected 503 with no accounts, got %v", err)
	}
}
