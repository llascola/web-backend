---
name: error-handling
description: Guidelines for idiomatic error handling in the Ports and Adapters architecture. Use this skill when creating new endpoints, services, or adapters that need to return or handle errors across architectural boundaries.
metadata:
  author: llascola
  version: "1.0"
---

# Error Handling in Ports & Adapters

Errors in this project flow **inward to outward**, always speaking the Domain language. External layers (HTTP, gRPC) translate Domain errors into transport-specific codes. Services and adapters NEVER import `net/http` or return raw status codes.

## 1. Domain Error Types

All error types live in `internal/app/domain/errors.go`. Each is a pointer-based struct implementing the `error` interface.

| Type | Semantic Meaning | HTTP Translation |
|---|---|---|
| `*ErrValidation` | Business rule failed (bad input, weak password) | `400 Bad Request` |
| `*ErrNotFound` | Requested resource does not exist | `404 Not Found` |
| `*ErrConflict` | State violation (duplicate email, already exists) | `409 Conflict` |
| `*ErrUnauthorized` | Authentication/authorization failed | `401 Unauthorized` |
| `*ErrInternal` | Unexpected system fault (DB down, I/O failure) | `500 Internal Server Error` |

### Creating Errors
Always return pointer instances with a descriptive `Message`:
```go
return &domain.ErrConflict{Message: "user already exists"}
return &domain.ErrNotFound{Message: "user not found"}
return &domain.ErrValidation{Message: "password must be at least 8 characters"}
```

### Domain-Level Sentinel Errors
For reusable validation rules, define package-level typed sentinels:
```go
var ErrPasswordWeak = &ErrValidation{Message: "password must be at least 8 characters"}
var ErrImageTooLarge = &ErrValidation{Message: "image size exceeds maximum limit"}
```

## 2. Service Layer Rules

Services (`internal/app/services/`) orchestrate domain logic and out-ports. They must:

- **Return domain errors** for all business failures, never `errors.New("raw string")`
- **Propagate adapter errors** as-is when they are already typed (e.g. repository returns `&domain.ErrNotFound{}`)
- **Wrap unknown failures** from infrastructure in `&domain.ErrInternal{}` only if the error is unexpected

```go
// ✅ Correct
if err == nil {
    return &domain.ErrConflict{Message: "user already exists"}
}

// ❌ Violation
return errors.New("user already exists")
```

## 3. Adapter Layer Rules

### Driven Adapters (Repositories, External APIs)
Translate infrastructure-specific errors into domain types at the boundary:
```go
// In postgres/user_repository.go
if ent.IsNotFound(err) {
    return nil, &domain.ErrNotFound{Message: "user not found"}
}
return nil, err // Unknown DB errors bubble up as-is
```

### Driving Adapters (HTTP Handlers)
**Never manually pick HTTP status codes for service errors.** Always delegate to the centralized translator:
```go
// ✅ Correct — in any handler
if err := h.authService.Register(ctx, email, password); err != nil {
    HandleError(ctx, err)
    return
}

// ❌ Violation — hardcoded status
ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
```

## 4. The HandleError Translator

Located in `internal/adapters/driving/rest/handlers/errors.go`. Uses `errors.As()` chains to map domain types to HTTP codes:

```go
func HandleError(ctx *gin.Context, err error) {
    var errValidation *domain.ErrValidation
    if errors.As(err, &errValidation) {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": errValidation.Message})
        return
    }
    // ... ErrNotFound -> 404, ErrConflict -> 409, ErrUnauthorized -> 401
    // Fallback: unknown errors -> 500
    ctx.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
}
```

### Adding a New Error Type
1. Define the struct + `Error()` method in `domain/errors.go`
2. Add an `errors.As()` branch in `HandleError`
3. Use it in your services/adapters

## 5. Testing Errors

Always assert the **exact type**, not just the message string:
```go
// ✅ Correct — proves the HTTP layer will map it correctly
var target *domain.ErrConflict
require.ErrorAs(t, err, &target)

// ❌ Weak — passes even if the type is wrong
assert.ErrorContains(t, err, "already exists")
```
