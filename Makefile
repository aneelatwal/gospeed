BINARY_NAME=gospeed
VERSION?=$(shell git describe --tags --always)
BUILD_DIR=bin

build:
	npx tailwindcss -i ./internal/web/tailwind.css -o ./internal/web/style.css --minify
	go build -o gospeed ./cmd/server
	@echo "🚀 Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/server
	@echo "🚀 Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/server

run:
	./gospeed

clean:
	rm -rf $(BUILD_DIR)

release: build
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)