# Recipes Desk Backend

[![CI](https://github.com/t-oma/recipes-desk-backend/actions/workflows/ci.yml/badge.svg)](https://github.com/t-oma/recipes-desk-backend/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.26-blue.svg)](https://golang.org)

A production-ready RESTful API for managing recipes built with Go, Gin, MongoDB, and following Clean Architecture principles.

## Features

- **Authentication & Authorization**
  - JWT-based authentication with access and refresh tokens
  - User registration and login
  - Protected routes with middleware
  - Refresh token rotation (single-use)
  - HTTP-only cookies for token storage
  - Secure password hashing with bcrypt

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
│   └── auth/             # Authentication module
│   │   ├── domain/       # User & RefreshToken entities
│   │   ├── service/      # Auth business logic & JWT service
│   │   ├── repository/   # MongoDB repositories
│   │   ├── handler/      # HTTP handlers & middleware
│   │   └── module.go     # Module initialization
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

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/auth/register` | Register new user | No |
| POST | `/api/v1/auth/login` | User login | No |
| POST | `/api/v1/auth/refresh` | Refresh access token | No |
| POST | `/api/v1/auth/logout` | User logout | Yes |
| GET | `/api/v1/auth/me` | Get current user | Yes |

**Authentication Flow:**

1. **Registration/Login**: Server returns user data and sets two HTTP-only cookies:
   - `access_token` - Short-lived (15 min), used for authentication
   - `refresh_token` - Long-lived (7 days), used to obtain new access tokens

2. **Protected Routes**: Include access token in cookies or use `Authorization: Bearer <token>` header

3. **Token Refresh**: When access token expires, call `/auth/refresh` endpoint with refresh token cookie

4. **Logout**: Clears both cookies

**Testing with cURL:**

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","firstName":"John","lastName":"Doe"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  -c cookies.txt

# Access protected endpoint
curl http://localhost:8080/api/v1/auth/me \
  -b cookies.txt

# Refresh tokens
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -b cookies.txt \
  -c cookies.txt

# Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -b cookies.txt
```

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

**Unit Tests:** Fast, no Docker required (91 tests)

```bash
make test
```

**Integration Tests:** Requires Docker, uses Testcontainers (19 tests)

```bash
make test-integration
```

**All Tests:**

```bash
make test-all
```

### Test Coverage

The project includes comprehensive test coverage:

- **Domain Layer**: Entity validation and business rules
- **Service Layer**: Business logic with mocked dependencies
- **Handler Layer**: HTTP handlers and middleware
- **Repository Layer**: Integration tests with real MongoDB (Testcontainers)

**Auth Module Tests:**
- Unit: 91 test cases covering validation, service logic, handlers, and middleware
- Integration: 19 test cases covering MongoDB repositories

Run specific test suites:

```bash
# Auth module tests only
go test ./internal/modules/auth/...

# Integration tests only
go test -tags=integration ./internal/modules/auth/repository/...

# With coverage
make coverage
```

## Authentication Architecture

### Security Features

- **HTTP-only Cookies**: Tokens stored in HTTP-only cookies (not accessible via JavaScript)
- **Secure Password Hashing**: bcrypt with cost factor 14
- **Token Rotation**: Refresh tokens are single-use and rotated on each refresh
- **TTL Indexing**: Expired refresh tokens automatically cleaned up by MongoDB
- **CORS Protection**: Configured for specific origins with credentials support

### Token Lifecycle

```
┌─────────────────┐     ┌─────────────────┐
│  Access Token   │     │  Refresh Token  │
│   (15 minutes)  │     │   (7 days)      │
│   HTTP-only     │     │   HTTP-only     │
│   Cookie        │     │   Cookie        │
└────────┬────────┘     └────────┬────────┘
         │                       │
         │    ┌─────────────┐    │
         └───►│   MongoDB   │◄───┘
              │  (TTL index)│
              └─────────────┘
```

### Module Structure

```
internal/modules/auth/
├── domain/
│   ├── user.go              # User entity & validation
│   ├── refresh_token.go     # RefreshToken entity
│   └── user_repository.go   # Repository interfaces
├── service/
│   ├── service.go           # Auth business logic
│   ├── tokens.go            # JWT token service
│   └── bcrypt_hasher.go     # Password hashing
├── repository/
│   ├── mongo.go             # User repository (MongoDB)
│   └── mongo_refresh.go     # Refresh token repository
├── handler/
│   ├── handler.go           # HTTP handlers
│   ├── middleware.go        # Auth middleware
│   ├── request.go           # Request DTOs
│   ├── response.go          # Response DTOs
│   └── mapper.go            # DTO mappers
└── module.go                # Module initialization
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
