package producer

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-common/types"
	"github.com/muktiarafi/ticketing-payments/internal/entity"
)

type PaymentProducerImpl struct {
	message.Publisher
}

func NewPaymentProducer(publisher message.Publisher) PaymentProducer {
	return &PaymentProducerImpl{
		Publisher: publisher,
	}
}

func (p *PaymentProducerImpl) Created(payment *entity.Payment) error {
	paymentCreatedEventData := &types.PaymentCreatedEvent{
		ID:       payment.ID,
		StripeID: payment.StripeID,
		OrderID:  payment.Order.ID,
	}
	buf, err := paymentCreatedEventData.Marshal()
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), buf)
	return p.Publish(common.PaymentCreated, msg)
}
