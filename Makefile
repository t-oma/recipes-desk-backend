.PHONY: build run \ 
		test test-unit \ 
		test-integration test-integration-recipes test-integration-auth test-integration-all test-all \ 
		coverage coverage-integration coverage-recipes coverage-auth coverage-all \
        clean \ 
		docker-up docker-down docker-logs \ 
		install \ 
		swag lint

# =============================================================================
# BUILD
# =============================================================================

# Build application binary
build:
	go build -o bin/api cmd/api/main.go

# Run application locally with hot reload
run:
	go run cmd/api/main.go

# =============================================================================
# UNIT TESTS (fast, no Docker required)
# =============================================================================

# Run all unit tests (excludes integration tests)
test:
	go test -v ./...

# Alias for test (explicit)
test-unit:
	go test -v ./...

# =============================================================================
# INTEGRATION TESTS (slow, requires Docker)
# =============================================================================

# Run ALL integration tests in the project
test-integration:
	go test -v -tags=integration -run '^TestIntegration_' ./...

# Run integration tests for recipes module only
test-integration-recipes:
	go test -v -tags=integration -run '^TestIntegration_' ./internal/modules/recipes/...

test-integration-auth:
	go test -v -tags=integration -run '^TestIntegration_' ./internal/modules/auth/...

# =============================================================================
# ALL TESTS (unit + integration)
# =============================================================================

# Run all tests (unit + integration)
test-all: test test-integration

# =============================================================================
# COVERAGE
# =============================================================================

# Coverage for unit tests only
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Coverage for integration tests only
coverage-integration:
	go test -tags=integration -run '^TestIntegration_' -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Integration coverage report generated: coverage.html"

coverage-recipes:
	go test -tags=integration -run '^TestIntegration_' -coverprofile=coverage.out ./internal/modules/recipes/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Recipes coverage report generated: coverage.html"

coverage-auth:
	go test -tags=integration -run '^TestIntegration_' -coverprofile=coverage.out ./internal/modules/auth/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Auth coverage report generated: coverage.html"

# Coverage for all tests
coverage-all:
	go test -tags=integration -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Full coverage report generated: coverage.html"

# =============================================================================
# CLEANUP
# =============================================================================

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# =============================================================================
# DEPENDENCIES
# =============================================================================

install:
	go mod download
	go mod tidy

install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# =============================================================================
# CODE QUALITY
# =============================================================================

lint:
	golangci-lint run

# Generate swagger documentation
swag:
	swag init -g cmd/api/main.go -o docs/swagger

# =============================================================================
# DOCKER
# =============================================================================

docker-up:
	@if [ ! -f scripts/mongo-keyfile ]; then \
		echo "🔑 Generating MongoDB keyFile..."; \
		openssl rand -base64 756 > scripts/mongo-keyfile && chmod 400 scripts/mongo-keyfile; \
	fi
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-build:
	docker compose build

docker-clean:
	docker compose down -v --remove-orphans

# =============================================================================
# SETUP
# =============================================================================

setup: install install-tools docker-up
	@echo "✅ Setup complete!"
	@echo "API will be available at http://localhost:8080"
	@echo "MongoDB available at localhost:27017"
	@echo "Mongo Express available at http://localhost:8081"
