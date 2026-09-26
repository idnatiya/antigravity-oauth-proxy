package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/antigravity"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/logger"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/openai"
)

const defaultImageModel = "gemini-3.1-flash-image"

// resolveImageModel resolves or normalizes the model name for image generation/editing.
func resolveImageModel(model string) string {
	m := strings.TrimSpace(model)
	if m == "" {
		return defaultImageModel
	}
	mLower := strings.ToLower(m)
	// Map OpenAI image models (dall-e-2, dall-e-3)
	if strings.Contains(mLower, "dall-e") {
		return defaultImageModel
	}
	// If a user passed a text/chat LLM model like gemini-3.8-flash, gemini-2.5-flash, etc.
	if !strings.Contains(mLower, "-image") && !strings.Contains(mLower, "imagen") {
		return defaultImageModel
	}
	return m
}

// buildImagePrompt appends aspect ratio and style cues to the prompt if provided.
func buildImagePrompt(prompt, size, style string) string {
	p := strings.TrimSpace(prompt)
	var cues []string

	s := strings.ToLower(strings.TrimSpace(size))
	switch s {
	case "16:9", "1792x1024", "1024x576", "1280x720", "1920x1080":
		cues = append(cues, "Aspect ratio 16:9 widescreen")
	case "9:16", "1024x1792", "576x1024", "720x1280", "1080x1920":
		cues = append(cues, "Aspect ratio 9:16 portrait vertical")
	case "4:3", "1024x768", "800x600":
		cues = append(cues, "Aspect ratio 4:3 landscape")
	case "3:4", "768x1024", "600x800":
		cues = append(cues, "Aspect ratio 3:4 portrait")
	case "1:1", "1024x1024", "512x512", "256x256":
		cues = append(cues, "Aspect ratio 1:1 square")
	default:
		if s != "" {
			cues = append(cues, fmt.Sprintf("Aspect ratio %s", s))
		}
	}

	st := strings.ToLower(strings.TrimSpace(style))
	if st != "" && st != "natural" {
		cues = append(cues, fmt.Sprintf("Style: %s", st))
	}

	if len(cues) > 0 {
		pLower := strings.ToLower(p)
		if !strings.Contains(pLower, "aspect ratio") && !strings.Contains(pLower, "ratio:") {
			p = fmt.Sprintf("%s (%s)", p, strings.Join(cues, ", "))
		}
	}

	return p
}

// parseDataURIOrBase64 extracts mimeType and base64 string from a raw base64 or data URI.
func parseDataURIOrBase64(raw string) (mimeType, b64Data string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", errors.New("image data is empty")
	}

	if strings.HasPrefix(raw, "data:") {
		commaIdx := strings.Index(raw, ",")
		if commaIdx == -1 {
			return "", "", errors.New("invalid data URI format: missing comma")
		}
		header := raw[:commaIdx]
		b64Data = raw[commaIdx+1:]

		header = strings.TrimPrefix(header, "data:")
		parts := strings.Split(header, ";")
		if len(parts) > 0 && parts[0] != "" {
			mimeType = strings.TrimSpace(parts[0])
		} else {
			mimeType = "image/png"
		}
	} else {
		b64Data = raw
		mimeType = "image/png"
	}

	return mimeType, b64Data, nil
}

// extractImageData parses GenerateContentResponse looking for inlineData (image bytes) and any text.
func extractImageData(resp *antigravity.GenerateContentResponse) (mimeType, b64Data, text string, err error) {
	if resp == nil || resp.Response == nil {
		return "", "", "", errors.New("empty response from upstream")
	}
	candidates, ok := resp.Response["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "", "", "", errors.New("no candidates returned from upstream")
	}
	first, ok := candidates[0].(map[string]interface{})
	if !ok {
		return "", "", "", errors.New("invalid candidate format")
	}

	var parts []interface{}
	if content, ok := first["content"].(map[string]interface{}); ok {
		if ps, ok := content["parts"].([]interface{}); ok {
			parts = ps
		}
	}
	if len(parts) == 0 {
		if ps, ok := first["parts"].([]interface{}); ok {
			parts = ps
		}
	}

	for _, p := range parts {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		if inline, ok := pm["inlineData"].(map[string]interface{}); ok {
			if m, ok := inline["mimeType"].(string); ok {
				mimeType = m
			}
			if d, ok := inline["data"].(string); ok {
				b64Data = d
			}
		}
		if txt, ok := pm["text"].(string); ok && txt != "" {
			if isThought, ok := pm["thought"].(bool); !ok || !isThought {
				if text != "" {
					text += "\n"
				}
				text += txt
			}
		}
	}

	if b64Data == "" {
		if text != "" {
			return "", "", text, fmt.Errorf("upstream returned text instead of image: %s", text)
		}
		return "", "", "", errors.New("no image data returned in candidate parts")
	}
	if mimeType == "" {
		mimeType = "image/png"
	}
	return mimeType, b64Data, text, nil
}

