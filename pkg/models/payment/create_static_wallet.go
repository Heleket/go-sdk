package payment

type CreateStaticWalletRequest struct {
	Currency         string `json:"currency"`
	Network          string `json:"network"`
	OrderID          string `json:"order_id"`
	UrlCallback      string `json:"url_callback"`
	FromReferralCode string `json:"from_referral_code"`
}

type CreateStaticWalletResponse struct {
	State   int                      `json:"state"`
	Result  CreateStaticWalletResult `json:"result"`
	Errors  map[string]interface{}   `json:"errors,omitempty"`
	Message string                   `json:"message,omitempty"`
}

type CreateStaticWalletResult struct {
	WalletUuid string `json:"wallet_uuid"`
	Uuid       string `json:"uuid"`
	Address    string `json:"address"`
	Network    string `json:"network"`
	Currency   string `json:"currency"`
	Url        string `json:"url"`
}
