package payment

import "time"

type HistoryRequest struct {
	DateFrom *time.Time `json:"date_from,omitempty"`
	DateTo   *time.Time `json:"date_to,omitempty"`
}

type HistoryResponse struct {
	State   int                    `json:"state"`
	Result  HistoryResult          `json:"result"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type HistoryResult struct {
	Items    []HistoryInvoiceItem `json:"items"`
	Paginate HistoryPagination    `json:"paginate"`
}

type HistoryPagination struct {
	Count          int     `json:"count"`
	HasPages       bool    `json:"hasPages"`
	NextCursor     *string `json:"nextCursor"`
	PreviousCursor *string `json:"previousCursor"`
	PerPage        int     `json:"perPage"`
}

type HistoryInvoiceItem struct {
	Uuid            string      `json:"uuid"`
	OrderId         string      `json:"order_id"`
	Amount          string      `json:"amount"`
	PaymentAmount   string      `json:"payment_amount"`
	PayerAmount     string      `json:"payer_amount"`
	DiscountPercent int         `json:"discount_percent"`
	Discount        string      `json:"discount"`
	PayerCurrency   string      `json:"payer_currency"`
	Currency        string      `json:"currency"`
	MerchantAmount  string      `json:"merchant_amount"`
	Comments        interface{} `json:"comments"`
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
