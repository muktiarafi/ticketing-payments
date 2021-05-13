package repository

import "github.com/muktiarafi/ticketing-payments/internal/entity"

type OrderRepository interface {
	Insert(order *entity.Order) (*entity.Order, error)
	FindOne(orderID int64) (*entity.Order, error)
	Update(order *entity.Order) (*entity.Order, error)
}
