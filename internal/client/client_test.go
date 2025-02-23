package client_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/gatewayerrors"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
 At the moment I am using the testserver but I could of used the fake bank to test the client instead.  But I still think this is a good way to test the client as it is a simple client and we can test all the possible responses from the bank but it would have been more realistic to use the fake bank.

 This is not exhaustive either, if i had more time I would test all the possible responses from the bank.
*/

func TestHTTPClient_PostBankPayment(t *testing.T) {
	// Create a test server that returns a 200 OK response with a valid JSON body
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&models.PostPaymentBankResponse{
			Authorised:        true,
			AuthorizationCode: "123456",
		})
	}))
	defer testServer.Close()

	// Create the HTTP client with the test server's URL
	httpClient := client.NewClient(testServer.URL, 5*time.Second)

	postPayment := models.PostPaymentBankRequest{
		CardNumber: "2222405343248877",
		ExpiryDate: "4/2025",
		Currency:   "GBP",
		Amount:     100,
		CVV:        "123",
	}

	resp, err := httpClient.PostBankPayment(&postPayment)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.True(t, resp.Authorised)
	assert.NotEmpty(t, resp.AuthorizationCode)

}

func TestHTTPClient_PostBankPayment_ServiceUnavailable(t *testing.T) {
	// Create a test server that returns a 503 Service Unavailable response
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer testServer.Close()

	// Create the HTTP client with the test server's URL
	httpClient := client.NewClient(testServer.URL, 5*time.Second)

	postPayment := models.PostPaymentBankRequest{
		CardNumber: "2222405343248877",
		ExpiryDate: "4/2025",
		Currency:   "GBP",
		Amount:     100,
		CVV:        "123",
	}

	// Make the request using the HTTP client
	resp, err := httpClient.PostBankPayment(&postPayment)
	require.Error(t, err)
	require.Nil(t, resp)

	var bankErr *gatewayerrors.BankError
	errors.As(err, &bankErr)

	assert.Equal(t, http.StatusServiceUnavailable, bankErr.StatusCode)
}
