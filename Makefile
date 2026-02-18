.PHONY: build run test clean docker-up docker-down docker-logs install swag

# Build application
build:
	go build -o bin/api cmd/api/main.go

# Run application locally
run:
	go run cmd/api/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
install:
	go mod download
	go mod tidy

# Install development tools
install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Lint code
lint:
	golangci-lint run

# Generate swagger docs
swag:
	swag init -g cmd/api/main.go -o docs/swagger

# Docker commands
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-build:
	docker compose build

docker-clean:
	docker compose down -v --remove-orphans

# Full development setup
setup: install install-tools docker-up
	@echo "Setup complete! API will be available at http://localhost:8080"
	@echo "MongoDB available at localhost:27017"
	@echo "Mongo Express available at http://localhost:8081"
