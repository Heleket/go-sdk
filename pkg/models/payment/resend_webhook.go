package payment

type ResendWebhookRequest struct {
	Uuid    string `json:"uuid"`
	OrderId string `json:"order_id"`
}

type ResendWebhookResponse struct {
	State   int                    `json:"state"`
	Result  interface{}            `json:"result,omitempty"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}
