package antigravity

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

func TestPrepareAntigravityRequestMatchesCLIShape(t *testing.T) {
	req := &GenerateContentRequest{
		Project: "test-project",
		Model:   "gemini-pro-agent",
		Request: GeminiInternalRequest{
			Contents: []Content{{
				Role:  "user",
				Parts: []ContentPart{{Text: "hello"}},
			}},
			SystemInstruction: &SystemInstruction{
				Role:  "user",
				Parts: []ContentPart{{Text: "client system"}},
			},
		},
	}

	prepareAntigravityRequest(req)

	if req.UserAgent != RequestUserAgent {
		t.Fatalf("UserAgent = %q, want %q", req.UserAgent, RequestUserAgent)
	}
	if req.RequestType != RequestTypeAgent {
		t.Fatalf("RequestType = %q, want %q", req.RequestType, RequestTypeAgent)
	}

	requestIDPattern := regexp.MustCompile(`^agent/[0-9a-f-]{36}/[0-9]{13}/[0-9a-f-]{36}/1$`)
	if !requestIDPattern.MatchString(req.RequestID) {
		t.Fatalf("RequestID = %q, want Antigravity CLI shape", req.RequestID)
	}

	if req.Request.SystemInstruction == nil {
		t.Fatal("SystemInstruction is nil")
	}
	if req.Request.SystemInstruction.Role != "user" {
		t.Fatalf("SystemInstruction.Role = %q, want user", req.Request.SystemInstruction.Role)
	}
	parts := req.Request.SystemInstruction.Parts
	if len(parts) != 2 {
		t.Fatalf("SystemInstruction parts = %d, want 2", len(parts))
	}
	if !strings.Contains(parts[0].Text, "<identity>") || !strings.Contains(parts[0].Text, "You are Antigravity") {
		t.Fatalf("first system part does not contain Antigravity identity")
	}
	if strings.Contains(parts[0].Text, "Please ignore the following [ignore]") {
		t.Fatalf("first system part contains legacy ignore injection")
	}
	if parts[1].Text != "client system" {
		t.Fatalf("existing system part = %q, want client system", parts[1].Text)
	}
}

func TestPrepareAntigravityRequestDefaultsThinkingConfig(t *testing.T) {
	req := &GenerateContentRequest{
		Model: "gemini-3.1-pro-low",
		Request: GeminiInternalRequest{
			Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
		},
	}

	prepareAntigravityRequest(req)

	if req.Request.GenerationConfig == nil || req.Request.GenerationConfig.ThinkingConfig == nil {
		t.Fatal("ThinkingConfig was not defaulted")
	}
	thinkingConfig := req.Request.GenerationConfig.ThinkingConfig
	if thinkingConfig.IncludeThoughts == nil || !*thinkingConfig.IncludeThoughts {
		t.Fatalf("IncludeThoughts = %v, want true", thinkingConfig.IncludeThoughts)
	}
	if thinkingConfig.ThinkingBudget == nil || *thinkingConfig.ThinkingBudget != 1024 {
		t.Fatalf("ThinkingBudget = %v, want 1024", *thinkingConfig.ThinkingBudget)
	}

	// Test that high models default to 10001
	highReq := &GenerateContentRequest{
		Model: "gemini-3.1-pro-high",
		Request: GeminiInternalRequest{
			Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
		},
	}
	prepareAntigravityRequest(highReq)
	if highReq.Request.GenerationConfig.ThinkingConfig.ThinkingBudget == nil || *highReq.Request.GenerationConfig.ThinkingConfig.ThinkingBudget != 10001 {
		t.Fatalf("ThinkingBudget = %v, want 10001 for high model", *highReq.Request.GenerationConfig.ThinkingConfig.ThinkingBudget)
	}
}

