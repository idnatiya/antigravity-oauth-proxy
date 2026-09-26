package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/antigravity"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/auth"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/credentials"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/logger"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/project"
)

const (
	defaultQuotaCooldown       = time.Hour
	quotaCacheTTL              = time.Minute
	quotaFetchTimeout          = 15 * time.Second
	credentialsFailureCooldown = 30 * time.Minute
)

var (
	// errNoAccounts is an UpstreamError so handlers forward it like any upstream failure.
	errNoAccounts = &antigravity.UpstreamError{
		StatusCode:  http.StatusServiceUnavailable,
		ContentType: "application/json",
		Body:        []byte(`{"error":{"code":503,"message":"No Google account available. Add one at /dashboard/accounts","status":"UNAVAILABLE"}}`),
	}
	retryDelayPattern = regexp.MustCompile(`"(?:retryDelay|quotaResetDelay)"\s*:\s*"([^"]+)"`)
)

type account struct {
	id                string // email, or "default" for the legacy oauth_creds.json account
	path              string // "" when credentials come from CLOUDCODE_OAUTH_CREDS; cannot be removed
	projectID         string
	client            *antigravity.Client
	coolUntil         time.Time                 // guarded by AccountPool.mu
	coolReason        string                    // guarded by AccountPool.mu
	quota             *antigravity.QuotaSummary // guarded by AccountPool.mu
	quotaErr          string                    // guarded by AccountPool.mu
	quotaAt           time.Time                 // guarded by AccountPool.mu
	needsVerification bool                      // guarded by AccountPool.mu
	validationURL     string                    // guarded by AccountPool.mu
}

// AccountPool holds the Google accounts used for upstream requests. Accounts
// are tried in order; one that hits its quota is skipped until its cooldown ends.
type AccountPool struct {
	mu       sync.Mutex
	dir      string
	accounts []*account
	session  []byte // pending web OAuth session; in memory only
	now      func() time.Time
}

// NewAccountPool loads every account file in dir. A missing dir is fine.
func NewAccountPool(dir string) *AccountPool {
	p := &AccountPool{dir: dir, now: time.Now}
	paths, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	sort.Strings(paths)
	for _, path := range paths {
		provider := credentials.NewFileProviderAt(path)
		creds, err := provider.GetCredentials()
		if err != nil {
			logger.Get().Warn().Err(err).Str("path", path).Msg("Skipping unreadable account file")
			continue
		}
		id := creds.Email
		if id == "" {
			id = strings.TrimSuffix(filepath.Base(path), ".json")
		}
		projectID := creds.ProjectID
		client := antigravity.NewClient(provider)
		if projectID == "" {
			if projectID, err = discoverProject(provider, client); err != nil {
				logger.Get().Warn().Err(err).Str("account", id).Msg("Skipping account without project ID")
				continue
			}
		}
		p.accounts = append(p.accounts, &account{id: id, path: path, projectID: projectID, client: client})
	}
	return p
}

// AddDefault puts the single-account provider first in the pool.
func (p *AccountPool) AddDefault(provider credentials.CredentialsProvider, projectID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	a := &account{id: "default", projectID: projectID, client: antigravity.NewClient(provider)}
	if fp, ok := provider.(*credentials.FileProvider); ok {
		a.path = fp.Path() // "" when credentials come from CLOUDCODE_OAUTH_CREDS
	}
	p.accounts = append([]*account{a}, p.accounts...)
}

// candidates returns ready accounts in order. When every account is cooling
// down, it returns the one that recovers first so the client sees the real
// upstream error instead of a synthetic one.
func (p *AccountPool) candidates() []*account {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := p.now()
	var ready []*account
	var soonest *account
	for _, a := range p.accounts {
		if !now.Before(a.coolUntil) {
			ready = append(ready, a)
		} else if soonest == nil || a.coolUntil.Before(soonest.coolUntil) {
			soonest = a
		}
	}
	if len(ready) == 0 && soonest != nil {
		return []*account{soonest}
	}
	return ready
}

