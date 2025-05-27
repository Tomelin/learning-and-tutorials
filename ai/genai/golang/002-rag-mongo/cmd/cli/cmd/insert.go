package cmd

import (
	"context"
	"fmt"
	"io/ioutil" // For reading file content
	"log"       // For logging errors from RunE
	"os"        // For os.Exit and file operations if needed

	"learn-ai/ragmongo/config"
	"learn-ai/ragmongo/pkg/db"
	"learn-ai/ragmongo/pkg/llm"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

var textContent string
var filePath string

// insertCmd represents the insert command
var insertCmd = &cobra.Command{
	Use:   "insert",
	Short: "Insert a document (text and its embedding) into MongoDB for RAG.",
	Long: `Insert a document into MongoDB. The document consists of the provided text
and its automatically generated embedding using Google Gemini.
You must provide either --text or --file flag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Validate flags: only one of text or file should be provided
		if (textContent == "" && filePath == "") || (textContent != "" && filePath != "") {
			return fmt.Errorf("you must provide either --text or --file flag, but not both")
		}

		// 2. Load Config
		// Assuming config is in the executable's dir or parent.
		// config.LoadConfig() uses env vars or a default path, so "." is not strictly needed.
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// 3. Read input text
		var finalContent string
		if filePath != "" {
			data, errReadFile := ioutil.ReadFile(filePath)
			if errReadFile != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, errReadFile)
			}
			finalContent = string(data)
		} else {
			finalContent = textContent
		}

		if finalContent == "" {
			return fmt.Errorf("input content cannot be empty")
		}

		// 4. Initialize Clients
		ctx := context.Background()

		geminiAPIKey := cfg.GeminiAPIKey
		if geminiAPIKey == "" {
			return fmt.Errorf("GEMINI_API_KEY is not set in config")
		}

		// Initialize the genai.Client first
		genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(geminiAPIKey))
		if err != nil {
			return fmt.Errorf("failed to create genai client: %w", err)
		}
		defer genaiClient.Close()

		// cfg.EmbeddingModel should be "text-embedding-004" or similar
		// Hardcoding "text-embedding-004" as cfg.EmbeddingModel is not yet in config.Configs
		embeddingModelName := "text-embedding-004"
		log.Printf("Using embedding model: %s (Note: Currently hardcoded, consider adding to config.Configs)", embeddingModelName)
		embeddingClient := genaiClient.EmbeddingModel(embeddingModelName)
		if embeddingClient == nil {
			return fmt.Errorf("failed to initialize embedding model '%s'", embeddingModelName)
		}

		// MongoDB Client
		mongoClient, err := db.NewMongoClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to connect to MongoDB: %w", err)
		}
		defer func() {
			if err := mongoClient.Disconnect(ctx); err != nil {
				log.Printf("Failed to disconnect MongoDB client: %v", err)
			}
		}()

		// 5. Generate Embedding
		fmt.Println("Generating embedding for the input text...")
		embeddingVector, err := llm.GetEmbedding(ctx, embeddingClient, finalContent)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}

		// 6. Insert Document
		fmt.Println("Inserting document into MongoDB...")
		err = db.InsertRAGDocument(ctx, mongoClient, cfg, finalContent, embeddingVector)
		if err != nil {
			return fmt.Errorf("failed to insert document: %w", err)
		}

		fmt.Println("Successfully inserted document into MongoDB.")
		return nil
	},
}

func init() {
	// This function will be called by Cobra to initialize the command and its flags.
	// It's important that it's added to the rootCmd in root.go

	// Add flags to insertCmd
	insertCmd.Flags().StringVarP(&textContent, "text", "t", "", "Text content to insert directly.")
	insertCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to a text file to insert.")

	// Add insertCmd to rootCmd
	// This is typically done in the init() of the root command or main.go
	// For example: rootCmd.AddCommand(insertCmd)
	// Ensure this command is added to the root command to be available.
	// For now, we define it here. If root.go is in the same package, it can access it.
	// Otherwise, a function like NewInsertCmd() *cobra.Command { return insertCmd } might be needed.
	rootCmd.AddCommand(insertCmd) // Assuming rootCmd is accessible, as per typical Cobra structure.
}
