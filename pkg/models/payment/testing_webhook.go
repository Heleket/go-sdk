package payment

type TestingWebhookRequest struct {
	UrlCallback string `json:"url_callback"`
	Currency    string `json:"currency"`
	Network     string `json:"network"`
	Uuid        string `json:"uuid"`
	OrderId     string `json:"order_id"`
	Status      string `json:"status"`
}

type TestingWebhookResponse struct {
	State   int                    `json:"state"`
	Result  interface{}            `json:"result,omitempty"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}