func TestPrepareAntigravityRequestClearsThinkingLevelForEncodedModel(t *testing.T) {
	testCases := []struct {
		model      string
		wantBudget int
	}{
		{model: "gemini-3.1-pro-high", wantBudget: 10001},
		{model: "gemini-3.6-flash-high", wantBudget: 10001},
		{model: "gemini-3.6-flash-medium", wantBudget: 4096},
		{model: "gemini-3.6-flash-low", wantBudget: 1024},
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			includeThoughts := false
			req := &GenerateContentRequest{
				Model: tc.model,
				Request: GeminiInternalRequest{
					Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
					GenerationConfig: &GeminiGenerationConfig{
						ThinkingConfig: &ThinkingConfig{
							IncludeThoughts: &includeThoughts,
							ThinkingLevel:   "LOW",
						},
					},
				},
			}

			prepareAntigravityRequest(req)

			thinkingConfig := req.Request.GenerationConfig.ThinkingConfig
			if thinkingConfig.ThinkingLevel != "" {
				t.Fatalf("ThinkingLevel = %q, want empty", thinkingConfig.ThinkingLevel)
			}
			if thinkingConfig.IncludeThoughts == nil || *thinkingConfig.IncludeThoughts {
				t.Fatalf("IncludeThoughts = %v, want false", thinkingConfig.IncludeThoughts)
			}
			if thinkingConfig.ThinkingBudget == nil || *thinkingConfig.ThinkingBudget != tc.wantBudget {
				t.Fatalf("ThinkingBudget = %v, want %d", *thinkingConfig.ThinkingBudget, tc.wantBudget)
			}
		})
	}
}

func TestPrepareAntigravityRequestStripsThinkingConfigForGptOss(t *testing.T) {
	req := &GenerateContentRequest{
		Model: "gpt-oss-120b-medium",
		Request: GeminiInternalRequest{
			Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
		},
	}

	prepareAntigravityRequest(req)

	if req.Request.GenerationConfig != nil && req.Request.GenerationConfig.ThinkingConfig != nil {
		t.Fatalf("ThinkingConfig should be nil for gpt-oss models")
	}
}

func TestPrepareAntigravityRequestAppliesGemini37ThinkingPreset(t *testing.T) {
	req := &GenerateContentRequest{
		Model: "gemini-3.7-flash-high",
		Request: GeminiInternalRequest{
			Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
		},
	}

	prepareAntigravityRequest(req)

	thinkingConfig := req.Request.GenerationConfig.ThinkingConfig
	if thinkingConfig.ThinkingLevel != "high" {
		t.Fatalf("ThinkingLevel = %q, want %q", thinkingConfig.ThinkingLevel, "high")
	}
}

func TestPrepareAntigravityRequestAppliesGemini38ThinkingPreset(t *testing.T) {
	testCases := []struct {
		name          string
		model         string
		incomingLevel string
		wantLevel     string
	}{
		{
			name:      "defaults from suffix",
			model:     "gemini-3.8-flash-high",
			wantLevel: "high",
		},
		{
			name:          "suffix overrides incoming level",
			model:         "gemini-3.8-flash-medium",
			incomingLevel: "MINIMAL",
			wantLevel:     "medium",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &GenerateContentRequest{
				Model: tc.model,
				Request: GeminiInternalRequest{
					Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
				},
			}
			if tc.incomingLevel != "" {
				req.Request.GenerationConfig = &GeminiGenerationConfig{
					ThinkingConfig: &ThinkingConfig{ThinkingLevel: tc.incomingLevel},
				}
			}

			prepareAntigravityRequest(req)

			thinkingConfig := req.Request.GenerationConfig.ThinkingConfig
			if thinkingConfig.ThinkingLevel != tc.wantLevel {
				t.Fatalf("ThinkingLevel = %q, want %q", thinkingConfig.ThinkingLevel, tc.wantLevel)
			}
		})
	}
}

