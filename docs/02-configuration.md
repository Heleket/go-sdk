# 02 — Configuration

## Two API keys, two clients

Heleket has two distinct API keys:

| Key | Used by | SDK factory |
|---|---|---|
| Payment API key | `/v1/payment/*`, `/v1/wallet/*`, `/v1/balance`, `/v1/exchange-rate/*` | `heleket.NewPaymentClient()` |
| Payout API key  | `/v1/payout/*`, `/v1/transfer/*` | `heleket.NewPayoutClient()` |

Each key is also used to verify **its own** webhooks (payment webhooks → payment key; payout webhooks → payout key).

## Default construction

```go
import (
    "log"
    "log/slog"
    "time"

    heleket "github.com/heleket/go-sdk"
)

payment, err := heleket.NewPaymentClient(merchantID, paymentKey)
if err != nil {
    log.Fatal(err)
}

payout, err := heleket.NewPayoutClient(merchantID, payoutKey)
if err != nil {
    log.Fatal(err)
}
```

## Functional options

```go
client, err := heleket.NewPaymentClient(merchantID, paymentKey,
    heleket.WithDebug(true),
    heleket.WithTimeout(60 * time.Second),
    heleket.WithLogger(slog.Default()),
    heleket.WithUserAgent("myapp/1.0"),
    heleket.WithMaxRetries(3),
)
```

Available options:

- `heleket.WithBaseURL(string)` — override the API base URL. Must be `https://`, except `http://localhost` / `127.0.0.1` for local testing.
- `heleket.WithTimeout(time.Duration)` — per-request timeout (default 30s)
- `heleket.WithDebug(bool)` — toggle debug logging
- `heleket.WithLogger(*slog.Logger)` — pass your own logger (the SDK uses Debug-level messages)
- `heleket.WithHTTPClient(*http.Client)` — drop in a custom `*http.Client` (TLS, proxy, instrumentation)
- `heleket.WithTransport(heleket.Transport)` — swap out the entire HTTP layer (e.g. with a `FakeTransport` in tests)
- `heleket.WithUserAgent(string)` — append tokens to the User-Agent header (the SDK always sends `heleket-go/<Version>` plus your additions)
- `heleket.WithMaxRetries(int)` — number of retries on transport errors and HTTP 5xx (default 3, set to 0 to disable)
- `heleket.WithMaxResponseBytes(int64)` — cap on response body size (default 16 MiB)

## Concurrency and immutability

`PaymentClient` and `PayoutClient` are safe for concurrent use across goroutines. The Config you pass to `NewPaymentClient` / `NewPayoutClient` is snapshotted into the client at construction time — later mutations to your Config struct are not reflected in the live client.

## Custom HTTP client

Use a tuned `*http.Client` while keeping the SDK's default wire behaviour:

```go
client, err := heleket.NewPaymentClient(merchantID, paymentKey,
    heleket.WithHTTPClient(&http.Client{
        Timeout: 90 * time.Second,
        Transport: &http.Transport{
            Proxy: http.ProxyFromEnvironment,
        },
    }),
)
```

Note: when you supply your own `*http.Client`, you are responsible for the redirect policy. If you do not set `CheckRedirect`, Go's default follows up to 10 redirects across hosts — which would leak the signed `sign` header to whatever host the redirect resolves to. The SDK's default-built client blocks all redirects; preserve that guarantee in your custom client.

## Retries

Transport errors (DNS, timeouts, broken connections) and HTTP 5xx responses are retried 3 times by default with exponential backoff (100ms → 200ms → 400ms, capped at 5s). Retries respect the context deadline. Heleket rejects duplicate `OrderID`s and returns the existing record, so retrying create-* calls is safe.

Disable retries entirely if you handle them upstream:

```go
client, err := heleket.NewPaymentClient(merchantID, paymentKey,
    heleket.WithMaxRetries(0),
)
```

## Environment variables (recommended)

The example bootstrap reads these:

```env
HELEKET_MERCHANT_ID=...
HELEKET_PAYMENT_KEY=...
HELEKET_PAYOUT_KEY=...
HELEKET_DEBUG=0
```

Use the same names in your production environment for consistency.

## Timeouts

Default is 30 seconds. Heleket's `/v1/payout` and `/v1/payment` endpoints can take several seconds during blockchain confirmations — don't drop below 10s.

```go
client, err := heleket.NewPaymentClient(merchantID, paymentKey,
    heleket.WithTimeout(60 * time.Second),
)
```

## Next

→ [03 — Architecture](03-architecture.md) or jump to [04 — Payments API](04-payments.md).
