package openai_compatible

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	systemPrompt = ""
)

type Client struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

func NewClient(apiKey, endpoint string) *Client {
	return &Client{
		apiKey:   apiKey,
		endpoint: endpoint,
		client:   &http.Client{Timeout: 60 * time.Second},
	}
}

// AnalyzeResult содержит результат анализа AI и usage info
type AnalyzeResult struct {
	Content          string
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Строгие структуры для парсинга ответа OpenAI

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
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

func (c *Client) Analyze(ctx context.Context, message string, imageDatas [][]byte) (*AnalyzeResult, error) {
	content := []map[string]interface{}{}

	if message != "" {
		content = append(content, map[string]interface{}{
			"type": "text",
			"text": message,
		})
	}
	for _, imageData := range imageDatas {
		if len(imageData) == 0 {
			continue
		}
		b64 := base64.StdEncoding.EncodeToString(imageData)
		content = append(content, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": "data:image/png;base64," + b64,
			},
		})
	}

	messages := []map[string]interface{}{
		// {"role": "system", "content": SystemPrompt},
		{"role": "user", "content": content},
	}

	payload := map[string]interface{}{
		"model":             "gpt-4o",
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
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", c.endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var result openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &AnalyzeResult{
		Content:          result.Choices[0].Message.Content,
		Model:            result.Model,
		PromptTokens:     result.Usage.PromptTokens,
		CompletionTokens: result.Usage.CompletionTokens,
		TotalTokens:      result.Usage.TotalTokens,
	}, nil
}
