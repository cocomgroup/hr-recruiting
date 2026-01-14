package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// HubHRMSClient is a GraphQL client for Hub-HRMS
type HubHRMSClient struct {
	url        string
	apiKey     string
	httpClient *http.Client
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}    `json:"data"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// NewHubHRMSClient creates a new Hub-HRMS client
func NewHubHRMSClient(url, apiKey string) *HubHRMSClient {
	return &HubHRMSClient{
		url:    url,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Query executes a GraphQL query
func (c *HubHRMSClient) Query(ctx context.Context, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Hub-HRMS returned status %d: %s", resp.StatusCode, string(body))
	}

	var gqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		log.Printf("GraphQL errors: %+v", gqlResp.Errors)
	}

	return &gqlResp, nil
}

// Mutate executes a GraphQL mutation
func (c *HubHRMSClient) Mutate(ctx context.Context, mutation string, variables map[string]interface{}) (*GraphQLResponse, error) {
	return c.Query(ctx, mutation, variables)
}

// ProxyHandler proxies GraphQL requests to Hub-HRMS
func (c *HubHRMSClient) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse GraphQL request
	var gqlReq GraphQLRequest
	if err := json.Unmarshal(body, &gqlReq); err != nil {
		http.Error(w, "Invalid GraphQL request", http.StatusBadRequest)
		return
	}

	// Forward to Hub-HRMS
	req, err := http.NewRequestWithContext(r.Context(), "POST", c.url, bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	// Copy user auth token from original request if present
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		req.Header.Set("X-User-Token", authHeader)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Error proxying to Hub-HRMS: %v", err)
		http.Error(w, "Failed to execute request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error copying response: %v", err)
	}
}

// Health checks Hub-HRMS connectivity
func (c *HubHRMSClient) Health(ctx context.Context) error {
	query := `query { __typename }`
	_, err := c.Query(ctx, query, nil)
	return err
}