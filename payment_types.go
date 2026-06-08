package heleket

// CreateInvoiceRequest is the body of POST /v1/payment.
// Only Amount, Currency, and OrderID are required. See
// https://doc.heleket.com/methods/payments/creating-invoice for field semantics.
type CreateInvoiceRequest struct {
	// Required.
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	OrderID  string `json:"order_id"`

	// Optional.
	Network                string            `json:"network,omitempty"`
	URLReturn              string            `json:"url_return,omitempty"`
	URLSuccess             string            `json:"url_success,omitempty"`
	URLCallback            string            `json:"url_callback,omitempty"`
	IsPaymentMultiple      *bool             `json:"is_payment_multiple,omitempty"`
	Lifetime               int               `json:"lifetime,omitempty"`
	ToCurrency             string            `json:"to_currency,omitempty"`
	Subtract               int               `json:"subtract,omitempty"`
	AccuracyPaymentPercent string            `json:"accuracy_payment_percent,omitempty"`
	AdditionalData         string            `json:"additional_data,omitempty"`
	Currencies             []CurrencyNetwork `json:"currencies,omitempty"`
	ExceptCurrencies       []CurrencyNetwork `json:"except_currencies,omitempty"`
	CourseSource           CourseSource      `json:"course_source,omitempty"`
	FromReferralCode       string            `json:"from_referral_code,omitempty"`
	DiscountPercent        int               `json:"discount_percent,omitempty"`
	IsRefresh              *bool             `json:"is_refresh,omitempty"`
	PayerEmail             string            `json:"payer_email,omitempty"`
}

// CurrencyNetwork pairs a currency with an optional network, as accepted by
// CreateInvoiceRequest.Currencies / ExceptCurrencies.
type CurrencyNetwork struct {
	Currency string `json:"currency"`
	Network  string `json:"network,omitempty"`
}

// Invoice is the response of POST /v1/payment and POST /v1/payment/info.
type Invoice struct {
	UUID             string        `json:"uuid"`
	OrderID          string        `json:"order_id"`
	Amount           string        `json:"amount"`
	PaymentAmount    *string       `json:"payment_amount,omitempty"`
	PaymentAmountUSD *string       `json:"payment_amount_usd,omitempty"`
	PayerAmount      string        `json:"payer_amount,omitempty"`
	PayerCurrency    string        `json:"payer_currency,omitempty"`
	Currency         string        `json:"currency"`
	MerchantAmount   *string       `json:"merchant_amount,omitempty"`
	Network          string        `json:"network,omitempty"`
	Address          string        `json:"address,omitempty"`
	From             *string       `json:"from,omitempty"`
	TxID             *string       `json:"txid,omitempty"`
	PaymentStatus    PaymentStatus `json:"payment_status,omitempty"`
	Status           PaymentStatus `json:"status,omitempty"`
	URL              string        `json:"url,omitempty"`
	ExpiredAt        int64         `json:"expired_at,omitempty"`
	IsFinal          bool          `json:"is_final"`
	AdditionalData   *string       `json:"additional_data,omitempty"`
	CreatedAt        string        `json:"created_at,omitempty"`
	UpdatedAt        string        `json:"updated_at,omitempty"`
	Commission       string        `json:"commission,omitempty"`
	AddressQRCode    string        `json:"address_qr_code,omitempty"`
	DiscountPercent  int           `json:"discount_percent,omitempty"`
	Discount         string        `json:"discount,omitempty"`
	Convert          *Conversion   `json:"convert,omitempty"`
}

// Conversion describes an auto-conversion applied by Heleket.
type Conversion struct {
	ToCurrency string `json:"to_currency"`
	Commission string `json:"commission"`
	Rate       string `json:"rate"`
	Amount     string `json:"amount"`
}

// InfoOptions identifies an invoice (or payout) for read endpoints.
// Set exactly one of UUID or OrderID. The server prioritises OrderID if both
// are sent — the SDK still rejects empty input client-side.
type InfoOptions struct {
	UUID    string `json:"uuid,omitempty"`
	OrderID string `json:"order_id,omitempty"`
}

func (o InfoOptions) validate() error {
	if o.UUID == "" && o.OrderID == "" {
		return ErrIdentifierRequired
	}
	return nil
}

// HistoryOptions parameterises the history endpoints. Date format is
// "YYYY-MM-DD HH:MM:SS". Cursor comes from a prior page's Pagination.NextCursor.
type HistoryOptions struct {
	DateFrom string `json:"date_from,omitempty"`
	DateTo   string `json:"date_to,omitempty"`
	Cursor   string `json:"-"` // sent as ?cursor= query string
}

// HistoryPage is the paginated response shape for /v1/payment/list and
// /v1/payout/list. T is the per-item type (Invoice or Payout).
type HistoryPage[T any] struct {
	Items    []T        `json:"items"`
	Paginate Pagination `json:"paginate"`
}

