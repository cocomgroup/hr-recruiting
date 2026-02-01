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

// Health returns basic health status - always returns 200 if service is running
// This is used by load balancers and should NOT depend on external services
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "hr-recruiting-api",
	})
}

// Liveness is a simple liveness probe - indicates the service is alive
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Readiness checks if the service is ready to serve traffic
// This DOES check external dependencies like Hub-HRMS
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	health := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().Format(time.RFC3339),
		"checks": map[string]interface{}{
			"api": "healthy",
		},
	}

	// Check Hub-HRMS connectivity
	if err := h.client.Health(ctx); err != nil {
		health["checks"].(map[string]interface{})["hubhrms"] = "unhealthy"
		health["checks"].(map[string]interface{})["hubhrms_error"] = err.Error()
		health["status"] = "not ready"
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(health)
		return
	}

	health["checks"].(map[string]interface{})["hubhrms"] = "healthy"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}
