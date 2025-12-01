package withdrawal

type InformationRequest struct {
	Uuid    string `json:"uuid"`
	OrderId string `json:"order_id,omitempty"`
}

type InformationResponse struct {
	State   int                    `json:"state"`
	Result  InformationResult      `json:"result,omitempty"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type InformationResult struct {
	Uuid      string `json:"uuid"`
	Status    string `json:"status"`
	Currency  string `json:"currency"`
	Network   string `json:"network"`
	Amount    string `json:"amount"`
	ToAddress string `json:"to_address"`
}
