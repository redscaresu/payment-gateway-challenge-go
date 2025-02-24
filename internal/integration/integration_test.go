package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/handlers"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestPostGetPaymentHandler_Integration(t *testing.T) {
	ctx := context.Background()
	api := api.New()

	go func() {
		api.Run(ctx, ":8090")
	}()

	postPayment := &models.PostPaymentHandlerRequest{
		CardNumber:  2222405343248877,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         123,
	}

	body, err := json.Marshal(postPayment)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "http://localhost:8090/api/payments", bytes.NewBuffer(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.PostPaymentResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	_, err = uuid.Parse(response.Id)
	assert.NilError(t, err)
	assert.Equal(t, "authorized", response.PaymentStatus)

	fourChar := getLastFourCharacters(t, postPayment.CardNumber)
	fourInt, err := strconv.Atoi(fourChar)
	require.NoError(t, err)
	assert.Equal(t, fourInt, response.CardNumberLastFour)
	assert.Equal(t, postPayment.ExpiryMonth, response.ExpiryMonth)
	assert.Equal(t, postPayment.ExpiryYear, response.ExpiryYear)
	assert.Equal(t, postPayment.Currency, response.Currency)
	assert.Equal(t, postPayment.Amount, response.Amount)

	reqGet, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8090/api/payments/%s", response.Id), bytes.NewBuffer(body))
	require.NoError(t, err)

	respGet, err := http.DefaultClient.Do(reqGet)
	require.NoError(t, err)

	var getHandlerResponse models.GetPaymentHandlerResponse
	err = json.NewDecoder(respGet.Body).Decode(&getHandlerResponse)
	require.NoError(t, err)

	assert.Equal(t, getHandlerResponse.Id, response.Id)
	assert.Equal(t, "authorized", response.PaymentStatus)
	assert.Equal(t, 8877, response.CardNumberLastFour)
	assert.Equal(t, 4, response.ExpiryMonth)
	assert.Equal(t, 2025, response.ExpiryYear)
	assert.Equal(t, "GBP", response.Currency)
	assert.Equal(t, 100, response.Amount)
}

func TestPostPaymentHandler_IntegrationCardNumberValidationError(t *testing.T) {
	ctx := context.Background()
	api := api.New()

	go func() {
		api.Run(ctx, ":8090")
	}()

	postPayment := &models.PostPaymentHandlerRequest{
		CardNumber:  1,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         123,
	}

	body, err := json.Marshal(postPayment)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "http://localhost:8090/api/payments", bytes.NewBuffer(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	var response models.PostPayment400Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	_, err = uuid.Parse(response.Id)
	assert.NilError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "rejected", response.PaymentStatus)
}

func TestPostPaymentHandler_IntegrationBankError(t *testing.T) {
	ctx := context.Background()
	api := api.New()

	go func() {
		api.Run(ctx, ":8090")
	}()

	postPayment := &models.PostPaymentHandlerRequest{
		CardNumber:  2222405343248870,
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         123,
	}

	body, err := json.Marshal(postPayment)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "http://localhost:8090/api/payments", bytes.NewBuffer(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	var response handlers.HandlerErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Equal(t, "The acquiring bank is currently unavailable. Please try again later.", response.Message)
}

func getLastFourCharacters(t *testing.T, i int) string {
	t.Helper()

	s := strconv.Itoa(i)
	require.Equal(t, 16, len(s))
	return s[len(s)-4:]
}
