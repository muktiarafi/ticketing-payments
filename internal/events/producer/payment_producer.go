package producer

import "github.com/muktiarafi/ticketing-payments/internal/entity"

type PaymentProducer interface {
	Created(payment *entity.Payment) error
}
