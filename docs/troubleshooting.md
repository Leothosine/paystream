# Troubleshooting Guide

Common issues when running or integrating with PayStream, and how to resolve them.

## API server won't start

**Symptom:** `cmd/paystream-api` exits immediately on startup.

- Check that all required environment variables from `.env.example` are set — the process fails fast if configuration is missing rather than starting in a partial state.
- Confirm the database is reachable: `psql "$DATABASE_URL" -c 'select 1'`.
- Port `8080` may already be in use by another process; the API does not currently support a configurable port at startup, so free the port or stop the conflicting process.

## Worker isn't processing payouts

**Symptom:** Batches stay in `pending` status indefinitely (see [architecture.md](architecture.md) for the API/worker split).

- Verify `cmd/paystream-worker` is actually running — the API only writes jobs, it never submits Stellar transactions itself.
- Check the worker's Horizon connectivity. If Horizon is unreachable, submitted transactions can't be confirmed and batches will appear stuck in `signing`.
- If running multiple worker replicas, confirm advisory locks aren't being held by a crashed process — a worker that dies mid-transaction can leave a lock until its session ends.

## Webhooks aren't arriving

**Symptom:** Registered endpoint never receives `payout.completed` / `payout.failed` events.

- Confirm the endpoint returns a `2xx` status quickly. Slow or non-2xx responses are treated as delivery failures and retried with exponential backoff, not an infinite hang.
- Verify the HMAC-SHA256 signature on your end using the shared webhook secret — a signature mismatch usually means the raw request body was re-serialized (e.g. by a JSON middleware) before verifying, which changes the byte content being signed.
- Check your endpoint isn't behind a firewall/allowlist blocking PayStream's egress IPs.

## Health check failing

**Symptom:** `GET /health` returns non-200 or times out.

- A non-200 usually means the process is still starting up or shutting down.
- A timeout with no response at all typically indicates the process has hung or crashed without exiting — check process logs and restart.

## Getting more help

If an issue isn't covered here, open a GitHub issue with:

- The command or request that triggered the problem
- Relevant log output
- Whether the issue is reproducible or intermittent
