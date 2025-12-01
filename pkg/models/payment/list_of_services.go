package payment

type ListOfServicesResponse struct {
	State   int                    `json:"state"`
	Result  []Result               `json:"result"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type Result struct {
	Network     string `json:"network"`
	Currency    string `json:"currency"`
	IsAvailable bool   `json:"is_available"`
	Limit       struct {
		MinAmount string `json:"min_amount"`
		MaxAmount string `json:"max_amount"`
	} `json:"limit"`
	Commission struct {
		FeeAmount string `json:"fee_amount"`
		Percent   string `json:"percent"`
	} `json:"commission"`
}
