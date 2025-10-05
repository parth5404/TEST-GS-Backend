package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/parth5404/TEST-GS-Backend/go_email/services"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User represents a user document from MongoDB (matching Node.js schema)
type User struct {
	ID                  string   `bson:"_id,omitempty" json:"id,omitempty"`
	FirstName           string   `bson:"firstName" json:"firstName"`
	LastName            string   `bson:"lastName" json:"lastName"`
	Email               string   `bson:"email" json:"email"`
	Role                string   `bson:"role" json:"role"`
	Avatar              string   `bson:"avatar" json:"avatar"`
	Active              bool     `bson:"active" json:"active"`
	Approved            bool     `bson:"approved" json:"approved"`
	Profile             string   `bson:"profile" json:"profile"`
	Courses             []string `bson:"courses" json:"courses"`
	CourseProgress      []string `bson:"courseProgress" json:"courseProgress"`
	Reviews             []string `bson:"reviews" json:"reviews"`
	Token               string   `bson:"token,omitempty" json:"token,omitempty"`
	ResetPasswordToken  string   `bson:"resetPasswordToken,omitempty" json:"resetPasswordToken,omitempty"`
	ResetPasswordExpire string   `bson:"resetPasswordExpire,omitempty" json:"resetPasswordExpire,omitempty"`
	CreatedAt           string   `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt           string   `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}

// StartDailyEmailScheduler starts the cron job for daily emails at 9 AM
func StartDailyEmailScheduler() (*mongo.Client, error) {
	// Load Mongo URI
	mongoURI := os.Getenv("MONGO_URI")
	// if mongoURI == "" {
	// 	log.Println("Warning: MONGO_URI not set, using default localhost")
	// 	mongoURI = "mongodb+srv://parthlahoti5404:sher5404@cluster0.wbav0jr.mongodb.net/"
	// }

	// Add database name to URI if not present
	if !strings.Contains(mongoURI, "mongodb.net/") || strings.HasSuffix(mongoURI, "mongodb.net/") {
		if strings.HasSuffix(mongoURI, "/") {
			mongoURI += "test" // Add default database name
		} else {
			mongoURI += "/test" // Add default database name
		}
		log.Printf("📍 Updated MONGO_URI to include database: %s", mongoURI)
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Configure client options with authentication
	clientOpts := options.Client().ApplyURI(mongoURI)

	// Add additional options for better connection handling
	clientOpts.SetMaxPoolSize(10)
	clientOpts.SetMinPoolSize(5)

	log.Printf("🔌 Attempting to connect to MongoDB...")
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect error: %w", err)
	}

	// Verify connection with retry
	log.Printf("🏓 Pinging MongoDB to verify connection...")
	if err = client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		log.Printf("⚠️  MongoDB connection failed: %v", err)
		log.Println("📧 Cron job will be disabled until MongoDB is fixed")
		return nil, fmt.Errorf("mongo ping error: %w", err)
	}
	log.Println("✅ Connected to MongoDB for cron jobs")

	// Create cron scheduler with timezone support
	location, _ := time.LoadLocation("Asia/Kolkata") // IST timezone
	c := cron.New(cron.WithLocation(location))

	// Daily at 9:00 AM IST - "0 9 * * *" means every day at 9 AM
	_, err = c.AddFunc("0 9 * * *", func() {
		log.Println("🕘 Running daily email job at:", time.Now().Format("2006-01-02 15:04:05"))
		if err := sendDailyNewsletterToAll(client); err != nil {
			log.Printf("❌ Daily email job error: %v", err)
		} else {
			log.Println("✅ Daily email job completed successfully")
		}
	})

	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to add cron job: %w", err)
	}

	// Start the cron scheduler
	c.Start()
	log.Println("🚀 Daily email cron scheduler started - will send emails daily at 9:00 AM IST")

	return client, nil
}

