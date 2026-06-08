# 05 — Payouts API

Withdrawals from the merchant balance. Uses the **payout API key** — different from the payment key.

```go
client := heleket.NewPayoutClient(payoutKey, merchantID)
ctx := context.Background()
```

Error model identical to `PaymentClient` — see [09 — Error handling](09-error-handling.md).

## CreatePayout

`CreatePayout(ctx, CreatePayoutRequest) (*Payout, error)`  →  `POST /v1/payout`

Required: `Amount`, `Currency`, `OrderID`, `Address`, `IsSubtract`.

Important fields:

| Field | Meaning |
|---|---|
| `IsSubtract` | `true` → commission deducted from `Amount`; `false` → commission added on top |
| `Network` | Required for multi-network coins (USDT etc.). Omit for BTC and other single-network chains |
| `Priority` | `economy`, `high`, `highest`, or `recommended` (default) — BTC/ETH/POLY/BSC only |
| `ToCurrency` | Required if `Currency` is fiat (e.g. USD → USDT) |
| `Memo` | Required for TON (1–30 chars) |

```go
payout, err := client.CreatePayout(ctx, heleket.CreatePayoutRequest{
    Amount:      "5.00",
    Currency:    "USDT",
    Network:     "TRON",
    OrderID:     "payout-001",
    Address:     "TDD97yguPESTpcrJMqU6h2ozZbibv4Vaqm",
    IsSubtract:  true,
    Priority:    "recommended",
    URLCallback: "https://your.site/heleket-payout-webhook",
})

fmt.Println(payout.UUID)
fmt.Println(payout.Status)  // typically PayoutStatusProcess initially
fmt.Println(payout.Balance) // remaining balance after deduction
```

## GetInfo

`GetInfo(ctx, InfoOptions) (*Payout, error)`  →  `POST /v1/payout/info`

```go
info, _ := client.GetInfo(ctx, heleket.InfoOptions{UUID: payout.UUID})
```

## ListHistory

`ListHistory(ctx, HistoryOptions) (*HistoryPage[Payout], error)`  →  `POST /v1/payout/list`

Same shape as `PaymentClient.ListHistory`.

```go
page, _ := client.ListHistory(ctx, heleket.HistoryOptions{
    DateFrom: "2026-01-01 00:00:00",
})
for _, p := range page.Items { /* ... */ }
```

## CalculateWithdrawal

`CalculateWithdrawal(ctx, CalculateRequest) (*Calculation, error)`  →  `POST /v1/payout/calculate`

Preview the commission and final amount before committing.

```go
preview, _ := client.CalculateWithdrawal(ctx, heleket.CalculateRequest{
    Currency:   "USDT",
    Network:    "TRON",
    Amount:     "100",
    IsSubtract: true,
})
fmt.Println(preview.Commission)
fmt.Println(preview.MerchantAmount)
```

## ListServices

`ListServices(ctx) ([]Service, error)`  →  `POST /v1/payout/services`

Catalogue of supported (currency, network) pairs for withdrawals.

```go
services, _ := client.ListServices(ctx)
```

## TransferToPersonal

`TransferToPersonal(ctx, amount, currency string) (*Transfer, error)`  →  `POST /v1/transfer/to-personal`

Move funds from the business balance to the personal wallet (same Heleket account).

```go
client.TransferToPersonal(ctx, "10.00", "USDT")
```

## TransferToBusiness

`TransferToBusiness(ctx, amount, currency string) (*Transfer, error)`  →  `POST /v1/transfer/to-business`

The reverse transfer.

```go
client.TransferToBusiness(ctx, "10.00", "USDT")
```

## Next

→ [06 — Webhooks](06-webhooks.md) — **read before going to production**.
