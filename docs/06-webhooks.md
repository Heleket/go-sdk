# 06 — Webhooks

Webhooks deliver payment and payout status updates to your server. **Always** verify the signature before trusting the payload.

## How Heleket signs webhooks

For every webhook event Heleket POSTs a JSON body containing your fields plus a `sign` field. The signature is computed exactly like outbound request signatures:

```
sign = md5( base64( json_body_without_sign_field ) . apiKey )
```

The key depends on the webhook type:

| Webhook type | Key to verify with |
|---|---|
| `payment`, `wallet` | Payment API key |
| `payout` | Payout API key |

If you mix them up, every webhook fails.

## Verifying a webhook

Use `webhook.NewVerifier`. Two entry points:

### Verify from the raw request body

```go
import (
    "errors"
    "io"
    "log/slog"
    "net/http"

    heleket "github.com/heleket/go-sdk"
    "github.com/heleket/go-sdk/webhook"
)

verifier := webhook.NewVerifier(paymentKey)

http.HandleFunc("/heleket-webhook", func(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "cannot read body", http.StatusBadRequest)
        return
    }

    payload, err := verifier.VerifyRaw(body)
    if err != nil {
        var se *heleket.SignatureError
        if errors.As(err, &se) {
            slog.Warn("heleket signature mismatch", slog.String("reason", se.Reason))
        }
        http.Error(w, "invalid signature", http.StatusBadRequest)
        return
    }

    if payload.IsSuccessful() {
        markOrderPaid(payload.OrderID, payload.Amount)
    }

    w.WriteHeader(http.StatusOK)
})
```

### Verify from an already-decoded map

```go
payload, err := verifier.Verify(decoded) // decoded is map[string]any
```

## The slash-escaping caveat

Go's `encoding/json` does **not** escape forward slashes. PHP's `json_encode` does. Heleket signs using PHP-style encoding. Both `Verify` and `VerifyRaw` decode and re-encode the payload, then hash — this works as long as both sides agree on slash handling.

Heleket currently sends payloads compatible with what Go produces on re-encode. If verification fails despite a correct key, this is the next thing to check (see [11 — Troubleshooting](11-troubleshooting.md)).

## Reading the verified payload

`webhook.Payload` exposes typed fields plus convenience accessors:

```go
payload.Type           // "payment", "wallet", or "payout"
payload.IsPayment()    // true for payment/wallet
payload.IsPayout()
payload.UUID
payload.OrderID
payload.Status         // "paid", "wrong_amount", "process", ...
payload.IsFinal        // server-provided
payload.IsFinalStatus()// falls back to status enum if IsFinal absent
payload.IsSuccessful() // true for paid / paid_over
payload.Amount
payload.TxID
payload.Network
payload.Raw            // map[string]any with every field, including sign
```

## Recommended handler shape

```go
http.HandleFunc("/heleket-webhook", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK) // claim early; do work async

    body, err := io.ReadAll(r.Body)
    if err != nil { return }

    payload, err := verifier.VerifyRaw(body)
    if err != nil { return }

    queue.Enqueue(payload.Raw) // hand off to background worker
})
```

Respond fast (Heleket retries until you 200). Push heavy work to a queue.

## Idempotency and replay protection

Heleket may deliver the same status more than once. Common causes:

- **Retries** — if you respond non-2xx (or time out), Heleket retries with the same payload.
- **Manual replays** — operators (and your own code via `ResendWebhook`) can re-send historical events.
- **Network drops** — your edge accepted the request but the response never reached Heleket, so it retries.

A signature-valid webhook captured off the wire can also be replayed by an attacker who proxies it back to your endpoint. The signature alone does **not** prove freshness.

**You must de-duplicate.** Treat each `(uuid, status)` pair as a unique event and store it before doing any side-effect work.

### Minimal idempotent handler

```go
http.HandleFunc("/heleket-webhook", func(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    payload, err := verifier.VerifyRaw(body)
    if err != nil {
        http.Error(w, "bad signature", http.StatusBadRequest)
        return
    }

    eventKey := payload.UUID + ":" + payload.Status

    // Atomic "claim" — returns false if the key already exists.
    claimed, err := db.ClaimWebhookEvent(r.Context(), eventKey)
    if err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    if !claimed {
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, "OK (duplicate)")
        return
    }

    if err := processWebhook(r.Context(), payload); err != nil {
        // Release the claim so Heleket's retry can succeed.
        db.ReleaseWebhookEvent(r.Context(), eventKey)
        http.Error(w, "processing failed", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "OK")
})
```

### Pattern with database/sql

```sql
CREATE TABLE heleket_webhook_events (
    event_key    VARCHAR(128) PRIMARY KEY,
    payload      JSONB NOT NULL,
    received_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);
```

```go
// Postgres: ON CONFLICT DO NOTHING returns RowsAffected=0 when the row already exists.
func (db *DB) ClaimWebhookEvent(ctx context.Context, key string) (bool, error) {
    res, err := db.ExecContext(ctx,
        `INSERT INTO heleket_webhook_events (event_key, payload) VALUES ($1, '{}') ON CONFLICT DO NOTHING`,
        key,
    )
    if err != nil { return false, err }
    n, err := res.RowsAffected()
    return n == 1, err
}
```

### Pattern with Redis

```go
// SET key value NX EX 86400 — set only if not exists, expire in 24h.
ok, err := redis.SetNX(ctx, "heleket:webhook:"+eventKey, "1", 24*time.Hour).Result()
if !ok || err != nil { /* duplicate */ }
```

The Redis variant is simpler but forgets after the TTL — pick a window longer than Heleket's retry window (24h is a safe default).

### Why `(uuid, status)`, not just `uuid`

A single invoice can legitimately progress through several final statuses — for example `check` → `paid`, or `wrong_amount_waiting` → `paid`. Deduping by `uuid` alone would drop the second event and your accounting would diverge. The `(uuid, status)` pair is unique per real state transition.

## IP allowlisting

Heleket sends webhooks from `31.133.220.8`. Allow only that IP at your reverse proxy if possible. **Don't rely on IP alone** — signature verification is the authoritative trust boundary.

## Local development

Use a tunnel to expose your local machine:

```bash
make example-webhook
# In another terminal:
ngrok http 8000
```

Then set the resulting URL as `url_callback` when creating an invoice, or call `TestWebhook`:

```go
client.TestWebhook(ctx, heleket.TestWebhookRequest{
    Type:        "payment",
    URLCallback: "https://abc123.ngrok.io",
    Currency:    "USD",
    Network:     "tron",
    Status:      heleket.PaymentStatusPaid,
    OrderID:     "order-42",
})
```

## When signatures fail — debug recipe

Capture the failing payload, then:

```bash
echo '<paste the JSON here>' | bin/heleket-webhook-inspect --key=$HELEKET_PAYMENT_KEY
```

See [07 — Debugging](07-debugging.md) for the inspector CLI in detail.

## Next

→ [07 — Debugging](07-debugging.md)
