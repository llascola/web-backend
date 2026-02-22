---
name: owasp-security
description: OWASP security guidelines and best practices for developing secure applications. Use this when implementing authentication, authorization, data validation, encryption, or security-sensitive features.
metadata:
  version: "1.0"
---

# OWASP Security Guidelines

This skill provides mandatory security practices for backend development, heavily inspired by the OWASP Top 10 and recent security hardening efforts in this project. Apply these rules whenever building or refactoring auth flows, API endpoints, or database interactions.

## 1. Authentication & Session Management

- **Password Hashing**: Always use modern adaptive hashing algorithms like `bcrypt` (with a cost of 12 or higher) or `argon2id`. Do not use MD5 or SHA for passwords.
- **Anti-Enumeration**: Ensure login endpoints return generic error messages (e.g., `invalid email or password`). Execute dummy hash comparisons (constant-time operations) for invalid users to prevent timing attacks.
- **Token Revocation**: JWTs are stateless. Always implement a Token Blocklist (e.g., using Redis with TTL matching the JWT expiration) to instantly revoke tokens on logout or security events.
- **Refresh Tokens**: Use Short-Lived Access Tokens (e.g., 15m) paired with Long-Lived Refresh Tokens (e.g., 7d). Store refresh tokens securely (hashed in the database). Define opaque random strings for refresh tokens rather than encoded JWTs to limit exposure.

## 2. JWT (JSON Web Token) Security

When issuing or validating JWTs:
- **Claims Verification**: Always validate all standard claims:
  - `exp` (Expiration Time): Ensure tokens are short-lived.
  - `iss` (Issuer): Verify the token came from your trusted authorization server.
  - `aud` (Audience): Ensure the token is intended for this specific API.
  - `jti` (JWT ID): Issue unique IDs to track individual tokens and facilitate blocklisting.
- **Signature Algorithm**: Use strong asymmetric algorithms like `RS256` or `ES256`. Do not allow `none` algorithm.

## 3. Data Validation & Injection Prevention

- **Input Validation**: Strictly validate all untrusted input against an OpenAPI schema (or similar strict types) before processing.
- **SQL Injection**: Always use ORMs (like `ent`) or parameterized queries. Never concatenate strings into raw SQL queries.

## 4. Secure API Design

- **Least Privilege**: Protect administrative or sensitive routes with appropriate Role-Based Access Control (RBAC). Use multi-level scopes if applicable.
- **Error Handling**: Do not expose internal system errors, stack traces, or database implementation details in API responses. Over-sharing aids attackers in fingerprinting the system.

## 5. Security Checklist for New Endpoints

Before finalizing any new feature, verify:
- [ ] Is authentication required and enforced via middleware?
- [ ] Are permissions verified at the business logic layer (not just route layer)?
- [ ] Are request payloads rigidly checked against their type schemas?
- [ ] Does the endpoint leak sensitive PII or system architectures in error scenarios?

## Examples

### Good: Generic Login Response (Anti-Enumeration)
```go
func Login(email, password string) error {
    user, err := repo.FindByEmail(email)
    if err != nil {
        domain.CompareDummyPassword(password) // Thwart timing attacks
        return errors.New("invalid credentials")
    }
    if !user.CheckPassword(password) {
        return errors.New("invalid credentials")
    }
    return nil
}
```

### Bad: Leaky Error Response
```go
func Login(email, password string) error {
    user, err := repo.FindByEmail(email)
    if err != nil {
        return errors.New("user not found") // EXPOSES VALIDITY OF EMAIL
    }
    // ...
}
```
