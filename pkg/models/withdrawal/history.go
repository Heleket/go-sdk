package withdrawal

import "time"

type HistoryRequest struct {
	DateFrom *time.Time `json:"date_from,omitempty" form:"date_from"`
	DateTo   *time.Time `json:"date_to,omitempty" form:"date_to"`
}

type HistoryResponse struct {
	State   int                    `json:"state"`
	Result  HistoryResult          `json:"result"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type HistoryItem struct {
	Uuid      string    `json:"uuid"`
	Amount    string    `json:"amount"`
	Currency  string    `json:"currency"`
	Network   string    `json:"network"`
	Address   string    `json:"address"`
	Txid      string    `json:"txid,omitempty"`
	Status    string    `json:"status"`
	IsFinal   bool      `json:"is_final"`
	Balance   string    `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HistoryPagination struct {
	Count          int         `json:"count"`
	HasPages       bool        `json:"hasPages"`
	NextCursor     string      `json:"nextCursor"`
	PreviousCursor interface{} `json:"previousCursor"`
	PerPage        int         `json:"perPage"`
}

type HistoryResult struct {
	Items    []HistoryItem     `json:"items"`
	Paginate HistoryPagination `json:"paginate"`
}
