package main

import (
	"context"
	"errors"
	"fmt" // For constructing the prompt
	openai "github.com/sashabaranov/go-openai"
)

// GetLLMAnswerWithKey queries the LLM for an answer using an API key.
// It's a wrapper around GetLLMAnswer.
func GetLLMAnswerWithKey(apiKey string, sentence string, tags []string) (string, error) {
	client := openai.NewClient(apiKey)
	return GetLLMAnswer(client, sentence, tags)
}

// GetLLMAnswer queries the LLM for an answer using a provided OpenAI client.
// This is the core function, designed for testability by allowing client injection.
// It takes a sentence string, and a slice of tags as input.
// It returns the LLM's answer as a string and an error.
func GetLLMAnswer(client *openai.Client, sentence string, tags []string) (string, error) {
	// Construct the prompt message
	prompt := fmt.Sprintf("User query: %s", sentence)
	if len(tags) > 0 {
		prompt += fmt.Sprintf("\nRelevant tags: %v", tags)
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are a helpful assistant designed to answer user queries based on the provided information.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo, // Or another suitable model
		Messages: messages,
		// MaxTokens: // Optionally set MaxTokens
		// Temperature: // Optionally set Temperature
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		// log.Printf("Error creating chat completion: %v\n", err) // Optional
		return "", err
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		// log.Println("No answer received from LLM") // Optional
		return "", errors.New("no answer received from LLM")
	}

	return resp.Choices[0].Message.Content, nil
}

// GetEmbeddingWithKey generates a vector embedding for the given sentence using an API key.
// It's a wrapper around GetEmbedding.
func GetEmbeddingWithKey(apiKey string, sentence string) ([]float64, error) {
	client := openai.NewClient(apiKey)
	return GetEmbedding(client, sentence)
}

// GetEmbedding generates a vector embedding for the given sentence using a provided OpenAI client.
// This is the core function, designed for testability by allowing client injection.
// It returns a slice of float64 (the embedding) and an error.
func GetEmbedding(client *openai.Client, sentence string) ([]float64, error) {
	req := openai.EmbeddingRequest{
		Input: []string{sentence},
		Model: openai.AdaEmbeddingV2, // Or another model like openai.SmallEmbedding3
	}

	resp, err := client.CreateEmbeddings(context.Background(), req)
	if err != nil {
		// log.Printf("Error creating embedding: %v
", err) // Optional for debugging
		return nil, err
	}

	if len(resp.Data) == 0 || len(resp.Data[0].Embedding) == 0 {
		// log.Println("No embedding data received from OpenAI") // Optional
		return nil, errors.New("no embedding data received")
	}

	// Convert []float32 to []float64
	float32Embedding := resp.Data[0].Embedding
	float64Embedding := make([]float64, len(float32Embedding))
	for i, v := range float32Embedding {
		float64Embedding[i] = float64(v)
	}

	return float64Embedding, nil
}