func TestPrepareAntigravityRequestPreservesThinkingConfig(t *testing.T) {
	includeThoughts := false
	thinkingBudget := 123
	req := &GenerateContentRequest{
		Model: "gemini-3.1-pro-low",
		Request: GeminiInternalRequest{
			Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
			GenerationConfig: &GeminiGenerationConfig{
				ThinkingConfig: &ThinkingConfig{
					IncludeThoughts: &includeThoughts,
					ThinkingBudget:  &thinkingBudget,
				},
			},
		},
	}

	prepareAntigravityRequest(req)

	thinkingConfig := req.Request.GenerationConfig.ThinkingConfig
	if thinkingConfig.IncludeThoughts == nil || *thinkingConfig.IncludeThoughts {
		t.Fatalf("IncludeThoughts = %v, want false", thinkingConfig.IncludeThoughts)
	}
	if thinkingConfig.ThinkingBudget == nil || *thinkingConfig.ThinkingBudget != 123 {
		t.Fatalf("ThinkingBudget = %v, want 123", thinkingConfig.ThinkingBudget)
	}
}

func TestPrepareAntigravityRequestThinkingDisabled(t *testing.T) {
	t.Run("zero budget preserves zero budget and disables thoughts", func(t *testing.T) {
		zero := 0
		req := &GenerateContentRequest{
			Model: "gemini-3.8-flash-tiered",
			Request: GeminiInternalRequest{
				Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
				GenerationConfig: &GeminiGenerationConfig{
					ThinkingConfig: &ThinkingConfig{
						ThinkingBudget: &zero,
					},
				},
			},
		}

		prepareAntigravityRequest(req)

		tc := req.Request.GenerationConfig.ThinkingConfig
		if tc.ThinkingBudget == nil || *tc.ThinkingBudget != 0 {
			t.Fatalf("ThinkingBudget = %v, want 0", tc.ThinkingBudget)
		}
		if tc.IncludeThoughts == nil || *tc.IncludeThoughts {
			t.Fatalf("IncludeThoughts = %v, want false", tc.IncludeThoughts)
		}
	})

	t.Run("none thinkingLevel converts to zero budget on 3.8 flash", func(t *testing.T) {
		req := &GenerateContentRequest{
			Model: "gemini-3.8-flash-tiered",
			Request: GeminiInternalRequest{
				Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
				GenerationConfig: &GeminiGenerationConfig{
					ThinkingConfig: &ThinkingConfig{
						ThinkingLevel: "none",
					},
				},
			},
		}

		prepareAntigravityRequest(req)

		tc := req.Request.GenerationConfig.ThinkingConfig
		if tc.ThinkingBudget == nil || *tc.ThinkingBudget != 0 {
			t.Fatalf("ThinkingBudget = %v, want 0", tc.ThinkingBudget)
		}
		if tc.IncludeThoughts == nil || *tc.IncludeThoughts {
			t.Fatalf("IncludeThoughts = %v, want false", tc.IncludeThoughts)
		}
		if tc.ThinkingLevel != "" {
			t.Fatalf("ThinkingLevel = %q, want empty", tc.ThinkingLevel)
		}
	})

	t.Run("off thinkingLevel converts to UNSPECIFIED on flash-lite", func(t *testing.T) {
		req := &GenerateContentRequest{
			Model: "gemini-3.5-flash-lite",
			Request: GeminiInternalRequest{
				Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
				GenerationConfig: &GeminiGenerationConfig{
					ThinkingConfig: &ThinkingConfig{
						ThinkingLevel: "off",
					},
				},
			},
		}

		prepareAntigravityRequest(req)

		tc := req.Request.GenerationConfig.ThinkingConfig
		if tc.ThinkingLevel != "THINKING_LEVEL_UNSPECIFIED" {
			t.Fatalf("ThinkingLevel = %q, want THINKING_LEVEL_UNSPECIFIED", tc.ThinkingLevel)
		}
		if tc.ThinkingBudget != nil {
			t.Fatalf("ThinkingBudget = %v, want nil", tc.ThinkingBudget)
		}
		if tc.IncludeThoughts == nil || *tc.IncludeThoughts {
			t.Fatalf("IncludeThoughts = %v, want false", tc.IncludeThoughts)
		}
	})
}

