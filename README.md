# Event Processor

HTTP API that stores events in PostgreSQL and publishes them to Kafka. A consumer in the same process reads the `events` topic and logs each event. Built as a learning project around clean architecture in Go: domain, application, infrastructure, and HTTP interfaces stay separate. The application depends on an `EventPublisher` port; the Kafka client stays in infrastructure.

## Requirements

- Go 1.25+
- Docker (for PostgreSQL and Kafka)

## Project layout

```
cmd/api/                              # application entrypoint
internal/
  application/                        # use cases and validation
  config/                             # environment loading
  domain/                             # Event entity
  infrastructure/kafka/               # Kafka producer and consumer
  infrastructure/migrations/          # embedded SQL migrations, applied on startup
  infrastructure/postgres/            # database connection and repository
  interfaces/http/                    # HTTP server and handlers
```

## Setup

Start PostgreSQL and Kafka:

```bash
docker compose up -d
```

Export the required environment variables:

```bash
export APP_ENV=development
export HTTP_PORT=8080
export DATABASE_URL="postgres://event_processor:event_processor@localhost:5432/event_processor?sslmode=disable"
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC=events
export KAFKA_GROUP_ID=event-processor
```

Run the API:

```bash
go run ./cmd/api
```

The process applies embedded SQL migrations on startup, publishes each created event to `KAFKA_TOPIC`, and consumes that topic in the same process. It listens on `HTTP_PORT` and shuts down on `SIGINT` / `SIGTERM`.

`POST /api/v1/events` persists the event and then publishes it. The message key is the event ID. The value is JSON with the same fields as the API response, and `payload` stays as JSON (it is not base64). The consumer logs `event processed` after it reads the message.

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
go test ./internal/application ./internal/interfaces/http ./internal/infrastructure/kafka
```

Repository integration tests need a running database and `DATABASE_URL`:

```bash
DATABASE_URL="postgres://event_processor:event_processor@localhost:5432/event_processor?sslmode=disable" \
  go test ./internal/infrastructure/postgres
```

`go test ./...` also runs the integration test, so set `DATABASE_URL` first.
