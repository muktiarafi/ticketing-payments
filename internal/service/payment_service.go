package service

import (
	"github.com/muktiarafi/ticketing-payments/internal/entity"
)

type PaymentService interface {
	Create(token string, userID, orderID int64) (*entity.Payment, error)
}
