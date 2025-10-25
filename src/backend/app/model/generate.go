package model

type AiGenerateResult struct {
	Content          string
	Model            string
	PromptTokens     int64
	CompletionTokens int64
	TotalTokens      int64
}

type AiGenerateForm struct {
	ApiKey       string
	Prompt       string
	SystemPrompt string
}
