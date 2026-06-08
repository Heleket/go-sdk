# 07 — Debugging

Two tools ship with the SDK: a **debug flag** for runtime tracing via `log/slog`, and a **webhook inspector CLI** for ad-hoc payload verification.

## Debug mode

Pass `heleket.WithDebug(true)` to the client constructor, or set `HELEKET_DEBUG=1` in your `.env` if you use the example bootstrap.

```go
client := heleket.NewPaymentClient(merchantID, paymentKey,
    heleket.WithDebug(true),
    heleket.WithLogger(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))),
)

_, _ = client.CreateInvoice(ctx, heleket.CreateInvoiceRequest{...})
```

Output (via `slog`):

```
time=2026-05-20T10:00:00Z level=DEBUG msg="heleket request" method=POST url=https://api.heleket.com/v1/payment body={"amount":"15.00","currency":"USD","order_id":"order-42"}
time=2026-05-20T10:00:00Z level=DEBUG msg="heleket response" status=200 body={"state":0,"result":{"uuid":"...", ...}}
```

> ⚠️ Debug output contains the **request body** and **response body**. The API key and the `sign` header are **never** passed to the logger. Still, scrub debug logs before sharing.

If you don't pass a logger, the SDK falls back to a stderr TextHandler when debug is on.

## Webhook inspector CLI

`bin/heleket-webhook-inspect` reads a JSON webhook payload, prints the parsed fields, and verifies the signature.

Build it once:

```bash
make build
```

### Usage

```bash
# From stdin (most common)
cat webhook.json | bin/heleket-webhook-inspect --key=$HELEKET_PAYMENT_KEY

# From a file
bin/heleket-webhook-inspect --key=$KEY --file=webhook.json

# Hint the expected type
bin/heleket-webhook-inspect --key=$KEY --type=payout < webhook.json
```

### Sample output

```
Heleket webhook inspector
----------------------------------------
  type       payment
  uuid       1ec87133-b22d-4643-988f-cac29a6ac85d
  order_id   order-42
  status     paid
  amount     15.00
  network    tron
  txid       deadbeef
  is_final   yes
  sign       4a8f2c2c4a8f2c2c…

signature: valid
```

If the signature does not verify, the tool prints `signature: INVALID` plus a checklist of common causes.

### Exit codes

| Code | Meaning |
|---|---|
| 0 | Payload valid, signature verified |
| 1 | Input not parseable as JSON |
| 2 | Signature mismatch |
| 3 | Missing arguments |

### Where to get the payload

Capture it in your handler before verification fails:

```go
payload, err := verifier.VerifyRaw(body)
if err != nil {
    _ = os.WriteFile("/tmp/heleket-failed.json", body, 0o600)
    /* ... */
}
```

Then pipe that file into the inspector.

## Common error shapes

| Symptom | Likely cause |
|---|---|
| Always `signature: INVALID` | Wrong API key — using payment key for a payout webhook (or vice versa) |
| Signature valid in `--file` but invalid in production | Your middleware mutated the body before your handler read it. Read `r.Body` directly. |
| `*ValidationError` from `CreateInvoice` | A required field is missing — `.Fields` lists them |
| `*APIError` with `"Server error, #1"` | Heleket-side outage; retry after a short backoff |
| `*HTTPError` containing `context deadline exceeded` | Network timeout — raise `WithTimeout` or check egress firewalls |

## Verbose HTTP trace (advanced)

`HTTPTransport` doesn't expose request/response dumps directly. To capture wire-level traces, wrap or replace it:

```go
type tracingTransport struct{ next heleket.Transport }

func (t *tracingTransport) RoundTrip(ctx context.Context, method, url string, headers http.Header, body []byte) (*heleket.Response, error) {
    log.Printf("→ %s %s\n%s", method, url, body)
    resp, err := t.next.RoundTrip(ctx, method, url, headers, body)
    if err == nil {
        log.Printf("← %d\n%s", resp.StatusCode, resp.Body)
    }
    return resp, err
}

client := heleket.NewPaymentClient(merchantID, paymentKey,
    heleket.WithTransport(&tracingTransport{next: heleket.NewHTTPTransport(nil)}),
)
```

## Next

→ [08 — Testing](08-testing.md)
