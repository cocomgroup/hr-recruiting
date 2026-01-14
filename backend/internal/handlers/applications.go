package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"hr-recruiting/internal/gateway"
	"hr-recruiting/internal/services"
)

// ApplicationHandler handles application-related requests
type ApplicationHandler struct {
	client        *gateway.HubHRMSClient
	uploadService *services.UploadService
	emailService  *services.EmailService
}

// NewApplicationHandler creates a new application handler
func NewApplicationHandler(
	client *gateway.HubHRMSClient,
	uploadService *services.UploadService,
	emailService *services.EmailService,
) *ApplicationHandler {
	return &ApplicationHandler{
		client:        client,
		uploadService: uploadService,
		emailService:  emailService,
	}
}

// SubmitApplication handles job application submission
func (h *ApplicationHandler) SubmitApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	defer r.Body.Close()

	// Validate required fields
	requiredFields := []string{"jobId", "firstName", "lastName", "email", "phone", "resumeUrl", "currentLocation", "availability"}
	for _, field := range requiredFields {
		if _, ok := input[field]; !ok {
			respondError(w, http.StatusBadRequest, "Missing required field: "+field, nil)
			return
		}
	}

	// Set default values
	if _, ok := input["willingToRelocate"]; !ok {
		input["willingToRelocate"] = false
	}

	variables := map[string]interface{}{
		"input": input,
	}

	resp, err := h.client.Mutate(ctx, gateway.SubmitApplicationMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to submit application", err)
		return
	}

	// Send confirmation email asynchronously
	go h.emailService.SendApplicationConfirmation(
		input["email"].(string),
		input["firstName"].(string),
		input["jobId"].(string),
	)

	respondJSON(w, http.StatusCreated, resp.Data)
}

// ListApplications returns a list of applications
func (h *ApplicationHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	jobID := r.URL.Query().Get("jobId")
	status := r.URL.Query().Get("status")
	dateFrom := r.URL.Query().Get("dateFrom")
	dateTo := r.URL.Query().Get("dateTo")
	minScoreStr := r.URL.Query().Get("minScore")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Build filters
	filters := make(map[string]interface{})
	if jobID != "" {
		filters["jobId"] = jobID
	}
	if status != "" {
		filters["status"] = status
	}
	if dateFrom != "" {
		filters["dateFrom"] = dateFrom
	}
	if dateTo != "" {
		filters["dateTo"] = dateTo
	}
	if minScoreStr != "" {
		if minScore, err := strconv.ParseFloat(minScoreStr, 64); err == nil {
			filters["minScore"] = minScore
		}
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

	variables := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}
	if len(filters) > 0 {
		variables["filters"] = filters
	}

	resp, err := h.client.Query(ctx, gateway.GetApplicationsQuery, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch applications", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// GetApplication returns a single application by ID
func (h *ApplicationHandler) GetApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	appID := chi.URLParam(r, "id")
	
	if appID == "" {
		respondError(w, http.StatusBadRequest, "Application ID is required", nil)
		return
	}

	variables := map[string]interface{}{
		"id": appID,
	}

	resp, err := h.client.Query(ctx, gateway.GetApplicationQuery, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch application", err)
		return
	}

	if resp.Data == nil {
		respondError(w, http.StatusNotFound, "Application not found", nil)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// UpdateStatus updates an application's status
func (h *ApplicationHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	appID := chi.URLParam(r, "id")
	
	if appID == "" {
		respondError(w, http.StatusBadRequest, "Application ID is required", nil)
		return
	}

	var input struct {
		Status string `json:"status"`
		Note   string `json:"note,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	defer r.Body.Close()

	if input.Status == "" {
		respondError(w, http.StatusBadRequest, "Status is required", nil)
		return
	}

	variables := map[string]interface{}{
		"id":     appID,
		"status": input.Status,
	}
	if input.Note != "" {
		variables["note"] = input.Note
	}

	resp, err := h.client.Mutate(ctx, gateway.UpdateApplicationStatusMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update application status", err)
		return
	}

	// Send status update email asynchronously
	go h.emailService.SendStatusUpdate(appID, input.Status)

	respondJSON(w, http.StatusOK, resp.Data)
}

// BulkUpdateStatus updates multiple applications' status
func (h *ApplicationHandler) BulkUpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input struct {
		IDs    []string `json:"ids"`
		Status string   `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	defer r.Body.Close()

	if len(input.IDs) == 0 {
		respondError(w, http.StatusBadRequest, "Application IDs are required", nil)
		return
	}
	if input.Status == "" {
		respondError(w, http.StatusBadRequest, "Status is required", nil)
		return
	}

	variables := map[string]interface{}{
		"ids":    input.IDs,
		"status": input.Status,
	}

	resp, err := h.client.Mutate(ctx, gateway.BulkUpdateApplicationStatusMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update application statuses", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// AddNote adds a note to an application
func (h *ApplicationHandler) AddNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	appID := chi.URLParam(r, "id")
	
	if appID == "" {
		respondError(w, http.StatusBadRequest, "Application ID is required", nil)
		return
	}

	var input struct {
		Content    string `json:"content"`
		IsInternal bool   `json:"isInternal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	defer r.Body.Close()

	if input.Content == "" {
		respondError(w, http.StatusBadRequest, "Note content is required", nil)
		return
	}

	variables := map[string]interface{}{
		"applicationId": appID,
		"content":       input.Content,
		"isInternal":    input.IsInternal,
	}

	resp, err := h.client.Mutate(ctx, gateway.AddApplicationNoteMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add note", err)
		return
	}

	respondJSON(w, http.StatusCreated, resp.Data)
}

// ScoreApplication triggers AI scoring for an application
func (h *ApplicationHandler) ScoreApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	appID := chi.URLParam(r, "id")
	
	if appID == "" {
		respondError(w, http.StatusBadRequest, "Application ID is required", nil)
		return
	}

	variables := map[string]interface{}{
		"applicationId": appID,
	}

	resp, err := h.client.Mutate(ctx, gateway.ScoreApplicationMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to score application", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// GetCandidate returns candidate information
func (h *ApplicationHandler) GetCandidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	candidateID := chi.URLParam(r, "id")
	
	if candidateID == "" {
		respondError(w, http.StatusBadRequest, "Candidate ID is required", nil)
		return
	}

	variables := map[string]interface{}{
		"id": candidateID,
	}

	resp, err := h.client.Query(ctx, gateway.GetCandidateQuery, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch candidate", err)
		return
	}

	if resp.Data == nil {
		respondError(w, http.StatusNotFound, "Candidate not found", nil)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}

// UpdateCandidate updates candidate profile
func (h *ApplicationHandler) UpdateCandidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	candidateID := chi.URLParam(r, "id")
	
	if candidateID == "" {
		respondError(w, http.StatusBadRequest, "Candidate ID is required", nil)
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	defer r.Body.Close()

	variables := map[string]interface{}{
		"id":    candidateID,
		"input": input,
	}

	resp, err := h.client.Mutate(ctx, gateway.UpdateCandidateProfileMutation, variables)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update candidate", err)
		return
	}

	respondJSON(w, http.StatusOK, resp.Data)
}