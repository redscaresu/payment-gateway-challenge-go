package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	payments := NewPaymentsHandler(ps)

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
	t.Run("postPayment", func(t *testing.T) {
		// Create a new HTTP request for testing with a non-existing payment ID
		req, err := http.NewRequest("POST", "/api/payments", nil)
		require.NoError(t, err)

		// Create a new HTTP request recorder for recording the response
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		// Check the HTTP status code in the response
		assert.Equal(t, http.StatusNotImplemented, w.Code)
	})
}
