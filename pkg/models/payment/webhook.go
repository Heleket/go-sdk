package payment

type WebhookData struct {
	Type              string      `json:"type"`
	Uuid              string      `json:"uuid"`
	OrderId           string      `json:"order_id"`
	Amount            string      `json:"amount"`
	PaymentAmount     string      `json:"payment_amount"`
	PaymentAmountUsd  string      `json:"payment_amount_usd"`
	MerchantAmount    string      `json:"merchant_amount"`
	Commission        string      `json:"commission"`
	IsFinal           bool        `json:"is_final"`
	Status            string      `json:"status"`
	From              string      `json:"from"`
	WalletAddressUuid interface{} `json:"wallet_address_uuid"`
	Network           string      `json:"network"`
	Currency          string      `json:"currency"`
	PayerCurrency     string      `json:"payer_currency"`
	AdditionalData    interface{} `json:"additional_data"`
	Convert           struct {
		ToCurrency string      `json:"to_currency"`
		Commission interface{} `json:"commission"`
		Rate       string      `json:"rate"`
		Amount     string      `json:"amount"`
	} `json:"convert"`
	Txid string `json:"txid"`
	Sign string `json:"sign"`
}
