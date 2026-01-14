package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"hr-recruiting/internal/config"
	"hr-recruiting/internal/gateway"
	"hr-recruiting/internal/handlers"
	appMiddleware "hr-recruiting/internal/middleware"
	"hr-recruiting/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize services
	hubHRMSClient := gateway.NewHubHRMSClient(cfg.HubHRMS.URL, cfg.HubHRMS.APIKey)
	uploadService := services.NewUploadService(cfg.AWS.S3Bucket, cfg.AWS.Region)
	emailService := services.NewEmailService(cfg.Email.SendGridKey)
	
	// Initialize handlers
	jobHandler := handlers.NewJobHandler(hubHRMSClient)
	applicationHandler := handlers.NewApplicationHandler(hubHRMSClient, uploadService, emailService)
	analyticsHandler := handlers.NewAnalyticsHandler(hubHRMSClient)
	healthHandler := handlers.NewHealthHandler(hubHRMSClient)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Custom middleware
	r.Use(appMiddleware.AuthMiddleware)

	// Health check (no auth required)
	r.Get("/health", healthHandler.Health)
	r.Get("/health/live", healthHandler.Liveness)
	r.Get("/health/ready", healthHandler.Readiness)

	// GraphQL proxy to Hub-HRMS
	r.Post("/graphql", hubHRMSClient.ProxyHandler)

	// API Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			// Jobs
			r.Get("/jobs", jobHandler.ListJobs)
			r.Get("/jobs/{id}", jobHandler.GetJob)
			r.Post("/jobs/{id}/view", jobHandler.IncrementView)

			// Applications (public submission)
			r.Post("/applications", applicationHandler.SubmitApplication)

			// File upload (public for candidates)
			r.Post("/upload/resume", uploadService.UploadResume)
			r.Post("/upload/presigned-url", uploadService.GetPresignedURL)
		})

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.RequireAuth)

			// Job management (recruiters/admins)
			r.Post("/jobs", jobHandler.CreateJob)
			r.Put("/jobs/{id}", jobHandler.UpdateJob)
			r.Post("/jobs/{id}/publish", jobHandler.PublishJob)
			r.Post("/jobs/{id}/close", jobHandler.CloseJob)
			r.Delete("/jobs/{id}", jobHandler.DeleteJob)
			r.Post("/jobs/generate-description", jobHandler.GenerateDescription)

			// Application management (recruiters)
			r.Get("/applications", applicationHandler.ListApplications)
			r.Get("/applications/{id}", applicationHandler.GetApplication)
			r.Put("/applications/{id}/status", applicationHandler.UpdateStatus)
			r.Post("/applications/{id}/notes", applicationHandler.AddNote)
			r.Post("/applications/{id}/score", applicationHandler.ScoreApplication)
			r.Post("/applications/bulk-update", applicationHandler.BulkUpdateStatus)

			// Analytics (recruiters/admins)
			r.Get("/analytics/metrics", analyticsHandler.GetMetrics)
			r.Get("/analytics/jobs/{id}/performance", analyticsHandler.GetJobPerformance)
			r.Get("/analytics/pipeline", analyticsHandler.GetPipeline)
			r.Get("/analytics/trends", analyticsHandler.GetTrends)

			// Candidate management
			r.Get("/candidates/{id}", applicationHandler.GetCandidate)
			r.Put("/candidates/{id}", applicationHandler.UpdateCandidate)
		})
	})

	// Static file serving (optional)
	workDir, _ := os.Getwd()
	filesDir := http.Dir(workDir + "/static")
	FileServer(r, "/static", filesDir)

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 HR Recruiting API server starting on port %s", cfg.Server.Port)
		log.Printf("📡 Hub-HRMS endpoint: %s", cfg.HubHRMS.URL)
		log.Printf("🌍 Environment: %s", cfg.Server.Environment)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	<-done
	log.Println("🛑 Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

// FileServer conveniently sets up a http.FileServer handler to serve static files
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}