package payment

type GenerateQrRequest struct {
	WalletAddressUUID string `json:"wallet_address_uuid"`
}

type GenerateQrResponse struct {
	State   int                    `json:"state"`
	Result  GenerateQrResult       `json:"result"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type GenerateQrResult struct {
	Image string `json:"image"`
}
