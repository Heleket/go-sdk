# 03 — Architecture

## High-level diagram

```
              ┌──────────────────────────────┐
              │  heleket.New{Payment,Payout} │  factory constructors
              └──────────┬───────────────────┘
                         │
        ┌────────────────┴────────────────┐
        ▼                                 ▼
┌──────────────────┐               ┌──────────────────┐
│  PaymentClient   │               │  PayoutClient    │
└────────┬─────────┘               └────────┬─────────┘
         │  embeds                          │ embeds
         └────────────┬─────────────────────┘
                      ▼
              ┌───────────────┐
              │  baseClient   │   one place for: encode JSON, sign,
              │               │   dispatch, debug-log, decode envelope
              └───────┬───────┘
   Config ──────────► │
   Transport ───────► │
   *slog.Logger ────► │
                      ▼
              ┌────────────────────┐
              │   HTTPTransport    │   ← default, swappable
              └────────────────────┘
```

## Package layout

| Path | Purpose |
|---|---|
| `github.com/heleket/go-sdk` | Root package. Public types, constructors, signature helpers, error types. |
| `github.com/heleket/go-sdk/webhook` | Webhook signature verification + payload value type. Imported separately so webhook-only consumers don't pull the full client. |
| `github.com/heleket/go-sdk/internal/testutil` | `FakeTransport` for offline tests. Importable from inside the module. |
| `github.com/heleket/go-sdk/cmd/heleket-webhook-inspect` | CLI binary. |
| `github.com/heleket/go-sdk/examples/...` | Runnable demo programs. |

## Source files (root package)

| File | Responsibility |
|---|---|
| `heleket.go`, `doc.go` | Package documentation, top-level types live here |
| `config.go` | `Config` struct, `Option` constructors, `newConfig` validation |
| `signature.go` | `Sign(body, key)` + `SignatureEqual(a, b)` |
| `errors.go` | `APIError`, `ValidationError`, `HTTPError`, `SignatureError` |
| `transport.go` | `Transport` interface + default `HTTPTransport` |
| `debug.go` | `slog`-based request/response trace helpers |
| `status.go` | `PaymentStatus`, `PayoutStatus`, `CourseSource` string types |
| `client.go` | `baseClient.post` — the one place that performs HTTP |
| `payment.go`, `payment_types.go` | `PaymentClient` + every payment endpoint + request/response structs |
| `payout.go`, `payout_types.go` | `PayoutClient` + every payout endpoint |

## Request flow

1. Caller invokes `client.CreateInvoice(ctx, req)`.
2. `PaymentClient.CreateInvoice` calls `baseClient.post(ctx, "/v1/payment", req, &invoice)`.
3. `baseClient.post`:
   - Marshals the request struct to JSON (empty struct → empty bytes).
   - Signs the JSON with the API key via `Sign`.
   - Builds headers: `merchant`, `sign`, `Content-Type: application/json`.
   - Calls `debugRequest` (no-op when `Debug=false`).
   - Calls `Transport.RoundTrip(ctx, "POST", url, headers, body)`.
   - Calls `debugResponse`.
   - Decodes the response envelope (`state`, `result`, `errors`).
4. On `state=0`, the `result` JSON is decoded into the typed pointer. On any failure, a typed error is returned.

## Signature flow

```
req struct       apiKey
    │              │
    ▼              ▼
json.Marshal()    ─ │
    │              │
    ▼              │
body []byte ──────►│
    │              │
    ▼              │
base64(body) + key ┘
    │
    ▼
md5(...) hex
    │
    ▼
sign  ──►  HTTP header "sign: <hex>"
```

The same `Sign` function is used to **produce** outgoing request signatures and to **verify** incoming webhook signatures (in the `webhook` subpackage).

## Concurrency

`PaymentClient` and `PayoutClient` are **safe for concurrent use** across goroutines. They hold an immutable `Config` and an `*http.Client`, both of which are concurrency-safe. The `FakeTransport` test double is also goroutine-safe (it uses `sync.Mutex` internally).

You can share a single client across an HTTP server's handlers — there's no need to construct a client per request.

## Design choices

- **No third-party deps.** Stdlib only. Easier to vendor, audit, and install in restricted environments.
- **Transport is an interface.** Tests inject `FakeTransport`. Merchants who want a custom HTTP backend can implement the interface.
- **Configuration is immutable.** All "withers" return new `Config` values. Easier to reason about request-scoped clients.
- **Typed structs, not maps.** Idiomatic Go, IDE autocomplete, harder to misspell field names. Forward compatibility is handled via `json.RawMessage` in the unmarshal path — the SDK never errors on unknown fields.
- **Functional options for `Config`.** Keeps the constructor signature stable while leaving room to grow.
- **Generic `HistoryPage[T]`.** One shape for payment and payout history.

## Next

→ [04 — Payments API](04-payments.md)
