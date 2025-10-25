package main

import (
	"log"

	"github.com/myjupyter/testifai/src/backend/app/ai/openai_compatible"
	"github.com/myjupyter/testifai/src/backend/config"
	"github.com/myjupyter/testifai/src/backend/server"
	"github.com/myjupyter/testifai/src/backend/service/prompt_builder"
	"github.com/myjupyter/testifai/src/backend/service/router"
)

// @title          Testifai backend
// @version         1.0
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  iauglov@gmail.com
// @BasePath  /
func main() {
	cfg := &config.Config{
		Host:        "127.0.0.1",
		ListenAddr:  ":6667",
		LLMEndpoint: "http://llm-manager.k.avito.ru/v1/chat/completions",
		LLMAPIKey:   "sk-e59959b5-483c-4671-abc1-bf0be292ab6f",
		LLMProvider: "openai",
	}

	promptBuilderSrv, err := prompt_builder.New()
	if err != nil {
		log.Fatal(err)
	}
	routerSrv, err := router.New(
		promptBuilderSrv,
		openai_compatible.New(
			cfg.LLMProvider,
			cfg.LLMEndpoint,
			"Qwen/Qwen3-Coder-480B-A35B-Instruct-FP8",
			cfg.LLMAPIKey,
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	handler := server.NewHandler(routerSrv)
	srv := server.NewServer(cfg, handler)
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
