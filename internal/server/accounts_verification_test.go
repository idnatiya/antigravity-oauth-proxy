package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/antigravity"
)

func TestCooldownFor_VerificationRequired(t *testing.T) {
	err := &antigravity.UpstreamError{
		StatusCode: http.StatusForbidden,
		Body: []byte(`{
			"error": {
				"code": 403,
				"message": "Verify your account to continue.",
				"status": "PERMISSION_DENIED",
				"details": [
					{
						"@type": "type.googleapis.com/google.rpc.ErrorInfo",
						"reason": "VALIDATION_REQUIRED",
						"domain": "googleapis.com",
						"metadata": {
							"validation_url": "https://accounts.google.com/signin/continue?sarp=1"
						}
					}
				]
			}
		}`),
	}

	dur, ok := cooldownFor(err)
	if !ok {
		t.Fatal("expected cooldownFor to return true for verification required error")
	}
	if dur != 30*time.Minute {
		t.Fatalf("expected cooldown of 30m, got %v", dur)
	}
}

func TestUpstreamErrorStatus_VerificationRequired(t *testing.T) {
	err := &antigravity.UpstreamError{
		StatusCode: http.StatusForbidden,
		Body: []byte(`{
			"error": {
				"code": 403,
				"message": "Verify your account to continue.",
				"status": "PERMISSION_DENIED",
				"details": [
					{
						"@type": "type.googleapis.com/google.rpc.ErrorInfo",
						"reason": "VALIDATION_REQUIRED",
						"domain": "googleapis.com",
						"metadata": {
							"validation_url": "https://accounts.google.com/signin/continue?challenge=xyz"
						}
					}
				]
			}
		}`),
	}

	status, msg := upstreamErrorStatus(err, "fallback")
	if status != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", status)
	}
	if !strings.Contains(msg, "https://accounts.google.com/signin/continue?challenge=xyz") {
		t.Fatalf("expected msg to contain validation_url, got %q", msg)
	}
}

func TestAccountView_VerificationJSON(t *testing.T) {
	v := accountView{
		ID:                "test@gmail.com",
		ProjectID:         "test-proj",
		Removable:         true,
		NeedsVerification: true,
		ValidationURL:     "https://accounts.google.com/signin/continue?challenge=123",
	}

	bytes, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed["needsVerification"] != true {
		t.Fatalf("expected needsVerification to be true, got %v", parsed["needsVerification"])
	}
	if parsed["validationUrl"] != "https://accounts.google.com/signin/continue?challenge=123" {
		t.Fatalf("expected validationUrl to match, got %v", parsed["validationUrl"])
	}
}
