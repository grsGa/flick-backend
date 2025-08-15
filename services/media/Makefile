# Media Service Makefile

.PHONY: help build test test-unit test-integration clean proto docker run dev

# Default target
help:
	@echo "Available targets:"
	@echo "  build           - Build the media service binary"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration- Run integration tests only"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  clean           - Clean build artifacts"
	@echo "  proto           - Generate protobuf files"
	@echo "  docker          - Build Docker image"
	@echo "  run             - Run the service locally"
	@echo "  dev             - Run in development mode with hot reload"
	@echo "  lint            - Run linter"
	@echo "  format          - Format code"

# Build the service
build:
	@echo "Building media service..."
	go build -o bin/media-service main.go

# Run all tests
test:
	@echo "Running all tests..."
	go test -v ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v -run Integration ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/media.proto

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t flick/media-service -f ../../deploy/docker/Dockerfile ../../

# Run the service locally
run: build
	@echo "Running media service..."
	./bin/media-service

# Development mode with environment variables
dev:
	@echo "Running in development mode..."
	@export MEDIA_SERVICE_PORT=50054 && \
	export MINIO_ENDPOINT=localhost:9000 && \
	export MINIO_ACCESS_KEY=minioadmin && \
	export MINIO_SECRET_KEY=minioadmin123 && \
	export MINIO_USE_SSL=false && \
	export MINIO_BUCKET_NAME=flick-media && \
	export CONSUL_AGENT_ADDR=localhost:8500 && \
	go run main.go

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	mockgen -source=internal/repository/interfaces.go -destination=internal/repository/mocks/mock_repository.go
	mockgen -source=internal/storage/interfaces.go -destination=internal/storage/mocks/mock_storage.go

# Run tests in Docker environment
test-docker:
	@echo "Running tests in Docker environment..."
	docker-compose -f ../../deploy/docker/docker-compose.test.yml up --build --abort-on-container-exit

# Database migrations (if needed in future)
migrate-up:
	@echo "Running database migrations..."
	# migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up

migrate-down:
	@echo "Rolling back database migrations..."
	# migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" down

# Health check
health:
	@echo "Checking service health..."
	grpc_health_probe -addr=localhost:50054 -service=media-service

# Benchmark tests
bench:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./...

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	# Add documentation generation commands here

# All-in-one development setup
setup: deps proto build test
	@echo "Development setup complete!"

# Production build
build-prod:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/media-service main.go
