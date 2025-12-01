package payment

type RefundBlockedRequest struct {
	Uuid    string `json:"uuid"`
	OrderId string `json:"order_id,omitempty"`
	Address string `json:"address,omitempty"`
}

type RefundBlockedResponse struct {
	State   int                    `json:"state"`
	Result  RefundBlockedResult    `json:"result,omitempty"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type RefundBlockedResult struct {
	Commission string `json:"commission"`
	Amount     string `json:"amount"`
}