// GetAllUsers fetches all active users from MongoDB
func GetAllUsers(client *mongo.Client) ([]User, error) {
	//dbName := os.Getenv("DB_NAME")
	dbName := "test"

	collection := client.Database(dbName).Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find all active and approved users (matching your Node.js logic)
	filter := bson.M{
		"active":   true,
		"approved": true,
	}

	// Projection to only get required fields for email sending
	projection := bson.M{
		"firstName": 1,
		"lastName":  1,
		"email":     1,
		"role":      1,
		"active":    1,
		"approved":  1,
	}

	opts := options.Find().SetProjection(projection)
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	log.Printf("📧 Found %d active and approved users for daily email", len(users))

	// Log some sample user info (without sensitive data)
	if len(users) > 0 {
		log.Printf("📝 Sample user: %s %s (%s) - Role: %s",
			users[0].FirstName, users[0].LastName, users[0].Email, users[0].Role)
	}

	return users, nil
}

// sendDailyNewsletterToAll sends daily newsletter to all users
func sendDailyNewsletterToAll(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get all active users
	users, err := GetAllUsers(client)
	if err != nil {
		return fmt.Errorf("GetAllUsers failed: %w", err)
	}

	if len(users) == 0 {
		log.Println("⚠️  No active users found for daily email")
		return nil
	}

	// Use worker pool to send emails concurrently
	const maxWorkers = 10
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	for _, u := range users {
		// Check if context is cancelled
		if ctx.Err() != nil {
			log.Println("⚠️  Context cancelled, stopping email sends")
			break
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore slot

		// Capture user for goroutine
		user := u
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore slot

			// Create daily email content
			subject := "Daily LMS Update - " + time.Now().Format("January 2, 2006")
			body := createDailyEmailBody(user)

			// Send email using the services.SendEmail function
			if err := services.SendEmail(user.Email, subject, body); err != nil {
				log.Printf("❌ Failed to send email to %s (%s): %v", user.FirstName, user.Email, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}

			log.Printf("✅ Email sent successfully to %s (%s)", user.FirstName, user.Email)
			mu.Lock()
			successCount++
			mu.Unlock()
		}()
	}

	// Wait for all emails to be sent
	wg.Wait()

	log.Printf("📊 Daily email summary: %d successful, %d failed, %d total", successCount, errorCount, len(users))
	return nil
}

// createDailyEmailBody creates the HTML body for daily email
func createDailyEmailBody(user User) string {
	currentDate := time.Now().Format("Monday, January 2, 2006")

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Daily LMS Update</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .highlight { background: #e3f2fd; padding: 15px; border-left: 4px solid #2196f3; margin: 20px 0; border-radius: 5px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        .button { background: #4CAF50; color: white; padding: 12px 25px; text-decoration: none; border-radius: 5px; display: inline-block; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎓 Daily LMS Update</h1>
        <p>%s</p>
    </div>
    <div class="content">
        <h2>Hello %s! 👋</h2>
        
        <div class="highlight">
            <h3>📚 Today's Learning Reminder</h3>
            <p>Don't forget to continue your learning journey! Every day is an opportunity to grow and acquire new skills.</p>
        </div>
        
        <h3>🔥 Quick Tips for Today:</h3>
        <ul>
            <li>📖 Spend at least 30 minutes on your enrolled courses</li>
            <li>💡 Practice what you learned yesterday</li>
            <li>🤝 Engage with your fellow learners in discussions</li>
            <li>🎯 Set a small learning goal for today</li>
        </ul>
        
        <div class="highlight">
            <h3>💪 Your Learning Streak</h3>
            <p>Keep up the great work! Consistency is key to mastering new skills. Remember, small daily progress leads to big achievements.</p>
        </div>
        
        <center>
            <a href="#" class="button">Continue Learning 🚀</a>
        </center>
        
        <div class="footer">
            <p>This is your daily motivation email from LMS Team</p>
            <p>Stay curious, stay learning! 🌟</p>
            <hr>
            <small>You're receiving this because you're an active member of our learning community.</small>
        </div>
    </div>
</body>
</html>`, currentDate, user.FirstName)
}
