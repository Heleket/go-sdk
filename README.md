# Heleket Go integration (reference)

A production-grade reference Go SDK for the [Heleket](https://heleket.com) cryptocurrency payment API. Covers the full documented surface (payments, payouts, balance, services, exchange rates), ships with typed request/response structs, unit tests (including `-race`), runnable examples, a debug flag wired into `log/slog`, automatic retry on transport / 5xx errors, a webhook inspector CLI, and a Docker harness.

Built to be `go get`'d directly into your project. Zero runtime dependencies — only the Go standard library.

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"

    heleket "github.com/heleket/go-sdk"
)

func main() {
    client, err := heleket.NewPaymentClient(merchantID, paymentKey)
    if err != nil {
        log.Fatal(err)
    }

    invoice, err := client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
        Amount:   "15.00",
        Currency: "USD",
        OrderID:  "order-42",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(invoice.URL) // → https://pay.heleket.com/pay/<uuid>
}
```

## Install

```bash
go get github.com/heleket/go-sdk
```

Requirements: Go 1.22+.

## Documentation

Full reference lives in [`docs/`](docs/README.md):

- [01 — Installation](docs/01-installation.md)
- [02 — Configuration](docs/02-configuration.md)
- [03 — Architecture](docs/03-architecture.md)
- [04 — Payments API](docs/04-payments.md)
- [05 — Payouts API](docs/05-payouts.md)
- [06 — Webhooks](docs/06-webhooks.md) ⚑ critical reading
- [07 — Debugging](docs/07-debugging.md)
- [08 — Testing](docs/08-testing.md)
- [09 — Error handling](docs/09-error-handling.md)
- [10 — Reference (statuses, currencies, endpoints)](docs/10-reference.md)
- [11 — Troubleshooting](docs/11-troubleshooting.md)

## What's in the box

```
go.mod / *.go            Production code — zero deps beyond the standard library
webhook/                 Subpackage for incoming webhook verification
internal/testutil/       FakeTransport for offline tests
examples/                Ten runnable programs covering every endpoint
cmd/heleket-webhook-inspect/  CLI for verifying and dumping any webhook payload
docker/                  golang:1.22-alpine multi-stage build
docs/                    Full module documentation
```

## Common tasks

```bash
make install              # go mod download
make test                 # go test ./...
make race                 # go test -race ./...
make vet                  # go vet ./...
make staticcheck          # staticcheck ./...
make fmt                  # gofmt -w .
make qa                   # All quality gates
make example-invoice      # Create a real invoice (needs .env)
make example-webhook      # Run the webhook listener on :8000
make docker-shell         # Drop into a containerized Go shell
make build                # Compile heleket-webhook-inspect to bin/
make help                 # Full target list
```

## Built-in resilience

- **Retries.** Transport errors (DNS, timeouts, broken connections) and HTTP 5xx responses are retried up to 3 times by default with exponential backoff. Tune via `heleket.WithMaxRetries(n)` or disable with `n = 0`. Heleket rejects duplicate `OrderID`s and returns the existing record, so retrying create-* calls is safe.
- **Response body cap.** The SDK refuses to read more than 16 MiB per response by default to protect against memory-exhaustion from a misbehaving server. Tune via `heleket.WithMaxResponseBytes`.
- **No cross-host redirects.** The default `*http.Client` blocks all redirects so the signed `sign` header never reaches an unexpected host.
- **HTTPS-only base URL.** `WithBaseURL` accepts `https://` for production and `http://localhost` / `127.0.0.1` for local testing — nothing else.
- **User-Agent.** Every request carries `heleket-go/<version>`; append your own identifier via `heleket.WithUserAgent("myapp/1.0")`.

## Security notes (read this)

- **Always verify webhook signatures.** See [docs/06-webhooks.md](docs/06-webhooks.md). Never trust the payload otherwise.
- **De-duplicate replays.** Use a `(uuid, status)` key in your DB before doing side-effect work — pattern documented in [docs/06-webhooks.md](docs/06-webhooks.md#idempotency-and-replay-protection).
- **Whitelist Heleket's webhook source IP `31.133.220.8`** at your reverse proxy or firewall.
- **Two separate API keys** — payments and payouts. Mixing them breaks webhook verification.
- **The SDK never logs API keys.** Debug-mode output via `log/slog` includes URL, method, and body — but the `sign` header and API key are explicitly excluded.

## License

MIT.
