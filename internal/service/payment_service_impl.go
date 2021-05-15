package service

import (
	"errors"

	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-payments/internal/entity"
	"github.com/muktiarafi/ticketing-payments/internal/events/producer"
	"github.com/muktiarafi/ticketing-payments/internal/repository"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/charge"
)

type PaymentServiceImpl struct {
	producer.PaymentProducer
	repository.OrderRepository
	repository.PaymentRepository
}

func NewPaymentService(
	paymentProducer producer.PaymentProducer,
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository) PaymentService {
	return &PaymentServiceImpl{
		PaymentProducer:   paymentProducer,
		OrderRepository:   orderRepo,
		PaymentRepository: paymentRepo,
	}
}

func (s *PaymentServiceImpl) Create(token string, userID, orderID int64) (*entity.Payment, error) {
	order, err := s.OrderRepository.FindOne(orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, &common.Error{
			Code:    common.ECONCLICT,
			Op:      "PaymentServiceImpl.Create",
			Message: "Not Authorized",
			Err:     errors.New("trying to access not owned order"),
		}
	}

	if order.Status == "CANCELLED" {
		return nil, &common.Error{
			Code:    common.EINVALID,
			Op:      "PaymentServiceImpl.Create",
			Message: "Order is Cancelled",
			Err:     errors.New("trying to make payment to cancelled order"),
		}
	}

	charge, err := charge.New(&stripe.ChargeParams{
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		Amount:   stripe.Int64(int64(order.Price * 100)),
		Source: &stripe.SourceParams{
			Token: stripe.String(token),
		},
	})
	if err != nil {
		return nil, err
	}

	payment := &entity.Payment{
		StripeID: charge.ID,
		Order:    order,
	}
	newPayment, err := s.PaymentRepository.Insert(payment)
	if err != nil {
		return nil, err
	}

	if err := s.PaymentProducer.Created(payment); err != nil {
		return nil, err
	}

	return newPayment, nil
}
