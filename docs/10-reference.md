# 10 — Reference

Quick lookup tables.

## Payment statuses (`heleket.PaymentStatus`)

| Constant | Wire value | Final? | Successful? |
|---|---|---|---|
| `PaymentStatusPaid` | `paid` | ✓ | ✓ |
| `PaymentStatusPaidOver` | `paid_over` | ✓ | ✓ |
| `PaymentStatusWrongAmount` | `wrong_amount` | ✓ | — |
| `PaymentStatusWrongAmountWaiting` | `wrong_amount_waiting` | — | — |
| `PaymentStatusCheck` | `check` | — | — |
| `PaymentStatusConfirmCheck` | `confirm_check` | — | — |
| `PaymentStatusProcess` | `process` | — | — |
| `PaymentStatusFail` | `fail` | ✓ | — |
| `PaymentStatusCancel` | `cancel` | ✓ | — |
| `PaymentStatusSystemFail` | `system_fail` | ✓ | — |
| `PaymentStatusLocked` | `locked` | ✓ | — |
| `PaymentStatusRefundProcess` | `refund_process` | — | — |
| `PaymentStatusRefundPaid` | `refund_paid` | ✓ | — |
| `PaymentStatusRefundFail` | `refund_fail` | ✓ | — |

Helpers:

```go
heleket.PaymentStatus("paid").IsFinal()      // true
heleket.PaymentStatus("paid").IsSuccessful() // true
```

## Payout statuses (`heleket.PayoutStatus`)

| Constant | Wire value | Final? | Successful? |
|---|---|---|---|
| `PayoutStatusProcess` | `process` | — | — |
| `PayoutStatusCheck` | `check` | — | — |
| `PayoutStatusPaid` | `paid` | ✓ | ✓ |
| `PayoutStatusFail` | `fail` | ✓ | — |
| `PayoutStatusCancel` | `cancel` | ✓ | — |
| `PayoutStatusSystemFail` | `system_fail` | ✓ | — |

## AML link statuses (`heleket.AmlLinkStatus`)

Returned per item by `GetAmlLinks` for a blocked (locked) payment.

| Constant | Wire value | Final? | Successful? |
|---|---|---|---|
| `AmlLinkStatusInit` | `init` | — | — |
| `AmlLinkStatusPending` | `pending` | — | — |
| `AmlLinkStatusCompleted` | `completed` | ✓ | ✓ |
| `AmlLinkStatusExpired` | `expired` | ✓ | — |

## Exchange-rate sources (`heleket.CourseSource`)

- `CourseSourceBinance` (`"Binance"`)
- `CourseSourceBinanceP2P` (`"BinanceP2P"`)
- `CourseSourceExmo` (`"Exmo"`)
- `CourseSourceKucoin` (`"Kucoin"`)

## Endpoint table

| Method | Path | Surface |
|---|---|---|
| `CreateInvoice` | `POST /v1/payment` | PaymentClient |
| `GetInfo` (payment) | `POST /v1/payment/info` | PaymentClient |
| `GetAmlLinks` | `POST /v1/payment/aml-links` | PaymentClient |
| `ListHistory` (payment) | `POST /v1/payment/list` | PaymentClient |
| `Refund` | `POST /v1/payment/refund` | PaymentClient |
| `ResendWebhook` | `POST /v1/payment/resend` | PaymentClient |
| `ListServices` (payment) | `POST /v1/payment/services` | PaymentClient |
| `CreateStaticWallet` | `POST /v1/wallet` | PaymentClient |
| `GenerateQRCode` | `POST /v1/wallet/qr` | PaymentClient |
| `BlockStaticWallet` | `POST /v1/wallet/block-address` | PaymentClient |
| `RefundBlockedWallet` | `POST /v1/wallet/blocked-address-refund` | PaymentClient |
| `TestWebhook` | `POST /v1/test-webhook/{payment\|wallet}` | PaymentClient |
| `GetBalance` | `POST /v1/balance` | PaymentClient |
| `GetExchangeRates` | `GET /v1/exchange-rate/{currency}/list` | PaymentClient |
| `CreatePayout` | `POST /v1/payout` | PayoutClient |
| `GetInfo` (payout) | `POST /v1/payout/info` | PayoutClient |
| `ListHistory` (payout) | `POST /v1/payout/list` | PayoutClient |
| `CalculateWithdrawal` | `POST /v1/payout/calculate` | PayoutClient |
| `ListServices` (payout) | `POST /v1/payout/services` | PayoutClient |
| `TransferToPersonal` | `POST /v1/transfer/to-personal` | PayoutClient |
| `TransferToBusiness` | `POST /v1/transfer/to-business` | PayoutClient |

## Currency and network codes

Heleket's catalogue evolves. Always source the authoritative list from `ListServices` at runtime rather than hard-coding values. Examples seen at the time of writing:

- Currencies: `BTC`, `ETH`, `USDT`, `USDC`, `DAI`, `LTC`, `BCH`, `XRP`, `TRX`, `TON`, `BNB`, `MATIC`, `DOGE`, `SHIB`, `DASH`, `XMR`
- Networks: `bitcoin`, `ethereum`, `tron`, `bsc`, `polygon`, `ton`, `litecoin`, `bch`, `ripple`, `dogecoin`, `dash`, `monero`

## HTTP / signing reminders

- Most requests use `POST` (even most reads). The only `GET` is `GetExchangeRates` (`/v1/exchange-rate/{currency}/list`).
- All requests include `merchant`, `sign`, `Content-Type: application/json` headers.
- `sign` = `md5(base64(json_body) . apiKey)`. For no-arg endpoints, body and sign are computed over the empty byte slice.
- Webhook payload's `sign` uses the same formula against the corresponding API key.

## Next

→ [11 — Troubleshooting](11-troubleshooting.md).
