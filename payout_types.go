package heleket

import "encoding/json"

// CreatePayoutRequest is the body of POST /v1/payout. Required fields:
// Amount, Currency, OrderID, Address, IsSubtract.
type CreatePayoutRequest struct {
	// Required.
	Amount     string `json:"amount"`
	Currency   string `json:"currency"`
	OrderID    string `json:"order_id"`
	Address    string `json:"address"`
	IsSubtract bool   `json:"is_subtract"`

	// Optional.
	Network      string `json:"network,omitempty"`
	URLCallback  string `json:"url_callback,omitempty"`
	ToCurrency   string `json:"to_currency,omitempty"`
	CourseSource string `json:"course_source,omitempty"`
	FromCurrency string `json:"from_currency,omitempty"`
	Priority     string `json:"priority,omitempty"`
	Memo         string `json:"memo,omitempty"`
}

// Payout is the response of POST /v1/payout and POST /v1/payout/info.
//
// Balance and PayerAmount are typed as json.RawMessage because the Heleket
// server returns them in different shapes depending on the endpoint and
// payment state (object, string, or null). Decode into your own struct or
// keep as bytes for logging.
type Payout struct {
	UUID           string          `json:"uuid"`
	Amount         string          `json:"amount"`
	Currency       string          `json:"currency"`
	Commission     string          `json:"commission,omitempty"`
	MerchantAmount string          `json:"merchant_amount,omitempty"`
	Network        string          `json:"network,omitempty"`
	Address        string          `json:"address,omitempty"`
	TxID           *string         `json:"txid,omitempty"`
	Status         PayoutStatus    `json:"status"`
	IsFinal        bool            `json:"is_final"`
	Balance        json.RawMessage `json:"balance,omitempty"`
	PayerCurrency  string          `json:"payer_currency,omitempty"`
	PayerAmount    json.RawMessage `json:"payer_amount,omitempty"`
	OrderID        string          `json:"order_id,omitempty"`
	CreatedAt      string          `json:"created_at,omitempty"`
	UpdatedAt      string          `json:"updated_at,omitempty"`
}

// CalculateRequest is the body of POST /v1/payout/calculate.
type CalculateRequest struct {
	Currency   string `json:"currency"`
	Network    string `json:"network"`
	Amount     string `json:"amount"`
	IsSubtract bool   `json:"is_subtract,omitempty"`
}

// Calculation is the result of CalculateWithdrawal.
type Calculation struct {
	Commission     string `json:"commission"`
	MerchantAmount string `json:"merchant_amount,omitempty"`
	Amount         string `json:"amount,omitempty"`
	Network        string `json:"network,omitempty"`
	Currency       string `json:"currency,omitempty"`
}

// TransferRequest is the body of POST /v1/transfer/to-personal and
// /v1/transfer/to-business.
type TransferRequest struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// Transfer is the response of TransferToPersonal / TransferToBusiness.
type Transfer struct {
	UserUUID     string `json:"user_wallet_uuid,omitempty"`
	MerchantUUID string `json:"merchant_wallet_uuid,omitempty"`
	Amount       string `json:"amount,omitempty"`
	Currency     string `json:"currency,omitempty"`
}
