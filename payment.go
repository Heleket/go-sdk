package heleket

import (
	"context"
	"net/url"
)

// PaymentClient is the client for the Heleket Payments API plus the
// balance and exchange-rate endpoints. Construct via NewPaymentClient.
//
// PaymentClient is safe for concurrent use across goroutines.
type PaymentClient struct {
	baseClient
}

// NewPaymentClient constructs a PaymentClient with the given merchant
// credentials.
//
//	client, err := heleket.NewPaymentClient(merchantID, paymentKey,
//	    heleket.WithDebug(true),
//	    heleket.WithTimeout(60 * time.Second),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewPaymentClient(merchantID, apiKey string, opts ...Option) (*PaymentClient, error) {
	cfg, err := newConfig(merchantID, apiKey, opts...)
	if err != nil {
		return nil, err
	}
	return &PaymentClient{baseClient: baseClient{config: cfg}}, nil
}

// CreateInvoice creates a new invoice.
func (c *PaymentClient) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
	var out Invoice
	if err := c.post(ctx, "/v1/payment", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetInfo looks up an invoice by UUID or OrderID. Set exactly one.
func (c *PaymentClient) GetInfo(ctx context.Context, opts InfoOptions) (*Invoice, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	var out Invoice
	if err := c.post(ctx, "/v1/payment/info", opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAmlLinks returns the AML/KYC/SoF questionnaire links for a blocked
// (locked) payment. Set exactly one of UUID or OrderID; the server prioritises
// OrderID when both are sent.
//
// Each returned AmlLink carries the questionnaire URL to hand to the end user,
// its expiry, and a status (see AmlLinkStatus). Completing the questionnaires
// unblocks a held payment.
func (c *PaymentClient) GetAmlLinks(ctx context.Context, opts InfoOptions) ([]AmlLink, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	var out []AmlLink
	if err := c.post(ctx, "/v1/payment/aml-links", opts, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListHistory returns paginated payment history.
func (c *PaymentClient) ListHistory(ctx context.Context, opts HistoryOptions) (*HistoryPage[Invoice], error) {
	path := "/v1/payment/list"
	if opts.Cursor != "" {
		path += "?cursor=" + url.QueryEscape(opts.Cursor)
	}
	var out HistoryPage[Invoice]
	if err := c.post(ctx, path, opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateStaticWallet creates a persistent top-up wallet bound to an order ID.
func (c *PaymentClient) CreateStaticWallet(ctx context.Context, req CreateStaticWalletRequest) (*StaticWallet, error) {
	var out StaticWallet
	if err := c.post(ctx, "/v1/wallet", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GenerateQRCode returns a base64 QR-code image for an existing static wallet.
func (c *PaymentClient) GenerateQRCode(ctx context.Context, req GenerateQRCodeRequest) (*QRCode, error) {
	var out QRCode
	if err := c.post(ctx, "/v1/wallet/qr", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BlockStaticWallet stops a static wallet from accepting further transfers.
func (c *PaymentClient) BlockStaticWallet(ctx context.Context, req BlockStaticWalletRequest) (*BlockedWallet, error) {
	if err := req.validate(); err != nil {
		return nil, err
	}
	var out BlockedWallet
	if err := c.post(ctx, "/v1/wallet/block-address", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RefundBlockedWallet sends the contents of a blocked wallet to a recovery address.
func (c *PaymentClient) RefundBlockedWallet(ctx context.Context, req RefundBlockedWalletRequest) (*RefundResult, error) {
	var out RefundResult
	if err := c.post(ctx, "/v1/wallet/blocked-address-refund", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Refund is no longer available on PaymentClient.
//
// The /v1/payment/refund endpoint is now signed with the PAYOUT API key, which a
// PaymentClient does not hold, so it can no longer produce a valid signature for
// a refund. This stub returns [ErrRefundMoved] without issuing a request, rather
// than silently signing with the wrong key. Use [PayoutClient.Refund] instead:
//
//	payout, _ := heleket.NewPayoutClient(merchantID, payoutKey)
//	payout.Refund(ctx, req)
//
// Deprecated: refunds moved to [PayoutClient.Refund] in v0.2.0. See UPGRADING.md.
func (c *PaymentClient) Refund(context.Context, RefundRequest) (*RefundResult, error) {
	return nil, ErrRefundMoved
}

// ResendWebhook asks Heleket to redeliver the last webhook for the invoice.
func (c *PaymentClient) ResendWebhook(ctx context.Context, opts InfoOptions) error {
	if err := opts.validate(); err != nil {
		return err
	}
	return c.post(ctx, "/v1/payment/resend", opts, nil)
}

// TestWebhook sends a synthetic event to the merchant's callback URL.
func (c *PaymentClient) TestWebhook(ctx context.Context, req TestWebhookRequest) error {
	if err := req.validate(); err != nil {
		return err
	}
	return c.post(ctx, "/v1/test-webhook/"+req.Type, req, nil)
}

// ListServices returns the catalogue of supported (currency, network) pairs
// for payments.
func (c *PaymentClient) ListServices(ctx context.Context) ([]Service, error) {
	var out []Service
	if err := c.post(ctx, "/v1/payment/services", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBalance returns merchant + personal balances by currency.
func (c *PaymentClient) GetBalance(ctx context.Context) (*Balance, error) {
	// The API returns: { state, result: [ { balance: { merchant: [...], user: [...] } } ] }
	// We decode into a wrapper and pull out the balance object.
	type wrapper struct {
		Balance Balance `json:"balance"`
	}
	var wrappers []wrapper
	if err := c.post(ctx, "/v1/balance", nil, &wrappers); err != nil {
		return nil, err
	}
	if len(wrappers) == 0 {
		return &Balance{}, nil
	}
	out := wrappers[0].Balance
	return &out, nil
}

// GetExchangeRates returns rates for the given fiat currency.
//
// Read-only endpoint: issued as a GET with the currency in the path.
func (c *PaymentClient) GetExchangeRates(ctx context.Context, currency string) ([]ExchangeRate, error) {
	var out []ExchangeRate
	path := "/v1/exchange-rate/" + url.PathEscape(currency) + "/list"
	if err := c.get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}
