package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"hr-recruiting/internal/gateway"
)

// AnalyticsHandler handles analytics-related requests
type AnalyticsHandler struct {
	client *gateway.HubHRMSClient
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(client *gateway.HubHRMSClient) *AnalyticsHandler {
	return &AnalyticsHandler{client: client}
}

// GetMetrics returns recruitment metrics
func (h *AnalyticsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse date range from query params
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	// Default to last 30 days if not provided
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	variables := map[string]interface{}{
		"dateRange": map[string]string{
			"start": startDate.Format(time.RFC3339),
			"end":   endDate.Format(time.RFC3339),
		},
	}

	resp, err := h.client.Query(ctx, gateway.GetRecruitmentMetricsQuery, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch metrics", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// GetJobPerformance returns performance metrics for a specific job
func (h *AnalyticsHandler) GetJobPerformance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := chi.URLParam(r, "id")
	
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required", nil)
		return
	}

	variables := map[string]interface{}{
		"jobId": jobID,
	}

	resp, err := h.client.Query(ctx, gateway.GetJobPerformanceQuery, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch job performance", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// GetPipeline returns the application pipeline
func (h *AnalyticsHandler) GetPipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	jobID := r.URL.Query().Get("jobId")

	variables := make(map[string]interface{})
	if jobID != "" {
		variables["jobId"] = jobID
	}

	resp, err := h.client.Query(ctx, gateway.GetApplicationPipelineQuery, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch pipeline", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// GetTrends returns application trends over time
func (h *AnalyticsHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse date range
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	endDate := time.Now()
	startDate := endDate.AddDate(0, -3, 0) // Default to last 3 months

	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	variables := map[string]interface{}{
		"dateRange": map[string]string{
			"start": startDate.Format(time.RFC3339),
			"end":   endDate.Format(time.RFC3339),
		},
	}

	resp, err := h.client.Query(ctx, gateway.GetRecruitmentMetricsQuery, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch trends", err)
		return
	}

	// Extract just the trend data
	data := resp.Data.(map[string]interface{})
	if metrics, ok := data["recruitmentMetrics"].(map[string]interface{}); ok {
		if trends, ok := metrics["applicationTrend"]; ok {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"applicationTrend": trends,
			})
			return
		}
	}

	respondJSON(w, http.StatusOK, resp.Data)
}