# Error Handling Guide

This guide documents PayStream's error response format and the error codes clients should handle.

## Response format

All API errors share a single JSON shape:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "amount must be greater than 0"
  }
}
```

- `code` is a stable, machine-readable string — safe to branch on in client code.
- `message` is a human-readable description and may change wording over time; do not match on it.

## HTTP status codes

| Status | Meaning | Retry? |
|---|---|---|
| `400` | Malformed request body or invalid field value | No — fix the request |
| `401` | Missing or invalid API key | No — fix credentials |
| `404` | Resource not found | No |
| `409` | Conflicting state (e.g. duplicate schedule name, batch already cancelled) | No — resolve the conflict first |
| `429` | Rate limited | Yes — with backoff, honoring `Retry-After` if present |
| `500` | Internal server error | Yes — with backoff |
| `503` | Service temporarily unavailable (e.g. database failover in progress) | Yes — with backoff |

## Common error codes

| Code | Cause |
|---|---|
| `invalid_request` | A field failed validation (missing, wrong type, out of range) |
| `unauthorized` | API key missing, malformed, or revoked |
| `not_found` | The referenced resource (recipient, batch, schedule) doesn't exist |
| `conflict` | The operation would violate a uniqueness or state constraint |
| `rate_limited` | Too many requests in the current window |
| `internal_error` | Unexpected server-side failure — safe to retry, and worth reporting if persistent |

## Payout batch failure states

A batch can reach `status: "failed"` after being accepted (see [api.md](api.md) for the batch lifecycle). This is reported asynchronously, not as an HTTP error, since batch processing happens in the worker after the initial `202 Accepted` response. Poll `GET /payouts/{batch_id}` or subscribe to the `payout.failed` webhook event to observe this.

Common causes of a failed batch:

- Insufficient balance in the source Stellar account
- A recipient's `stellar_address` is invalid or the destination account requires a trustline that doesn't exist
- Horizon rejected the submitted transaction (e.g. sequence number conflict, fee too low)

## Webhook delivery errors

Webhook delivery failures do not surface to the API caller — they're handled by the worker's retry logic (exponential backoff; see [architecture.md](architecture.md)). If your endpoint is consistently failing, deliveries will eventually stop retrying; check your endpoint's logs and the HMAC signature verification first, as covered in [troubleshooting.md](troubleshooting.md).

## Client-side recommendations

- Always check `error.code`, never just the HTTP status, when branching on specific failure handling.
- Treat `500` and `503` as retryable with exponential backoff and jitter; treat `400`, `401`, `404`, and `409` as non-retryable without changing the request.
- Log the full error body when reporting bugs — the `message` field often contains the specific field or constraint that failed.
