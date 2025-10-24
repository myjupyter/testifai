package gemini

import (
	"context"
	"fmt"

	"github.com/myjupyter/testifai/src/backend/app/model"
	"google.golang.org/genai"
)

type Client struct {
}

const geminiEndpoint = "https://generativelanguage.googleapis.com"

func New() *Client {
	genai.SetDefaultBaseURLs(genai.BaseURLParameters{
		GeminiURL: geminiEndpoint,
	})
	return &Client{}
}

func (c *Client) Generate(ctx context.Context, form model.AiGenerateForm) (model.AiGenerateResult, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  form.ApiKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		return model.AiGenerateResult{}, fmt.Errorf("error create gemeni client: %w", err)
	}

	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-pro",
		genai.Text(form.Prompt),
		nil,
	)

	if err != nil {
		return model.AiGenerateResult{}, fmt.Errorf("error api request: %w", err)
	}

	return model.AiGenerateResult{
		Content:          resp.Text(),
		Model:            resp.ModelVersion,
		PromptTokens:     int64(resp.UsageMetadata.PromptTokenCount),
		CompletionTokens: int64(resp.UsageMetadata.CandidatesTokenCount),
		TotalTokens:      int64(resp.UsageMetadata.TotalTokenCount),
	}, nil

}

func (c *Client) ProviderName() string {
	return "gemini"
}
