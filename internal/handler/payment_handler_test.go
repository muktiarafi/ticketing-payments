package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-payments/internal/entity"
	"github.com/muktiarafi/ticketing-payments/internal/model"
)

func TestPaymentHandlerNew(t *testing.T) {
	user := &common.UserPayload{1, "bambank@gmail.com"}
	cookie := signIn(user)

	t.Run("create payment normally", func(t *testing.T) {
		order := entity.Order{
			ID:      1,
			Price:   12,
			Status:  "CREATED",
			Version: 1,
			UserID:  1,
		}
		newOrder, err := orderRepository.Insert(&order)
		if err != nil {
			t.Error(err)
		}

		paymentDTO := model.PaymentDTO{
			Token:   "wer3424erwe",
			OrderID: newOrder.ID,
		}
		paymentJSON, _ := json.Marshal(paymentDTO)
		request := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(paymentJSON))
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusCreated, response.Code)

		responseBody, _ := ioutil.ReadAll(response.Body)
		apiResponse := struct {
			Data *entity.Payment `json:"data"`
		}{}
		json.Unmarshal(responseBody, &apiResponse)

		if len(apiResponse.Data.StripeID) < 1 {
			t.Error("expecting to get stripe id with length of 8")
		}
	})

	t.Run("create payment on nonexistent order", func(t *testing.T) {
		paymentDTO := model.PaymentDTO{
			Token:   "wkewrewrwer",
			OrderID: 23132313,
		}
		paymentJSON, _ := json.Marshal(paymentDTO)
		request := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(paymentJSON))
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusNotFound, response.Code)
	})

	t.Run("make payment on not owned order", func(t *testing.T) {
		order := entity.Order{
			ID:      12,
			Price:   12,
			Status:  "CREATED",
			Version: 1,
			UserID:  12,
		}
		newOrder, err := orderRepository.Insert(&order)
		if err != nil {
			t.Error(err)
		}

		paymentDTO := model.PaymentDTO{
			Token:   "wkewrewrwer",
			OrderID: newOrder.ID,
		}
		paymentJSON, _ := json.Marshal(paymentDTO)
		request := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(paymentJSON))
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusConflict, response.Code)
	})

	t.Run("make payment on cancelled order", func(t *testing.T) {
		order := entity.Order{
			ID:      14,
			Price:   12,
			Status:  "CANCELLED",
			Version: 1,
			UserID:  1,
		}
		newOrder, err := orderRepository.Insert(&order)
		if err != nil {
			t.Error(err)
		}

		paymentDTO := model.PaymentDTO{
			Token:   "wqeqweyyyyyyy",
			OrderID: newOrder.ID,
		}
		paymentJSON, _ := json.Marshal(paymentDTO)
		request := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(paymentJSON))
		request.AddCookie(cookie)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusBadRequest, response.Code)
	})
}