// try runs call with each candidate account until one succeeds or fails with
// an error that is not account-specific.
func (p *AccountPool) try(req *antigravity.GenerateContentRequest, call func(*antigravity.Client, *antigravity.GenerateContentRequest) error) error {
	accounts := p.candidates()
	if len(accounts) == 0 {
		return errNoAccounts
	}
	var err error
	for _, a := range accounts {
		r := *req
		r.Project = a.projectID
		if err = call(a.client, &r); err == nil {
			return nil
		}
		var upstreamErr *antigravity.UpstreamError
		if errors.As(err, &upstreamErr) && upstreamErr.IsVerificationRequired() {
			vURL := upstreamErr.ExtractValidationURL()
			p.mu.Lock()
			a.needsVerification = true
			a.validationURL = vURL
			p.mu.Unlock()
			logger.Get().Warn().
				Str("account", a.id).
				Str("validation_url", vURL).
				Msg("Google account requires verification (403 VALIDATION_REQUIRED)")
		}
		d, ok := cooldownFor(err)
		if !ok {
			return err
		}
		p.mu.Lock()
		a.coolUntil = p.now().Add(d)
		if a.needsVerification {
			a.coolReason = "Account verification required by Google"
		} else {
			a.coolReason = err.Error()
		}
		p.mu.Unlock()
		logger.Get().Warn().Err(err).Str("account", a.id).Dur("cooldown", d).Msg("Account unavailable, trying next account")
	}
	return err
}

// cooldownFor reports whether err means "this account can't serve now" and for how long.
func cooldownFor(err error) (time.Duration, bool) {
	var upstreamErr *antigravity.UpstreamError
	if errors.As(err, &upstreamErr) {
		if upstreamErr.IsVerificationRequired() {
			return 30 * time.Minute, true
		}
		if upstreamErr.StatusCode == 429 {
			return parseRetryDelay(upstreamErr.Body), true
		}
		if upstreamErr.StatusCode >= 500 && upstreamErr.StatusCode <= 504 {
			return 15 * time.Second, true
		}
	}
	var credsErr *antigravity.CredentialsError
	if errors.As(err, &credsErr) {
		return credentialsFailureCooldown, true
	}
	return 0, false
}

// parseRetryDelay reads google.rpc.RetryInfo.retryDelay or ErrorInfo
// quotaResetDelay ("3723.4s", "1h2m3s") from a 429 body.
func parseRetryDelay(body []byte) time.Duration {
	if m := retryDelayPattern.FindSubmatch(body); m != nil {
		if d, err := time.ParseDuration(string(m[1])); err == nil && d > 0 {
			return d
		}
	}
	return defaultQuotaCooldown
}

// ready returns how many accounts can serve requests now, and the total.
func (p *AccountPool) ready() (ready, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := p.now()
	for _, a := range p.accounts {
		if !now.Before(a.coolUntil) {
			ready++
		}
	}
	return ready, len(p.accounts)
}

type accountView struct {
	ID            string `json:"id"`
	ProjectID     string `json:"projectId"`
	Removable     bool   `json:"removable"`
	CoolingUntil  string `json:"coolingUntil,omitempty"`
	CoolingReason string `json:"coolingReason,omitempty"`

	Quota      *antigravity.QuotaSummary `json:"quota,omitempty"`
	QuotaError string                    `json:"quotaError,omitempty"`

	NeedsVerification bool   `json:"needsVerification,omitempty"`
	ValidationURL     string `json:"validationUrl,omitempty"`
}

// refreshQuotas re-reads usage for accounts whose cached summary is older than
// quotaCacheTTL, so reloading the dashboard doesn't hammer Google.
func (p *AccountPool) refreshQuotas(ctx context.Context) {
	p.mu.Lock()
	now := p.now()
	var stale []*account
	for _, a := range p.accounts {
		if now.Sub(a.quotaAt) >= quotaCacheTTL {
			stale = append(stale, a)
		}
	}
	p.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, quotaFetchTimeout)
	defer cancel()
	var wg sync.WaitGroup
	for _, a := range stale {
		wg.Add(1)
		go func() {
			defer wg.Done()
			summary, err := a.client.RetrieveUserQuotaSummary(ctx, a.projectID)
			p.mu.Lock()
			defer p.mu.Unlock()
			a.quotaAt = p.now()
			if err != nil {
				var upstreamErr *antigravity.UpstreamError
				if errors.As(err, &upstreamErr) && upstreamErr.IsVerificationRequired() {
					vURL := upstreamErr.ExtractValidationURL()
					a.needsVerification = true
					a.validationURL = vURL
					a.quotaErr = "Account verification required by Google"
					a.coolUntil = p.now().Add(30 * time.Minute)
					a.coolReason = "Account verification required by Google"
					logger.Get().Warn().Str("account", a.id).Str("validation_url", vURL).Msg("Quota check: account requires verification")
					return
				}
				errMsg := err.Error()
				if strings.Contains(errMsg, "403") || strings.Contains(strings.ToLower(errMsg), "permission") || strings.Contains(strings.ToLower(errMsg), "subscription") {
					a.quotaErr = "No Google AI Pro subscription on this account"
				} else {
					a.quotaErr = errMsg
				}
				logger.Get().Warn().Err(err).Str("account", a.id).Msg("Failed to read account quota")
				return
			}
			a.quota, a.quotaErr = summary, ""
			if a.needsVerification {
				a.needsVerification = false
				a.validationURL = ""
			}
		}()
	}
	wg.Wait()
}

