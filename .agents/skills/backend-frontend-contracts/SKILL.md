---
name: backend-frontend-contracts
description: Defines the single source of truth and workflow for HTTP contracts between backend and frontend using OpenAPI. Use when adding, changing, or consuming any endpoint that crosses the backend/frontend boundary, or when coordinating parallel agents working on both sides.
---

# Backend–Frontend Contracts

This skill formalizes how the backend and frontend interact via HTTP, so that two agents can work in parallel while staying compatible.

## 1. Contract Source of Truth

- The only source of truth for the HTTP contract is the OpenAPI spec:
  - Current location: `backend/openapi/openapi.yml`.
- A contract is defined by:
  - `operationId`
  - HTTP method + path (e.g. `POST /auth/login`)
  - Request schema (body, params, headers)
  - Response schema(s) and status codes
  - Security (auth / roles)

**Rule:** If it is not in the OpenAPI spec, it is not part of the contract.

## 2. Unique Contract Identifiers (operationId)

For every interaction, use the OpenAPI `operationId` as the unique contract ID.

Examples:

- `Login` → `POST /auth/login`, `AuthRequest` → `LoginResponse`
- `UploadImage` → `POST /api/admin/upload-image`, `multipart/form-data` → `UploadImageResponse`

When agents coordinate, they must talk in terms of operationIds:

- “Backend: implement `UploadImage`”
- “Frontend: call `UploadImage` and handle `UploadImageResponse`”

## 3. Backend Agent Rules

When the task is in `backend/` and touches frontend-facing behavior:

1. Add or change endpoints only via the spec
   - Edit `backend/openapi/openapi.yml`:
     - Add or modify `paths[...]` and `components.schemas[...]`
     - Ensure each operation has a stable `operationId`.
2. Regenerate backend code
   - Run the project’s OpenAPI codegen command (for example `make openapi`).
   - Fix all compile errors by implementing generated interfaces.
3. Respect schemas
   - Handlers must:
     - Accept exactly the request shape defined by the spec.
     - Return only the response shapes and status codes defined by the spec.
4. Communicate changes
   - When changing a contract:
     - Mention the affected `operationId`s.
     - Prefer additive changes (new fields, new responses) over breaking ones.
     - For breaking changes, coordinate a versioning or migration strategy.

Never change response shapes or paths in code without updating the OpenAPI spec first.

## 4. Frontend Agent Rules

When the task is in `frontend/` and touches backend calls:

1. Treat the OpenAPI spec as read-only contract
   - Do not invent endpoints, fields, or status codes that are not in the spec.
2. Generate types and client from the spec
   - Use an OpenAPI-to-TypeScript tool (for example `openapi-typescript`, `orval`) to generate:
     - Types for request and response schemas (for example `AuthRequest`, `LoginResponse`).
     - Optionally, typed API client functions.
   - Store generated types in a single contracts module, for example:
     - `frontend/src/contracts/openapi-types.ts`.
3. Use generated types in the API layer
   - In `src/lib/api*.ts` and feature API files:
     - Import contract types instead of redefining interfaces by hand.
     - Use functions that map 1:1 to OpenAPI operations, aligned with `operationId`s.
4. React to contract changes
   - If the backend agent changes `openapi.yml`:
     - Regenerate TypeScript types.
     - Update all usages of the affected `operationId`s and types.

Never bypass the contract with `any` or untyped `fetch` that uses magic strings for URLs and payloads.

## 5. Parallel Agent Workflow

When running a backend agent and a frontend agent in parallel:

1. Agree on the contract first
   - Decide which `operationId`s are in scope.
   - Ensure they are present and correct in `openapi/openapi.yml`.
2. Backend agent focus
   - Implements or updates handlers for those `operationId`s.
   - Ensures responses conform to the schemas.
3. Frontend agent focus
   - Regenerates frontend types from the same spec.
   - Uses those types in API wrappers and components.
4. Verification
   - Backend: tests that the handler behavior matches the spec.
   - Frontend: tests that the client and UI behave according to the contract types.

If spec and generated types compile and tests pass on both sides, the contract is respected.

## 6. When to Use This Skill

Use this skill whenever:

- Adding or changing a backend endpoint that the frontend will call.
- Adding or changing frontend code that calls the backend.
- Running backend and frontend agents in parallel on a feature that crosses the boundary.
- Investigating mismatches between backend responses and frontend expectations.

If in doubt, check or update the OpenAPI spec first, then implement on each side.