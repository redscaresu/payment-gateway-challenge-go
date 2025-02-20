package domain

import (
	"strconv"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/google/uuid"
)

type Domain struct {
	PaymentService PaymentService
}

func NewDomain(repo *repository.PaymentsRepository, client client.Client, postPaymentService PaymentService) *Domain {
	return &Domain{
		PaymentService: postPaymentService,
	}
}

type PaymentService interface {
	PostPayment(request *models.PostPaymentHandlerRequest) (*models.PostPaymentResponse, error)
}

type PaymentServiceImpl struct {
	repo               *repository.PaymentsRepository
	PostPaymentService PaymentService
	client             client.Client
}

func NewPaymentServiceImpl(repo *repository.PaymentsRepository, client client.Client) *PaymentServiceImpl {
	return &PaymentServiceImpl{
		repo:   repo,
		client: client,
	}
}

func (p *PaymentServiceImpl) PostPayment(request *models.PostPaymentHandlerRequest) (*models.PostPaymentResponse, error) {

	expiryDate := strconv.Itoa(request.ExpiryMonth) + "/" + strconv.Itoa(request.ExpiryYear)
	cvv := strconv.Itoa(request.Cvv)
	cardNumber := strconv.Itoa(request.CardNumber)

	//validate the curreny code

	PostPaymentBankRequest := &models.PostPaymentBankRequest{
		CardNumber: cardNumber,
		ExpiryDate: expiryDate,
		Currency:   request.Currency,
		Amount:     request.Amount,
		CVV:        cvv,
	}

	// make a call to the external payment API
	bankResponse, err := p.client.PostBankPayment(PostPaymentBankRequest)
	if err != nil {
		return nil, err
	}

	cardNumberLastFour, err := strconv.Atoi(getLastFourCharacters(cardNumber))
	if err != nil {
		return nil, err
	}

	uuid := uuid.New().String()
	paymentResponse := &models.PostPaymentResponse{
		Id:                 uuid,
		PaymentStatus:      strconv.FormatBool(bankResponse.Authorised),
		CardNumberLastFour: cardNumberLastFour,
		ExpiryMonth:        request.ExpiryMonth,
		ExpiryYear:         request.ExpiryYear,
		Currency:           request.Currency,
		Amount:             request.Amount,
	}

	// save the payment in the repository
	p.repo.AddPayment(*paymentResponse)

	return paymentResponse, nil
}

func getLastFourCharacters(s string) string {
	if len(s) < 4 {
		return s
	}
	return s[len(s)-4:]
}
