# AGENTS.md - Recipes Desk Backend

Guidelines for agentic coding agents working in this Go backend repository.

## Project Overview

Go 1.26 backend API with Gin framework, MongoDB, and modular architecture.

- Framework: gin-gonic/gin
- Database: MongoDB (go.mongodb.org/mongo-driver)
- Logging: rs/zerolog
- Config: spf13/viper
- Testing: stretchr/testify, testcontainers for integration tests

## Build, Lint, Test Commands

### Build & Run

```bash
make build          # Build binary to bin/api
make run            # Run with hot reload (go run)
make install        # Download dependencies
make install-tools  # Install swag, golangci-lint
```

### Linting

```bash
make lint           # Run golangci-lint
golangci-lint run   # Alternative
```

### Unit Tests (fast, no Docker)

```bash
make test           # All unit tests
make test-unit      # Alias for test
go test -v ./...    # Direct command
```

### Integration Tests (requires Docker)

```bash
make test-integration              # All integration tests
make test-integration-recipes      # Recipes module only
```

### Coverage

```bash
make coverage           # Unit test coverage
make coverage-integration
make coverage-all
```

### Docker

```bash
make docker-up      # Start MongoDB, Mongo Express
make docker-down    # Stop containers
make docker-logs    # Follow logs
```

## Code Style Guidelines

### Imports

Use gci formatter with this order: standard, default, blank, dot, alias, localmodule.

```go
import (
    "context"
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/rs/zerolog"

    "recipes-desk/internal/modules/recipes/domain"
)
```

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `handler`, `service`, `domain`)
- **Types**: PascalCase, exported with doc comments (e.g., `RecipeService`, `MongoRepository`)
- **Interfaces**: noun or verb+er (e.g., `Repository`, `RecipeService`)
- **Constants**: camelCase for private, PascalCase for exported
- **Errors**: `Err` prefix for sentinel errors (e.g., `ErrNotFound`, `ErrValidation`)

### Struct Tags

Follow tagliatelle rules:

- JSON: camelCase (`json:"cookingTime"`)
- BSON: camelCase (`bson:"cookingTime"`)
- YAML: camelCase
- Mapstructure: kebab-case

### Error Handling

Use sentinel errors in domain layer with `fmt.Errorf` wrapping:

```go
// Define sentinel errors
var (
    ErrNotFound   = errors.New("recipe not found")
    ErrValidation = errors.New("validation error")
)

// Wrap with context
var ErrEmptyTitle = fmt.Errorf("%w: title cannot be empty", ErrValidation)

// Check with errors.Is
if errors.Is(err, domain.ErrNotFound) {
    // handle not found
}
```

### No Global Variables or init()

The linter enforces no global variables and no `init()` functions. Use dependency injection.

## Testing Conventions

### Test Package Naming

Tests MUST use `_test` suffix package for black-box testing:

```go
package handler_test  // NOT package handler
```

### Table-Driven Tests

Use table-driven tests with descriptive names:

```go
func TestHandler_GetByID(t *testing.T) {
    tests := []struct {
        name           string
        id             string
        mockSetup      func(*mockService)
        wantStatusCode int
    }{
        {
            name:           "success",
            id:             recipeID.Hex(),
            mockSetup:      func(m *mockService) { /* ... */ },
            wantStatusCode: http.StatusOK,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

### Integration Tests

Integration tests require build tag and `TestIntegration_` prefix:

```go
//go:build integration

func TestIntegration_MongoRepository_Create(t *testing.T) {
    // uses testcontainers
}
```

### Assertions

Use testify with `require` for fatal assertions and `assert` for non-fatal:

```go
require.NoError(t, err)        // Stops test on failure
assert.Equal(t, expected, got) // Continues on failure
assert.ErrorIs(t, err, domain.ErrNotFound)
```

## Pre-commit Checklist

Before committing, always run:

```bash
make lint   # Must pass with no errors
make test   # All unit tests must pass
```
