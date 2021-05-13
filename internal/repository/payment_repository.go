package repository

import "github.com/muktiarafi/ticketing-payments/internal/entity"

type PaymentRepository interface {
	Insert(payment *entity.Payment) (*entity.Payment, error)
}
