package utils

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User struct (adjust according to your schema)
type User struct {
	ID        string `bson:"_id,omitempty" json:"id,omitempty"`
	FirstName string `bson:"firstName,omitempty" json:"firstName,omitempty"`
	LastName  string `bson:"lastName,omitempty" json:"lastName,omitempty"`
	Email     string `bson:"email,omitempty" json:"email,omitempty"`
}

// GetAllUsers connects to MongoDB and returns all users as a slice
func GetAllUsers(client *mongo.Client, dbName string, collName string) ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := client.Database(dbName).Collection(collName)

	// empty filter => get all docs
	cursor, err := collection.Find(ctx, bson.M{}, options.Find())
	if err != nil {
		return nil, fmt.Errorf("error fetching users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []User
	for cursor.Next(ctx) {
		var u User
		if err := cursor.Decode(&u); err != nil {
			return nil, fmt.Errorf("decode error: %w", err)
		}
		users = append(users, u)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return users, nil
}
