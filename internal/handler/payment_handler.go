package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-payments/internal/events/producer"
	"github.com/muktiarafi/ticketing-payments/internal/model"
	"github.com/muktiarafi/ticketing-payments/internal/service"
)

type PaymentHandler struct {
	producer.PaymentProducer
	service.PaymentService
}

func NewPaymentHandler(paymentProducer producer.PaymentProducer, paymentSrv service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		PaymentProducer: paymentProducer,
		PaymentService:  paymentSrv,
	}
}

func (h *PaymentHandler) Route(router *echo.Echo) {
	payments := router.Group("/api/payments", common.RequireAuth)
	payments.POST("", h.New)
}

func (h *PaymentHandler) New(c echo.Context) error {
	userPayload, ok := c.Get("userPayload").(*common.UserPayload)
	const op = "PaymentHandler.New"
	if !ok {
		return &common.Error{
			Op:  op,
			Err: errors.New("missing payload in context"),
		}
	}

	paymentDTO := new(model.PaymentDTO)
	if err := c.Bind(paymentDTO); err != nil {
		return &common.Error{Op: op, Err: err}
	}

	if err := c.Validate(paymentDTO); err != nil {
		return err
	}

	payment, err := h.PaymentService.Create(
		paymentDTO.Token,
		int64(userPayload.ID),
		paymentDTO.OrderID,
	)
	if err != nil {
		return err
	}

	if err := h.PaymentProducer.Created(payment); err != nil {
		return err
	}

	return common.NewResponse(http.StatusCreated, "Created", payment).SendJSON(c)
}
