install_tooling:
	@go install github.com/swaggo/swag/cmd/swag@latest

swag:
	@go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g cmd/backend/main.go -o api

.PHONY: testifai-cli
testifai-cli:
	go build -o bin/testifai-cli ./cmd/frontend/main.go