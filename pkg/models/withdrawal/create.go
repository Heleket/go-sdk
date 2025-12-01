package withdrawal

type CreateWithdrawalRequest struct {
	Amount       string  `json:"amount" validate:"required"`
	Currency     string  `json:"currency" validate:"required"`
	OrderID      string  `json:"order_id" validate:"required,min=1,max=100,alpha_dash"`
	Address      string  `json:"address" validate:"required"`
	IsSubtract   bool    `json:"is_subtract" validate:"required"`
	Network      string  `json:"network" validate:"required"`
	URLCallback  *string `json:"url_callback,omitempty"`
	ToCurrency   *string `json:"to_currency,omitempty"`
	CourseSource *string `json:"course_source,omitempty" validate:"omitempty,oneof=Binance BinanceP2p Exmo Kucoin Garantexio"`
	FromCurrency *string `json:"from_currency,omitempty"`
	Priority     *string `json:"priority,omitempty" validate:"omitempty,min=4,max=11,oneof=recommended economy high highest"`
	Memo         *string `json:"memo,omitempty" validate:"omitempty,min=1,max=30"`
}

type CreateWithdrawalResponse struct {
	State  int                    `json:"state"`
	Result CreateWithdrawalResult `json:"result,omitempty"`
	Errors map[string]interface{} `json:"errors,omitempty"`
}

type CreateWithdrawalResult struct {
	Uuid      string `json:"uuid"`
	Status    string `json:"status"`
	Currency  string `json:"currency"`
	Network   string `json:"network"`
	Amount    string `json:"amount"`
	ToAddress string `json:"to_address"`
}
