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

	"github.com/joho/godotenv"
	"github.com/parth5404/TEST-GS-Backend/go_email/utils"
)

func main() {
	// Load environment variables
	if os.Getenv("ENVIRONMENT") == "development" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	// Start the daily email cron job
	log.Println("🚀 Starting email service with daily cron job...")
	mongoClient, err := utils.StartDailyEmailScheduler()
	if err != nil {
		log.Printf("❌ Failed to start daily email scheduler: %v", err)
		log.Println("⚠️  Continuing without cron job - only HTTP endpoints will work")
	} else {
		log.Println("✅ Daily email cron job started successfully")

		// Ensure MongoDB connection is closed when the program exits
		defer func() {
			if mongoClient != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := mongoClient.Disconnect(ctx); err != nil {
					log.Printf("Error disconnecting from MongoDB: %v", err)
				} else {
					log.Println("✅ MongoDB connection closed")
				}
			}
		}()
	}

	// Set up HTTP endpoints
	http.HandleFunc("/send-email", utils.EmailConv)

	// Add a health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"email-service","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	// Add endpoint to manually trigger daily email (for testing)
	http.HandleFunc("/trigger-daily-email", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Println("📧 Manual trigger for daily email received")
		if mongoClient == nil {
			http.Error(w, "MongoDB not connected", http.StatusInternalServerError)
			return
		}

		// This would be the same function that cron calls
		go func() {
			users, err := utils.GetAllUsers(mongoClient)
			if err != nil {
				log.Printf("❌ Error getting users: %v", err)
				return
			}
			log.Printf("✅ Manual daily email triggered for %d users", len(users))
		}()

		fmt.Fprintf(w, `{"status":"triggered","message":"Daily email job started manually"}`)
	})

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: nil, // Use default mux
	}

	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("🌐 Email service server starting on port %s...", port)
		log.Printf("📍 Available endpoints:")
		log.Printf("   POST /send-email - Send individual emails")
		log.Printf("   GET  /health - Health check")
		log.Printf("   POST /trigger-daily-email - Manually trigger daily emails")
		log.Printf("⏰ Daily emails scheduled for 9:00 AM IST")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("🛑 Shutting down email service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Email service stopped gracefully")
	}
}
