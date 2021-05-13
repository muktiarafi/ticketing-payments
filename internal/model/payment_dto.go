package model

type PaymentDTO struct {
	Token   string `json:"token" validate:"required"`
	OrderID int64  `json:"orderId" validate:"required"`
}