func TestPrepareAntigravityRequestClampsMaxOutputTokens(t *testing.T) {
	testCases := []struct {
		model      string
		maxTokens  int
		wantTokens int
	}{
		{"gemini-3.1-flash-lite", 65536, 32768},
		{"gemini-3.1-flash-lite", 100000, 32768},
		{"gemini-3.1-flash-lite", 4096, 4096},
		{"gemini-3.1-pro-low", 65536, 32768},
		{"gemini-2.5-flash", 65536, 32768},
		{"gemini-3-flash-agent", 65536, 65536},
		{"gemini-3.7-flash-high", 65536, 65536},
		{"gemini-3.8-flash-high", 65536, 65536},
		{"gemini-3-flash", 65536, 65536},
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			req := &GenerateContentRequest{
				Model: tc.model,
				Request: GeminiInternalRequest{
					Contents: []Content{{Role: "user", Parts: []ContentPart{{Text: "hello"}}}},
					GenerationConfig: &GeminiGenerationConfig{
						MaxOutputTokens: tc.maxTokens,
					},
				},
			}

			prepareAntigravityRequest(req)

			if req.Request.GenerationConfig.MaxOutputTokens != tc.wantTokens {
				t.Fatalf("MaxOutputTokens = %d, want %d", req.Request.GenerationConfig.MaxOutputTokens, tc.wantTokens)
			}
		})
	}
}

func TestLoadCodeAssistResponseParsesPaidTier(t *testing.T) {
	body := []byte(`{
		"currentTier":{"id":"free-tier","name":"Antigravity","description":"Gemini-powered code suggestions"},
		"allowedTiers":[{"id":"free-tier","name":"Antigravity","isDefault":true}],
		"cloudaicompanionProject":"aesthetic-container-3v00q",
		"gcpManaged":false,
		"upgradeSubscriptionUri":"https://codeassist.google.com/upgrade",
		"paidTier":{
			"id":"g1-pro-tier",
			"name":"Google AI Pro",
			"description":"Google AI Pro",
			"upgradeSubscriptionUri":"https://antigravity.google/g1-upgrade",
			"upgradeSubscriptionText":"You can upgrade to a Google AI Ultra plan.",
			"availableCredits":[{"creditType":"GOOGLE_ONE_AI","creditAmount":"1000","minimumCreditAmountForUsage":"50"}]
		}
	}`)

	var resp LoadCodeAssistResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.UpgradeSubscriptionURI != "https://codeassist.google.com/upgrade" {
		t.Fatalf("UpgradeSubscriptionURI = %q", resp.UpgradeSubscriptionURI)
	}
	if resp.PaidTier == nil {
		t.Fatal("PaidTier is nil")
	}
	if resp.PaidTier.ID != "g1-pro-tier" {
		t.Fatalf("PaidTier.ID = %q", resp.PaidTier.ID)
	}
	if len(resp.PaidTier.AvailableCredits) != 1 || resp.PaidTier.AvailableCredits[0].CreditAmount != "1000" {
		t.Fatalf("AvailableCredits = %#v", resp.PaidTier.AvailableCredits)
	}
}

func TestApplyHeadersMatchesAntigravityCLI(t *testing.T) {
	header := http.Header{}
	ApplyHeaders(header, "token", "application/json")

	if got := header.Get("Authorization"); got != "Bearer token" {
		t.Fatalf("Authorization = %q", got)
	}
	if got := header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := header.Get("User-Agent"); !strings.HasPrefix(got, "antigravity/cli/1.1.13 (aidev_client;") {
		t.Fatalf("User-Agent = %q", got)
	}
	if got := header.Get("X-Goog-Api-Client"); got != "" {
		t.Fatalf("X-Goog-Api-Client = %q, want empty", got)
	}
	if got := header.Get("Client-Metadata"); got != "" {
		t.Fatalf("Client-Metadata = %q, want empty", got)
	}
	if got := header.Get("Accept"); got != "" {
		t.Fatalf("Accept = %q, want empty", got)
	}
}

