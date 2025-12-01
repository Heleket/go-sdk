package payment

type BlockStaticWalletRequest struct {
	UUID          string `json:"uuid"`
	OrderId       string `json:"order_id"`
	IsForceRefund bool   `json:"is_force_refund"`
}

type BlockStaticWalletResponse struct {
	State   int                     `json:"state"`
	Result  BlockStaticWalletResult `json:"result,omitempty"`
	Errors  map[string]interface{}  `json:"errors,omitempty"`
	Message string                  `json:"message,omitempty"`
}

type BlockStaticWalletResult struct {
	Uuid   string `json:"uuid"`
	Status string `json:"status"`
}
