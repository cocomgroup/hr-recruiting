package config

import (
	"os"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	Server  ServerConfig
	HubHRMS HubHRMSConfig
	AWS     AWSConfig
	Email   EmailConfig
	CORS    CORSConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        string
	Environment string
}

// HubHRMSConfig holds Hub-HRMS integration configuration
type HubHRMSConfig struct {
	URL    string
	APIKey string
}

// AWSConfig holds AWS configuration
type AWSConfig struct {
	Region   string
	S3Bucket string
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	SendGridKey string
	FromEmail   string
	FromName    string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		HubHRMS: HubHRMSConfig{
			URL:    getEnv("HUBHRMS_GRAPHQL_URL", ""),
			APIKey: getEnv("HUBHRMS_API_KEY", ""),
		},
		AWS: AWSConfig{
			Region:   getEnv("AWS_REGION", "us-east-1"),
			S3Bucket: getEnv("AWS_S3_BUCKET", "hr-recruiting-resumes"),
		},
		Email: EmailConfig{
			SendGridKey: getEnv("SENDGRID_API_KEY", ""),
			FromEmail:   getEnv("EMAIL_FROM", "noreply@company.com"),
			FromName:    getEnv("EMAIL_FROM_NAME", "HR Recruiting"),
		},
		CORS: CORSConfig{
			AllowedOrigins: strings.Split(
				getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000"),
				",",
			),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}