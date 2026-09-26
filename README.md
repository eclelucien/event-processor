# Event Processor

HTTP API that stores and retrieves events in PostgreSQL. Built as a learning project around clean architecture in Go: domain, application, infrastructure, and HTTP interfaces stay separate.

## Requirements

- Go 1.25+
- Docker (for PostgreSQL)

## Project layout

```
cmd/api/                              # application entrypoint
internal/
  application/                        # use cases and validation
  config/                             # environment loading
  domain/                             # Event entity
  infrastructure/migrations/          # embedded SQL migrations, applied on startup
  infrastructure/postgres/            # database connection and repository
  interfaces/http/                    # HTTP server and handlers
```

## Setup

Start PostgreSQL:

```bash
docker compose up -d
```

Export the required environment variables (all three are mandatory):

```bash
export APP_ENV=development
export HTTP_PORT=8080
export DATABASE_URL="postgres://event_processor:event_processor@localhost:5432/event_processor?sslmode=disable"
```

Run the API:

```bash
go run ./cmd/api
```

The process applies embedded SQL migrations on startup, listens on `HTTP_PORT`, and shuts down on `SIGINT` / `SIGTERM`.

## API

### Health

```http
GET /health
```

```json
{"status":"ok"}
```

### Create event

```http
POST /api/v1/events
Content-Type: application/json
```

```json
{
  "type": "payment.created",
  "source": "revofin",
  "payload": {
    "payment_id": "123",
    "amount": 150.00
  }
}
```

`201 Created`:

```json
{
  "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "type": "payment.created",
  "source": "revofin",
  "payload": {
    "payment_id": "123",
    "amount": 150.00
  },
  "created_at": "2026-09-21T21:00:00.000000000Z"
}
```

`type`, `source`, and `payload` are required. Invalid JSON or missing fields return `400` with `{"error":"..."}`.

### Get event by ID

```http
GET /api/v1/events/{id}
```

`200 OK` returns the same event shape as create. Unknown IDs return `404`; unexpected failures return `500`.

### List events

```http
GET /api/v1/events?type=payment.created&source=revofin&limit=20&offset=0
```

`type` and `source` are optional filters. `limit` defaults to `20` and must be between `1` and `100`. `offset` defaults to `0` and must be zero or greater.

`200 OK` returns a JSON array of events, newest first. An empty result is `[]`. Invalid `limit` or `offset` returns `400`; unexpected failures return `500`.

## Tests

Unit tests (no database):

```bash
go test ./internal/application ./internal/interfaces/http
```

Repository integration tests need a running database and `DATABASE_URL`:

```bash
DATABASE_URL="postgres://event_processor:event_processor@localhost:5432/event_processor?sslmode=disable" \
  go test ./internal/infrastructure/postgres
```

`go test ./...` also runs the integration test, so set `DATABASE_URL` first.
