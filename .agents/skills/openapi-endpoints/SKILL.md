---
name: openapi-endpoints
description: The complete workflow for adding new API endpoints to the backend. Use whenever the user asks to create, add, or implement a new API route, handler, or endpoint in this project.
metadata:
  author: llascola
  version: "1.0"
---

# Creating OpenAPI Endpoints

This project uses an **OpenAPI-first code generation workflow**. You must NEVER manually wire routes in `router.go`. The source of truth for all API definitions is the OpenAPI spec.

Follow these 3 steps exactly to add a new endpoint.

## Step 1: Define the Spec

The entire OpenAPI specification is centralized in a single file to guarantee robust code generation and avoid `$ref` resolution bugs across tools.

1. **Open Spec File**: Open `openapi/openapi.yml`.
2. **Schemas (Bottom)**: First, define your request and response payloads under `components/schemas`.
3. **Paths (Middle)**: Add your endpoint path and HTTP method under `paths`. Reference your newly created schemas using `$ref: '#/components/schemas/<SchemaName>'`.
4. **Security**: If the endpoint needs authentication, add `security: - BearerAuth: []` to the path method. For admin-only endpoints, use `security: - BearerAuth: [admin]`.

Example path addition (`openapi/openapi.yml`):
```yaml
paths:
  /api/profile:
    get:
      summary: Get current user profile
      operationId: GetProfile
      tags:
        - users
      security:
        - BearerAuth: []  # Requires JWT, any role
      responses:
        '200':
          description: User profile
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserProfile'
```

## Step 2: Regenerate Code

Always run the code generation after modifying ANY spec file.

```bash
make openapi
```

This updates:
- `internal/adapters/driving/rest/openapi/types.gen.go` (Data models)
- `internal/adapters/driving/rest/openapi/server.gen.go` (Server interfaces and routing hooks)

## Step 3: Implement the Handler

The Go compiler will now fail because the `Handler` struct no longer fully implements `ServerInterface`. Implement the required method in the appropriate `handlers/<domain>_handler.go` file.

**Rules for Handlers:**
1. **Typed Responses**: Use the generated structs from the `openapi` package for all JSON responses (e.g. `openapi.UserProfile`) instead of `gin.H` maps.
2. **Auth Context**: Security middleware automatically validates JWTs and enforces RBAC based on the OpenAPI spec. Handlers can safely retrieve user information from the context:
   ```go
   userIDStr, _ := ctx.Get("userID") // String
   userRole, _ := ctx.Get("role")    // String
   ```
3. **HTTP Status**: Ensure you return the same HTTP status codes as defined in the OpenAPI spec.

## Common Pitfalls & Edge Cases

- **"Type not found" during Go Compilation**: This happens if you add a new route to `openapi.yml` but forget to run `make openapi` before writing the Go handler, or if the generator failed silently due to a YAML syntax error. Always check the output of `make openapi`.
- **Swagger UI grouping mismatch**: Ensure the `tags` array on your operation matches the root `tags` defined at the top of `openapi.yml` to group them correctly in the `/docs` UI.
- **Incorrect `$ref` paths**: Always use the exact `#` pattern (`$ref: '#/components/schemas/MyModel'`) to reference components. Do not use relative file paths.
