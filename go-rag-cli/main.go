package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	// We will call functions from store.go and llm.go which have their own imports
)

func main() {
	// Define flags
	sentencePtr := flag.String("sentence", "", "The sentence/query for the RAG CLI (required)")
	tagsPtr := flag.String("tags", "", "Comma-separated tags (up to 3) to refine search (e.g., 'go,mongodb,vector')")

	flag.Parse()

	// Validate sentence
	if *sentencePtr == "" {
		log.Println("Error: -sentence flag is required.")
		flag.Usage()
		os.Exit(1)
	}
	sentence := *sentencePtr

	// Process tags
	var tags []string
	if *tagsPtr != "" {
		tags = strings.Split(*tagsPtr, ",")
		for i, t := range tags {
			tags[i] = strings.TrimSpace(t) // Clean up spaces
		}
		if len(tags) > 3 {
			log.Println("Error: Maximum of 3 tags allowed.")
			flag.Usage()
			os.Exit(1)
		}
	}

	log.Printf("Received sentence: %s\n", sentence)
	log.Printf("Received tags: %v\n", tags)

	// Retrieve API keys and URIs from environment variables
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		log.Fatal("Error: OPENAI_API_KEY environment variable not set.")
	}

	mongoDBURI := os.Getenv("MONGODB_URI")
	if mongoDBURI == "" {
		log.Fatal("Error: MONGODB_URI environment variable not set.")
	}
    
    log.Println("Configuration loaded (API keys and URI).")

	// --- Main Workflow ---
	// 1. Connect to MongoDB
	client, err := Connect(mongoDBURI) // Assumes Connect is in package main (store.go)
	if err != nil {
	    log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer Disconnect(client) // Assumes Disconnect is in package main (store.go)
	log.Println("Connected to MongoDB.")

	// 2. Get embedding for the sentence
	queryVector, err := GetEmbedding(openaiAPIKey, sentence) // Assumes GetEmbedding is in package main (llm.go)
	if err != nil {
	    log.Fatalf("Failed to get embedding: %v", err)
	}
	log.Printf("Generated embedding for sentence (vector length: %d)\n", len(queryVector))

	// 3. Search in MongoDB using tags and (conceptually) vector
	foundEntries, err := SearchEntries(client, queryVector, tags) // queryVector is for future use by SearchEntries
	if err != nil {
	    // Non-fatal error for search, as we can fall back to LLM
	    log.Printf("Error searching entries in MongoDB: %v. Proceeding to LLM.\n", err)
	}

	// 4. If a relevant entry is found, print its answer and exit
	if err == nil && len(foundEntries) > 0 {
	    // TODO: Define "relevance" more strictly if multiple entries are returned.
	    // For now, if any entry matches tags (and vector search was a TODO), we use the first one.
	    // In a true RAG, we'd check vector similarity here against queryVector.
	    log.Printf("Found existing answer in RAG store for sentence '%s' with tags %v.\n", sentence, tags)
	    fmt.Println("Answer from RAG store:", foundEntries[0].Answer)
	    os.Exit(0) // Successfully found and printed from RAG, so exit.
	}
	
	if err == nil { 
		log.Println("No sufficiently relevant answer found in RAG store by tags. Will query LLM.")
	} else {
        // This case means SearchEntries itself had an error. We already logged it.
        // We still proceed to LLM as a fallback.
        log.Println("Due to search error, proceeding to LLM.")
    }

	// 5. If not found (or if search failed), query LLM
	log.Println("Querying LLM...")
	llmAnswer, err := GetLLMAnswer(openaiAPIKey, sentence, tags) // Assumes GetLLMAnswer is in package main
	if err != nil {
	    log.Fatalf("Failed to get answer from LLM: %v", err)
	}

	// 6. Save the new sentence, vector, LLM answer, and tags to MongoDB
    if queryVector != nil { // Ensure we have a vector to save
        newEntry := RAGEntry{ // Assumes RAGEntry is in package main
            Sentence: sentence,
            Vector:   queryVector, 
            Answer:   llmAnswer,
            Tags:     tags,
        }
        errSave := SaveEntry(client, newEntry) // Assumes SaveEntry is in package main
        if errSave != nil { // Use a different error variable to avoid conflict with 'err' from SearchEntries
            log.Printf("Warning: Failed to save new entry to RAG store: %v\n", errSave)
            // Non-fatal, as we still have the answer for the user
        } else {
            log.Println("New entry saved to RAG store.")
        }
    } else {
        log.Println("Warning: queryVector is nil, skipping save to RAG store. This shouldn't happen if GetEmbedding was successful.")
    }

	// 7. Print LLM answer to the user
	fmt.Println("Answer from LLM:", llmAnswer)
    log.Println("Successfully provided answer from LLM.")
}