// Pagination is the cursor block returned alongside Items.
type Pagination struct {
	Count          int    `json:"count"`
	HasPages       bool   `json:"hasPages"`
	NextCursor     string `json:"nextCursor,omitempty"`
	PreviousCursor string `json:"previousCursor,omitempty"`
	PerPage        int    `json:"perPage"`
}

// CreateStaticWalletRequest is the body of POST /v1/wallet.
type CreateStaticWalletRequest struct {
	Currency    string `json:"currency"`
	Network     string `json:"network"`
	OrderID     string `json:"order_id"`
	URLCallback string `json:"url_callback,omitempty"`
}

// StaticWallet is the response of POST /v1/wallet.
type StaticWallet struct {
	UUID     string `json:"uuid"`
	WalletID string `json:"wallet_uuid,omitempty"`
	OrderID  string `json:"order_id"`
	Address  string `json:"address"`
	Network  string `json:"network"`
	Currency string `json:"currency"`
	URL      string `json:"url,omitempty"`
}

// QRCode is the response of POST /v1/wallet/qr. Image is a base64 data URI.
type QRCode struct {
	Image string `json:"image"`
}

// GenerateQRCodeRequest is the body of POST /v1/wallet/qr.
type GenerateQRCodeRequest struct {
	MerchantPaymentUUID string `json:"merchant_payment_uuid"`
}

// RefundBlockedWalletRequest is the body of POST /v1/wallet/blocked-address-refund.
type RefundBlockedWalletRequest struct {
	UUID    string `json:"uuid"`
	Address string `json:"address"`
}

// BlockStaticWalletRequest is the body of POST /v1/wallet/block-address.
type BlockStaticWalletRequest struct {
	UUID     string `json:"uuid,omitempty"`
	OrderID  string `json:"order_id,omitempty"`
	IsRefund bool   `json:"is_refund,omitempty"`
}

func (r BlockStaticWalletRequest) validate() error {
	if r.UUID == "" && r.OrderID == "" {
		return ErrIdentifierRequired
	}
	return nil
}

// BlockedWallet is the response of POST /v1/wallet/block-address.
type BlockedWallet struct {
	UUID    string `json:"uuid"`
	Status  string `json:"status"`
	OrderID string `json:"order_id,omitempty"`
}

// RefundRequest is the body of POST /v1/payment/refund.
type RefundRequest struct {
	UUID       string `json:"uuid,omitempty"`
	OrderID    string `json:"order_id,omitempty"`
	Address    string `json:"address"`
	IsSubtract bool   `json:"is_subtract"`
}

func (r RefundRequest) validate() error {
	if r.UUID == "" && r.OrderID == "" {
		return ErrIdentifierRequired
	}
	return nil
}

// RefundResult is returned by refund endpoints.
type RefundResult struct {
	Commission string `json:"commission,omitempty"`
	Amount     string `json:"amount,omitempty"`
	Address    string `json:"address,omitempty"`
}

// TestWebhookRequest triggers a synthetic webhook to the merchant's callback URL.
// Type must be "payment" or "wallet".
type TestWebhookRequest struct {
	Type        string        `json:"-"` // path segment, not body
	URLCallback string        `json:"url_callback"`
	Currency    string        `json:"currency"`
	Network     string        `json:"network"`
	Status      PaymentStatus `json:"status"`
	UUID        string        `json:"uuid,omitempty"`
	OrderID     string        `json:"order_id,omitempty"`
}

func (r TestWebhookRequest) validate() error {
	if r.Type != "payment" && r.Type != "wallet" {
		return ErrInvalidTestWebhookType
	}
	if r.UUID == "" && r.OrderID == "" {
		return ErrIdentifierRequired
	}
	return nil
}

// Service describes a (currency, network) pair available for payments or payouts.
type Service struct {
	Network     string            `json:"network"`
	Currency    string            `json:"currency"`
	IsAvailable bool              `json:"is_available"`
	Limit       ServiceLimit      `json:"limit,omitempty"`
	Commission  ServiceCommission `json:"commission,omitempty"`
}

// ServiceLimit describes the per-pair amount bounds.
type ServiceLimit struct {
	MinAmount string `json:"min_amount,omitempty"`
	MaxAmount string `json:"max_amount,omitempty"`
}

// ServiceCommission describes the per-pair fees.
type ServiceCommission struct {
	FeeAmount string `json:"fee_amount,omitempty"`
	Percent   string `json:"percent,omitempty"`
}

// Balance is the response of POST /v1/balance.
type Balance struct {
	Merchant []WalletBalance `json:"merchant"`
	User     []WalletBalance `json:"user"`
}

// WalletBalance is a per-currency entry in a Balance.
type WalletBalance struct {
	UUID         string `json:"uuid,omitempty"`
	Balance      string `json:"balance"`
	CurrencyCode string `json:"currency_code"`
}

// ExchangeRate is one entry in the response of /v1/exchange-rate/{currency}/list.
type ExchangeRate struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Course string `json:"course"`
	Source string `json:"source,omitempty"`
}
