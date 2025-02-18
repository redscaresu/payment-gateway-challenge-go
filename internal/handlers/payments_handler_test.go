package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain/mocks"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetPaymentHandler(t *testing.T) {
	expectedPayment := models.PostPaymentResponse{
		Id:                 "test-id",
		PaymentStatus:      "test-successful-status",
		CardNumberLastFour: 1234,
		ExpiryMonth:        10,
		ExpiryYear:         2035,
		Currency:           "GBP",
		Amount:             100,
	}
	ps := repository.NewPaymentsRepository()
	ps.AddPayment(expectedPayment)

	domain := domain.NewDomain(ps)

	payments := NewPaymentsHandler(ps, domain)

	r := chi.NewRouter()
	r.Get("/api/payments/{id}", payments.GetHandler())
	r.Post("/api/payments", payments.PostHandler())

	httpServer := &http.Server{
		Addr:    ":8091",
		Handler: r,
	}

	go func() error {
		return httpServer.ListenAndServe()
	}()

	t.Run("PaymentFound", func(t *testing.T) {
		// Create a new HTTP request for testing
		req, err := http.NewRequest("GET", "/api/payments/test-id", nil)
		require.NoError(t, err)

		// Create a new HTTP request recorder for recording the response
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Check the body is not nil
		require.NotNil(t, w.Body)

		var response models.PostPaymentResponse
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Check the response body is what we expect
		assert.Equal(t, expectedPayment, response)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("PaymentNotFound", func(t *testing.T) {
		// Create a new HTTP request for testing with a non-existing payment ID
		req, err := http.NewRequest("GET", "/api/payments/NonExistingID", nil)
		require.NoError(t, err)

		// Create a new HTTP request recorder for recording the response
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Check the HTTP status code in the response
		assert.Equal(t, w.Code, http.StatusNotFound)
	})
}

func TestPostPaymentHandler(t *testing.T) {
	expectedPayment := models.PostPaymentResponse{
		Id:                 "test-id",
		PaymentStatus:      "test-successful-status",
		CardNumberLastFour: 1234,
		ExpiryMonth:        10,
		ExpiryYear:         2035,
		Currency:           "GBP",
		Amount:             100,
	}
	ps := repository.NewPaymentsRepository()
	ps.AddPayment(expectedPayment)
	ctrl := gomock.NewController(t)
	mockPostPaymentService := mocks.NewMockPostPaymentService(ctrl)
	defer ctrl.Finish()

	mockDomain := &domain.Domain{
		PostPaymentService: mockPostPaymentService,
	}

	payments := NewPaymentsHandler(ps, mockDomain)

	r := chi.NewRouter()
	r.Get("/api/payments/{id}", payments.GetHandler())
	r.Post("/api/payments", payments.PostHandler())

	httpServer := &http.Server{
		Addr:    ":8091",
		Handler: r,
	}

	go func() error {
		return httpServer.ListenAndServe()
	}()

	// Arrange
	postPayment := &models.PostPaymentRequest{
		CardNumberLastFour: 8877,
		ExpiryMonth:        4,
		ExpiryYear:         2025,
		Currency:           "GBP",
		Amount:             100,
		Cvv:                123,
	}

	body, err := json.Marshal(postPayment)
	require.NoError(t, err)

	mockDomain.PostPaymentService.(*mocks.MockPostPaymentService).EXPECT().PostPayment(postPayment).Return(&models.PostPaymentResponse{
		Id:                 "test-id",
		PaymentStatus:      "test-successful-status",
		CardNumberLastFour: 8877,
		ExpiryMonth:        4,
		ExpiryYear:         2025,
		Currency:           "GBP",
		Amount:             100,
	}, nil)

	// Act
	req, err := http.NewRequest("POST", "/api/payments", bytes.NewBuffer(body))
	require.NoError(t, err)

	// Create a new HTTP request recorder for recording the response
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check the HTTP status code in the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PostPaymentResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Assert

	assert.Equal(t, "test-id", response.Id)
	assert.Equal(t, postPayment.CardNumberLastFour, response.CardNumberLastFour)
	assert.Equal(t, postPayment.ExpiryMonth, response.ExpiryMonth)
	assert.Equal(t, postPayment.ExpiryYear, response.ExpiryYear)
	assert.Equal(t, postPayment.Currency, response.Currency)
	assert.Equal(t, postPayment.Amount, response.Amount)
	assert.Equal(t, "test-successful-status", response.PaymentStatus)

}

// t.Run("postPaymentNoBody", func(t *testing.T) {
// 	// Create a new HTTP request for testing with a non-existing payment ID
// 	req, err := http.NewRequest("POST", "/api/payments", nil)
// 	require.NoError(t, err)

// 	// Create a new HTTP request recorder for recording the response
// 	w := httptest.NewRecorder()

// 	r.ServeHTTP(w, req)

// 	// Check the HTTP status code in the response
// 	assert.Equal(t, http.StatusBadRequest, w.Code)
// })
