/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"learn-ai/ragmongo/config"
	"learn-ai/ragmongo/pkg/db"
	"learn-ai/ragmongo/pkg/llm"

	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

var query string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ragcli",
	Short: "A CLI tool to perform RAG queries using MongoDB and Gemini.",
	Long: `This application allows you to ask questions (queries) that will be
answered by a Gemini language model, with context retrieved from a MongoDB
vector search.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load Configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}

		// Get Query
		queryString, _ := cmd.Flags().GetString("query")
		if queryString == "" {
			log.Fatalf("Error: query flag cannot be empty. Use -q or --query.")
		}

		ctx := context.Background()

		// Initialize Gemini Client (main client)
		genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
		if err != nil {
			log.Fatalf("Failed to create GenAI client: %v", err)
		}
		defer genaiClient.Close()

		// Initialize Gemini Generative Model
		// Model names can be like "gemini-pro", "gemini-1.0-pro", "gemini-1.5-flash-latest" etc.
		// Using a common one, can be made configurable if needed.
		generativeModel := genaiClient.GenerativeModel("gemini-pro")

		// Initialize Gemini Embedding Model
		// Model names can be like "text-embedding-004", "embedding-001" etc.
		// Using a common one, can be made configurable.
		embeddingModel := genaiClient.EmbeddingModel("text-embedding-004")
		if embeddingModel == nil { // Simple check as NewEmbeddingModelClient is not directly used
			log.Fatalf("Failed to initialize embedding model.")
		}

		// Initialize MongoDB Client
		mongoClient, err := db.NewMongoClient(cfg)
		if err != nil {
			log.Fatalf("Failed to initialize MongoDB client: %v", err)
		}
		defer func() {
			if err := mongoClient.Disconnect(ctx); err != nil {
				log.Printf("Failed to disconnect MongoDB client: %v", err)
			}
		}()

		// Perform RAG
		log.Println("Performing vector search...")
		retrievedDocs, err := db.VectorSearch(ctx, mongoClient, embeddingModel, cfg, queryString)
		if err != nil {
			log.Fatalf("Error during vector search: %v", err)
		}

		if len(retrievedDocs) == 0 {
			log.Println("No relevant documents found in MongoDB for the query.")
			// Decide if you want to proceed to LLM without context or inform user
			// For this example, we'll proceed, and the LLM will be informed via prompt.
		} else {
			log.Printf("Retrieved %d documents from MongoDB.", len(retrievedDocs))
		}

		log.Println("Generating content with context...")
		finalResponse, err := llm.GenerateContentWithContext(ctx, generativeModel, queryString, retrievedDocs)
		if err != nil {
			log.Fatalf("Error generating content with context: %v", err)
		}

		// Print Response
		fmt.Println("\nResponse:")
		fmt.Println(finalResponse)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&query, "query", "q", "", "Query string for RAG (required)")
	// Mark persistent flag as required if using a newer version of Cobra that supports it directly
	// For older versions, the check is done in Run.
	// Example for newer Cobra: rootCmd.MarkPersistentFlagRequired("query")

	// The existing interactive flag:
	rootCmd.Flags().BoolP("interactive", "i", false, "Enable interactive session (not used in this RAG command)")
}
