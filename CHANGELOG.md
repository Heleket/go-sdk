# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

While the SDK is in `0.x` the public API may still change between minor
versions; it will be frozen at `1.0.0`.

## [0.2.0] - 2026-06-25

### Added
- `PayoutClient.Refund(ctx, RefundRequest) (*RefundResult, error)` →
  `POST /v1/payment/refund`. The endpoint is now signed with the **payout** API
  key, so refunds live on `PayoutClient`. Construct it with
  `NewPayoutClient(merchantID, payoutKey)`.
- `examples/refund` (and `make example-refund`) demonstrating the payout-key
  refund flow.

### Deprecated
- `PaymentClient.Refund` is now a stub that returns the new `ErrRefundMoved`
  sentinel **without issuing a request** — a payment client cannot sign
  `/v1/payment/refund` with the payout key, so it fails loudly instead of
  signing with the wrong one. Use `PayoutClient.Refund`. See
  [`UPGRADING.md`](UPGRADING.md).

## [0.1.1] - 2026-06-17

### Added
- GitHub Actions CI running the `make qa` gates (build, vet, gofmt, race tests,
  staticcheck) on the supported Go versions for every push and pull request.

## [0.1.0] - 2026-06-17

Initial public release — a production-grade Go SDK for the Heleket
cryptocurrency payment API, with zero dependencies beyond the standard library.

### Added
- `NewPaymentClient` / `NewPayoutClient` constructors configured with functional
  options (`WithDebug`, `WithTimeout`, `WithTransport`, `WithUserAgent`,
  `WithMaxRetries`, `WithMaxResponseBytes`, `WithBaseURL`).
- Payments API: invoices, info, AML/KYC/SoF links, history, static wallets, QR
  codes, wallet blocking and blocked-wallet refunds, refunds, webhook resend,
  test webhooks, and the services catalogue.
- Payouts API: payouts, info, history, fee calculation, transfers, and services.
- Balance and exchange-rate endpoints.
- Signed webhook verification (`webhook` subpackage) using constant-time
  signature comparison.
- Swappable HTTP transport (`Transport` interface, default `HTTPTransport`, and
  `internal/testutil.FakeTransport` for offline tests).
- Automatic retry with exponential backoff on transport errors and HTTP 5xx.
- Typed errors (`APIError`, `ValidationError`, `HTTPError`) with `errors.Is` /
  `errors.As` sentinels.
- Status enums (`PaymentStatus`, `PayoutStatus`, `AmlLinkStatus`) with
  `IsFinal()` / `IsSuccessful()` helpers.
- Debug logging via `log/slog` that never emits API keys or the `sign` header.
- `heleket-webhook-inspect` CLI, eleven runnable examples, and a Docker harness.

[0.2.0]: https://github.com/heleket/go-sdk/releases/tag/v0.2.0
[0.1.1]: https://github.com/heleket/go-sdk/releases/tag/v0.1.1
[0.1.0]: https://github.com/heleket/go-sdk/releases/tag/v0.1.0
