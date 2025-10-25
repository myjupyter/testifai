PATH := $(PATH):$(PWD)/bin

install_tooling:
	@go install github.com/swaggo/swag/cmd/swag@latest

swag:
	@go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g cmd/backend/main.go -o api

.PHONY: build 
build:
	go build -o bin/testifai ./cmd/frontend/main.go

.PHONY: test
test:
	go generate ./...