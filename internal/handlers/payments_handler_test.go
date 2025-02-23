package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain/mocks"
	genericerrors "github.com/cko-recruitment/payment-gateway-challenge-go/internal/errors"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetPaymentHandler(t *testing.T) {
	savedPayment := models.PostPaymentResponse{
		Id:                 "test-id",
		PaymentStatus:      "test-successful-status",
		CardNumberLastFour: 1234,
		ExpiryMonth:        10,
		ExpiryYear:         2035,
		Currency:           "GBP",
		Amount:             100,
	}
	ps := repository.NewPaymentsRepository()
	ps.AddPayment(savedPayment)

	expectedPayment := models.GetPaymentHandlerResponse{
		Id:                 "test-id",
		Status:             "test-successful-status",
		LastFourCardDigits: 1234,
		ExpiryMonth:        10,
		ExpiryYear:         2035,
		Currency:           "GBP",
		Amount:             100,
	}

	payments := NewPaymentsHandler(ps, nil)

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

		var response models.GetPaymentHandlerResponse
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
	t.Run("MissingID", func(t *testing.T) {
		// Create a new HTTP request for testing with a non-existing payment ID
		req, err := http.NewRequest("GET", "/api/payments/", nil)
		require.NoError(t, err)

		// Create a new HTTP request recorder for recording the response
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Check the HTTP status code in the response
		assert.Equal(t, http.StatusNotFound, w.Code)
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
	mockPaymentService := mocks.NewMockPaymentService(ctrl)
	defer ctrl.Finish()

	mockDomain := &domain.Domain{
		PaymentService: mockPaymentService,
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
	postPayment := &models.PostPaymentHandlerRequest{
		CardNumber:  2222405343248877,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         123,
	}

	body, err := json.Marshal(postPayment)
	require.NoError(t, err)

	postPaymentResponseID := uuid.New().String()
	mockDomain.PaymentService.(*mocks.MockPaymentService).EXPECT().PostPayment(postPayment).Return(&models.PostPaymentResponse{
		Id:                 postPaymentResponseID,
		PaymentStatus:      "authorized",
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

	lastFourCharacters, err := strconv.Atoi(getLastFourCharacters(t, postPayment.CardNumber))
	require.NoError(t, err)

	// Assert
	assert.Equal(t, postPaymentResponseID, response.Id)
	assert.Equal(t, lastFourCharacters, response.CardNumberLastFour)
	assert.Equal(t, postPayment.ExpiryMonth, response.ExpiryMonth)
	assert.Equal(t, postPayment.ExpiryYear, response.ExpiryYear)
	assert.Equal(t, postPayment.Currency, response.Currency)
	assert.Equal(t, postPayment.Amount, response.Amount)
	assert.Equal(t, "authorized", response.PaymentStatus)

}

func TestPostPaymentHandler_NoBody(t *testing.T) {

	payments := NewPaymentsHandler(nil, nil)

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

	// Create a new HTTP request for testing with a non-existing payment ID
	req, err := http.NewRequest("POST", "/api/payments", nil)
	require.NoError(t, err)

	// Create a new HTTP request recorder for recording the response
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check the HTTP status code in the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPostPaymentHandler_InvalidJson(t *testing.T) {

	payments := NewPaymentsHandler(nil, nil)

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

	// Create a new HTTP request for testing with a non-existing payment ID
	req, err := http.NewRequest("POST", "/api/payments", bytes.NewBuffer([]byte("invalid json")))
	require.NoError(t, err)

	// Create a new HTTP request recorder for recording the response
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check the HTTP status code in the response
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBankError_DomainError(t *testing.T) {

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
	mockPaymentService := mocks.NewMockPaymentService(ctrl)
	defer ctrl.Finish()

	mockDomain := &domain.Domain{
		PaymentService: mockPaymentService,
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
	postPayment := &models.PostPaymentHandlerRequest{
		CardNumber:  2222405343248877,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         123,
	}

	body, err := json.Marshal(postPayment)
	require.NoError(t, err)

	mockDomain.PaymentService.(*mocks.MockPaymentService).EXPECT().PostPayment(postPayment).Return(nil, errors.New("boom"))

	// Act
	req, err := http.NewRequest("POST", "/api/payments", bytes.NewBuffer(body))
	require.NoError(t, err)

	// Create a new HTTP request recorder for recording the response
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestBankError_ServiceUnavailable(t *testing.T) {

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
	mockPaymentService := mocks.NewMockPaymentService(ctrl)
	defer ctrl.Finish()

	mockDomain := &domain.Domain{
		PaymentService: mockPaymentService,
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
	postPayment := &models.PostPaymentHandlerRequest{
		CardNumber:  2222405343248877,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         123,
	}

	body, err := json.Marshal(postPayment)
	require.NoError(t, err)

	mockedError := genericerrors.NewBankError(
		errors.New("acquiring bank unavailble"),
		http.StatusServiceUnavailable,
	)
	mockDomain.PaymentService.(*mocks.MockPaymentService).EXPECT().PostPayment(postPayment).Return(nil, mockedError)

	// Act
	req, err := http.NewRequest("POST", "/api/payments", bytes.NewBuffer(body))
	require.NoError(t, err)

	// Create a new HTTP request recorder for recording the response
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	var response HandlerErrorResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, "The acquiring bank is currently unavailable. Please try again later.", response.Message)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func getLastFourCharacters(t *testing.T, i int) string {
	t.Helper()

	s := strconv.Itoa(i)
	require.Equal(t, 16, len(s))
	return s[len(s)-4:]
}
