# 08 — Testing

## Running the SDK's own tests

```bash
make test         # go test ./...
make race         # go test -race ./...
make vet          # go vet ./...
make staticcheck  # staticcheck ./...
make fmt-check    # gofmt -l . must be empty
make qa           # all of the above
```

Install staticcheck once:

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
```

In Docker:

```bash
make docker-qa
```

## Writing tests against the SDK

The SDK is fully testable offline thanks to `heleket.Transport`. Inject `internal/testutil.FakeTransport` (or your own implementation) and the client will skip the network entirely.

```go
import (
    "context"
    "encoding/json"
    "testing"

    heleket "github.com/heleket/go-sdk"
    "github.com/heleket/go-sdk/internal/testutil"
)

func TestMyCheckout(t *testing.T) {
    fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
        "state":  0,
        "result": map[string]any{"uuid": "uuid-1", "url": "https://pay/..."},
    }, 200)

    client, _ := heleket.NewPaymentClient("merchant", "key",
        heleket.WithTransport(fake),
    )

    invoice, err := client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
        Amount: "1", Currency: "USD", OrderID: "x",
    })
    if err != nil { t.Fatal(err) }
    if invoice.UUID != "uuid-1" { t.Errorf("got %q", invoice.UUID) }

    // Inspect the request the SDK sent
    req := fake.LastRequest()
    if req.Method != "POST" { t.Errorf("method = %q", req.Method) }
    var body map[string]any
    _ = json.Unmarshal(req.Body, &body)
    if body["order_id"] != "x" { t.Errorf("body = %v", body) }
}
```

## FakeTransport API

| Method | Purpose |
|---|---|
| `Enqueue(*heleket.Response)` | Queue a pre-built response |
| `EnqueueJSON(any, statusCode)` | Convenience: JSON-encode the value |
| `FailNext(error)` | Make the next call return a `*heleket.HTTPError` wrapping the given error |
| `Requests() []RecordedRequest` | All recorded calls in order |
| `LastRequest() RecordedRequest` | Most recent recorded call |

`RecordedRequest` exposes `Method`, `URL`, `Headers`, `Body`.

FakeTransport is goroutine-safe — share one instance across parallel tests if needed.

## Integration tests against real Heleket

Gate them on env vars and skip when missing:

```go
func TestCreateInvoice_Integration(t *testing.T) {
    if os.Getenv("HELEKET_PAYMENT_KEY") == "" {
        t.Skip("set HELEKET_PAYMENT_KEY to run integration tests")
    }
    client, _ := heleket.NewPaymentClient(
        os.Getenv("HELEKET_MERCHANT_ID"),
        os.Getenv("HELEKET_PAYMENT_KEY"),
    )
    /* ... */
}
```

Tag the file with `//go:build integration` to keep it out of the default `go test ./...` run.

## CI example (GitHub Actions)

```yaml
name: Go SDK
on: [push, pull_request]
jobs:
  qa:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go install honnef.co/go/tools/cmd/staticcheck@latest
      - run: make qa
```

## Next

→ [09 — Error handling](09-error-handling.md)
