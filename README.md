# Portfolio Backend API

A production-grade Go backend service built with **Ports & Adapters** (Hexagonal Architecture), **OpenAPI-first** design, and a comprehensive **Test Pyramid**.

## Tech Stack

| Layer | Technology |
|---|---|
| **Language** | Go 1.25 |
| **HTTP Router** | Gin |
| **Database** | PostgreSQL 15 (via Ent ORM) |
| **Cache** | Redis 7 |
| **Object Storage** | MinIO (S3-compatible) |
| **Auth** | JWT (Access + Refresh Tokens, Blocklist) |
| **API Spec** | OpenAPI 3.0 (code-gen via `oapi-codegen`) |
| **Testing** | Testify, Mockery, testcontainers-go |
| **Containerization** | Docker, Docker Compose |

## Architecture

The project follows strict **Ports & Adapters** principles — dependencies always point inward toward the domain:

```
internal/
├── app/
│   ├── domain/          # Pure business entities & validation rules
│   ├── inports/         # Use-case interfaces (what the app CAN do)
│   ├── outports/        # Infrastructure interfaces (what the app NEEDS)
│   └── services/        # Business logic orchestrating domain + outports
├── adapters/
│   ├── driving/         # HTTP handlers, middleware, OpenAPI wiring
│   └── driven/          # Postgres, Redis, MinIO, JWT implementations
└── config/              # Environment-based configuration
```

> The core application knows nothing about HTTP, PostgreSQL, or JWTs. All infrastructure is injected through interfaces.

## API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | — | Health check |
| `POST` | `/auth/register` | — | Register a new user |
| `POST` | `/auth/login` | — | Login and receive tokens |
| `POST` | `/auth/logout` | Bearer | Revoke access token |
| `POST` | `/auth/refresh` | — | Rotate refresh token |
| `GET` | `/api/profile` | Bearer | Get authenticated user profile |
| `DELETE` | `/api/admin/users/:id` | Bearer (Admin) | Delete a user |
| `POST` | `/api/admin/upload-image` | Bearer (Admin) | Upload an image to storage |

Full OpenAPI specification: [`openapi/openapi.yml`](openapi/openapi.yml) · Interactive docs at `/docs` when running.

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose

### 1. Configure Environment

```bash
cp .env.example .env
# Edit .env with your values
```

### 2. Start Infrastructure

```bash
docker compose up -d postgres redis minio
```

### 3. Run the Server

```bash
go run . serve
```

The API will be available at `http://localhost:8080`.

## Development

### Code Generation

```bash
# Regenerate OpenAPI types and server stubs
make openapi

# Regenerate Ent ORM schema
make ent

# Regenerate Mockery test mocks
make mocks
```

### Running Tests

```bash
# Run all tests (unit + E2E + integration)
go test ./...

# Run only fast tests (skip Docker-based integration tests)
go test -short ./...

# Run with verbose output
go test -v ./...
```

## Test Pyramid

The project implements a full **Test Pyramid** strategy:

| Layer | Location | Strategy | Speed |
|---|---|---|---|
| **Unit** (Base) | `services/*_test.go` | Mock-driven via Mockery + Table-driven domain tests | ~ms |
| **E2E** (Middle) | `rest/e2e_*_test.go` | Real Gin router with in-memory fake adapters | ~ms |
| **Integration** (Peak) | `postgres/*_test.go`, `cache/*_test.go` | Ephemeral Docker containers via testcontainers-go | ~2s |

## Error Handling

All errors flow as **typed Domain structs** through the architecture:

| Domain Type | HTTP Status |
|---|---|
| `ErrValidation` | `400 Bad Request` |
| `ErrUnauthorized` | `401 Unauthorized` |
| `ErrNotFound` | `404 Not Found` |
| `ErrConflict` | `409 Conflict` |
| `ErrInternal` | `500 Internal Server Error` |

A centralized `HandleError()` function in the HTTP adapter translates these automatically — handlers never hardcode status codes.

## Docker Deployment

```bash
# Build and run everything
docker compose up --build

# Production-only (multi-stage Alpine image)
docker build -t web-backend .
```

## License

[MIT](LICENSE)
