package domain

import (
	"errors"
	"strconv"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/gatewayerrors"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"

	"github.com/google/uuid"
)

/*

I create a Domain that we are able to inject payment services into, the reason why I did this was so that we could inject mocked objects into the payment service.  These are not as strong as the integration tests but they do at least allow us in the future to include more exhaustive testing against each of the validations, if I had more time I would create table tests against each validation so its not just.

Arguable YAGNI but I created a domain that can contain n services, for example if end up having to include a new service to integrate with in the future its just a simple case of adding an additional service to the Domain struct and creating the requisitive methods and interfaces.  It also allows us to split the implementation away from the interface of the services here.

In future anymore calls to the mountebank just need to be added to the payments service interface and we can just easily mock it.

*/

type Domain struct {
	PaymentService PaymentService
}

func NewDomain(paymentService PaymentService) *Domain {
	return &Domain{
		PaymentService: paymentService,
	}
}

type PaymentService interface {
	Create(request *models.PostPaymentHandlerRequest) (*models.PostPaymentResponse, error)
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

func (p *PaymentServiceImpl) Create(request *models.PostPaymentHandlerRequest) (*models.PostPaymentResponse, error) {

	uuid := uuid.New().String()
	err := validateCardNumber(strconv.Itoa(request.CardNumber), uuid)
	if err != nil {
		return nil, err
	}
	cardNumber := strconv.Itoa(request.CardNumber)

	expiryDate, err := validateExpiryDate(request.ExpiryMonth, request.ExpiryYear, uuid)
	if err != nil {
		return nil, err
	}

	err = validateCurrencyISO(request.Currency, uuid)
	if err != nil {
		return nil, err
	}

	err = validateAmount(request.Amount, uuid)
	if err != nil {
		return nil, err
	}

	err = validateCVV(request.Cvv, uuid)
	if err != nil {
		return nil, err
	}

	cvvString := strconv.Itoa(request.Cvv)

	PostPaymentBankRequest := &models.PostPaymentBankRequest{
		CardNumber: cardNumber,
		ExpiryDate: expiryDate,
		Currency:   request.Currency,
		Amount:     request.Amount,
		CVV:        cvvString,
	}

	bankResponse, err := p.client.PostBankPayment(PostPaymentBankRequest)
	if err != nil {
		return nil, err
	}

	cardNumberLastFour, err := strconv.Atoi(getLastFourCharacters(cardNumber))
	if err != nil {
		return nil, err
	}

	paymentStatus := "declined"
	if bankResponse.Authorised {
		paymentStatus = "authorized"
	}

	paymentResponse := &models.PostPaymentResponse{
		Id:                 uuid,
		PaymentStatus:      paymentStatus,
		CardNumberLastFour: cardNumberLastFour,
		ExpiryMonth:        request.ExpiryMonth,
		ExpiryYear:         request.ExpiryYear,
		Currency:           request.Currency,
		Amount:             request.Amount,
	}

	p.repo.AddPayment(*paymentResponse)

	return paymentResponse, nil
}

func getLastFourCharacters(s string) string {
	if len(s) < 4 {
		return s
	}
	return s[len(s)-4:]
}

func validateCardNumber(cardNumber, id string) error {
	if len(cardNumber) < 14 || len(cardNumber) > 19 {
		return gatewayerrors.NewValidationError(
			errors.New("incorrect card length"),
			id,
			"card_number",
		)
	}

	return nil
}

func validateExpiryDate(requestMonth, requestYear int, id string) (string, error) {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	if requestMonth < month || requestMonth > 12 || requestMonth < 1 {
		return "", gatewayerrors.NewValidationError(
			errors.New("invalid expiry month"),
			id,
			"expiry_month",
		)
	}

	if requestYear < year {
		return "", gatewayerrors.NewValidationError(
			errors.New("year in past"),
			id,
			"expiry_year",
		)
	}

	if requestMonth < month {
		return "", gatewayerrors.NewValidationError(
			errors.New("month in past"),
			id,
			"expiry_month",
		)
	}

	return strconv.Itoa(requestMonth) + "/" + strconv.Itoa(requestYear), nil
}

var validCurrencyCodes = map[string]bool{
	"USD": true,
	"EUR": true,
	"GBP": true,
}

func validateCurrencyISO(currency, id string) error {
	_, isValid := validCurrencyCodes[currency]
	if !isValid {
		return gatewayerrors.NewValidationError(
			errors.New("unsupported Currency"),
			id,
			"currency",
		)
	}
	return nil
}

func validateAmount(amount int, id string) error {
	if amount <= 0 {
		return gatewayerrors.NewValidationError(
			errors.New("invalid amount"),
			id,
			"amount",
		)
	}
	return nil
}

func validateCVV(cvv int, id string) error {
	if cvv < 100 || cvv > 9999 {
		return gatewayerrors.NewValidationError(
			errors.New("invalid cvv"),
			id,
			"cvv",
		)
	}

	return nil
}
