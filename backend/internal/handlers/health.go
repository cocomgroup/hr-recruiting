package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"hr-recruiting/internal/gateway"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	client *gateway.HubHRMSClient
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(client *gateway.HubHRMSClient) *HealthHandler {
	return &HealthHandler{client: client}
}

// Health returns the overall health status
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"checks": map[string]interface{}{
			"api": "healthy",
		},
	}

	// Check Hub-HRMS connectivity
	if err := h.client.Health(ctx); err != nil {
		health["checks"].(map[string]interface{})["hubhrms"] = "unhealthy"
		health["checks"].(map[string]interface{})["hubhrms_error"] = err.Error()
		health["status"] = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		health["checks"].(map[string]interface{})["hubhrms"] = "healthy"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// Liveness is a simple liveness probe
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Readiness checks if the service is ready to serve traffic
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Check Hub-HRMS connectivity
	if err := h.client.Health(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "not ready",
			"reason": "Hub-HRMS unreachable",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}