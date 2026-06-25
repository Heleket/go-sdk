# Upgrading

## 0.1 → 0.2

`Refund` moved from `PaymentClient` to `PayoutClient`. The Heleket API now signs
`POST /v1/payment/refund` with the **payout** API key, so a payment client can no
longer produce a valid signature for it.

`PaymentClient.Refund` is kept as a deprecated stub that returns `ErrRefundMoved`
without issuing a request — it fails loudly rather than silently signing with the
wrong key. Switch to `PayoutClient.Refund`:

```go
// Before (0.1 and earlier)
payment, _ := heleket.NewPaymentClient(merchantID, paymentKey)
payment.Refund(ctx, heleket.RefundRequest{
    UUID:       invoiceUUID,
    Address:    "TBaCkAdDrEsS",
    IsSubtract: true,
})

// After (0.2)
payout, _ := heleket.NewPayoutClient(merchantID, payoutKey)
payout.Refund(ctx, heleket.RefundRequest{
    UUID:       invoiceUUID,
    Address:    "TBaCkAdDrEsS",
    IsSubtract: true,
})
```

Detect the deprecated path at runtime with the sentinel:

```go
if errors.Is(err, heleket.ErrRefundMoved) {
    // you called PaymentClient.Refund — switch to PayoutClient.Refund
}
```

Everything else is unchanged — this is an additive minor release.
