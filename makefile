.PHONY: all build test lint mocks tidy

all: test lint build

run:
	@echo "==> Running application..."
	go run ./cmd/main.go --force_recrawl

build:
	@echo "==> Building application..."
	go build -o ./bin/crawler ./cmd/main.go

test:
	@echo "==> Running tests..."
	go test -v ./...

lint:
	@echo "==> Running linter..."
	golangci-lint run

mocks:
	@echo "==> Generating mocks..."
	go generate ./... 

tidy:
	@echo "==> Tidying go modules..."
	go mod tidy