// openAIImagesGenerationsHandler handles POST /v1/images/generations.
func (s *Server) openAIImagesGenerationsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Error reading image generation request body")
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req openai.ImageGenerationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		logger.Get().Error().Err(err).Msg("Error parsing image generation request body")
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		writeAPIError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	model := resolveImageModel(req.Model)
	finalPrompt := buildImagePrompt(req.Prompt, req.Size, req.Style)

	logger.Get().Info().
		Str("model", model).
		Str("original_prompt", req.Prompt).
		Str("final_prompt", finalPrompt).
		Str("size", req.Size).
		Msg("Generating image with upstream model")

	gemReq := &antigravity.GenerateContentRequest{
		Project: s.projectID,
		Model:   model,
		Request: antigravity.GeminiInternalRequest{
			Contents: []antigravity.Content{
				{
					Role: "user",
					Parts: []antigravity.ContentPart{
						{Text: finalPrompt},
					},
				},
			},
			GenerationConfig: &antigravity.GeminiGenerationConfig{},
		},
	}

	resp, err := s.generate(gemReq)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Upstream image generation failed")
		status, msg := upstreamErrorStatus(err, "Error generating image from upstream")
		s.recordUsage("/v1/images/generations", model, false, status, time.Since(startTime), 0, 0, err.Error())
		writeAPIError(w, status, msg)
		return
	}

	mimeType, b64Data, text, err := extractImageData(resp)
	if err != nil {
		logger.Get().Error().Err(err).Str("text_fallback", text).Msg("Failed to extract image data from upstream response")
		s.recordUsage("/v1/images/generations", model, false, http.StatusInternalServerError, time.Since(startTime), 0, 0, err.Error())
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to extract image from upstream response: %v", err))
		return
	}

	revisedPrompt := req.Prompt
	if text != "" {
		revisedPrompt = text
	}

	imageResp := openai.ImageResponse{
		Created: time.Now().Unix(),
		Data: []openai.ImageData{
			{
				B64JSON:       b64Data,
				URL:           fmt.Sprintf("data:%s;base64,%s", mimeType, b64Data),
				RevisedPrompt: revisedPrompt,
			},
		},
	}

	s.recordUsage("/v1/images/generations", model, false, http.StatusOK, time.Since(startTime), 0, 0, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(imageResp); err != nil {
		logger.Get().Error().Err(err).Msg("Error encoding image response")
	}
}

// openAIImagesEditsHandler handles POST /v1/images/edits.
func (s *Server) openAIImagesEditsHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var prompt string
	var model string
	var size string
	var mimeType string
	var b64Data string

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Parse multipart form (max 32MB)
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			logger.Get().Error().Err(err).Msg("Error parsing multipart form")
			http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusBadRequest)
			return
		}

		prompt = r.FormValue("prompt")
		model = r.FormValue("model")
		size = r.FormValue("size")

		file, header, err := r.FormFile("image")
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "image file is required in multipart form")
			return
		}
		defer file.Close()

		imgBytes, err := io.ReadAll(file)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "failed to read image file: "+err.Error())
			return
		}

		mimeType = header.Header.Get("Content-Type")
		if mimeType == "" || mimeType == "application/octet-stream" {
			mimeType = http.DetectContentType(imgBytes)
		}
		if !strings.HasPrefix(mimeType, "image/") {
			mimeType = "image/png"
		}
		b64Data = base64.StdEncoding.EncodeToString(imgBytes)
	} else {
		// JSON payload
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req openai.ImageEditRequest
		if err := json.Unmarshal(body, &req); err != nil {
			logger.Get().Error().Err(err).Msg("Error parsing image edit JSON body")
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			return
		}

		prompt = req.Prompt
		model = req.Model
		size = req.Size

		var parseErr error
		mimeType, b64Data, parseErr = parseDataURIOrBase64(req.Image)
		if parseErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid image data: "+parseErr.Error())
			return
		}
	}

	if strings.TrimSpace(prompt) == "" {
		writeAPIError(w, http.StatusBadRequest, "prompt is required")
		return
	}
	if b64Data == "" {
		writeAPIError(w, http.StatusBadRequest, "image is required")
		return
	}

	model = resolveImageModel(model)
	finalPrompt := buildImagePrompt(prompt, size, "")

	logger.Get().Info().
		Str("model", model).
		Str("prompt", prompt).
		Str("size", size).
		Str("mime_type", mimeType).
		Int("base64_len", len(b64Data)).
		Msg("Editing image with upstream model")

	gemReq := &antigravity.GenerateContentRequest{
		Project: s.projectID,
		Model:   model,
		Request: antigravity.GeminiInternalRequest{
			Contents: []antigravity.Content{
				{
					Role: "user",
					Parts: []antigravity.ContentPart{
						{
							InlineData: &antigravity.InlineData{
								MimeType: mimeType,
								Data:     b64Data,
							},
						},
						{Text: finalPrompt},
					},
				},
			},
			GenerationConfig: &antigravity.GeminiGenerationConfig{},
		},
	}

	resp, err := s.generate(gemReq)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Upstream image edit failed")
		status, msg := upstreamErrorStatus(err, "Error editing image from upstream")
		s.recordUsage("/v1/images/edits", model, false, status, time.Since(startTime), 0, 0, err.Error())
		writeAPIError(w, status, msg)
		return
	}

	resMimeType, resB64Data, text, err := extractImageData(resp)
	if err != nil {
		logger.Get().Error().Err(err).Str("text_fallback", text).Msg("Failed to extract edited image data from upstream response")
		s.recordUsage("/v1/images/edits", model, false, http.StatusInternalServerError, time.Since(startTime), 0, 0, err.Error())
		writeAPIError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to extract edited image from upstream response: %v", err))
		return
	}

	revisedPrompt := prompt
	if text != "" {
		revisedPrompt = text
	}

	imageResp := openai.ImageResponse{
		Created: time.Now().Unix(),
		Data: []openai.ImageData{
			{
				B64JSON:       resB64Data,
				URL:           fmt.Sprintf("data:%s;base64,%s", resMimeType, resB64Data),
				RevisedPrompt: revisedPrompt,
			},
		},
	}

	s.recordUsage("/v1/images/edits", model, false, http.StatusOK, time.Since(startTime), 0, 0, "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(imageResp); err != nil {
		logger.Get().Error().Err(err).Msg("Error encoding image response")
	}
}
