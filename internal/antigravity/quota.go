package antigravity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// QuotaSummary is the retrieveUserQuotaSummary response, the same source the
// Antigravity CLI's /usage view reads: model groups (Gemini, Claude+GPT) that
// each share a weekly and a five-hour window.
type QuotaSummary struct {
	Groups      []QuotaGroup `json:"groups"`
	Description string       `json:"description,omitempty"`
}

type QuotaGroup struct {
	DisplayName string        `json:"displayName"`
	Description string        `json:"description,omitempty"`
	Buckets     []QuotaBucket `json:"buckets"`
}

// QuotaBucket is one window of a group. RemainingFraction is a pointer so a
// missing field is not mistaken for a drained window.
type QuotaBucket struct {
	BucketID          string   `json:"bucketId,omitempty"`
	DisplayName       string   `json:"displayName,omitempty"`
	Window            string   `json:"window,omitempty"` // "5h" or "weekly"
	RemainingFraction *float64 `json:"remainingFraction,omitempty"`
	ResetTime         string   `json:"resetTime,omitempty"`
}

func (c *Client) RetrieveUserQuotaSummary(ctx context.Context, project string) (*QuotaSummary, error) {
	body := map[string]string{}
	if project != "" {
		body["project"] = project
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var lastErr error
	for _, endpoint := range Endpoints {
		url := fmt.Sprintf("%s/v1internal:retrieveUserQuotaSummary", endpoint)
		resp, err := c.doRequest(ctx, http.MethodPost, url, bodyBytes, "application/json")
		if err != nil {
			lastErr = err
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("could not read response body: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = &UpstreamError{
				StatusCode:  resp.StatusCode,
				Body:        respBody,
				ContentType: resp.Header.Get("Content-Type"),
				Endpoint:    endpoint,
			}
			continue
		}

		var result QuotaSummary
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("could not unmarshal response body: %w", err)
		}
		return &result, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("retrieveUserQuotaSummary failed with no endpoints available")
}
