package domain

import (
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/google/uuid"
)

// PostPaymentService is a domain service that processes a payment request.
type PostPaymentService interface {
	// PostPayment processes a payment request and returns a payment response.
	PostPayment(request *models.PostPaymentRequest) (*models.PostPaymentResponse, error)
}

// Domain is a domain service that processes a payment request.
type Domain struct {
	repo *repository.PaymentsRepository
	PostPaymentService
}

// NewDomain creates a new PostPaymentService.
func NewDomain(repo *repository.PaymentsRepository) *Domain {
	return &Domain{
		repo: repo,
	}
}

// PostPayment processes a payment request and returns a payment response.
func (d *Domain) PostPayment(request *models.PostPaymentRequest) (*models.PostPaymentResponse, error) {

	uuid := uuid.New().String()
	paymentResponse := &models.PostPaymentResponse{
		Id:                 uuid,
		PaymentStatus:      "test-successful-status",
		CardNumberLastFour: request.CardNumberLastFour,
		ExpiryMonth:        request.ExpiryMonth,
		ExpiryYear:         request.ExpiryYear,
		Currency:           request.Currency,
		Amount:             request.Amount,
	}

	// save the payment in the repository
	d.repo.AddPayment(*paymentResponse)

	return paymentResponse, nil
}
