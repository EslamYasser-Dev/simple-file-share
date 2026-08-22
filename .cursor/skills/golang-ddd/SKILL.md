---
name: golang-ddd
description: >-
  Build and refactor Go backends using Domain-Driven Design and hexagonal
  architecture. Use when working on Go services in this repo, structuring
  layers, ports/adapters, domain models, or application use cases.
---

# Go DDD — file-share backend

## Layer rules

```
cmd/           → composition root only (wire deps, no logic)
domain/        → entities, value objects, domain errors, ports (interfaces)
application/   → use cases (one service per operation, depends on ports only)
infrastructure/adapters/primary/   → HTTP handlers, DTOs (inbound adapters)
infrastructure/adapters/secondary/ → FS, auth, config (outbound adapters)
```

**Dependency direction:** primary → application → domain ← secondary. Domain imports nothing outside stdlib + own package.

## Domain

- **Entities / value objects** live in `domain/models` or `domain/valueobjects`.
- **Ports** (repository, auth, config) in `domain/ports` — interfaces only.
- **Errors** in `domain/errors` — typed errors (`NotFoundError`, `ValidationError`).
- No HTTP, JSON tags, URLs, or presentation fields in domain models.

## Application (use cases)

- One struct per operation: `ListFilesService`, `UploadService`, etc.
- Constructor takes port interfaces: `NewXxxService(repo ports.FileRepository)`.
- `Execute(...)` is the single entry point.
- Validate input via domain value objects before calling repositories.

## Primary adapters (HTTP)

- Handlers in `infrastructure/adapters/primary/http/handlers/` — thin, one responsibility.
- DTOs in `infrastructure/adapters/primary/http/dto/` — JSON mapping only.
- Shared HTTP helpers in `handlers/common.go`.
- Middleware chain in `server.go`: logging → auth (if enabled) → handler.
- Never put business logic in handlers.

## Secondary adapters

- Implement domain ports; hide OS/HTTP details.
- Path safety enforced at repository boundary (`resolve` must stay inside root).
- Anti-corruption layers for external formats (e.g. `UploadPartAdapter`).

## Conventions

- Use `ports.ConfigProvider` for all config; env keys: `PORT`, `ROOT_DIR` (or `FILE_SHARE_ROOT`), `USERNAME`/`FILE_SHARE_USERNAME`, `PASSWORD`/`FILE_SHARE_PASSWORD`, `JWT_SECRET`, `MAX_UPLOAD_BYTES`, `APP_ENV`, `ENABLE_TLS`, `ENABLE_AUTH`.
- **Zero external dependencies** — stdlib only in `go.mod`.
- File metadata search uses in-memory `FileIndexRepository`; rebuilt on startup, kept in sync via `IndexedFileRepository`.
- Production defaults `ROOT_DIR=/data` when unset.
- Auth disabled in dev (`EnableAuth() == false`), enabled in production.
- Return JSON from API handlers; map domain errors to HTTP status in `respondWithError`.
- Close resources in loop body, never `defer` inside loops.
- JWT `exp`/`iat` use Unix seconds (`Unix()`, not `UnixMilli()`).

## Checklist for new features

1. Domain model / VO + port method if needed
2. Application service (`Execute`)
3. Secondary adapter implementation
4. HTTP handler + DTO
5. Register route + middleware in `server.go`
6. Wire in `cmd/server/main.go`
7. Test domain/adapter boundary (path traversal, auth)
