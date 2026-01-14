package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// UploadService handles file uploads to S3
type UploadService struct {
	client *s3.Client
	bucket string
}

// NewUploadService creates a new upload service
func NewUploadService(bucket, region string) *UploadService {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to load AWS config: %v", err))
	}

	return &UploadService{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}
}

// UploadResume handles direct resume file uploads
func (s *UploadService) UploadResume(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}
	
	contentType, allowed := allowedExts[ext]
	if !allowed {
		http.Error(w, "Invalid file type. Only PDF, DOC, and DOCX are allowed", http.StatusBadRequest)
		return
	}

	// Validate file size (max 10MB)
	if header.Size > 10<<20 {
		http.Error(w, "File too large. Maximum size is 10MB", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("resumes/%s/%s%s", 
		time.Now().Format("2006/01"), 
		uuid.New().String(), 
		ext,
	)

	// Upload to S3
	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"original-filename": header.Filename,
			"uploaded-at":       time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload file: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate public URL
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, filename)

	// Return response
	response := map[string]interface{}{
		"success":          true,
		"url":              url,
		"filename":         filename,
		"originalFilename": header.Filename,
		"size":             header.Size,
		"contentType":      contentType,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetPresignedURL generates a presigned URL for direct upload
func (s *UploadService) GetPresignedURL(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Filename    string `json:"filename"`
		ContentType string `json:"contentType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate content type
	allowedTypes := map[string]bool{
		"application/pdf": true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	}
	if !allowedTypes[input.ContentType] {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	// Generate unique key
	ext := filepath.Ext(input.Filename)
	key := fmt.Sprintf("resumes/%s/%s%s", 
		time.Now().Format("2006/01"), 
		uuid.New().String(), 
		ext,
	)

	// Create presigned request
	presignClient := s3.NewPresignClient(s.client)
	presignedReq, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(input.ContentType),
		Metadata: map[string]string{
			"original-filename": input.Filename,
		},
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate presigned URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate final URL
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key)

	// Return response
	response := map[string]interface{}{
		"success":   true,
		"uploadUrl": presignedReq.URL,
		"key":       key,
		"url":       url,
		"expiresIn": 900, // 15 minutes
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteFile deletes a file from S3
func (s *UploadService) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// GetFileURL returns the public URL for a file
func (s *UploadService) GetFileURL(key string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key)
}