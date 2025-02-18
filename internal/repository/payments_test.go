package repository_test

import (
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestGetPayment(t *testing.T) {

	// arrange
	expectedPayment := models.PostPaymentResponse{
		Id:                 "test-id",
		PaymentStatus:      "test-successful-status",
		CardNumberLastFour: 1234,
		ExpiryMonth:        10,
		ExpiryYear:         2035,
		Currency:           "GBP",
		Amount:             100,
	}

	repository := repository.NewPaymentsRepository()
	repository.AddPayment(expectedPayment)

	// act
	payment := repository.GetPayment(expectedPayment.Id)

	// assert
	assert.Equal(t, expectedPayment, *payment)
}

func TestAddPayment(t *testing.T) {

	// arrange
	expectedPayment := models.PostPaymentResponse{
		Id:                 "test-id",
		PaymentStatus:      "test-successful-status",
		CardNumberLastFour: 1234,
		ExpiryMonth:        10,
		ExpiryYear:         2035,
		Currency:           "GBP",
		Amount:             100,
	}

	repository := repository.NewPaymentsRepository()

	// act
	repository.AddPayment(expectedPayment)

	// assert
	assert.Equal(t, &expectedPayment, repository.GetPayment(expectedPayment.Id))
}
