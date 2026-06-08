package heleket

import (
	"context"
	"net/url"
)

// PayoutClient is the client for the Heleket Payouts API and balance transfers.
// Construct via NewPayoutClient. Uses the payout API key — distinct from the
// payment key.
//
// PayoutClient is safe for concurrent use across goroutines.
type PayoutClient struct {
	baseClient
}

// NewPayoutClient constructs a PayoutClient with the given merchant credentials.
func NewPayoutClient(merchantID, apiKey string, opts ...Option) (*PayoutClient, error) {
	cfg, err := newConfig(merchantID, apiKey, opts...)
	if err != nil {
		return nil, err
	}
	return &PayoutClient{baseClient: baseClient{config: cfg}}, nil
}

// CreatePayout creates a withdrawal.
func (c *PayoutClient) CreatePayout(ctx context.Context, req CreatePayoutRequest) (*Payout, error) {
	var out Payout
	if err := c.post(ctx, "/v1/payout", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetInfo looks up a payout by UUID or OrderID.
func (c *PayoutClient) GetInfo(ctx context.Context, opts InfoOptions) (*Payout, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}
	var out Payout
	if err := c.post(ctx, "/v1/payout/info", opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListHistory returns paginated payout history.
func (c *PayoutClient) ListHistory(ctx context.Context, opts HistoryOptions) (*HistoryPage[Payout], error) {
	path := "/v1/payout/list"
	if opts.Cursor != "" {
		path += "?cursor=" + url.QueryEscape(opts.Cursor)
	}
	var out HistoryPage[Payout]
	if err := c.post(ctx, path, opts, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CalculateWithdrawal previews the commission and final amount for a payout.
func (c *PayoutClient) CalculateWithdrawal(ctx context.Context, req CalculateRequest) (*Calculation, error) {
	var out Calculation
	if err := c.post(ctx, "/v1/payout/calculate", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListServices returns the catalogue of supported (currency, network) pairs
// for payouts.
func (c *PayoutClient) ListServices(ctx context.Context) ([]Service, error) {
	var out []Service
	if err := c.post(ctx, "/v1/payout/services", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TransferToPersonal moves funds from the business balance to the personal wallet.
func (c *PayoutClient) TransferToPersonal(ctx context.Context, req TransferRequest) (*Transfer, error) {
	var out Transfer
	if err := c.post(ctx, "/v1/transfer/to-personal", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TransferToBusiness moves funds from the personal balance to the business balance.
func (c *PayoutClient) TransferToBusiness(ctx context.Context, req TransferRequest) (*Transfer, error) {
	var out Transfer
	if err := c.post(ctx, "/v1/transfer/to-business", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
