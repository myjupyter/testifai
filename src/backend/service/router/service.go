package router

import (
	"context"

	"github.com/myjupyter/testifai/src/backend/app/model"
	"github.com/myjupyter/testifai/src/backend/service/prompt_builder"
)

type Provider interface {
	ProviderName() string
	Generate(ctx context.Context, form model.AiGenerateForm) (model.AiGenerateResult, error)
}

type Service struct {
	promptBuilderSrv *prompt_builder.Service
	provider         Provider
}

func New(
	promptBuilderSrv *prompt_builder.Service,
	provider Provider,
) (*Service, error) {
	return &Service{
		provider:         provider,
		promptBuilderSrv: promptBuilderSrv,
	}, nil
}

func (s *Service) Generate(ctx context.Context, request model.GenerateRequest) (model.AiGenerateResult, error) {
	prompt, err := s.promptBuilderSrv.BuildPrompt(request)
	if err != nil {
		return model.AiGenerateResult{}, err
	}
	generate, err := s.provider.Generate(ctx, model.AiGenerateForm{
		Prompt:       prompt,
		SystemPrompt: "",
	})

	if err != nil {
		return model.AiGenerateResult{}, err
	}
	return generate, nil

}
