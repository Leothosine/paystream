# Database Schema

PayStream stores all durable state in a single shared PostgreSQL database. The API writes rows describing work to be done (payouts, schedules, webhook registrations); the worker reads and updates them as it drives payments to settlement. This document describes the core tables and how they relate. Column lists are illustrative of the domain, not a literal migration diff — see `migrations/` for the authoritative, sequentially numbered schema.

## Entity overview

```
recipients ──< payout_methods
     │
     │            ┌──< payout_items ──┐
     ▼            │                   │
 schedules ──> payout_batches ──> payments ──> stripe/anchor refs
                    │
                    └──< webhook_deliveries >── webhook_endpoints
```

## Core tables

### `recipients`

One row per person or entity PayStream can pay.

| Column | Type | Notes |
|---|---|---|
| `id` | text (PK) | e.g. `rec_abc` |
| `name` | text | display name |
| `email` | text | notification address |
| `country` | text | ISO 3166-1 alpha-2, drives corridor selection |
| `created_at` | timestamptz | |

### `payout_methods`

A recipient can have more than one way to receive funds (Stellar wallet, ACH account, local bank via anchor).

| Column | Type | Notes |
|---|---|---|
| `id` | text (PK) | |
| `recipient_id` | text (FK → `recipients.id`) | |
| `type` | text | `stellar_wallet`, `ach`, `anchor` |
| `details` | jsonb | shape depends on `type` (Stellar public key, routing/account numbers, anchor-specific fields) |
| `is_default` | boolean | |

### `payout_batches`

A single funding request submitted via the API, corresponding to one Stellar multi-operation transaction (or a chain of them if it exceeds the 100-operation limit).

| Column | Type | Notes |
|---|---|---|
| `id` | text (PK) | e.g. `bat_1` |
| `name` | text | human label, e.g. "May 2026 contractors" |
| `asset` | text | e.g. `USDC` |
| `status` | text | `pending`, `signing`, `submitted`, `settled`, `failed` |
| `idempotency_key` | text (unique) | dedupes replayed requests within 24h |
| `schedule_id` | text (FK → `schedules.id`, nullable) | set when the batch was produced by a recurring schedule rather than a direct API call |
| `created_at` | timestamptz | |

### `payout_items`

One row per recipient payment within a batch.

| Column | Type | Notes |
|---|---|---|
| `id` | text (PK) | |
| `batch_id` | text (FK → `payout_batches.id`) | |
| `recipient_id` | text (FK → `recipients.id`) | |
| `amount` | numeric(20,7) | Stellar-precision decimal |
| `status` | text | `pending`, `settled`, `returned` |
| `stellar_tx_hash` | text, nullable | set once submitted to Horizon |

### `schedules`

Recurring payout definitions driven by the schedule engine (`internal/schedule`).

| Column | Type | Notes |
|---|---|---|
| `id` | text (PK) | |
| `recipient_id` | text (FK → `recipients.id`) | |
| `amount` | numeric(20,7) | |
| `interval` | text | `daily`, `weekly`, `monthly` |
| `status` | text | `active`, `paused`, `canceled` |
| `next_run_at` | timestamptz | polled by the worker's schedule engine |

### `webhook_endpoints` / `webhook_deliveries`

Registered customer endpoints and the individual delivery attempts made to them (see `internal/webhook`), including HMAC signature and retry/backoff bookkeeping.

## Conventions

- Primary keys are prefixed, human-readable text IDs (`rec_`, `bat_`, `sch_`, `ach_`, …), not raw UUIDs, to make logs and support tickets easier to read.
- Monetary columns are `numeric`, never floating point, at the storage layer.
- Every state-changing write is expected to also emit an audit log row per the compliance posture described in the top-level `README.md`.
- Migrations live in `migrations/`, are numbered sequentially, and should be reversible where possible (see the PR checklist in `CONTRIBUTING.md`).
