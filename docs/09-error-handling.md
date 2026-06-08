# 09 — Error handling

## Error types

| Type | When |
|---|---|
| `*heleket.HTTPError` | Transport-layer failure: DNS, TCP, TLS, timeout |
| `*heleket.ValidationError` | HTTP 422 with a field-level errors map (wraps `*APIError`) |
| `*heleket.APIError` | Any other `state != 0` from Heleket |
| `*heleket.SignatureError` | Webhook signature mismatch |
| `error` (plain) | Pre-flight argument checks (e.g. empty `InfoOptions`), JSON decode failures |

All SDK error types implement `error`. Inspect with `errors.As` to recover the typed value, or `errors.Is` against a sentinel to match a category:

```go
import "errors"

if errors.Is(err, heleket.ErrValidation) { /* 422 — field-level errors */ }
if errors.Is(err, heleket.ErrTransport)  { /* DNS, TCP, TLS, timeout */ }
if errors.Is(err, heleket.ErrAPI)        { /* any business error from server */ }
if errors.Is(err, heleket.ErrSignature)  { /* webhook signature mismatch */ }
```

`ErrValidation` also satisfies `ErrAPI` because `ValidationError` is-a `APIError`.

## Quick reference

| Error | Thrown by | What to do |
|---|---|---|
| `*HTTPError` | Any client method | Retry with backoff |
| `*ValidationError` | Any client method | Surface `Fields` to the caller; do not retry |
| `*APIError` | Any client method | Inspect `Message` + `HTTPStatus` + `RawBody`; retry only on 5xx |
| `*SignatureError` | `webhook.Verifier` methods | Drop the request, return HTTP 400 |

## Recommended catch shape

```go
import (
    "errors"

    heleket "github.com/heleket/go-sdk"
)

invoice, err := client.CreateInvoice(ctx, req)
if err != nil {
    var ve *heleket.ValidationError
    var ae *heleket.APIError
    var he *heleket.HTTPError
    switch {
    case errors.As(err, &ve):
        // 422 — never retry. Show errors to the user.
        for field, msgs := range ve.Fields {
            log.Printf("%s: %v", field, msgs)
        }
    case errors.As(err, &he):
        // Transport — usually safe to retry.
        scheduleRetry()
    case errors.As(err, &ae):
        // Business error. Retry only on 5xx + idempotent operation.
        if ae.HTTPStatus >= 500 {
            scheduleRetry()
        } else {
            log.Printf("api error: %s (raw=%s)", ae.Message, ae.RawBody)
        }
    default:
        log.Printf("unexpected: %v", err)
    }
}
```

Note the order: `*ValidationError` first because it embeds `*APIError` — checking `*APIError` first would always match.

## Retry guidance

| Status / error | Safe to retry? |
|---|---|
| `*HTTPError` (timeout, DNS, TLS) | Yes — exponential backoff |
| `*APIError` HTTP 5xx | Yes — exponential backoff |
| `*APIError` HTTP 4xx (non-422) | No — fix the request first |
| `*ValidationError` (422) | No |
| `*SignatureError` | No — never trust the payload |

For idempotency on retries of create-* calls, **reuse the same `OrderID`**. Heleket rejects duplicates and returns the existing record, which is what you want.

## Logging

The SDK doesn't log; it only writes via `*slog.Logger` when debug is on (see [07 — Debugging](07-debugging.md)). Use `errors.As` to surface enough context for forensics:

```go
case errors.As(err, &ae):
    logger.Warn("heleket api rejected create-invoice",
        slog.Int("status", ae.HTTPStatus),
        slog.String("message", ae.Message),
        slog.String("raw_body", string(ae.RawBody)),
        slog.String("order_id", req.OrderID),
    )
```

## Validation error shape

`ValidationError.Fields` is `map[string][]string`:

```go
ve.Fields == map[string][]string{
    "amount":   {"validation.required"},
    "currency": {"validation.required"},
}
```

Values are arrays of messages (sometimes more than one per field).

## Next

→ [10 — Reference](10-reference.md)
