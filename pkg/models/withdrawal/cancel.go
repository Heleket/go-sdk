package withdrawal

type CancelWithdrawalRequest struct {
	Uuid    string `json:"uuid"`
	OrderId string `json:"order_id,omitempty"`
}

type CancelWithdrawalResponse struct {
	State   int                    `json:"state"`
	Result  CancelWithdrawalResult `json:"result,omitempty"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type CancelWithdrawalResult struct {
	Uuid   string `json:"uuid"`
	Status string `json:"status"`
}
