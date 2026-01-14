package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
	Status  int         `json:"status"`
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// respondError writes an error response
func respondError(w http.ResponseWriter, status int, message string, err error) {
	response := ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
		Status:  status,
	}
	
	if err != nil {
		response.Details = err.Error()
		log.Printf("Error: %s - %v", message, err)
	}
	
	respondJSON(w, status, response)
}

// respondSuccess writes a success response with a message
func respondSuccess(w http.ResponseWriter, message string, data interface{}) {
	response := map[string]interface{}{
		"success": true,
		"message": message,
	}
	
	if data != nil {
		response["data"] = data
	}
	
	respondJSON(w, http.StatusOK, response)
}