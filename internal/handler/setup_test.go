package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-payments/internal/driver"
	"github.com/muktiarafi/ticketing-payments/internal/entity"
	"github.com/muktiarafi/ticketing-payments/internal/repository"
	"github.com/ory/dockertest/v3"
	"github.com/stripe/stripe-go/v72"
)

var (
	pool     *dockertest.Pool
	resource *dockertest.Resource
)

var router *echo.Echo
var orderRepository repository.OrderRepository

func TestMain(m *testing.M) {
	db := &driver.DB{
		SQL: newTestDatabase(),
	}

	router = echo.New()
	router.Use(middleware.Logger())

	val := validator.New()
	trans := common.NewDefaultTranslator(val)
	customValidator := &common.CustomValidator{val, trans}
	router.Validator = customValidator
	router.HTTPErrorHandler = common.CustomErrorHandler

	orderRepository = repository.NewOrderRepository(db)
	paymentRepository := repository.NewPaymentRepository(db)
	paymentServiceMock := &paymentServiceMock{orderRepository, paymentRepository}

	paymentProducer := &paymentProducerStub{}
	paymentHandler := NewPaymentHandler(paymentProducer, paymentServiceMock)
	paymentHandler.Route(router)

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func newTestDatabase() *sql.DB {
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err = pool.Run("postgres", "alpine", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=postgres"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	var db *sql.DB
	if err = pool.Retry(func() error {
		db, err = sql.Open(
			"pgx",
			fmt.Sprintf("host=localhost port=%s dbname=postgres user=postgres password=secret", resource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}

		migrationFilePath := filepath.Join("..", "..", "db", "migrations")
		return driver.Migration(migrationFilePath, db)
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return db
}

func assertResponseCode(t testing.TB, want, got int) {
	t.Helper()

	if got != want {
		t.Errorf("Expected status code %d, but got %d instead", want, got)
	}
}

func signIn(userPayload *common.UserPayload) *http.Cookie {

	token, _ := common.CreateToken(userPayload)

	cookie := http.Cookie{
		Name:    "session",
		Value:   token,
		Expires: time.Now().Add(10 * time.Minute),
		Path:    "/auth",
	}

	return &cookie
}

type paymentProducerStub struct{}

func (p *paymentProducerStub) Created(payment *entity.Payment) error {
	fmt.Println("payment producer stub producer payment created event")

	return nil
}

type paymentServiceMock struct {
	repository.OrderRepository
	repository.PaymentRepository
}

func (s *paymentServiceMock) Create(token string, userID, orderID int64) (*entity.Payment, error) {
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
	charge, err := func(params *stripe.ChargeParams) (*stripe.Charge, error) {
		fmt.Printf("currency: %s, amount %d, token: %s", *params.Currency, *params.Amount, *params.Source.Token)
		ch := stripe.Charge{
			ID: randStringBytes(8),
		}
		return &ch, nil
	}(&stripe.ChargeParams{
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

	return newPayment, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