func (p *AccountPool) list() []accountView {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := p.now()
	views := make([]accountView, 0, len(p.accounts))
	for _, a := range p.accounts {
		v := accountView{
			ID:                a.id,
			ProjectID:         a.projectID,
			Removable:         a.path != "",
			Quota:             a.quota,
			QuotaError:        a.quotaErr,
			NeedsVerification: a.needsVerification,
			ValidationURL:     a.validationURL,
		}
		if now.Before(a.coolUntil) {
			v.CoolingUntil = a.coolUntil.UTC().Format(time.RFC3339)
			v.CoolingReason = a.coolReason
		}
		views = append(views, v)
	}
	return views
}

// remove logs an account out by deleting its local token file. The token is
// deliberately not revoked at Google: it belongs to Antigravity's OAuth client,
// so revoking would also sign the user out of the Antigravity IDE.
func (p *AccountPool) remove(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, a := range p.accounts {
		if a.id != id {
			continue
		}
		if a.path == "" {
			return errors.New("account is configured via environment and cannot be logged out here")
		}
		if err := os.Remove(a.path); err != nil && !os.IsNotExist(err) {
			return err
		}
		p.accounts = append(p.accounts[:i], p.accounts[i+1:]...)
		return nil
	}
	return errors.New("account not found")
}

// GoogleAuthStore implementation: the web login adds a new account.

// GetCredentials returns an error so GoogleAuth never reuses another
// account's refresh token for a new login.
func (p *AccountPool) GetCredentials() (*credentials.OAuthCredentials, error) {
	return nil, errors.New("no current account")
}

func (p *AccountPool) LoadGoogleAuthSession() ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.session, nil
}

func (p *AccountPool) SaveGoogleAuthSession(session []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.session = session
	return nil
}

