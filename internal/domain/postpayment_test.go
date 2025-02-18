package domain_test

import (
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostPayment(t *testing.T) {
	postPayment := models.PostPaymentRequest{
		CardNumberLastFour: 8877,
		ExpiryMonth:        4,
		ExpiryYear:         2025,
		Currency:           "GBP",
		Amount:             100,
		Cvv:                123,
	}

	repo := repository.NewPaymentsRepository()
	domain := domain.NewDomain(repo)

	response, err := domain.PostPayment(&postPayment)
	require.NoError(t, err)

	_, err = uuid.Parse(response.Id)
	require.NoError(t, err)

	assert.Equal(t, "test-successful-status", response.PaymentStatus)
	assert.Equal(t, postPayment.CardNumberLastFour, response.CardNumberLastFour)
	assert.Equal(t, postPayment.ExpiryMonth, response.ExpiryMonth)
	assert.Equal(t, postPayment.ExpiryYear, response.ExpiryYear)
	assert.Equal(t, postPayment.Currency, response.Currency)
	assert.Equal(t, postPayment.Amount, response.Amount)

	// Check if the payment was saved in the repository
	dbPayment := repo.GetPayment(response.Id)
	assert.Equal(t, response.Id, dbPayment.Id)
}
