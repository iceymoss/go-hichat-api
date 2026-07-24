# Sync API Documentation

You are an API documentation generator for the go-hichat-api project. Your task is to read all Go-Zero `.api` definition files and generate a complete, up-to-date API reference document at `docs/api.md`.

## Steps

### 1. Discover all .api files

Search for all `.api` files under the `apps/` directory:
- `apps/user/api/user.api`
- `apps/user/api/domain.api`
- `apps/social/api/social.api`
- `apps/im/api/im.api`
- `apps/trend/api/trend.api`
- Any other `.api` files that may have been added

Read **every** `.api` file found.

### 2. Parse and extract

From each `.api` file, extract:
- **Service info**: title, author, version from the `info()` block
- **Routes**: HTTP method, path, prefix, group, JWT requirement, `@doc` description, handler name, request type, response type
- **Type definitions**: all `type` blocks including struct fields, JSON tags, comments, and `optional` markers

### 3. Generate `docs/api.md`

Generate a well-structured Chinese-language Markdown document with the following format:

```markdown
# HiChat API Reference

> Auto-generated from `.api` definition files. Do not edit manually.
> Last synced: {current date YYYY-MM-DD}

## Table of Contents
(auto-generate based on services)

---

## User Service (`/api/v1/user`)

### POST /api/v1/user/register - User Registration
(description from @doc)

**Auth:** None / JWT Required

**Request Body:**
| Field | Type | Required | JSON Key | Description |
|-------|------|----------|----------|-------------|
| ... | ... | ... | ... | (from comments) |

**Response:**
| Field | Type | JSON Key | Description |
|-------|------|----------|-------------|
| ... | ... | ... | ... |

---
(repeat for each endpoint)
```

### 4. Documentation rules

- Group endpoints by **service** (User, Social, IM, Trend)
- Within each service, group by **functional module** (e.g., Friend management, Group management, Comment, Like)
- For each endpoint include: HTTP method, full path (prefix + path), description (`@doc`), auth requirement, request fields, response fields
- Mark fields as **Required** or **Optional** based on the `optional` tag in the `.api` file
- Include field comments from the `.api` file as description
- Include shared/common type definitions (like `User`, `Friends`, `Groups`, `Discuss`, `Like`, `ChatLog`, `Conversation`, etc.) in a separate **Data Models** section at the end
- Use Chinese for descriptions where the source `.api` comments are in Chinese
- For enum-like fields (e.g., `sex: 0-unknown, 1-male, 2-female`), include the value mapping from comments

### 5. Diff awareness

Before writing the file:
- If `docs/api.md` already exists, read it first
- After generating the new content, write it to `docs/api.md`
- Summarize what changed (new endpoints, modified fields, removed endpoints) in your response to the user

### 6. Output

Write the generated documentation to `docs/api.md` and report a summary of changes to the user.
