package payment

import "time"

type CreateInvoiceRequest struct {
	Amount                 string                 `json:"amount"`
	Currency               string                 `json:"currency"`
	IsRefresh              bool                   `json:"is_refresh"`
	DiscountPercent        int64                  `json:"discount_percent,omitempty"`
	FromReferralCode       string                 `json:"from_referral_code,omitempty"`
	CourseSource           string                 `json:"course_source,omitempty"`
	ExpectCurrencies       []string               `json:"expect_currencies,omitempty"`
	Currencies             []string               `json:"currencies,omitempty"`
	AdditionalData         map[string]interface{} `json:"additional_data,omitempty"`
	AccuracyPaymentPercent int64                  `json:"accuracy_payment_percent,omitempty"`
	Subtract               int8                   `json:"subtract"`
	ToCurrency             string                 `json:"to_currency"`
	Lifetime               int16                  `json:"lifetime,omitempty"`
	IsPaymentMultiple      bool                   `json:"is_payment_multiple"`
	UrlCallback            string                 `json:"url_callback"`
	UrlSuccess             string                 `json:"url_success"`
	UrlReturn              string                 `json:"url_return"`
	Network                string                 `json:"network"`
	OrderID                string                 `json:"order_id"`
}

type CreateInvoiceResponse struct {
	State   int                    `json:"state"`
	Result  CreateInvoiceResult    `json:"result,omitempty"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type CreateInvoiceResult struct {
	Uuid            string      `json:"uuid"`
	OrderId         string      `json:"order_id"`
	Amount          string      `json:"amount"`
	PaymentAmount   interface{} `json:"payment_amount"`
	PayerAmount     interface{} `json:"payer_amount"`
	DiscountPercent interface{} `json:"discount_percent"`
	Discount        string      `json:"discount"`
	PayerCurrency   interface{} `json:"payer_currency"`
	Currency        string      `json:"currency"`
	MerchantAmount  interface{} `json:"merchant_amount"`
	Network         interface{} `json:"network"`
	Address         interface{} `json:"address"`
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
