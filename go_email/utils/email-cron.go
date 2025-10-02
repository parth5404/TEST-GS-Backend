package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setup() {
	// Load Mongo URI
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("mongo connect error: %v", err)
	}

	// Verify connection
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo ping error: %v", err)
	}
	log.Println("Connected to MongoDB")

	// Ensure client disconnect on exit
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	// Create cron scheduler
	c := cron.New()

	// Add function to run every Monday at 09:00 (server timezone)
	// If you want timezone aware schedules, use cron.New(cron.WithLocation(...))
	_, err = c.AddFunc("0 9 * * MON", func() {
		log.Println("Running weekly newsletter job:", time.Now())
		if err := sendWeeklyNewsletter(client); err != nil {
			log.Printf("weekly job error: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("failed to add cron job: %v", err)
	}

	c.Start()
	log.Println("Cron scheduler started")

	select {}
}


func sendWeeklyNewsletter(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	users, err := utils.GetAllUsers(client, "your_db_name", "users")
	if err != nil {
		return fmt.Errorf("GetAllUsers failed: %w", err)
	}

	for _, u := range users {
		subject := "Weekly Course Newsletter"
		body := "Hello " + u.FirstName + ",\n\nHere is your weekly course newsletter..."
		if err := utils.SendEmail(u.Email, subject, body); err != nil {
			log.Printf("failed to send email to %s: %v", u.Email, err)
			continue
		}
		log.Printf("email sent to %s", u.Email)
	}


	_ = ctx 
	return nil
}
