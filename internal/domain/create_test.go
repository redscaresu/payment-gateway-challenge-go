package domain_test

import (
	"strconv"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client/mocks"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/gatewayerrors"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPostPayment_Authorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	postPayment := models.PostPaymentHandlerRequest{
		CardNumber:  2222405343248877,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         123,
	}

	mockClient.EXPECT().PostBankPayment((&models.PostPaymentBankRequest{
		CardNumber: "2222405343248877",
		ExpiryDate: "4/2025",
		Currency:   "GBP",
		Amount:     100,
		CVV:        "123",
	})).Return((&models.PostPaymentBankResponse{
		Authorised:        true,
		AuthorizationCode: "abb53d1a-42dd-4ecc-9a25-dca064d35eb2",
	}), nil)

	repo := repository.NewPaymentsRepository()
	domain := domain.NewPaymentServiceImpl(repo, mockClient)

	response, err := domain.Create(&postPayment)
	require.NoError(t, err)

	_, err = uuid.Parse(response.Id)
	require.NoError(t, err)

	lastFourCharacters, err := strconv.Atoi(getLastFourCharacters(t, postPayment.CardNumber))
	require.NoError(t, err)

	assert.Equal(t, "authorized", response.PaymentStatus)
	assert.Equal(t, lastFourCharacters, response.CardNumberLastFour)
	assert.Equal(t, postPayment.ExpiryMonth, response.ExpiryMonth)
	assert.Equal(t, postPayment.ExpiryYear, response.ExpiryYear)
	assert.Equal(t, postPayment.Currency, response.Currency)
	assert.Equal(t, postPayment.Amount, response.Amount)

	// Check if the payment was saved in the repository
	dbPayment := repo.GetPayment(response.Id)
	assert.Equal(t, response.Id, dbPayment.Id)
}

func TestPostPayment_InvalidCardNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	postPayment := models.PostPaymentHandlerRequest{
		CardNumber:  123,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         123,
	}

	repo := repository.NewPaymentsRepository()
	domain := domain.NewPaymentServiceImpl(repo, mockClient)

	var validationError *gatewayerrors.ValidationError
	response, err := domain.Create(&postPayment)
	require.Nil(t, response)
	require.Error(t, err)
	require.ErrorAs(t, err, &validationError)

	_, err = uuid.Parse(validationError.ID)
	assert.NoError(t, err)
	assert.Equal(t, "incorrect card length", validationError.Error())
	assert.Equal(t, "card_number", validationError.GetFieldError())
}

func TestPostPayment_InvalidExpiryDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	postPayment := models.PostPaymentHandlerRequest{
		CardNumber:  2222405343248877,
		ExpiryMonth: 4,
		ExpiryYear:  1900,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         123,
	}

	repo := repository.NewPaymentsRepository()
	domain := domain.NewPaymentServiceImpl(repo, mockClient)

	var validationError *gatewayerrors.ValidationError
	response, err := domain.Create(&postPayment)
	require.Nil(t, response)
	require.Error(t, err)
	require.ErrorAs(t, err, &validationError)

	_, err = uuid.Parse(validationError.ID)
	assert.NoError(t, err)
	assert.Equal(t, "year in past", validationError.Error())
	assert.Equal(t, "expiry_year", validationError.GetFieldError())
}

func TestPostPayment_InvalidCurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	postPayment := models.PostPaymentHandlerRequest{
		CardNumber:  2222405343248877,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "invalid_currency",
		Amount:      100,
		Cvv:         123,
	}

	repo := repository.NewPaymentsRepository()
	domain := domain.NewPaymentServiceImpl(repo, mockClient)

	var validationError *gatewayerrors.ValidationError
	response, err := domain.Create(&postPayment)
	require.Nil(t, response)
	require.Error(t, err)
	require.ErrorAs(t, err, &validationError)

	_, err = uuid.Parse(validationError.ID)
	assert.NoError(t, err)
	assert.Equal(t, "unsupported Currency", validationError.Error())
	assert.Equal(t, "currency", validationError.GetFieldError())
}

func TestPostPayment_InvalidCVV(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	postPayment := models.PostPaymentHandlerRequest{
		CardNumber:  2222405343248877,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         0,
	}

	repo := repository.NewPaymentsRepository()
	domain := domain.NewPaymentServiceImpl(repo, mockClient)

	response, err := domain.Create(&postPayment)
	require.Error(t, err)

	_, err = uuid.Parse(response.Id)
	require.NoError(t, err)

	_, err = uuid.Parse(response.Id)
	require.NoError(t, err)

	assert.Equal(t, "rejected", response.PaymentStatus)
	assert.Equal(t, 0, response.CardNumberLastFour)
	assert.Equal(t, 0, response.ExpiryMonth)
	assert.Equal(t, 0, response.ExpiryYear)
	assert.Equal(t, "", response.Currency)
	assert.Equal(t, 0, response.Amount)
}

func TestPostPayment_NotAuthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockClient(ctrl)

	postPayment := models.PostPaymentHandlerRequest{
		CardNumber:  2222405343248877,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         123,
	}

	mockClient.EXPECT().PostBankPayment((&models.PostPaymentBankRequest{
		CardNumber: "2222405343248877",
		ExpiryDate: "4/2025",
		Currency:   "GBP",
		Amount:     100,
		CVV:        "123",
	})).Return((&models.PostPaymentBankResponse{
		// test that we can handle a declined payment
		Authorised:        false,
		AuthorizationCode: "abb53d1a-42dd-4ecc-9a25-dca064d35eb2",
	}), nil)

	repo := repository.NewPaymentsRepository()
	domain := domain.NewPaymentServiceImpl(repo, mockClient)

	response, err := domain.Create(&postPayment)
	require.NoError(t, err)

	_, err = uuid.Parse(response.Id)
	require.NoError(t, err)

	lastFourCharacters, err := strconv.Atoi(getLastFourCharacters(t, postPayment.CardNumber))
	require.NoError(t, err)

	assert.Equal(t, "declined", response.PaymentStatus)
	assert.Equal(t, lastFourCharacters, response.CardNumberLastFour)
	assert.Equal(t, postPayment.ExpiryMonth, response.ExpiryMonth)
	assert.Equal(t, postPayment.ExpiryYear, response.ExpiryYear)
	assert.Equal(t, postPayment.Currency, response.Currency)
	assert.Equal(t, postPayment.Amount, response.Amount)

	// Check if the payment was saved in the repository
	dbPayment := repo.GetPayment(response.Id)
	assert.Equal(t, response.Id, dbPayment.Id)
}

func getLastFourCharacters(t *testing.T, i int) string {
	t.Helper()

	s := strconv.Itoa(i)
	require.Equal(t, 16, len(s))
	return s[len(s)-4:]
}
