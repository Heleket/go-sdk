package withdrawal

type CalculateRequest struct {
	Amount       string  `json:"amount" binding:"required"`
	Address      string  `json:"address" binding:"required"`
	Currency     string  `json:"currency" binding:"required"`
	ToCurrency   *string `json:"to_currency,omitempty"`
	Network      *string `json:"network,omitempty"`
	IsSubtract   bool    `json:"is_subtract" binding:"required"`
	CourseSource *string `json:"course_source,omitempty" binding:"omitempty,oneof=Binance BinanceP2p Exmo Kucoin Garantexio"`
	Priority     *string `json:"priority,omitempty" binding:"omitempty,min=4,max=11,oneof=recommended economy high highest"`
}

type CalculateResponse struct {
	State   int                    `json:"state"`
	Result  CalculateResult        `json:"result,omitempty"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type CalculateResult struct {
	Commission     string `json:"commission"`
	MerchantAmount string `json:"merchant_amount"`
	PayoutAmount   string `json:"payout_amount"`
}
