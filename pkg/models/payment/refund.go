package payment

type RefundRequest struct {
	Address    string  `json:"address" binding:"required"`
	IsSubtract bool    `json:"is_subtract" binding:"required"`
	UUID       string  `json:"uuid,omitempty" binding:"required_without=OrderID"`
	OrderID    *string `json:"order_id,omitempty" binding:"required_without=UUID,min=1,max=128,alpha_dash"`
	Amount     *string `json:"amount,omitempty" binding:"max=40"`
}

type RefundResponse struct {
	State   int                    `json:"state"`
	Result  []interface{}          `json:"result,omitempty"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}
