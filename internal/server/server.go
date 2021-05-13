package server

import (
	"log"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-payments/internal/config"
	"github.com/muktiarafi/ticketing-payments/internal/driver"
	"github.com/muktiarafi/ticketing-payments/internal/events/consumer"
	"github.com/muktiarafi/ticketing-payments/internal/events/producer"
	"github.com/muktiarafi/ticketing-payments/internal/handler"
	custommiddleware "github.com/muktiarafi/ticketing-payments/internal/middleware"
	"github.com/muktiarafi/ticketing-payments/internal/repository"
	"github.com/muktiarafi/ticketing-payments/internal/service"
	"github.com/muktiarafi/ticketing-payments/internal/utils"
)

func SetupServer() *echo.Echo {
	e := echo.New()
	p := custommiddleware.NewPrometheus("echo", nil)
	p.Use(e)

	val := validator.New()
	trans := common.NewDefaultTranslator(val)
	customValidator := &common.CustomValidator{val, trans}
	e.Validator = customValidator
	e.HTTPErrorHandler = common.CustomErrorHandler
	e.Use(middleware.Logger())

	db, err := driver.ConnectSQL(config.PostgresDSN())
	if err != nil {
		log.Fatal(err)
	}

	utils.InitStripe()
	orderRepository := repository.NewOrderRepository(db)
	paymentRepository := repository.NewPaymentRepository(db)
	paymentService := service.NewPaymentService(orderRepository, paymentRepository)

	producerBrokers := []string{config.NewProducerBroker()}
	commonPublisher, err := common.NewPublisher(producerBrokers, watermill.NewStdLogger(false, false))
	if err != nil {
		log.Fatal(err)
	}
	paymentProducer := producer.NewPaymentProducer(commonPublisher)
	paymentHandler := handler.NewPaymentHandler(paymentProducer, paymentService)
	paymentHandler.Route(e)

	subscriberConfig := &common.SubscriberConfig{
		Brokers:       []string{config.NewConsumerBroker()},
		ConsumerGroup: "orders-service",
		FromBeginning: true,
		LoggerAdapter: watermill.NewStdLogger(false, false),
	}
	subscriber, err := common.NewSubscriber(subscriberConfig)
	if err != nil {
		log.Fatal(err)
	}
	paymentConsumer := consumer.NewPaymentConsumer(orderRepository)
	commonConsumer := common.NewConsumer(subscriber)
	commonConsumer.On(common.OrderCreated, paymentConsumer.OrderCreated)
	commonConsumer.On(common.OrderCancelled, paymentConsumer.OrderCancelled)

	return e
}
