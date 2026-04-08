.PHONY: build clean run test

APP_NAME=svnsearch
VERSION=1.0.0
BUILD_DIR=build

build:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(shell date +%Y-%m-%d_%H:%M:%S)" -o $(BUILD_DIR)/$(APP_NAME).exe ./cmd/svnsearch

build-linux:
	go build -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(shell date +%Y-%m-%d_%H:%M:%S)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/svnsearch

clean:
	rm -rf $(BUILD_DIR)

run:
	go run ./cmd/svnsearch

test:
	go test -v ./...

deps:
	go mod download
	go mod tidy
