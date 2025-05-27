package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/generative-ai-go/genai"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"learn-ai/ragmongo/config"
	"learn-ai/ragmongo/pkg/llm" // Added import for llm package
)

// NewMongoClient creates and returns a new MongoDB client.
// It uses the MongoURI from the provided configuration.
func NewMongoClient(cfg *config.Configs) (*mongo.Client, error) {
	if cfg.MongoURI == "" {
		return nil, fmt.Errorf("MongoDB URI is not configured")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(cfg.MongoURI).SetServerAPIOptions(serverAPI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the primary
	if err := client.Ping(ctx, nil); err != nil {
		// Disconnect if ping fails
		if disconnectErr := client.Disconnect(context.Background()); disconnectErr != nil {
			log.Printf("failed to disconnect MongoDB client after ping failure: %v", disconnectErr)
		}
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected to MongoDB and pinged successfully.")
	return client, nil
}

// VectorSearch performs a vector search in MongoDB using the provided query.
// It generates an embedding for the query, then uses that embedding to search
// a configured MongoDB collection.
func VectorSearch(ctx context.Context, mongoClient *mongo.Client, embeddingClient *genai.EmbeddingModel, cfg *config.Configs, query string) ([]string, error) {
	if mongoClient == nil {
		return nil, fmt.Errorf("MongoDB client is nil")
	}
	if embeddingClient == nil {
		return nil, fmt.Errorf("Gemini embedding client is nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("configuration is nil")
	}
	if cfg.MongoDatabase == "" || cfg.MongoCollection == "" || cfg.MongoVectorField == "" || cfg.MongoContentField == "" {
		return nil, fmt.Errorf("MongoDB database, collection, vector field, or content field is not configured")
	}
	if query == "" {
		return nil, fmt.Errorf("search query is empty")
	}

	log.Printf("Generating embedding for query: %s", query)
	embeddingVector, err := llm.GetEmbedding(ctx, embeddingClient, query) // Call the new function
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding for query '%s': %w", query, err)
	}
	// embeddingVector is now ready to be used
	log.Printf("Embedding generated successfully via llm.GetEmbedding. Vector dimension: %d", len(embeddingVector))

	collection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)

	// Hardcoded index name as per requirements.
	// Consider making numCandidates and limit configurable via cfg if needed.
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$vectorSearch", Value: bson.D{
				{Key: "index", Value: "vector_index"}, // Name of the vector search index in MongoDB Atlas
				{Key: "path", Value: cfg.MongoVectorField},
				{Key: "queryVector", Value: embeddingVector}, // This is the []float32 from Gemini
				{Key: "numCandidates", Value: int32(150)},    // Correctly use int32()
				{Key: "limit", Value: int32(10)},             // Correctly use int32()
			}},
		},
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0}, // Exclude the _id field
				{Key: cfg.MongoContentField, Value: 1},
				{Key: "score", Value: bson.D{{Key: "$meta", Value: "vectorSearchScore"}}},
			}},
		},
	}

	log.Printf("Executing vector search pipeline on collection: %s.%s", cfg.MongoDatabase, cfg.MongoCollection)
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to execute aggregation pipeline for vector search: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("Error closing MongoDB cursor: %v", err)
		}
	}()

	var results []string
	var documents []bson.M
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("failed to decode documents from cursor: %w", err)
	}
	
	log.Printf("Found %d documents from vector search.", len(documents))

	for _, doc := range documents {
		contentValue, ok := doc[cfg.MongoContentField]
		if !ok {
			log.Printf("Content field '%s' not found in document: %v", cfg.MongoContentField, doc)
			continue
		}

		contentStr, ok := contentValue.(string)
		if !ok {
			log.Printf("Content field '%s' is not a string in document: %v (type: %T)", cfg.MongoContentField, doc, contentValue)
			continue
		}
		results = append(results, contentStr)

		// Log score if present
		if scoreValue, ok := doc["score"]; ok {
			log.Printf("Document content: '%s', Score: %v", contentStr, scoreValue)
		} else {
			log.Printf("Document content: '%s' (Score not projected or found)", contentStr)
		}
	}

	if len(results) == 0 {
		log.Println("Vector search returned no matching documents.")
	}

	return results, nil
}

// InsertRAGDocument inserts a new document with its text content and vector embedding
// into the configured MongoDB collection.
func InsertRAGDocument(ctx context.Context, mongoClient *mongo.Client, cfg *config.Configs, textContent string, embeddingVector []float32) error {
	if mongoClient == nil {
		return fmt.Errorf("MongoDB client is nil")
	}
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if cfg.MongoDatabase == "" || cfg.MongoCollection == "" || cfg.MongoVectorField == "" || cfg.MongoContentField == "" {
		return fmt.Errorf("MongoDB database, collection, vector field, or content field is not configured")
	}
	if textContent == "" {
		return fmt.Errorf("textContent cannot be empty")
	}
	if len(embeddingVector) == 0 {
		return fmt.Errorf("embeddingVector cannot be empty")
	}

	collection := mongoClient.Database(cfg.MongoDatabase).Collection(cfg.MongoCollection)

	document := bson.M{
		cfg.MongoContentField: textContent,
		cfg.MongoVectorField:  embeddingVector,
		// Note: MongoDB will automatically generate an _id for this document.
	}

	log.Printf("Inserting document into %s.%s. Content field: '%s', Vector field: '%s'",
		cfg.MongoDatabase, cfg.MongoCollection, cfg.MongoContentField, cfg.MongoVectorField)

	_, err := collection.InsertOne(ctx, document)
	if err != nil {
		return fmt.Errorf("failed to insert document into MongoDB: %w", err)
	}

	log.Printf("Successfully inserted document.")
	return nil
}
