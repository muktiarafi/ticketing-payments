package repository

import (
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-payments/internal/driver"
	"github.com/muktiarafi/ticketing-payments/internal/entity"
)

type PaymentRepositoryImpl struct {
	*driver.DB
}

func NewPaymentRepository(db *driver.DB) PaymentRepository {
	return &PaymentRepositoryImpl{
		DB: db,
	}
}

func (r *PaymentRepositoryImpl) Insert(payment *entity.Payment) (*entity.Payment, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `INSERT INTO payments (stripe_id, order_id)
	VALUES ($1, $2)
	RETURNING id, stripe_id`

	if err := r.SQL.QueryRowContext(ctx, stmt, payment.StripeID, payment.Order.ID).Scan(
		&payment.ID,
		&payment.StripeID,
	); err != nil {
		return nil, &common.Error{Op: "PaymentRepositoryImpl.Insert", Err: err}
	}

	return payment, nil
}
