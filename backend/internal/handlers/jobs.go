package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"hr-recruiting/internal/gateway"
)

// JobHandler handles job-related requests
type JobHandler struct {
	client *gateway.HubHRMSClient
}

// NewJobHandler creates a new job handler
func NewJobHandler(client *gateway.HubHRMSClient) *JobHandler {
	return &JobHandler{client: client}
}

// ListJobs returns a list of jobs
func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := r.URL.Query().Get("q")
	department := r.URL.Query().Get("department")
	location := r.URL.Query().Get("location")
	employmentType := r.URL.Query().Get("employmentType")
	experienceLevel := r.URL.Query().Get("experienceLevel")
	remoteStr := r.URL.Query().Get("remote")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Build filters
	filters := make(map[string]interface{})
	if query != "" {
		filters["query"] = query
	}
	if department != "" {
		filters["departments"] = []string{department}
	}
	if location != "" {
		filters["locations"] = []string{location}
	}
	if employmentType != "" {
		filters["employmentTypes"] = []string{employmentType}
	}
	if experienceLevel != "" {
		filters["experienceLevels"] = []string{experienceLevel}
	}
	if remoteStr != "" {
		remote, _ := strconv.ParseBool(remoteStr)
		filters["remoteWork"] = remote
	}
	if status != "" {
		filters["status"] = status
	} else {
		// Default to published jobs for public API
		filters["status"] = "PUBLISHED"
	}

	// Parse pagination
	limit := 20
	offset := 0
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build variables
	variables := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}
	if len(filters) > 0 {
		variables["filters"] = filters
	}

	// Execute query
	resp, err := h.client.Query(ctx, gateway.GetJobsQuery, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch jobs", err)
		return
	}

	// Add total count header if available
	w.Header().Set("X-Total-Count", strconv.Itoa(len(resp.Data.(map[string]interface{})["jobs"].([]interface{}))))

	respondJSON(w, http.StatusOK, resp.Data)
}

// GetJob returns a single job by ID
func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := chi.URLParam(r, "id")
	
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required", nil)
		return
	}

	variables := map[string]interface{}{
		"id": jobID,
	}

	resp, err := h.client.Query(ctx, gateway.GetJobQuery, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch job", err)
		return
	}

	if resp.Data == nil {
		respondError(w, http.StatusNotFound, "Job not found", nil)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// CreateJob creates a new job posting
func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	defer r.Body.Close()

	// Validate required fields
	requiredFields := []string{"title", "department", "location", "employmentType", "experienceLevel", "description", "requirements", "skills"}
	for _, field := range requiredFields {
		if _, ok := input[field]; !ok {
			respondError(w, http.StatusBadRequest, "Missing required field: "+field, nil)
			return
		}
	}

	variables := map[string]interface{}{
		"input": input,
	}

	resp, err := h.client.Mutate(ctx, gateway.CreateJobMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create job", err)
		return
	}

	respondJSON(w, http.StatusCreated, resp.Data)
}

// UpdateJob updates an existing job
func (h *JobHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := chi.URLParam(r, "id")
	
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required", nil)
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	defer r.Body.Close()

	variables := map[string]interface{}{
		"id":    jobID,
		"input": input,
	}

	resp, err := h.client.Mutate(ctx, gateway.UpdateJobMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update job", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// PublishJob publishes a job posting
func (h *JobHandler) PublishJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := chi.URLParam(r, "id")
	
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required", nil)
		return
	}

	variables := map[string]interface{}{
		"id": jobID,
	}

	resp, err := h.client.Mutate(ctx, gateway.PublishJobMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to publish job", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// CloseJob closes a job posting
func (h *JobHandler) CloseJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := chi.URLParam(r, "id")
	
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required", nil)
		return
	}

	variables := map[string]interface{}{
		"id": jobID,
	}

	resp, err := h.client.Mutate(ctx, gateway.CloseJobMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to close job", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// DeleteJob deletes a job posting
func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := chi.URLParam(r, "id")
	
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required", nil)
		return
	}

	variables := map[string]interface{}{
		"id": jobID,
	}

	resp, err := h.client.Mutate(ctx, gateway.DeleteJobMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete job", err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Job deleted successfully",
		"data":    resp.Data,
	})
}

// IncrementView increments the view count for a job
func (h *JobHandler) IncrementView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	jobID := chi.URLParam(r, "id")
	
	if jobID == "" {
		respondError(w, http.StatusBadRequest, "Job ID is required", nil)
		return
	}

	variables := map[string]interface{}{
		"id": jobID,
	}

	resp, err := h.client.Mutate(ctx, gateway.IncrementJobViewMutation, variables)
	if err != nil {
		// Don't fail the request if view increment fails
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"message": "View count update failed",
		})
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// GenerateDescription generates a job description using AI
func (h *JobHandler) GenerateDescription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	defer r.Body.Close()

	// Validate required fields
	requiredFields := []string{"title", "department", "experienceLevel", "keySkills"}
	for _, field := range requiredFields {
		if _, ok := input[field]; !ok {
			respondError(w, http.StatusBadRequest, "Missing required field: "+field, nil)
			return
		}
	}

	variables := map[string]interface{}{
		"input": input,
	}

	resp, err := h.client.Mutate(ctx, gateway.GenerateJobDescriptionMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate job description", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}