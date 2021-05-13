package entity

type Order struct {
	ID      int64   `json:"id"`
	Price   float64 `json:"price"`
	Status  string  `json:"status"`
	Version int64   `json:"version"`
	UserID  int64   `json:"userId"`
}
