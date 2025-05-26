package gemini

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Gemini struct {
	APIKey string `yaml:"api_key" json:"api_key"`
	Model  string `yaml:"model" json:"model"`
}

type GenAIConfig struct {
	Gemini Gemini `yaml:"gemini" `
	client *genai.Client
	model  *genai.GenerativeModel
	Role   string `yaml:"role" json:"role"`
}

type AgentAI interface {
	Close() error
	Short(ctx context.Context, events *string) (*string, error)
}

func NewGeminiAgent(config *GenAIConfig, query *string) (AgentAI, error) {

	gai := config
	var err error
	if err = gai.validate(); err != nil {
		return nil, err
	}

	if query == nil || *query == "" {
		return nil, errors.New("query cannot be empty or nil")
	}

	ctx := context.Background()
	gai.client, err = genai.NewClient(ctx, option.WithAPIKey(gai.Gemini.APIKey))
	if err != nil {
		log.Fatalf("Erro ao criar cliente GenAI: %v", err)
	}

	gai.model = gai.client.GenerativeModel(gai.Gemini.Model)
	gai.Role = *query

	return gai, err
}

func (gai *GenAIConfig) validate() error {

	if gai == nil || gai.Gemini.APIKey == "" {
		return errors.New("genai config cannot be empty or nil")
	}

	switch gai.Gemini.Model {
	case "gemini-2.0-flash":
		gai.Gemini.Model = "gemini-2.0-flash"
	case "gemini-2.0-pro":
		gai.Gemini.Model = "gemini-2.0-pro"
	default:
		gai.Gemini.Model = "gemini-2.0-flash"
	}

	return nil
}

func (gai *GenAIConfig) Close() error {
	return gai.client.Close()
}

func (gai *GenAIConfig) Short(ctx context.Context, events *string) (*string, error) {

	if events == nil || *events == "" {
		return nil, errors.New("events cannot be empty or nil")
	}

	session := gai.model.StartChat()
	session.History = []*genai.Content{
		{
			Parts: []genai.Part{genai.Text(gai.Role)},
			Role:  "model",
		},
	}

	inputForLLM := fmt.Sprintf("De %s: %s", "user", *events)

	resp, err := session.SendMessage(ctx, genai.Text(inputForLLM))
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no response from model")
	}

	response := fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])

	return &response, nil

}
