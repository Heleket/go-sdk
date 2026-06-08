# 04 — Payments API

Every method on `PaymentClient`. Each section shows the signature, parameters, return shape, and a runnable example.

All examples assume:

```go
import (
    "context"
    heleket "github.com/heleket/go-sdk"
)

client := heleket.NewPaymentClient(paymentKey, merchantID)
ctx := context.Background()
```

Errors:

- `*heleket.ValidationError` on HTTP 422 (call `errors.As` to inspect `.Fields`)
- `*heleket.APIError` on any other API failure (with `.HTTPStatus`, `.RawBody`)
- `*heleket.HTTPError` on transport failure
- `errors.Is(err, heleket.ErrNeedOneIdentifier)`-style guards (currently a sentinel string) when you pass empty identifiers

See [09 — Error handling](09-error-handling.md) for catch strategies.

## CreateInvoice

`CreateInvoice(ctx, CreateInvoiceRequest) (*Invoice, error)`  →  `POST /v1/payment`

Required: `Amount`, `Currency`, `OrderID`. See `CreateInvoiceRequest` in [`payment_types.go`](../payment_types.go) for the full set of optional fields.

```go
invoice, err := client.CreateInvoice(ctx, heleket.CreateInvoiceRequest{
    Amount:      "15.00",
    Currency:    "USD",
    OrderID:     "order-42",
    Lifetime:    3600,
    URLCallback: "https://your.site/heleket-webhook",
})

fmt.Println(invoice.URL)           // payment page
fmt.Println(invoice.Address)       // wallet address (for payer)
fmt.Println(invoice.PaymentStatus) // initial status (typically "check")
```

## GetInfo

`GetInfo(ctx, InfoOptions) (*Invoice, error)`  →  `POST /v1/payment/info`

Set exactly one of `UUID` or `OrderID`. The server prioritises `OrderID` when both are sent.

```go
info, err := client.GetInfo(ctx, heleket.InfoOptions{UUID: invoiceUUID})
// or
info, err := client.GetInfo(ctx, heleket.InfoOptions{OrderID: "order-42"})

if info.Status.IsFinal() { /* ... */ }
```

## ListHistory

`ListHistory(ctx, HistoryOptions) (*HistoryPage[Invoice], error)`  →  `POST /v1/payment/list`

Date format: `YYYY-MM-DD HH:MM:SS`. Pagination cursors come from a previous response's `Paginate.NextCursor`.

```go
page, err := client.ListHistory(ctx, heleket.HistoryOptions{
    DateFrom: "2026-01-01 00:00:00",
    DateTo:   "2026-05-20 23:59:59",
})
for _, invoice := range page.Items { /* ... */ }

if page.Paginate.NextCursor != "" {
    next, _ := client.ListHistory(ctx, heleket.HistoryOptions{
        Cursor: page.Paginate.NextCursor,
    })
    _ = next
}
```

## CreateStaticWallet

`CreateStaticWallet(ctx, CreateStaticWalletRequest) (*StaticWallet, error)`  →  `POST /v1/wallet`

Persistent address bound to an order ID — incoming transfers are credited to the merchant.

```go
wallet, err := client.CreateStaticWallet(ctx, heleket.CreateStaticWalletRequest{
    Currency: "USDT",
    Network:  "tron",
    OrderID:  "topup-user-7",
})
fmt.Println(wallet.Address)
```

## GenerateQRCode

`GenerateQRCode(ctx, merchantPaymentUUID string) (*QRCode, error)`  →  `POST /v1/wallet/qr`

Returns a base64-encoded QR-code image (data URI) for the given static wallet.

```go
qr, err := client.GenerateQRCode(ctx, wallet.UUID)
// embed in HTML: <img src="{{.Image}}">
```

## BlockStaticWallet

`BlockStaticWallet(ctx, BlockStaticWalletRequest) (*BlockedWallet, error)`  →  `POST /v1/wallet/block-address`

Stop accepting transfers on a static wallet. Set `IsRefund: true` to release locked funds back to the original sender.

```go
client.BlockStaticWallet(ctx, heleket.BlockStaticWalletRequest{
    OrderID: "topup-user-7",
})
```

## RefundBlockedWallet

`RefundBlockedWallet(ctx, uuid, address string) (*RefundResult, error)`  →  `POST /v1/wallet/blocked-address-refund`

Send the contents of a blocked wallet to a recovery address.

```go
client.RefundBlockedWallet(ctx, wallet.UUID, "TXyz...recovery-address...")
```

## Refund

`Refund(ctx, RefundRequest) (*RefundResult, error)`  →  `POST /v1/payment/refund`

Required: `Address`, `IsSubtract`; one of `UUID` / `OrderID`.

```go
client.Refund(ctx, heleket.RefundRequest{
    UUID:       invoiceUUID,
    Address:    "TBaCkAdDrEsS",
    IsSubtract: true,
})
```

## ResendWebhook

`ResendWebhook(ctx, InfoOptions) error`  →  `POST /v1/payment/resend`

Ask Heleket to redeliver the last webhook for the invoice.

```go
client.ResendWebhook(ctx, heleket.InfoOptions{OrderID: "order-42"})
```

## TestWebhook

`TestWebhook(ctx, TestWebhookRequest) error`  →  `POST /v1/test-webhook/{payment|wallet}`

Send a synthetic webhook event to your callback URL — useful in development.

```go
client.TestWebhook(ctx, heleket.TestWebhookRequest{
    Type:        "payment",
    URLCallback: "https://your.site/heleket-webhook",
    Currency:    "USD",
    Network:     "tron",
    Status:      heleket.PaymentStatusPaid,
    OrderID:     "order-42",
})
```

## ListServices

`ListServices(ctx) ([]Service, error)`  →  `POST /v1/payment/services`

Catalogue of supported (currency, network) combinations with limits.

```go
services, _ := client.ListServices(ctx)
for _, s := range services {
    fmt.Printf("%s on %s: min=%s max=%s\n", s.Currency, s.Network, s.Limit.MinAmount, s.Limit.MaxAmount)
}
```

## GetBalance

`GetBalance(ctx) (*Balance, error)`  →  `POST /v1/balance`

```go
balance, _ := client.GetBalance(ctx)
for _, row := range balance.Merchant {
    fmt.Printf("%s: %s\n", row.CurrencyCode, row.Balance)
}
```

## GetExchangeRates

`GetExchangeRates(ctx, currency string) ([]ExchangeRate, error)`  →  `POST /v1/exchange-rate/{currency}/list`

```go
rates, _ := client.GetExchangeRates(ctx, "USD")
for _, r := range rates {
    fmt.Printf("%s -> %s = %s (from %s)\n", r.From, r.To, r.Course, r.Source)
}
```

## Next

→ [05 — Payouts API](05-payouts.md)
