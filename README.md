# go-starter-restapi

A small, batteries-included starter for building REST APIs in Go. It ships with a
layered architecture (handler → service → repository), Postgres access via `GORM`,
embedded SQL migrations, request validation, structured logging, graceful
shutdown, and interactive OpenAPI docs — using a `rooms` resource as a worked
example.

## Stack

| Concern        | Library |
| -------------- | ------- |
| Router         | [`go-chi/chi`](https://github.com/go-chi/chi) |
| Database       | [`GORM`](https://gorm.io) (with postgres driver) |
| Migrations     | [`golang-migrate`](https://github.com/golang-migrate/migrate) (embedded sources) |
| Validation     | [`go-playground/validator`](https://github.com/go-playground/validator) |
| Config         | [`caarlos0/env`](https://github.com/caarlos0/env) + [`joho/godotenv`](https://github.com/joho/godotenv) |
| API docs       | [`swaggo/http-swagger`](https://github.com/swaggo/http-swagger) (Swagger UI) |
| Logging        | `log/slog` (JSON) |
| Testing        | [`stretchr/testify`](https://github.com/stretchr/testify) |

Requires **Go 1.25+** and a **PostgreSQL** database.

## Project layout

```
cmd/
  api/         HTTP server entrypoint
  migrate/     migration CLI (up/down/version/force)
internal/
  app/         dependency wiring (pool, repo, service, handler, server)
  config/      environment config loading & validation
  domain/      entities, DTOs, response envelopes, typed errors
  handler/     HTTP handlers + JSON response helpers
  repository/  data access (GORM) + migrator wiring
  server/      router, middleware, server lifecycle
  service/     business logic
migrations/    embedded SQL migrations (*.sql)
docs/          embedded OpenAPI spec + Swagger UI route
```

## Getting started

### 1. Configure

Copy the example env file and fill in your database URL:

```sh
cp .env.example .env
```

| Variable           | Required | Default       | Description |
| ------------------ | -------- | ------------- | ----------- |
| `PORT`             | no       | `8080`        | HTTP listen port |
| `DATABASE_URL`     | **yes**  | —             | `postgres://user:pass@host:port/dbname` |
| `TEST_DATABASE_URL`| **yes**  | —             | `postgres://user:pass@host:port/dbname_test` |
| `DB_MAX_CONNS`     | no       | `10`          | Max pooled connections |
| `DB_MAX_IDLE_TIME` | no       | `15m`         | Max connection idle time |
| `ALLOWED_ORIGINS`  | no       | `*`           | Comma-separated CORS origins, e.g. `https://app.com,https://admin.app.com` |
| `ENVIRONMENT`      | no       | `development` | `development` \| `staging` \| `production` |
| `LOG_LEVEL`        | no       | `info`        | `debug` \| `info` \| `warn` \| `error` |
| `JWT_SECRET`       | **yes**  | —             | Secret key for signing JWT tokens |
| `JWT_EXPIRY_HOURS` | no       | `24`          | Token expiry time in hours |

### 2. Run migrations

```sh
go run ./cmd/migrate up        # apply all pending migrations
go run ./cmd/migrate down      # roll back one step
go run ./cmd/migrate version   # print current version
go run ./cmd/migrate force 1   # force-set version (recover from a dirty state)
```

Migrations live in `migrations/` and are embedded into the binary at build time.

### 3. Run the server

```sh
go run ./cmd/api
```

The server starts on `:$PORT` (8080 by default) and shuts down gracefully on
`SIGINT`/`SIGTERM`.

## API

Base path: `/api`. Interactive docs are served from the running app:

- **Swagger UI:** http://localhost:8080/api/docs/
- **Raw spec:** http://localhost:8080/api/docs/openapi.yaml

| Method   | Path                      | Description |
| -------- | ------------------------- | ----------- |
| `GET`    | `/api/healthz`            | Liveness check (returns `ok`) |
| `POST`   | `/api/register`           | Register a new user |
| `POST`   | `/api/login`              | Login, returns JWT token |
| `POST`   | `/api/logout`             | 🔒 Revoke current JWT token |
| `GET`    | `/api/rooms`              | List rooms (paginated via `?page=&limit=`) |
| `GET`    | `/api/rooms/cursor`       | List rooms (cursor pagination via `?cursor=&limit=`) |
| `POST`   | `/api/rooms`              | Create a room |
| `GET`    | `/api/rooms/{id}`         | Get a room by ID |
| `PUT`    | `/api/rooms/{id}`         | Update a room |
| `DELETE` | `/api/rooms/{id}`         | Soft-delete a room (sets `deleted_at`) |
| `POST`   | `/api/rooms/{id}/recover` | 🔒 Recover a soft-deleted room |

Deletes are soft: rows are marked with `deleted_at` and filtered out of reads,
and can be restored via the recover endpoint.

### Authentication

Register a new user, then log in. A default admin user (`admin` / `admin123`)
is seeded by the migration.

```sh
# Register
curl -X POST http://localhost:8080/api/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'

# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

curl -X POST "http://localhost:8080/api/rooms/1/recover" \
  -H "Authorization: Bearer $TOKEN"
```

Revoke a token via `/api/logout`:

```sh
curl -X POST http://localhost:8080/api/logout \
  -H "Authorization: Bearer $TOKEN"
```

Blacklist is in-memory (lost on restart).

### Response envelopes

Success responses are wrapped:

```json
{ "status": 200, "data": { "id": 1, "name": "Room A", "description": "..." } }
```

List responses add pagination metadata:

```json
{
  "status": 200,
  "data": [ { "id": 1, "name": "Room A", "description": "..." } ],
  "pagination": { "page": 1, "limit": 10, "total_pages": 5, "total": 42, "has_prev": false, "has_next": true }
}
```

Cursor-paginated list responses:

```json
{
  "status": 200,
  "data": [ { "id": 10, "name": "Room A", "description": "..." } ],
  "pagination": { "next_cursor": 5, "has_next": true, "limit": 10 }
}
```

Errors return:

```json
{ "error": "resource not found" }
```

Status codes follow HTTP semantics: `400` for invalid input/validation, `401`
for missing or invalid authentication, `404` when a resource doesn't exist,
`409` for conflicts (e.g. duplicate username, recovering a room that isn't
deleted), `500` for unexpected failures (the underlying error is logged,
not returned to the client).

### Example

```sh
curl -X POST http://localhost:8080/api/rooms \
  -H 'Content-Type: application/json' \
  -d '{"name":"Conference Room A","description":"Large room on the 3rd floor"}'
```

## Testing

Integration tests in `internal/handler` use [`testify`](https://github.com/stretchr/testify) for assertions and drive the real router → handler → service
→ repository → GORM stack via `httptest` against a real Postgres. They read
`TEST_DATABASE_URL` and **skip** when it is unset, so `go test ./...` stays green
without a database.

> The test database is migrated and `TRUNCATE`d between tests — point it at a
> throwaway DB, never your real one.

```sh
TEST_DATABASE_URL='postgres://user:pass@localhost:5432/app_test?sslmode=disable' \
  go test ./internal/handler -v
```

`TEST_DATABASE_URL` may also be set in `.env` (it is loaded automatically). The
suite covers health, full CRUD happy paths, soft-delete and recover (including
the `409` conflict and `404` cases), validation `400`s, `404`s, non-numeric ids,
and list pagination edges.

## Development

```sh
go build ./...   # build everything
go vet ./...     # static checks
go test ./...    # run tests (integration tests skip without TEST_DATABASE_URL)
```

## Adding a new resource

This project is meant to be copied and extended. To add a resource, mirror the
`rooms` example across the layers:

1. `internal/domain` — entity + DTOs.
2. `internal/repository` — interface + GORM implementation.
3. `internal/service` — business logic.
4. `internal/handler` — HTTP handlers.
5. `internal/server/routes.go` — register routes.
6. `migrations/` — add the `*.up.sql` / `*.down.sql` pair.
7. `docs/` — document the endpoints in the OpenAPI spec.
