package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/antigravity"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/logger"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/openai"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/transform"
)

type enhancePromptRequest struct {
	Prompt string `json:"prompt"`
	Style  string `json:"style,omitempty"`
}

type enhancePromptResponse struct {
	EnhancedPrompt string `json:"enhanced_prompt"`
}

// dashboardImageEnhancePromptHandler enhances image prompts using gemini-3.8-flash-high.
func (s *Server) dashboardImageEnhancePromptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req enhancePromptRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		writeAPIError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	sysPrompt := "You are an expert prompt engineer for AI image generators (like Imagen, Midjourney, and Stable Diffusion). " +
		"Expand the user's idea into a single, highly detailed, visually evocative English descriptive prompt covering the main subject, composition, camera angle, atmospheric lighting, color palette, fine textures, mood, and art style. " +
		"Output ONLY the enhanced descriptive prompt text. Do NOT include markdown quotes, explanations, prefixes like 'Prompt:', or any conversational text."

	userContent := req.Prompt
	if req.Style != "" {
		userContent = req.Prompt + " (Desired aesthetic/style: " + req.Style + ")"
	}

	gemReq, err := transform.ToGeminiRequest(&openai.ChatCompletionRequest{
		Model: "gemini-3.8-flash-high",
		Messages: []openai.Message{
			{
				Role:    "system",
				Content: sysPrompt,
			},
			{
				Role:    "user",
				Content: userContent,
			},
		},
	}, s.projectID)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to construct Gemini request for prompt enhancement")
		http.Error(w, "Failed to build request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := s.generate(gemReq)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to enhance prompt with Gemini 3.8")
		status, msg := upstreamErrorStatus(err, "Failed to enhance prompt with Gemini 3.8")
		writeAPIError(w, status, msg)
		return
	}

	enhancedText := extractTextFromGenerateResponse(resp)
	if enhancedText == "" {
		enhancedText = req.Prompt
	}

	// Clean any enclosing quotes or formatting if present
	enhancedText = strings.TrimSpace(enhancedText)
	enhancedText = strings.TrimPrefix(enhancedText, "\"")
	enhancedText = strings.TrimSuffix(enhancedText, "\"")
	enhancedText = strings.TrimPrefix(enhancedText, "`")
	enhancedText = strings.TrimSuffix(enhancedText, "`")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(enhancePromptResponse{
		EnhancedPrompt: enhancedText,
	})
}

func extractTextFromGenerateResponse(resp *antigravity.GenerateContentResponse) string {
	if resp == nil || resp.Response == nil {
		return ""
	}
	cands, ok := resp.Response["candidates"].([]interface{})
	if !ok || len(cands) == 0 {
		return ""
	}
	first, ok := cands[0].(map[string]interface{})
	if !ok {
		return ""
	}
	var parts []interface{}
	if c, ok := first["content"].(map[string]interface{}); ok {
		if ps, ok := c["parts"].([]interface{}); ok {
			parts = ps
		}
	}
	if len(parts) == 0 {
		if ps, ok := first["parts"].([]interface{}); ok {
			parts = ps
		}
	}

	var sb strings.Builder
	for _, p := range parts {
		if pm, ok := p.(map[string]interface{}); ok {
			if isThought, ok := pm["thought"].(bool); ok && isThought {
				continue
			}
			if txt, ok := pm["text"].(string); ok && txt != "" {
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(txt)
			}
		}
	}
	return sb.String()
}
