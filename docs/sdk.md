# SDK Documentation

PayStream provides client SDKs for Go, Node.js, Python, and Ruby under [`sdks/`](../sdks). Each SDK wraps the HTTP API described in [api.md](api.md).

> The SDKs are under active development (see each package's README). This page documents the intended usage shape so integrators can start building against it and track parity as each language catches up.

## Common concepts

Every SDK exposes the same operations, mapped to the underlying REST resources:

| Operation | REST endpoint |
|---|---|
| Create/get recipient | `POST /recipients`, `GET /recipients/{id}` |
| Create/get payout batch | `POST /payouts`, `GET /payouts/{batch_id}` |
| Create/get/delete schedule | `POST /schedules`, `GET /schedules/{id}`, `DELETE /schedules/{id}` |
| Register webhook | `POST /webhooks` |

All SDKs authenticate with an API key and raise a typed error carrying the API's `error.code` (see [error-handling.md](error-handling.md)) rather than a generic HTTP exception, so callers can branch on the code without parsing response bodies themselves.

## Go — [`sdks/go`](../sdks/go)

```go
import "github.com/breedar/paystream/sdks/go/paystream"

client := paystream.NewClient("sk_live_...")

batch, err := client.Payouts.Create(ctx, paystream.PayoutBatchRequest{
    Payouts: []paystream.Payout{
        {RecipientID: "rcp_01HZX3K9", Amount: "150.00", Asset: "USDC"},
    },
})
```

## Node.js — [`sdks/node`](../sdks/node)

```js
import { PayStream } from '@paystream/sdk';

const client = new PayStream('sk_live_...');

const batch = await client.payouts.create({
  payouts: [{ recipientId: 'rcp_01HZX3K9', amount: '150.00', asset: 'USDC' }],
});
```

## Python — [`sdks/python`](../sdks/python)

```python
from paystream import PayStream

client = PayStream("sk_live_...")

batch = client.payouts.create(
    payouts=[{"recipient_id": "rcp_01HZX3K9", "amount": "150.00", "asset": "USDC"}]
)
```

## Ruby — [`sdks/ruby`](../sdks/ruby)

```ruby
require "paystream"

client = PayStream::Client.new("sk_live_...")

batch = client.payouts.create(
  payouts: [{ recipient_id: "rcp_01HZX3K9", amount: "150.00", asset: "USDC" }]
)
```

## Choosing polling vs. webhooks

Payout batches settle asynchronously (see [architecture.md](architecture.md)). SDK consumers can either:

- Poll `payouts.get(batchId)` until `status` leaves `pending`/`signing`, or
- Register a webhook and react to `payout.completed` / `payout.failed` events, which avoids polling overhead for high batch volumes.

See [troubleshooting.md](troubleshooting.md) if webhook events aren't arriving.
