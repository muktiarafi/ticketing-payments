package consumer

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/muktiarafi/ticketing-common/types"
	"github.com/muktiarafi/ticketing-payments/internal/entity"
	"github.com/muktiarafi/ticketing-payments/internal/repository"
)

type PaymentConsumer struct {
	repository.OrderRepository
}

func NewPaymentConsumer(orderRepo repository.OrderRepository) *PaymentConsumer {
	return &PaymentConsumer{
		OrderRepository: orderRepo,
	}
}

func (c *PaymentConsumer) OrderCreated(msg *message.Message) error {
	orderCreatedEventData := new(types.OrderCreatedEvent)
	if err := orderCreatedEventData.Unmarshal(msg.Payload); err != nil {
		msg.Nack()
		return err
	}

	order := &entity.Order{
		ID:      orderCreatedEventData.ID,
		Price:   orderCreatedEventData.TicketPrice,
		Status:  orderCreatedEventData.Status,
		UserID:  orderCreatedEventData.UserID,
		Version: orderCreatedEventData.Version,
	}

	if _, err := c.OrderRepository.Insert(order); err != nil {
		msg.Nack()
		return err
	}

	msg.Ack()

	return nil
}

func (c *PaymentConsumer) OrderCancelled(msg *message.Message) error {
	orderCancelledEventData := new(types.OrderCancelledEvent)
	if err := orderCancelledEventData.Unmarshal(msg.Payload); err != nil {
		msg.Nack()
		return err
	}

	order := &entity.Order{
		ID:      orderCancelledEventData.ID,
		Status:  "CANCELLED",
		Version: orderCancelledEventData.Version,
	}

	if _, err := c.OrderRepository.Update(order); err != nil {
		msg.Nack()
		return err
	}

	msg.Ack()

	return nil
}
