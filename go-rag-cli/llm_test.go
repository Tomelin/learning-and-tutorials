package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai" // To use its types like EmbeddingResponse
	"github.com/stretchr/testify/assert"
)

func TestGetEmbedding_Success(t *testing.T) {
	// Mock OpenAI API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Check request method and path
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/embeddings", r.URL.Path)

		// 2. Check request body (optional, but good for robustness)
		var reqBody openai.EmbeddingRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Equal(t, []string{"test sentence"}, reqBody.Input)
		assert.Equal(t, openai.AdaEmbeddingV2, reqBody.Model)

		// 3. Send mock response
		mockResponse := openai.EmbeddingResponse{
			Data: []openai.Embedding{
				{
					Embedding: []float32{0.1, 0.2, 0.3},
					Index:     0,
					Object:    "embedding",
				},
			},
			Model: openai.AdaEmbeddingV2,
			Usage: openai.Usage{
				PromptTokens: 10,
				TotalTokens:  10,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(mockResponse)
		assert.NoError(t, err)
	}))
	defer server.Close()

	config := openai.DefaultConfig("dummy-api-key")
	config.BaseURL = server.URL + "/v1" // Point to mock server
	client := openai.NewClientWithConfig(config)

	// Call the refactored GetEmbedding function
	embedding, err := GetEmbedding(client, "test sentence") // Using the refactored GetEmbedding

	assert.NoError(t, err)
	assert.NotNil(t, embedding)
	expectedEmbedding := []float64{0.1, 0.2, 0.3}
	assert.True(t, reflect.DeepEqual(expectedEmbedding, embedding), fmt.Sprintf("Embeddings should be equal. Expected %v, got %v", expectedEmbedding, embedding))
}

func TestGetLLMAnswer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)

		var reqBody openai.ChatCompletionRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)

		assert.Equal(t, openai.GPT3Dot5Turbo, reqBody.Model) // Or model used in GetLLMAnswer
		assert.Len(t, reqBody.Messages, 2)
		assert.Equal(t, openai.ChatMessageRoleSystem, reqBody.Messages[0].Role)
		assert.Equal(t, openai.ChatMessageRoleUser, reqBody.Messages[1].Role)
		assert.Contains(t, reqBody.Messages[1].Content, "test query")
		assert.Contains(t, reqBody.Messages[1].Content, "tag1")

		mockResponse := openai.ChatCompletionResponse{
			ID:      "chatcmpl-test",
			Object:  "chat.completion",
			Created: 1677652288,
			Model:   openai.GPT3Dot5Turbo,
			Choices: []openai.ChatCompletionChoice{
				{
					Index: 0,
					Message: openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: "This is a mock LLM answer.",
					},
					FinishReason: openai.FinishReasonStop,
				},
			},
			Usage: openai.Usage{
				PromptTokens:     50,
				CompletionTokens: 50,
				TotalTokens:      100,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(mockResponse)
		assert.NoError(t, err)
	}))
	defer server.Close()

	config := openai.DefaultConfig("dummy-api-key")
	config.BaseURL = server.URL + "/v1" // Point to mock server
	client := openai.NewClientWithConfig(config)

	// Call the refactored GetLLMAnswer
	answer, err := GetLLMAnswer(client, "test query", []string{"tag1", "tag2"})

	assert.NoError(t, err)
	assert.Equal(t, "This is a mock LLM answer.", answer)
}
