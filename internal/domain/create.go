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

	expiryDate, err := validateExpiryDate(request.ExpiryMonth, request.ExpiryYear, uuid)
	if err != nil {
		return nil, err
	}

	if !validateCurrencyISO(request.Currency) {
		return &models.PostPaymentResponse{
			Id:            uuid,
			PaymentStatus: "rejected",
		}, errors.New("invalid currency")
	}

	totalAmount, err := validateAmount(request.Amount)
	if err != nil {
		return &models.PostPaymentResponse{
			Id:            uuid,
			PaymentStatus: "rejected",
		}, errors.New("invalid amount")
	}

	cvv, err := validateCVV(request.Cvv)
	if err != nil {
		return &models.PostPaymentResponse{
			Id:            uuid,
			PaymentStatus: "rejected",
		}, errors.New("invalid cvv")
	}

	cardNumber := strconv.Itoa(request.CardNumber)

	PostPaymentBankRequest := &models.PostPaymentBankRequest{
		CardNumber: cardNumber,
		ExpiryDate: expiryDate,
		Currency:   request.Currency,
		Amount:     totalAmount,
		CVV:        cvv,
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

func validateCurrencyISO(currency string) bool {
	_, isValid := validCurrencyCodes[currency]
	return isValid
}

func validateAmount(amount int) (int, error) {
	if amount < 0 {
		return 0, errors.New("invalid amount")
	}
	return amount, nil
}

func validateCVV(cvv int) (string, error) {
	if cvv < 100 || cvv > 9999 {
		return "", errors.New("invalid cvv")
	}

	return strconv.Itoa(cvv), nil
}
