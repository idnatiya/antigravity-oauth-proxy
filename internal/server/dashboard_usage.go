package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/logger"
	"github.com/dvcrn/antigravity-oauth-proxy/internal/usage"
)

// handleUsageStats handles GET /api/usage/stats.
func (s *Server) handleUsageStats(w http.ResponseWriter, r *http.Request) {
	if s.usageStore == nil {
		http.Error(w, `{"error":"Usage store not configured"}`, http.StatusInternalServerError)
		return
	}

	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "24h"
	}

	stats, err := s.usageStore.GetStats(r.Context(), timeRange)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to get usage stats")
		http.Error(w, `{"error":"Failed to get usage stats"}`, http.StatusInternalServerError)
		return
	}

	ready, total := s.accounts.ready()
	accountInfo := map[string]interface{}{
		"provider":       s.provider.Name(),
		"accounts_ready": ready,
		"accounts_total": total,
	}
	if accounts := s.accounts.candidates(); len(accounts) > 0 {
		accountInfo["project_id"] = accounts[0].projectID
	}

	if modelsResp, err := s.fetchModels(r.Context()); err == nil && modelsResp != nil {
		quotas := make(map[string]interface{})
		for mID, mData := range modelsResp.Models {
			if len(mData.QuotaInfo) > 0 {
				var q interface{}
				if json.Unmarshal(mData.QuotaInfo, &q) == nil {
					quotas[mID] = q
				}
			}
		}
		if len(quotas) > 0 {
			accountInfo["quotas"] = quotas
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"stats":   stats,
		"account": accountInfo,
		"range":   timeRange,
	})
}

// handleUsageRequests handles GET /api/usage/requests.
func (s *Server) handleUsageRequests(w http.ResponseWriter, r *http.Request) {
	if s.usageStore == nil {
		http.Error(w, `{"error":"Usage store not configured"}`, http.StatusInternalServerError)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	model := strings.TrimSpace(r.URL.Query().Get("model"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	status := 0
	if st := r.URL.Query().Get("status"); st != "" {
		if parsed, err := strconv.Atoi(st); err == nil {
			status = parsed
		}
	}

	filter := usage.RequestFilter{
		Limit:  limit,
		Offset: offset,
		Model:  model,
		Status: status,
		Search: search,
	}

	records, total, err := s.usageStore.GetRequests(r.Context(), filter)
	if err != nil {
		logger.Get().Error().Err(err).Msg("Failed to get requests history")
		http.Error(w, `{"error":"Failed to get requests history"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": records,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}
