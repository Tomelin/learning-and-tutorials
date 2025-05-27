package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

// GenerateContent sends a simple prompt to the Gemini model and returns the textual response.
// This function is similar to what might have existed in the 001-basic project.
func GenerateContent(ctx context.Context, genAIClient *genai.GenerativeModel, simplePrompt string) (string, error) {
	if genAIClient == nil {
		return "", fmt.Errorf("GenerativeModel client is nil")
	}
	if simplePrompt == "" {
		return "", fmt.Errorf("prompt is empty")
	}

	resp, err := genAIClient.GenerateContent(ctx, genai.Text(simplePrompt))
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response candidates found")
	}

	var builder strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					builder.WriteString(string(txt))
				}
			}
		}
	}
	return builder.String(), nil
}

// GenerateContentWithContext generates content using the Gemini model, with additional context
// provided from MongoDB documents (RAG pattern).
func GenerateContentWithContext(ctx context.Context, genAIClient *genai.GenerativeModel, userQuery string, mongoDocuments []string) (string, error) {
	if genAIClient == nil {
		return "", fmt.Errorf("GenerativeModel client is nil")
	}
	if userQuery == "" {
		return "", fmt.Errorf("user query is empty")
	}

	var promptBuilder strings.Builder
	promptBuilder.WriteString("Based on the following information:\n\n")

	if len(mongoDocuments) > 0 {
		for i, doc := range mongoDocuments {
			promptBuilder.WriteString(fmt.Sprintf("Context Document %d:\n%s\n\n", i+1, doc))
		}
	} else {
		promptBuilder.WriteString("No specific context documents were provided.\n\n")
	}

	promptBuilder.WriteString("Please answer the question: ")
	promptBuilder.WriteString(userQuery)

	constructedPrompt := promptBuilder.String()

	// For debugging or logging, you might want to see the constructed prompt
	// log.Printf("Constructed RAG Prompt:\n%s\n", constructedPrompt)

	resp, err := genAIClient.GenerateContent(ctx, genai.Text(constructedPrompt))
	if err != nil {
		return "", fmt.Errorf("error generating content with context: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response candidates found for RAG query")
	}

	var responseBuilder strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					responseBuilder.WriteString(string(txt))
				}
			}
		}
	}

	if responseBuilder.Len() == 0 {
		return "", fmt.Errorf("empty response from model after RAG query")
	}

	return responseBuilder.String(), nil
}

// GetEmbedding generates and returns the embedding for a given text using the provided EmbeddingModel.
func GetEmbedding(ctx context.Context, embeddingClient *genai.EmbeddingModel, text string) ([]float32, error) {
	if embeddingClient == nil {
		return nil, fmt.Errorf("embeddingClient is nil")
	}
	if text == "" {
		return nil, fmt.Errorf("text to embed cannot be empty")
	}

	res, err := embeddingClient.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("failed to embed content: %w", err)
	}

	if res == nil || res.Embedding == nil || len(res.Embedding.Values) == 0 {
		return nil, fmt.Errorf("received empty embedding from API")
	}

	return res.Embedding.Values, nil
}
