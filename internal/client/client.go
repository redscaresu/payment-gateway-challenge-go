package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/gatewayerrors"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

/*

here I include an interface so that we can mock the client but thinking about it I am not sure that is necessary we could have used the fake bank but at least now it means we can test all possible responses from the bank including errors like a 500.  I could of made this more generic by having a method called "DO" and then we could have passed in the method and the URL that way in future as we add more endpoints we can just call it with a different path as needed.  As it stands a new method would need to be created for each endpoint.

*/

type Client interface {
	PostBankPayment(request *models.PostPaymentBankRequest) (*models.PostPaymentBankResponse, error)
}

type HTTPClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    baseURL,
	}
}

func (c *HTTPClient) PostBankPayment(request *models.PostPaymentBankRequest) (*models.PostPaymentBankResponse, error) {
	url := fmt.Sprintf("%s/payments", c.baseURL)
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Log the JSON payload
	log.Printf("Sending request to %s with payload: %s", url, string(body))

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to make POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusServiceUnavailable {
		return nil, gatewayerrors.NewBankError(
			errors.New("acquiring bank unavailble"),
			http.StatusServiceUnavailable,
		)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var response models.PostPaymentBankResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}