func TestEnsureThoughtSignatures(t *testing.T) {
	t.Run("defaults missing thought signature to bypass sentinel", func(t *testing.T) {
		req := &GenerateContentRequest{
			Model: "gemini-3.8-flash-low",
			Request: GeminiInternalRequest{
				Contents: []Content{
					{
						Role: "user",
						Parts: []ContentPart{
							{Text: "cara hubungin ke telegram gmna?"},
						},
					},
					{
						Role: "model",
						Parts: []ContentPart{
							{
								FunctionCall: &FunctionCall{
									ID:   "call_44461383-79cc-4e9d-b7ad-6db8691e95b1",
									Name: "skill_view",
									Args: map[string]interface{}{"name": "hermes-agent"},
								},
							},
						},
					},
					{
						Role: "user",
						Parts: []ContentPart{
							{
								FunctionResponse: &FunctionResponse{
									ID:       "call_44461383-79cc-4e9d-b7ad-6db8691e95b1",
									Name:     "skill_view",
									Response: map[string]interface{}{"output": "ok"},
								},
							},
						},
					},
				},
			},
		}

		prepareAntigravityRequest(req)

		modelPart := req.Request.Contents[1].Parts[0]
		if modelPart.ThoughtSignature != DefaultBypassThoughtSignature {
			t.Fatalf("ThoughtSignature = %q, want %q", modelPart.ThoughtSignature, DefaultBypassThoughtSignature)
		}
		if modelPart.ThoughtSignatureSnake != DefaultBypassThoughtSignature {
			t.Fatalf("ThoughtSignatureSnake = %q, want %q", modelPart.ThoughtSignatureSnake, DefaultBypassThoughtSignature)
		}
		if modelPart.FunctionCall.ThoughtSignature != DefaultBypassThoughtSignature {
			t.Fatalf("FunctionCall.ThoughtSignature = %q, want %q", modelPart.FunctionCall.ThoughtSignature, DefaultBypassThoughtSignature)
		}
		if modelPart.FunctionCall.ThoughtSignatureSnake != DefaultBypassThoughtSignature {
			t.Fatalf("FunctionCall.ThoughtSignatureSnake = %q, want %q", modelPart.FunctionCall.ThoughtSignatureSnake, DefaultBypassThoughtSignature)
		}

		// Verify serialized JSON contains both thoughtSignature and thought_signature
		jsonBytes, err := json.Marshal(req.Request.Contents[1])
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		jsonStr := string(jsonBytes)
		if !strings.Contains(jsonStr, `"thoughtSignature":"skip_thought_signature_validator"`) {
			t.Fatalf("JSON missing camelCase thoughtSignature: %s", jsonStr)
		}
		if !strings.Contains(jsonStr, `"thought_signature":"skip_thought_signature_validator"`) {
			t.Fatalf("JSON missing snake_case thought_signature: %s", jsonStr)
		}
	})

	t.Run("preserves existing thought signature", func(t *testing.T) {
		existingSig := "valid_base64_signature_xyz"
		req := &GenerateContentRequest{
			Model: "gemini-3.8-flash-low",
			Request: GeminiInternalRequest{
				Contents: []Content{
					{
						Role: "model",
						Parts: []ContentPart{
							{
								ThoughtSignature: existingSig,
								FunctionCall: &FunctionCall{
									ID:   "call_123",
									Name: "read_file",
								},
							},
						},
					},
				},
			},
		}

		prepareAntigravityRequest(req)

		modelPart := req.Request.Contents[0].Parts[0]
		if modelPart.ThoughtSignature != existingSig {
			t.Fatalf("ThoughtSignature = %q, want %q", modelPart.ThoughtSignature, existingSig)
		}
		if modelPart.ThoughtSignatureSnake != existingSig {
			t.Fatalf("ThoughtSignatureSnake = %q, want %q", modelPart.ThoughtSignatureSnake, existingSig)
		}
		if modelPart.FunctionCall.ThoughtSignature != existingSig {
			t.Fatalf("FunctionCall.ThoughtSignature = %q, want %q", modelPart.FunctionCall.ThoughtSignature, existingSig)
		}
	})
}
