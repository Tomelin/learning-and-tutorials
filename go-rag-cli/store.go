package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RAGEntry struct {
	ID       string    `bson:"_id,omitempty"` // To hold MongoDB's _id
	Sentence string    `bson:"sentence"`
	Vector   []float64 `bson:"vector"` // Assuming vector is a slice of float64
	Answer   string    `bson:"answer"`
	Tags     []string  `bson:"tags"`
}

// Connect initializes a new MongoDB client.
func Connect(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Ping the primary
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	// log.Println("Connected to MongoDB!") // Optional: for debugging
	return client, nil
}

// Disconnect closes the MongoDB client.
func Disconnect(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		return err
	}
	// log.Println("Disconnected from MongoDB.") // Optional: for debugging
	return nil
}

// SaveEntry saves a RAGEntry to the database.
func SaveEntry(client *mongo.Client, entry RAGEntry) error {
	collection := client.Database("ragdb").Collection("entries") // Using "ragdb" and "entries"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, entry)
	if err != nil {
		// log.Printf("Failed to save entry: %v
", err) // Optional for debugging
		return err
	}
	// log.Printf("Saved entry with ID: %s
", result.InsertedID) // Optional
	return nil
}

// SearchEntries searches for RAGEntries by tags.
// True vector similarity search is a TODO and may require MongoDB Atlas Search
// or client-side computation. The queryVector parameter is for future use.
func SearchEntries(client *mongo.Client, queryVector []float64, tags []string) ([]RAGEntry, error) {
	collection := client.Database("ragdb").Collection("entries")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$all": tags} // Matches documents where 'tags' array contains all specified tags
	}

	// TODO: Implement actual vector similarity search if possible.
	// For now, this function primarily filters by tags.
	// The queryVector is present for future vector search implementation.
	// log.Printf("Searching with filter: %v and queryVector (unused currently): %v", filter, queryVector) // Optional for debugging

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		// log.Printf("Failed to find entries: %v
", err) // Optional
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []RAGEntry
	for cursor.Next(ctx) {
		var entry RAGEntry
		if err := cursor.Decode(&entry); err != nil {
			// log.Printf("Failed to decode entry: %v
", err) // Optional
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := cursor.Err(); err != nil {
		// log.Printf("Cursor error: %v
", err) // Optional
		return nil, err
	}

	return entries, nil
}
