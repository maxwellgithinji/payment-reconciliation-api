package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/insurance-portal/poc/internal/application/commands"
	"github.com/insurance-portal/poc/internal/application/queries"
	"github.com/insurance-portal/poc/internal/domain/reconciliation"
	httpHandler "github.com/insurance-portal/poc/internal/interfaces/http"
	// Import infrastructure implementations when ready
	// "github.com/insurance-portal/poc/internal/infrastructure/persistence/postgres"
	// "github.com/insurance-portal/poc/internal/infrastructure/auth"
)

const (
	defaultPort      = "8080"
	defaultJWTSecret = "your-secret-key-change-in-production"
	shutdownTimeout  = 15 * time.Second
)

func main() {
	// Load configuration
	port := getEnv("PORT", defaultPort)
	jwtSecret := getEnv("JWT_SECRET", defaultJWTSecret)

	// Initialize repositories (mock for POC)
	// In production, these would be actual database implementations
	log.Println("Initializing repositories...")
	
	// TODO: Initialize PostgreSQL connection and repositories
	// db := initDatabase()
	// userRepo := postgres.NewUserRepository(db)
	// customerRepo := postgres.NewCustomerRepository(db)
	// policyRepo := postgres.NewPolicyRepository(db)
	// paymentRepo := postgres.NewPaymentRepository(db)
	// reconciliationRepo := postgres.NewReconciliationRepository(db)
	
	// For POC, using in-memory repositories (to be implemented)
	// userRepo := memory.NewUserRepository()
	// customerRepo := memory.NewCustomerRepository()
	// policyRepo := memory.NewPolicyRepository()
	// paymentRepo := memory.NewPaymentRepository()
	// reconciliationRepo := memory.NewReconciliationRepository()
	
	log.Println("Note: Using placeholder repositories - implement actual persistence layer")

	// Initialize domain services
	log.Println("Initializing domain services...")
	reconRule := reconciliation.DefaultReconciliationRule()
	reconService := reconciliation.NewReconciliationService(reconRule)
	
	// TODO: Initialize user service
	// userService := auth.NewUserService()

	// Initialize command handlers (commented out until repos are ready)
	// createUserHandler := commands.NewCreateUserHandler(userRepo, userService)
	// updateUserHandler := commands.NewUpdateUserHandler(userRepo)
	// deactivateUserHandler := commands.NewDeactivateUserHandler(userRepo)
	// createPolicyHandler := commands.NewCreatePolicyHandler(policyRepo, customerRepo)
	// cancelPolicyHandler := commands.NewCancelPolicyHandler(policyRepo)
	// recordPaymentHandler := commands.NewRecordPolicyPaymentHandler(policyRepo)
	// processPaymentHandler := commands.NewProcessPaymentHandler(paymentRepo, policyRepo, reconciliationRepo, reconService)
	// manualReconHandler := commands.NewManualReconciliationHandler(reconciliationRepo, paymentRepo, policyRepo)

	// Initialize query handlers (commented out until repos are ready)
	// dashboardHandler := queries.NewPaymentTrackingDashboardHandler(policyRepo, paymentRepo, reconciliationRepo)
	// reconReportHandler := queries.NewReconciliationReportHandler(reconciliationRepo, paymentRepo, policyRepo, customerRepo)
	// policyDetailHandler := queries.NewPolicyDetailHandler(policyRepo, customerRepo, paymentRepo)

	// Initialize HTTP handlers (commented out until above are ready)
	// policyHandler := httpHandler.NewPolicyHandler(createPolicyHandler, cancelPolicyHandler, policyDetailHandler, policyRepo, customerRepo)
	// paymentHandler := httpHandler.NewPaymentHandler(processPaymentHandler, dashboardHandler, reconReportHandler, manualReconHandler)
	// customerHandler := httpHandler.NewCustomerHandler(customerRepo)
	// userHandler := httpHandler.NewUserHandler(createUserHandler, updateUserHandler, deactivateUserHandler, userRepo, userService, jwtSecret)
	// authMiddleware := httpHandler.NewAuthMiddleware(jwtSecret, userRepo)

	// Setup router (commented out until handlers are ready)
	// router := httpHandler.NewRouter(policyHandler, paymentHandler, customerHandler, userHandler, authMiddleware)

	// For now, create a basic health check server
	router := http.NewServeMux()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","message":"Insurance Portal POC API"}`)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		log.Printf("Dashboard: http://localhost:%s/api/v1/dashboard", port)
		log.Printf("Health check: http://localhost:%s/health", port)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}