# API Documentation

Base URL: `http://localhost:8080` (default port for `cmd/paystream-api`, see [architecture.md](architecture.md)).

All request and response bodies are JSON. All endpoints other than `GET /health` require an `Authorization: Bearer <api-key>` header.

## Health

### `GET /health`

Returns service liveness status.

**Response `200 OK`**

```json
{
  "status": "ok",
  "service": "paystream-api",
  "version": "0.1.0"
}
```

## Recipients

### `POST /recipients`

Register a new payout recipient.

**Request**

```json
{
  "name": "Jane Doe",
  "stellar_address": "GABC...WXYZ",
  "email": "jane@example.com"
}
```

**Response `201 Created`**

```json
{
  "id": "rcp_01HZX3K9",
  "name": "Jane Doe",
  "stellar_address": "GABC...WXYZ",
  "email": "jane@example.com",
  "created_at": "2026-09-09T12:00:00Z"
}
```

### `GET /recipients/{id}`

Fetch a recipient by ID. Returns `404 Not Found` if the recipient does not exist.

## Payout Batches

### `POST /payouts`

Create a payout batch. The API persists the batch and returns immediately — the worker signs and submits the underlying Stellar transaction(s) asynchronously (see [architecture.md](architecture.md)).

**Request**

```json
{
  "payouts": [
    { "recipient_id": "rcp_01HZX3K9", "amount": "150.00", "asset": "USDC" }
  ]
}
```

**Response `202 Accepted`**

```json
{
  "batch_id": "bat_01HZX4M2",
  "status": "pending",
  "payout_count": 1
}
```

### `GET /payouts/{batch_id}`

Fetch batch status. `status` is one of `pending`, `signing`, `submitted`, `settled`, `failed`.

## Schedules

### `POST /schedules`

Create a recurring payroll schedule.

**Request**

```json
{
  "name": "Monthly payroll",
  "cron": "0 0 1 * *",
  "payouts": [
    { "recipient_id": "rcp_01HZX3K9", "amount": "150.00", "asset": "USDC" }
  ]
}
```

### `GET /schedules/{id}` / `DELETE /schedules/{id}`

Fetch or cancel a schedule.

## Webhooks

### `POST /webhooks`

Register an endpoint to receive event notifications.

**Request**

```json
{
  "url": "https://example.com/hooks/paystream",
  "events": ["payout.completed", "payout.failed"]
}
```

Delivered payloads are signed with HMAC-SHA256; see the webhook delivery notes in [architecture.md](architecture.md).

## Errors

Errors are returned as:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "amount must be greater than 0"
  }
}
```

| HTTP status | Meaning |
|---|---|
| `400` | Malformed request body or invalid field |
| `401` | Missing or invalid API key |
| `404` | Resource not found |
| `409` | Conflicting state (e.g. duplicate schedule name) |
| `500` | Internal server error |
