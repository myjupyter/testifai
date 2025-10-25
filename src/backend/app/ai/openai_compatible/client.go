package openai_compatible

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"time"

	"github.com/myjupyter/testifai/src/backend/app/model"
)

const (
	systemPrompt = ""
)

type Client struct {
	providerName string
	endpoint     string
	model        string
	apiKey       string
	client       *http.Client
}

func New(providerName, endpoint, model, apiKey string) *Client {
	return &Client{
		providerName: providerName,
		endpoint:     endpoint,
		model:        model,
		apiKey:       apiKey,
		client:       &http.Client{Timeout: 60 * time.Second},
	}
}

// Строгие структуры для парсинга ответа OpenAI

type openAIUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type openAIChoice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type openAIResponse struct {
	Model   string         `json:"model"`
	Usage   openAIUsage    `json:"usage"`
	Choices []openAIChoice `json:"choices"`
}

func (c *Client) Generate(ctx context.Context, form model.AiGenerateForm) (model.AiGenerateResult, error) {
	content := []map[string]interface{}{}

	content = append(content, map[string]interface{}{
		"type": "text",
		"text": form.Prompt,
	})

	messages := []map[string]interface{}{
		{"role": "user", "content": content},
	}

	if form.SystemPrompt != "" {
		messages = append(messages, map[string]interface{}{"role": "system", "content": form.SystemPrompt})
	}

	payload := map[string]interface{}{
		"model":             c.model,
		"stream":            false,
		"messages":          messages,
		"temperature":       0.2,
		"top_p":             1,
		"presence_penalty":  0,
		"frequency_penalty": 0,
		"n":                 1,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return model.AiGenerateResult{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return model.AiGenerateResult{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return model.AiGenerateResult{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var result openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.AiGenerateResult{}, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return model.AiGenerateResult{}, fmt.Errorf("no choices in response")
	}

	return model.AiGenerateResult{
		Content:          result.Choices[0].Message.Content,
		Model:            result.Model,
		PromptTokens:     result.Usage.PromptTokens,
		CompletionTokens: result.Usage.CompletionTokens,
		TotalTokens:      result.Usage.TotalTokens,
	}, nil
}
func (c *Client) ProviderName() string {
	return c.providerName
}
