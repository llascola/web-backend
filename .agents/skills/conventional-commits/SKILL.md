---
name: conventional-commits
description: Rules for writing Conventional Commits when an AI agent creates git commits. Use when committing code changes, generating commit messages, or reviewing commit history.
metadata:
  author: llascola
  version: "1.0"
  spec: "https://www.conventionalcommits.org/en/v1.0.0/"
---

# Conventional Commits for AI Agents

When committing code on behalf of the user, you MUST follow the [Conventional Commits](https://www.conventionalcommits.org) specification.

## 1. Commit Message Format

```
<type>(<scope>): <description>

[optional body]
```

- **type** — required, lowercase, from the allowed list below
- **scope** — required, the feature or area affected (e.g., `auth`, `dashboard`, `api`, `skills`)
- **description** — required, imperative mood, lowercase, no period at end
- **body** — optional, explain *why* not *what* (the diff shows what)

## 2. Allowed Types

| Type | When to use |
|---|---|
| `feat` | New feature or capability |
| `fix` | Bug fix |
| `refactor` | Code restructuring, no behavior change |
| `test` | Adding or updating tests |
| `docs` | Documentation, comments, skills, READMEs |
| `style` | Formatting, whitespace, no logic change |
| `perf` | Performance improvement |
| `build` | Build config, dependencies, tooling |
| `ci` | CI/CD pipeline changes |
| `chore` | Maintenance (deps, configs, cleanup) |

## 3. Rules

- One commit per logical change — do NOT bundle unrelated changes
- Use `!` after scope for **breaking changes**: `feat(auth)!: switch to httpOnly cookies`
- Scope must match feature directories when applicable (`auth`, `dashboard`, `blog`)
- Use `lib` scope for shared infrastructure (`api-client`, `react-query`)
- Use `skills` scope for skill file changes
- Use `ui` scope for shared UI primitives (`src/components/ui/`)
- Keep the description under 72 characters

## 4. Examples

```sh
# Features
feat(auth): add mock login for local development
feat(dashboard): add image upload card component

# Refactors
refactor(api): consolidate adapters into three centralized files
refactor(auth): extract LoginForm into feature component

# Tests
test(dashboard): add TDD tests for StatusWidget and DashboardNav
test(login): add error handling integration tests

# Docs & Skills
docs(skills): update architecture skill with adapter pattern
docs(skills): create conventional-commits skill

# Build & Chore
chore(deps): bump vite to 6.1.0
build(config): add vitest path aliases

# Breaking Changes
feat(auth)!: replace localStorage token with httpOnly cookies
```

## 5. Multi-file Commits

When a single logical change spans multiple files, commit them together with a description that summarizes the intent:

```sh
# Good — one commit for the full adapter refactor
refactor(api): separate mock and production code with adapter pattern

# Bad — splitting a single refactor into micro-commits
refactor(api): create api.mock.ts
refactor(api): create api.prod.ts
refactor(api): update api.ts router
```
