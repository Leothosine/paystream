# Payment Flow

This document diagrams how a payout batch moves from API request to settled funds and a delivered webhook, and how the ACH and recurring-schedule rails feed into the same pipeline.

## End-to-end batch payout

```mermaid
sequenceDiagram
    participant Client
    participant API as API (:8080)
    participant DB as PostgreSQL
    participant Worker
    participant Horizon as Stellar Horizon
    participant Recipient

    Client->>API: POST /v1/payouts/batch (Idempotency-Key)
    API->>DB: insert payout_batch + payout_items (status=pending)
    API-->>Client: 202 Accepted {batch_id}

    Worker->>DB: poll pending batches
    Worker->>Worker: chunk items into ≤100-op transactions
    Worker->>Worker: sign multi-operation transaction
    Worker->>Horizon: submit transaction
    Worker->>DB: update batch status=submitted

    Horizon-->>Worker: settlement confirmation (polled)
    Worker->>DB: update payout_items status=settled
    Worker->>Recipient: webhook payout.completed (HMAC-signed)
```

## Recurring schedules feeding into batches

```mermaid
flowchart LR
    A[schedule created via API] --> B[internal/schedule engine]
    B -->|next_run_at is due| C[worker creates a payout_batch]
    C --> D[same batch pipeline as above]
    B -.pause/resume/cancel.-> B
```

A schedule (`internal/schedule`) never touches Stellar directly — when it becomes due, it produces an ordinary `payout_batch` row that flows through the same signing and settlement path as a batch submitted directly through the API.

## ACH as an alternate settlement rail

```mermaid
flowchart LR
    A[payout_item routed to an ACH payout_method] --> B[internal/ach: validate routing + account number]
    B --> C[Transfer created, status=pending]
    C --> D[Submit to ACH network, status=submitted]
    D -->|settles| E[status=settled]
    D -->|bank rejects| F[status=returned, ReturnCode recorded]
```

Recipients whose `payout_methods.type` is `ach` bypass Horizon entirely: `internal/ach` validates the routing number (ABA checksum) and account number, then tracks the transfer through `pending → submitted → settled`, or `returned` if the receiving bank rejects it (e.g. NACHA code `R01` for insufficient funds).

## Batch validation before signing

```mermaid
flowchart TD
    A[items array from request] --> B{count in 1..100?}
    B -- no --> Z[reject whole batch]
    B -- yes --> C{idempotency key seen before?}
    C -- yes, same batch --> R[return original result]
    C -- no --> D[validate each item]
    D --> E[accepted items summed into Total]
    D --> F[rejected items keep a per-item Reason]
    E --> G[hand accepted items to the signing queue]
```

`internal/batch` performs this per-item validation before any item reaches the Stellar signing queue, so one malformed recipient in a 50-recipient payroll run doesn't block the other 49.
