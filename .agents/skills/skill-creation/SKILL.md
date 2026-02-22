---
name: skill-creation
description: A step-by-step guide for creating new Agent Skills following the agentskills.io specification. Use when the user asks to create, scaffold, or define a new skill for this project.
metadata:
  author: llascola
  version: "1.0"
  spec-url: "https://agentskills.io/specification"
---

# Creating New Skills

Follow this guide when creating a new skill for the project. All skills MUST comply with the [Agent Skills specification](https://agentskills.io/specification).

## 1. Directory Structure

Skills live in `.agents/skills/`. Each skill is a folder containing at minimum a `SKILL.md` file:

```sh
.agents/skills/
└── my-skill-name/
    ├── SKILL.md          # Required — instructions + frontmatter
    ├── scripts/          # Optional — helper scripts
    ├── references/       # Optional — detailed docs (REFERENCE.md, etc.)
    └── assets/           # Optional — templates, images, data files
```

## 2. SKILL.md Format

### Frontmatter (required)

Every `SKILL.md` must start with YAML frontmatter:

```yaml
---
name: my-skill-name
description: What this skill does and when to use it. Include keywords that help agents identify relevant tasks.
metadata:
  author: llascola
  version: "1.0"
---
```

**Frontmatter rules:**
- `name` — **required**, 1-64 chars, lowercase alphanumeric + hyphens, must match folder name
- `description` — **required**, 1-1024 chars, describe what + when
- `license` — optional, keep short
- `compatibility` — optional, only if there are specific environment requirements
- `metadata` — optional, map of string keys to string values

**Name validation:**
- ✅ `my-skill`, `code-review`, `tdd-workflow`
- ❌ `My-Skill` (uppercase), `-pdf` (starts with hyphen), `pdf--processing` (consecutive hyphens)

### Body Content

After the frontmatter, write clear markdown instructions. Include:

1. **Step-by-step instructions** — numbered, actionable
2. **Examples** — concrete inputs and expected outputs
3. **Rules table** — project-specific constraints
4. **Edge cases** — common pitfalls and how to handle them

## 3. Progressive Disclosure

Skills are loaded in 3 stages — design content accordingly:

| Stage | What loads | Budget |
|---|---|---|
| 1. Metadata | `name` + `description` from frontmatter | ~100 tokens |
| 2. Instructions | Full `SKILL.md` body | < 5000 tokens recommended |
| 3. Resources | `scripts/`, `references/`, `assets/` | On demand only |

**Key principle**: Keep `SKILL.md` concise. Put detailed references in `references/` and link to them:

```md
See [the full API reference](references/REFERENCE.md) for details.
```

## 4. Checklist for New Skills

Before committing a new skill, verify:

- [ ] Folder name matches the `name` field in frontmatter
- [ ] Folder is inside `.agents/skills/`
- [ ] `description` explains both **what** and **when**
- [ ] Body content is under 5000 tokens
- [ ] Instructions are step-by-step and actionable
- [ ] Examples show concrete inputs and outputs
- [ ] Optional dirs (`scripts/`, `references/`, `assets/`) only created if needed

## 5. Existing Skills in This Project

| Skill | Purpose |
|---|---|
| `bulletproof-architecture` | Feature-based directory structure + adapter pattern |
| `bulletproof-components` | Component composition + styling guidelines |
| `bulletproof-performance` | Rendering optimization + state efficiency |
| `bulletproof-security` | Auth, RBAC, token management |
| `bulletproof-state` | Form state, server cache, app state |
| `bulletproof-testing` | Unit, integration, E2E testing guidelines |
| `tdd-workflow` | Red → Green → Refactor cycle with Vitest/RTL/MSW |
