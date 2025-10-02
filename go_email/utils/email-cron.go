package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func StartNewsletterScheduler() (*mongo.Client, error) {
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
		return nil, fmt.Errorf("mongo connect error: %w", err)
	}

	// Verify connection
	if err = client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping error: %w", err)
	}
	log.Println("Connected to MongoDB")

	// Create cron scheduler
	c := cron.New()

	// Add function to run every Monday at 09:00 (server timezone)
	_, err = c.AddFunc("0 9 * * MON", func() {
		log.Println("Running weekly newsletter job:", time.Now())
		if err := sendWeeklyNewsletter(client); err != nil {
			log.Printf("weekly job error: %v", err)
		}
	})
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to add cron job: %w", err)
	}

	c.Start()
	log.Println("Cron scheduler started")

	return client, nil
}

// sendWeeklyNewsletter fetches all users and sends them an email concurrently with bounded workers.
func sendWeeklyNewsletter(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	users, err := GetAllUsers(client, "your_db_name", "users")
	if err != nil {
		return fmt.Errorf("GetAllUsers failed: %w", err)
	}

	const maxWorkers = 10 
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, u := range users {
		// respect context cancellation
		if ctx.Err() != nil {
			log.Println("context cancelled, stopping sends")
			break
		}

		wg.Add(1)
		sem <- struct{}{} // acquire slot

		// capture u for goroutine
		user := u
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			subject := "Weekly Course Newsletter"
			body := "Hello " + user.FirstName + ",\n\nHere is your weekly course newsletter..."

			if err := SendEmail(user.Email, subject, body); err != nil {
				log.Printf("failed to send email to %s: %v", user.Email, err)
				return
			}
			log.Printf("email sent to %s", user.Email)
		}()
	}

	wg.Wait()

	
	return nil
}
