.PHONY: build test run docker-build lint clean

BINARY_NAME=gcs-emulator
GOOS?=darwin
GOARCH?=arm64
VERSION?=0.1.0

build:
	@echo "Building $(BINARY_NAME)..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/server

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

run:
	@echo "Starting emulator..."
	go run ./cmd/server

docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

lint:
	@echo "Running linter..."
	golangci-lint run ./...

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
