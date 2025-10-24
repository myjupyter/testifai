package router

import (
	"context"
	"errors"
	"fmt"

	"github.com/myjupyter/testifai/src/backend/app/model"
	"github.com/myjupyter/testifai/src/backend/service/prompt_builder"
)

type Provider interface {
	ProviderName() string
	Generate(ctx context.Context, form model.AiGenerateForm) (model.AiGenerateResult, error)
}

type Service struct {
	providers        []Provider
	promptBuilderSrv *prompt_builder.Service
}

func New(promptBuilderSrv *prompt_builder.Service, providers ...Provider) (*Service, error) {
	if len(providers) == 0 {
		return nil, errors.New("providers not provided")
	}
	return &Service{
		providers:        providers,
		promptBuilderSrv: promptBuilderSrv,
	}, nil
}

func (s *Service) Generate(ctx context.Context, request model.GenerateRequest) (model.AiGenerateResult, error) {
	for _, provider := range s.providers {
		if provider.ProviderName() == request.Provider {
			prompt, err := s.promptBuilderSrv.BuildPrompt(request)
			if err != nil {
				return model.AiGenerateResult{}, fmt.Errorf("router: build prompt err: %w", err)
			}
			generate, err := provider.Generate(ctx, model.AiGenerateForm{
				ApiKey:       request.ApiKey,
				Prompt:       prompt,
				SystemPrompt: "",
			})
			if err != nil {
				return model.AiGenerateResult{}, err
			}
			return generate, nil
		}
	}

	return model.AiGenerateResult{}, fmt.Errorf("router: provider [%s] not found", request.Provider)

}