func (p *AccountPool) CompleteGoogleAuth(creds *credentials.OAuthCredentials) error {
	ctx, cancel := context.WithTimeout(context.Background(), googleAuthTimeout)
	defer cancel()
	info, err := auth.FetchUserInfo(ctx, creds.AccessToken)
	if err != nil {
		return fmt.Errorf("fetch Google account email: %w", err)
	}
	email := strings.TrimSpace(info.Email)
	if email == "" || strings.ContainsAny(email, `/\`) || strings.HasPrefix(email, ".") {
		return fmt.Errorf("unexpected Google account email %q", email)
	}

	if p.dir != "" {
		_ = os.MkdirAll(p.dir, 0o700)
	}
	path := filepath.Join(p.dir, email+".json")
	provider := credentials.NewFileProviderAt(path)
	creds.Email = email
	if err := provider.SaveCredentials(creds); err != nil {
		return err
	}
	client := antigravity.NewClient(provider)
	projectID, err := discoverProject(provider, client)
	if err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("discover project for %s: %w", email, err)
	}
	creds.ProjectID = projectID
	if err := provider.SaveCredentials(creds); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	a := &account{id: email, path: path, projectID: projectID, client: client}
	for i, existing := range p.accounts {
		if existing.id == email {
			p.accounts[i] = a
			return nil
		}
	}
	p.accounts = append(p.accounts, a)
	logger.Get().Info().Str("account", email).Str("project_id", projectID).Msg("Added Google account")
	return nil
}

func discoverProject(provider credentials.CredentialsProvider, client *antigravity.Client) (string, error) {
	loadAssist, err := client.LoadCodeAssist()
	if err != nil {
		return "", err
	}
	return project.Discover(provider, "", loadAssist)
}

func (s *Server) generate(req *antigravity.GenerateContentRequest) (*antigravity.GenerateContentResponse, error) {
	var resp *antigravity.GenerateContentResponse
	err := s.accounts.try(req, func(c *antigravity.Client, r *antigravity.GenerateContentRequest) error {
		var err error
		resp, err = c.GenerateContent(r)
		return err
	})
	return resp, err
}

// stream can retry on another account because StreamGenerateContent only
// returns an error before it starts writing to out.
func (s *Server) stream(ctx context.Context, req *antigravity.GenerateContentRequest, out chan<- string) error {
	return s.accounts.try(req, func(c *antigravity.Client, r *antigravity.GenerateContentRequest) error {
		return c.StreamGenerateContent(ctx, r, out)
	})
}

// fetchModels lists models using the first ready account.
func (s *Server) fetchModels(ctx context.Context) (*antigravity.FetchAvailableModelsResponse, error) {
	accounts := s.accounts.candidates()
	if len(accounts) == 0 {
		return nil, errNoAccounts
	}
	return accounts[0].client.FetchAvailableModels(ctx)
}

func (s *Server) accountsHandler(w http.ResponseWriter, r *http.Request) {
	if !s.googleAuthRequestAllowed(w, r, http.MethodGet, http.MethodDelete) {
		return
	}
	if r.Method == http.MethodDelete {
		if err := s.accounts.remove(r.URL.Query().Get("id")); err != nil {
			writeAdminError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if r.Method == http.MethodGet {
		s.accounts.refreshQuotas(r.Context())
	}
	writeAdminJSON(w, map[string]any{
		"accounts":        s.accounts.list(),
		"webLoginEnabled": s.googleAuth != nil,
	})
}

// AccountTestResult holds the health check / test connection result for an account.
type AccountTestResult struct {
	ID                string `json:"id"`
	ProjectID         string `json:"projectId"`
	Success           bool   `json:"success"`
	LatencyMs         int64  `json:"latencyMs"`
	Error             string `json:"error,omitempty"`
	NeedsVerification bool   `json:"needsVerification,omitempty"`
	ValidationURL     string `json:"validationUrl,omitempty"`
}

// testAccount runs a live LoadCodeAssist check for an account to verify credentials and connectivity.
func (p *AccountPool) testAccount(ctx context.Context, id string) (*AccountTestResult, error) {
	p.mu.Lock()
	var target *account
	for _, a := range p.accounts {
		if a.id == id {
			target = a
			break
		}
	}
	p.mu.Unlock()

	if target == nil {
		return nil, errors.New("account not found")
	}

	start := time.Now()
	_, err := target.client.LoadCodeAssist()
	latency := time.Since(start).Milliseconds()

	res := &AccountTestResult{
		ID:        target.id,
		ProjectID: target.projectID,
		LatencyMs: latency,
	}

	if err != nil {
		res.Success = false
		res.Error = err.Error()
		var upstreamErr *antigravity.UpstreamError
		if errors.As(err, &upstreamErr) && upstreamErr.IsVerificationRequired() {
			vURL := upstreamErr.ExtractValidationURL()
			p.mu.Lock()
			target.needsVerification = true
			target.validationURL = vURL
			target.coolUntil = p.now().Add(30 * time.Minute)
			target.coolReason = "Account verification required by Google"
			p.mu.Unlock()
			res.NeedsVerification = true
			res.ValidationURL = vURL
			res.Error = "Account verification required by Google"
		}
	} else {
		res.Success = true
		p.mu.Lock()
		target.needsVerification = false
		target.validationURL = ""
		target.coolUntil = time.Time{}
		target.coolReason = ""
		p.mu.Unlock()
	}

	return res, nil
}

// testAllAccounts runs LoadCodeAssist for every account concurrently.
func (p *AccountPool) testAllAccounts(ctx context.Context) []AccountTestResult {
	p.mu.Lock()
	accounts := make([]*account, len(p.accounts))
	copy(accounts, p.accounts)
	p.mu.Unlock()

	var wg sync.WaitGroup
	results := make([]AccountTestResult, len(accounts))
	for i, a := range accounts {
		wg.Add(1)
		go func(idx int, acc *account) {
			defer wg.Done()
			start := time.Now()
			_, err := acc.client.LoadCodeAssist()
			latency := time.Since(start).Milliseconds()
			res := AccountTestResult{
				ID:        acc.id,
				ProjectID: acc.projectID,
				LatencyMs: latency,
			}
			if err != nil {
				res.Success = false
				res.Error = err.Error()
				var upstreamErr *antigravity.UpstreamError
				if errors.As(err, &upstreamErr) && upstreamErr.IsVerificationRequired() {
					vURL := upstreamErr.ExtractValidationURL()
					p.mu.Lock()
					acc.needsVerification = true
					acc.validationURL = vURL
					acc.coolUntil = p.now().Add(30 * time.Minute)
					acc.coolReason = "Account verification required by Google"
					p.mu.Unlock()
					res.NeedsVerification = true
					res.ValidationURL = vURL
					res.Error = "Account verification required by Google"
				}
			} else {
				res.Success = true
				p.mu.Lock()
				acc.needsVerification = false
				acc.validationURL = ""
				acc.coolUntil = time.Time{}
				acc.coolReason = ""
				p.mu.Unlock()
			}
			results[idx] = res
		}(i, a)
	}
	wg.Wait()
	return results
}

func (s *Server) handleTestAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" || id == "all" {
		results := s.accounts.testAllAccounts(r.Context())
		writeAdminJSON(w, map[string]any{"results": results})
		return
	}

	result, err := s.accounts.testAccount(r.Context(), id)
	if err != nil {
		writeAdminError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeAdminJSON(w, map[string]any{"result": result})
}
