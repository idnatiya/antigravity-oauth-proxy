package antigravity

import (
	"net/http"
	"testing"
)

func TestUpstreamError_VerificationDetection(t *testing.T) {
	// Sample error payload from Google CloudCode when verification challenge is triggered
	rawJSON := `{
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
						"validation_url": "https://accounts.google.com/signin/continue?sarp=1&scc=1&continue=https%3A%2F%2Fconsole.cloud.google.com"
					}
				}
			]
		}
	}`

	err := &UpstreamError{
		StatusCode:  http.StatusForbidden,
		Body:        []byte(rawJSON),
		ContentType: "application/json",
	}

	if !err.IsVerificationRequired() {
		t.Fatal("expected IsVerificationRequired() to be true")
	}

	url := err.ExtractValidationURL()
	expectedURL := "https://accounts.google.com/signin/continue?sarp=1&scc=1&continue=https%3A%2F%2Fconsole.cloud.google.com"
	if url != expectedURL {
		t.Fatalf("ExtractValidationURL() = %q; want %q", url, expectedURL)
	}

	// Non-verification error
	nonVerifErr := &UpstreamError{
		StatusCode:  http.StatusTooManyRequests,
		Body:        []byte(`{"error":{"code":429,"message":"Resource exhausted"}}`),
		ContentType: "application/json",
	}

	if nonVerifErr.IsVerificationRequired() {
		t.Fatal("expected IsVerificationRequired() to be false for 429")
	}
}

func TestUpstreamError_ExtractValidationURL_Fallback(t *testing.T) {
	// 403 without details array but with message
	rawJSON := `{
		"error": {
			"code": 403,
			"message": "Verify your account to continue.",
			"status": "PERMISSION_DENIED"
		}
	}`

	err := &UpstreamError{
		StatusCode: http.StatusForbidden,
		Body:       []byte(rawJSON),
	}

	if !err.IsVerificationRequired() {
		t.Fatal("expected IsVerificationRequired() to be true for message")
	}

	url := err.ExtractValidationURL()
	if url != "https://accounts.google.com/" {
		t.Fatalf("ExtractValidationURL() fallback = %q; want https://accounts.google.com/", url)
	}
}
