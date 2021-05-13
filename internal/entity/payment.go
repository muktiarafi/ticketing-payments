package entity

type Payment struct {
	ID       int64  `json:"id"`
	StripeID string `json:"stripeId"`
	*Order   `json:"order"`
}
