package domain

import (
	"strconv"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/google/uuid"
)

type Domain struct {
	repo               *repository.PaymentsRepository
	client             client.Client
	PostPaymentService PostPaymentService
}

func NewDomain(repo *repository.PaymentsRepository, client client.Client, postPaymentService PostPaymentService) *Domain {
	return &Domain{
		repo:               repo,
		client:             client,
		PostPaymentService: postPaymentService,
	}
}

type PostPaymentService interface {
	PostPayment(request *models.PostPaymentRequest) (*models.PostPaymentResponse, error)
}

type PostPaymentServiceImpl struct {
	repo               *repository.PaymentsRepository
	PostPaymentService PostPaymentService
	client             client.Client
}

func NewPostPaymentServiceImpl(repo *repository.PaymentsRepository, client client.Client) *PostPaymentServiceImpl {
	return &PostPaymentServiceImpl{
		repo:   repo,
		client: client,
	}
}

func (d *PostPaymentServiceImpl) PostPayment(request *models.PostPaymentRequest) (*models.PostPaymentResponse, error) {

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
