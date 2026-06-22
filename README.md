# go-starter-restapi

A small, batteries-included starter for building REST APIs in Go. It ships with a
layered architecture (handler → service → repository), Postgres access via `pgx`,
embedded SQL migrations, request validation, structured logging, graceful
shutdown, and interactive OpenAPI docs — using a `rooms` resource as a worked
example.

## Stack

| Concern        | Library |
| -------------- | ------- |
| Router         | [`go-chi/chi`](https://github.com/go-chi/chi) |
| Database       | [`jackc/pgx`](https://github.com/jackc/pgx) (v5, with `pgxpool`) |
| Migrations     | [`golang-migrate`](https://github.com/golang-migrate/migrate) (embedded sources) |
| Validation     | [`go-playground/validator`](https://github.com/go-playground/validator) |
| Config         | [`caarlos0/env`](https://github.com/caarlos0/env) + [`joho/godotenv`](https://github.com/joho/godotenv) |
| API docs       | [`swaggo/http-swagger`](https://github.com/swaggo/http-swagger) (Swagger UI) |
| Logging        | `log/slog` (JSON) |

Requires **Go 1.25+** and a **PostgreSQL** database.

## Project layout

```
cmd/
  api/         HTTP server entrypoint
  migrate/     migration CLI (up/down/version/force)
internal/
  config/      environment config loading & validation
  domain/      entities, DTOs, response envelopes, sentinel errors
  handler/     HTTP handlers + JSON response helpers
  repository/  data access (pgx) + migrator wiring
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
| `DB_MAX_CONNS`     | no       | `10`          | Max pooled connections |
| `DB_MAX_IDLE_TIME` | no       | `15m`         | Max connection idle time |
| `ENVIRONMENT`      | no       | `development` | `development` \| `staging` \| `production` |
| `LOG_LEVEL`        | no       | `info`        | `debug` \| `info` \| `warn` \| `error` |

> Note: `.env.example` currently lists `DB_MAX_CONS` — the variable the app
> actually reads is `DB_MAX_CONNS`.

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

| Method   | Path               | Description |
| -------- | ------------------ | ----------- |
| `GET`    | `/api/healthz`     | Liveness check (returns `ok`) |
| `GET`    | `/api/rooms`       | List rooms (paginated via `?page=&limit=`) |
| `POST`   | `/api/rooms`       | Create a room |
| `GET`    | `/api/rooms/{id}`  | Get a room by ID |
| `PUT`    | `/api/rooms/{id}`  | Update a room |
| `DELETE` | `/api/rooms/{id}`  | Delete a room |

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

Errors return:

```json
{ "error": "resource not found" }
```

Status codes follow HTTP semantics: `400` for invalid input/validation, `404`
when a resource doesn't exist, `500` for unexpected failures (the underlying
error is logged, not returned to the client).

### Example

```sh
curl -X POST http://localhost:8080/api/rooms \
  -H 'Content-Type: application/json' \
  -d '{"name":"Conference Room A","description":"Large room on the 3rd floor"}'
```

## Development

```sh
go build ./...   # build everything
go vet ./...     # static checks
go test ./...    # run tests
```

## Adding a new resource

This project is meant to be copied and extended. To add a resource, mirror the
`rooms` example across the layers:

1. `internal/domain` — entity + DTOs.
2. `internal/repository` — interface + pgx implementation.
3. `internal/service` — business logic.
4. `internal/handler` — HTTP handlers.
5. `internal/server/routes.go` — register routes.
6. `migrations/` — add the `*.up.sql` / `*.down.sql` pair.
7. `docs/` — document the endpoints in the OpenAPI spec.
