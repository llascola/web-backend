---
name: ports-and-adapters
description: Architecture guide for implementing new features using the Ports and Adapters (Hexagonal Architecture) pattern. Use this whenever asked to create new domain logic, services, repositories, or external integrations.
metadata:
  author: llascola
  version: "1.0"
---

# Implementing Features with Ports & Adapters

This project strictly follows the **Ports and Adapters** (Hexagonal Architecture) pattern. The core principle is that **dependencies always point inwards** toward the domain logic. The core application knows absolutely nothing about HTTP, PostgreSQL, MinIO, JWTs, or any other external technology.

When adding a new feature or integrating a new technology, follow this strict 5-layer workflow.

## The 5-Layer Workflow

### Layer 1: Core Domain (`internal/app/domain/`)

Start by defining the business entities and pure business rules.

- **What lives here:** Structs, domain validations, state changes (e.g., `User`, `Image`).
- **Rule:** This package CANNOT import anything from outside the standard library or other domain packages. No `gin`, `ent`, `aws-sdk`, `jwt`, etc.

### Layer 2: Drive the Application (`internal/app/inports/`)

Define what the application *can do* for the outside world.

- **What lives here:** Go interfaces defining use cases (e.g., `UserService`, `ImageService`).
- **Rule:** Input and output parameters must be domain entities, standard Go types, or dedicated DTOs defined in this package. Never use HTTP-specific types like `*gin.Context` (unless explicitly passed down for tracing/context cancellation, though prefer `context.Context`).

### Layer 3: Define External Needs (`internal/app/outports/`)

Define what the application *needs* from the outside world to do its job.

- **What lives here:** Go interfaces defining secondary actions like saving data or generating tokens (e.g., `UserRepository`, `TokenGenerator`).
- **Rule:** Again, these interfaces only speak the domain language. `UserRepository.Save(user *domain.User)` is correct. `UserRepository.Save(user *ent.User)` is a violation.

### Layer 4: Implement Business Logic (`internal/app/services/`)

Implement the In-Ports by orchestrating Domain logic and Out-Ports.

- **What lives here:** Structs that implement the `inports` interfaces (e.g., `AuthServiceImpl`).
- **Rule:** Services cannot directly import infrastructure libraries (no PostgreSQL, no JWT libraries). If a service needs to interact with the outside world, it must call an `outports` interface.

### Layer 5: Build Adapters (`internal/adapters/`)

Connect the outside world to the defined Ports.

1. **Driving / Primary Adapters (`internal/adapters/driving/`)**: These *use* the In-Ports. For HTTP APIs (like Gin), they receive requests, validate transport formats, and call the service layer (e.g., `rest/handlers/`).
2. **Driven / Secondary Adapters (`internal/adapters/driven/`)**: These *implement* the Out-Ports. This is where you put specific libraries.
   - Database layer using Ent? Put it in `adapters/driven/repository/postgres/`.
   - JWT generation using jwt-go? Put it in `adapters/driven/security/jwt_generator.go`.
   - File storage using MinIO? Put it in `adapters/driven/storage/minio_adapter.go`.

## Final Step: Dependency Wiring (`internal/app/app.go`)

Everything is wired together in the composition root (`NewApplication`).

When you create a new Service or Driven Adapter, you must wire it here:

1. Instantiate the specific Driven Adapter (e.g., `postgres.NewUserRepository(client)`).
2. Inject the Driven Adapter into the Service (e.g., `services.NewUserService(userRepo)`).
3. Expose the Service for Driving Adapters to use (e.g., attach it to `app.Service`).

## Quick Checklist for Violations

🔴 **Violation:** A file in `internal/app/services/` imports a specific database ORM (like `ent`).
✅ **Fix:** Create a `Repository` interface in `outports`, and create the Ent implementation in `adapters/driven/repository/`. The service calls the interface.

🔴 **Violation:** An entity in `internal/app/domain/` has JSON, DB, or XML struct tags (unless strictly for internal domain serialization).
✅ **Fix:** Move the struct tags to mapping structures defined in the appropriate adapter (e.g., the Ent schema or Go OpenAPI types). Map from domain to transport objects.

🔴 **Violation:** The `AuthService` generates JWT tokens directly using `github.com/golang-jwt/jwt/v5`.
✅ **Fix:** Create a `TokenGenerator` out-port interface, and move the JWT logic to an adapter in `internal/adapters/driven/security/`.
