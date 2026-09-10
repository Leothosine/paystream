# Getting Started

This guide walks a new contributor or evaluator through running PayStream locally and sending a first payout batch. It assumes no prior familiarity with the codebase.

## 1. Prerequisites

Install these before you begin:

- **Go 1.22+** — runs the API and worker binaries
- **Node.js 20+ and pnpm 9+** — runs the dashboard
- **PostgreSQL 14+** — shared datastore for the API and worker
- **A Stellar testnet wallet funded with testnet USDC** — create one at the [Stellar Laboratory](https://laboratory.stellar.org)
- Anchor credentials for at least one corridor (optional — mock anchors are included for local development)

## 2. Clone and bootstrap

```bash
git clone https://github.com/Breedar/paystream.git
cd paystream
make bootstrap     # installs Go deps, pnpm deps, runs DB migrations
```

`make bootstrap` also copies `.env.example` to `.env` if one doesn't already exist. Review that file and fill in your Postgres connection string and Stellar testnet secret key before continuing.

## 3. Start the API and dashboard

```bash
make dev           # starts API on :8080, dashboard on :3000
```

Open `http://localhost:3000` to confirm the dashboard loads, and `http://localhost:8080/health` to confirm the API is up.

## 4. Send your first payout batch

```bash
curl -X POST http://localhost:8080/v1/payouts/batch \
  -H "Authorization: Bearer $PAYSTREAM_API_KEY" \
  -H "Idempotency-Key: payroll-2026-05-01" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "May 2026 contractors",
    "asset": "USDC",
    "items": [
      {"recipient_id": "rec_abc", "amount": "1500.00"},
      {"recipient_id": "rec_def", "amount": "2200.00"}
    ]
  }'
```

The response includes a `batch_id`. PayStream signs and submits the underlying Stellar transactions, monitors settlement on Horizon, and emits a `payout.completed` webhook as each recipient's payment confirms. Always send a unique `Idempotency-Key` per logical batch — replays within 24 hours return the original response instead of double-paying recipients.

## 5. Where to go next

- [`docs/architecture.md`](architecture.md) — how the API, worker, and Stellar network fit together
- [`docs/api.md`](api.md) — full REST reference
- [`docs/error-handling.md`](error-handling.md) — error codes and retry semantics
- [`docs/troubleshooting.md`](troubleshooting.md) — common local setup issues
- [`CONTRIBUTING.md`](../CONTRIBUTING.md) — branch naming, commit conventions, and the PR checklist

## 6. Running tests before you contribute

```bash
make test             # full suite (Go + TypeScript)
make test-api         # Go tests only
make test-dashboard   # TypeScript/Vitest tests only
make lint             # golangci-lint + eslint + prettier
```
