package payment

import "time"

type InformationRequest struct {
	Address    string `json:"address"`
	IsSubtract string `json:"is_subtract"`
	Uuid       string `json:"uuid"`
	OrderId    string `json:"order_id"`
}

type InformationResponse struct {
	State   int                    `json:"state"`
	Result  InformationResult      `json:"result"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message"`
}

type InformationResult struct {
	Uuid            string      `json:"uuid"`
	OrderId         string      `json:"order_id"`
	Amount          string      `json:"amount"`
	PaymentAmount   string      `json:"payment_amount"`
	PayerAmount     string      `json:"payer_amount"`
	DiscountPercent int         `json:"discount_percent"`
	Discount        string      `json:"discount"`
	PayerCurrency   string      `json:"payer_currency"`
	Currency        string      `json:"currency"`
	Comments        interface{} `json:"comments"`
	MerchantAmount  string      `json:"merchant_amount"`
	Network         string      `json:"network"`
	Address         string      `json:"address"`
	From            interface{} `json:"from"`
	Txid            interface{} `json:"txid"`
	PaymentStatus   string      `json:"payment_status"`
	Url             string      `json:"url"`
	ExpiredAt       int         `json:"expired_at"`
	Status          string      `json:"status"`
	IsFinal         bool        `json:"is_final"`
	AdditionalData  interface{} `json:"additional_data"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}
