package domain

import (
	"strconv"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client"
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
	client client.Client
}

// NewDomain creates a new PostPaymentService.
func NewDomain(repo *repository.PaymentsRepository, client client.Client) *Domain {
	return &Domain{
		repo:   repo,
		client: client,
	}
}

// PostPayment processes a payment request and returns a payment response.
func (d *Domain) PostPayment(request *models.PostPaymentRequest) (*models.PostPaymentResponse, error) {

	expiryDate := strconv.Itoa(request.ExpiryMonth) + "/" + strconv.Itoa(request.ExpiryYear)
	cvv := strconv.Itoa(request.Cvv)

	//validate the curreny code

	PostPaymentBankRequest := &models.PostPaymentBankRequest{
		CardNumber: "2222405343248877",
		ExpiryDate: expiryDate,
		Currency:   request.Currency,
		Amount:     request.Amount,
		CVV:        cvv,
	}

	// make a call to the external payment API
	bankResponse, err := d.client.PostBankPayment(PostPaymentBankRequest)
	if err != nil {
		return nil, err
	}

	uuid := uuid.New().String()
	paymentResponse := &models.PostPaymentResponse{
		Id:                 uuid,
		PaymentStatus:      strconv.FormatBool(bankResponse.Authorised),
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
