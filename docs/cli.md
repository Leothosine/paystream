# CLI Tools

PayStream ships two binaries under `cmd/`.

## paystream-api

Serves the HTTP API.

```bash
go run ./cmd/paystream-api
```

- Listens on `:8080` (configurable via `API_PORT`, see [configuration](configuration.md)).
- `GET /health` — liveness check, backed by `internal/health`.
- `GET /version` — returns `{"version": "0.1.0"}` as JSON.

## paystream-worker

Runs background jobs (Horizon polling, signing queue, webhook delivery).

```bash
go run ./cmd/paystream-worker
```

- Logs a startup timestamp and blocks, processing jobs as they're wired in.
- Intended to run as a long-lived process alongside `paystream-api`.

## Building binaries

```bash
go build -o bin/paystream-api ./cmd/paystream-api
go build -o bin/paystream-worker ./cmd/paystream-worker
```

See [CONTRIBUTING.md](../CONTRIBUTING.md) for the full development workflow.
