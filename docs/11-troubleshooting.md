# 11 — Troubleshooting

The most common merchant problems and their fixes.

## 1. `signature: INVALID` on webhooks

**Most likely cause:** verifying with the wrong API key.

- Payment + static-wallet webhooks (`type: payment`, `type: wallet`) → **payment** API key
- Payout webhooks (`type: payout`) → **payout** API key

```go
verifier := webhook.NewVerifier(paymentApiKey) // for payment webhooks
verifier := webhook.NewVerifier(payoutApiKey)  // for payout webhooks
```

**Other causes:**

- Middleware mutated the body before your handler read it. Read `r.Body` directly with `io.ReadAll`.
- The payload was re-encoded somewhere in your pipeline with different settings (slash escaping, field ordering, whitespace).
- The webhook came from a sender that does not match Heleket's canonical encoding. Use `--type=payment|payout` with the inspector to confirm.

Run the inspector:

```bash
echo '<paste payload>' | bin/heleket-webhook-inspect --key=$KEY
```

## 2. `*ValidationError` when creating an invoice

`ve.Fields` lists the offending fields. The most-seen ones:

| Field | Common reason |
|---|---|
| `amount` | Empty, non-numeric, or below the minimum for the chosen currency |
| `order_id` | Duplicate (already exists in your merchant invoices) or contains unsupported characters |
| `currency` | Currency code not enabled for your merchant |
| `url_callback` | Not HTTPS, or shorter than 6 chars |

```go
var ve *heleket.ValidationError
if errors.As(err, &ve) {
    for field, msgs := range ve.Fields {
        fmt.Printf("%s: %v\n", field, msgs)
    }
}
```

## 3. `*APIError` with `Wrong key` / `You are forbidden`

The merchant account is restricted or in moderation. Check the dashboard at <https://dash.heleket.com>. New accounts need domain verification before invoices are accepted.

## 4. `context deadline exceeded` / `*HTTPError`

Heleket can take 5–10 seconds during heavy blockchain confirmation. Raise the timeout:

```go
client, _ := heleket.NewPaymentClient(merchantID, paymentKey,
    heleket.WithTimeout(60 * time.Second),
)
```

If the timeout fires on every call, your egress firewall is blocking `api.heleket.com:443`.

## 5. Payments stuck in `check` / `confirm_check`

- `check` = waiting for the transaction to appear in the mempool
- `confirm_check` = seen on-chain, awaiting confirmations (typically 1–6 depending on the network)

Wait. Don't poll faster than every 30 seconds.

`wrong_amount_waiting` means the payer underpaid — the invoice is still open and additional top-ups are accepted.

## 6. Payout API key recently rotated

> "Withdrawals will be temporarily blocked for 24 hours after generating a new payout API key."

Not a bug — schedule rotations accordingly.

## 7. Outbound signature mismatch

If Heleket rejects your **outbound** requests with a signature error, the most common cause is hand-rolled HTTP (the SDK handles signing correctly internally). The recipe is:

```go
body, _ := json.Marshal(params)               // exactly the bytes you'll send
sign := heleket.Sign(body, apiKey)            // md5(base64(body) + apiKey)
req.Header.Set("merchant", merchantID)
req.Header.Set("sign", sign)
req.Header.Set("Content-Type", "application/json")
req.Body = io.NopCloser(bytes.NewReader(body)) // same bytes
```

The body sent must be byte-identical to the body that was signed.

## 8. Webhook never received

- Confirm `url_callback` is publicly reachable (`curl` from a different network).
- Use `TestWebhook` to send a synthetic event without a real payment.
- Check your server logs for 4xx/5xx replies — Heleket retries until you 200; a 200 with broken logic is recorded as delivered.
- Whitelist `31.133.220.8` rather than blocking unknown IPs at the firewall.

## 9. `IsSubtract` confusion on payouts

| Value | Behaviour |
|---|---|
| `true` | Commission deducted from `Amount`. Recipient receives `Amount - commission`. |
| `false` | Commission added on top. Recipient receives `Amount`. Your balance is debited `Amount + commission`. |

Use `CalculateWithdrawal` to preview the numbers before committing.

## 10. Static wallet keeps emitting `is_final=true` per deposit

By design — each top-up generates its own webhook with `is_final=true` for that particular event. The wallet itself remains open and accepts further deposits. Don't close it after one event.

## Still stuck?

Capture the failing request/response with `WithDebug(true)`, then reach out to Heleket support at <https://heleket.com/contacts> with:

- The `order_id` involved
- Approximate timestamps (UTC+3)
- The error type (e.g. `*heleket.APIError`) and message
- A redacted excerpt of the debug output (strip API keys)

## End of documentation

Back to [docs index](README.md) or [README](../README.md).
