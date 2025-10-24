install_tooling:
	@go install github.com/swaggo/swag/cmd/swag@latest

swag:
	@go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g cmd/backend/main.go -o api