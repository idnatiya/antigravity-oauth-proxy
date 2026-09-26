package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/antigravity"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/openai"
)

func TestResolveImageModel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", defaultImageModel},
		{"dall-e-3", defaultImageModel},
		{"dall-e-2", defaultImageModel},
		{"gemini-3.8-flash-high", defaultImageModel},
		{"gemini-2.5-pro", defaultImageModel},
		{"gemini-3.1-flash-image", "gemini-3.1-flash-image"},
		{"imagen-3.0-generate-002", "imagen-3.0-generate-002"},
	}

	for _, tt := range tests {
		actual := resolveImageModel(tt.input)
		if actual != tt.expected {
			t.Errorf("resolveImageModel(%q) = %q; want %q", tt.input, actual, tt.expected)
		}
	}
}

func TestBuildImagePrompt(t *testing.T) {
	tests := []struct {
		prompt   string
		size     string
		style    string
		expected string
	}{
		{
			prompt:   "a red sports car",
			size:     "16:9",
			style:    "vivid",
			expected: "a red sports car (Aspect ratio 16:9 widescreen, Style: vivid)",
		},
		{
			prompt:   "a portrait of a woman",
			size:     "9:16",
			style:    "natural",
			expected: "a portrait of a woman (Aspect ratio 9:16 portrait vertical)",
		},
		{
			prompt:   "logo design",
			size:     "1024x1024",
			style:    "",
			expected: "logo design (Aspect ratio 1:1 square)",
		},
		{
			prompt:   "a futuristic city with aspect ratio 16:9",
			size:     "16:9",
			style:    "",
			expected: "a futuristic city with aspect ratio 16:9", // should not duplicate
		},
	}

	for _, tt := range tests {
		actual := buildImagePrompt(tt.prompt, tt.size, tt.style)
		if actual != tt.expected {
			t.Errorf("buildImagePrompt(%q, %q, %q) = %q; want %q", tt.prompt, tt.size, tt.style, actual, tt.expected)
		}
	}
}

func TestParseDataURIOrBase64(t *testing.T) {
	// Raw base64
	mime, b64, err := parseDataURIOrBase64("aGVsbG8=")
	if err != nil || mime != "image/png" || b64 != "aGVsbG8=" {
		t.Fatalf("unexpected raw base64 parsing: mime=%s, b64=%s, err=%v", mime, b64, err)
	}

	// Data URI png
	mime, b64, err = parseDataURIOrBase64("data:image/png;base64,iVBORw0KGgo=")
	if err != nil || mime != "image/png" || b64 != "iVBORw0KGgo=" {
		t.Fatalf("unexpected data URI parsing: mime=%s, b64=%s, err=%v", mime, b64, err)
	}

	// Data URI jpeg
	mime, b64, err = parseDataURIOrBase64("data:image/jpeg;base64,/9j/4AAQSkZJRg==")
	if err != nil || mime != "image/jpeg" || b64 != "/9j/4AAQSkZJRg==" {
		t.Fatalf("unexpected data URI parsing: mime=%s, b64=%s, err=%v", mime, b64, err)
	}

	// Empty
	_, _, err = parseDataURIOrBase64("")
	if err == nil {
		t.Fatal("expected error for empty image data")
	}
}

func TestExtractImageData(t *testing.T) {
	resp := &antigravity.GenerateContentResponse{
		Response: map[string]interface{}{
			"candidates": []interface{}{
				map[string]interface{}{
					"content": map[string]interface{}{
						"parts": []interface{}{
							map[string]interface{}{
								"inlineData": map[string]interface{}{
									"mimeType": "image/png",
									"data":     "iVBORw0KGgoAAAA==",
								},
							},
							map[string]interface{}{
								"text": "A cute red apple",
							},
						},
					},
				},
			},
		},
	}

	mime, b64, text, err := extractImageData(resp)
	if err != nil {
		t.Fatalf("extractImageData returned error: %v", err)
	}
	if mime != "image/png" {
		t.Errorf("expected mime image/png, got %s", mime)
	}
	if b64 != "iVBORw0KGgoAAAA==" {
		t.Errorf("expected b64 iVBORw0KGgoAAAA==, got %s", b64)
	}
	if text != "A cute red apple" {
		t.Errorf("expected text 'A cute red apple', got %s", text)
	}
}

func TestImagesGenerationsHandler_Validation(t *testing.T) {
	s := &Server{}

	// Missing prompt
	body, _ := json.Marshal(openai.ImageGenerationRequest{
		Prompt: "",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/images/generations", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.openAIImagesGenerationsHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty prompt, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "prompt is required") {
		t.Errorf("expected error message 'prompt is required', got %s", w.Body.String())
	}
}

func TestImagesEditsHandler_Validation(t *testing.T) {
	s := &Server{}

	// Missing image
	body, _ := json.Marshal(openai.ImageEditRequest{
		Prompt: "make it sunny",
		Image:  "",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/images/edits", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.openAIImagesEditsHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing image, got %d", w.Code)
	}
}
