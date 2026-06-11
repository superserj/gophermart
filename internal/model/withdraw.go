package model

// WithdrawRequest — тело запроса на списание.
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
