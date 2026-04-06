APP_NAME=rates-service
BIN_DIR=bin
APP_BIN=$(BIN_DIR)/app

.PHONY: build test run lint docker-build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(APP_BIN) ./cmd/app

test:
	go test ./...

run:
	go run ./cmd/app

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(APP_NAME) .
