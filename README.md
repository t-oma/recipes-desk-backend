# Recipes Desk Backend

[![CI](https://github.com/t-oma/recipes-desk-backend/actions/workflows/ci.yml/badge.svg)](https://github.com/t-oma/recipes-desk-backend/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.26-blue.svg)](https://golang.org)

A production-ready RESTful API for managing recipes built with Go, Gin, MongoDB, and following Clean Architecture principles.

## Features

- **Authentication & Authorization**
  - JWT-based authentication with access and refresh tokens
  - User registration and login
  - Protected routes

- **Recipes Management**
  - Create, read, update, delete recipes
  - Search recipes by title (case-insensitive)
  - Full CRUD operations with validation

- **Architecture**
  - Clean Architecture / Modular Monolith
  - Domain-Driven Design principles
  - Dependency Injection
  - Layered architecture (Domain, Service, Repository, Handler)

## Tech Stack

- **Language:** Go 1.26+
- **Web Framework:** Gin
- **Database:** MongoDB
- **Authentication:** JWT
- **Testing:** Testify, Testcontainers
- **CI/CD:** GitHub Actions
- **Documentation:** Swagger/OpenAPI

## Project Structure

```
internal/
├── modules/
│   ├── recipes/          # Recipes module
│   │   ├── domain/       # Entities and interfaces
│   │   ├── service/      # Business logic
│   │   ├── repository/   # Data access (MongoDB)
│   │   ├── handler/      # HTTP handlers
│   │   └── module.go     # Module initialization
│   └── auth/             # Auth module (WIP)
├── config/               # Configuration
├── infra/                # Infrastructure (DB, logger)
└── server/               # HTTP server setup
```

## Quick Start

### Prerequisites

- Go 1.26+
- Docker & Docker Compose
- Make

### Installation

```bash
# Clone repository
git clone <repository-url>
cd recipes-desk-backend

# Copy environment variables
cp .env.example .env

# Start MongoDB
docker-compose up -d

# Install dependencies
make install

# Run application
make run
```

The API will be available at `http://localhost:8080` by default

## API Endpoints

### Recipes

| Method | Endpoint                           | Description       | Auth |
| ------ | ---------------------------------- | ----------------- | ---- |
| GET    | `/api/v1/recipes`                  | List all recipes  | No   |
| GET    | `/api/v1/recipes/search?q={query}` | Search recipes    | No   |
| GET    | `/api/v1/recipes/{id}`             | Get recipe by ID  | No   |
| POST   | `/api/v1/recipes`                  | Create new recipe | Yes  |
| PUT    | `/api/v1/recipes/{id}`             | Update recipe     | Yes  |
| DELETE | `/api/v1/recipes/{id}`             | Delete recipe     | Yes  |

### Authentication

Coming soon...

## Development

### Available Commands

```bash
# Run application
make run

# Run unit tests
make test

# Run integration tests (requires Docker)
make test-integration

# Run all tests
make test-all

# Generate coverage report
make test-coverage

# Run linter
make lint

# Generate Swagger docs
make swag
```

### Testing

**Unit Tests:** Fast, no Docker required

```bash
make test
```

**Integration Tests:** Requires Docker, uses Testcontainers

```bash
make test-integration
```

**All Tests:**

```bash
make test-all
```

### Code Style

This project follows:

- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- Conventional Commits
- Clean Architecture principles

## CI/CD

The project uses GitHub Actions for continuous integration:

- **Lint:** Runs golangci-lint
- **Unit Tests:** Fast tests without external dependencies
- **Integration Tests:** Full tests with MongoDB in Docker
- **Build:** Verifies application builds successfully

## Author

Artem Levchenko
