package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/handlers"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

/* integration tests use the real service rather than mocks, I limited these due to time constraints but also because they take longer to run than the unit tests so it could end up slowing are pipeline if we make these complete.
As it stands now we test the happy path, 1 validation (in this case card number) and test when the upstream fails.
*/

func TestPostPaymentHandler_Integration(t *testing.T) {
	ctx, cli, containerID := startMountebankContainer(t)
	defer stopMountebankContainer(ctx, cli, containerID)

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
	ctx, cli, containerID := startMountebankContainer(t)
	defer stopMountebankContainer(ctx, cli, containerID)

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
	ctx, cli, containerID := startMountebankContainer(t)
	defer stopMountebankContainer(ctx, cli, containerID)

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

func startMountebankContainer(t *testing.T) (context.Context, *client.Client, string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "bbyars/mountebank:2.8.1",
		ExposedPorts: map[nat.Port]struct{}{
			"8085/tcp": {},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{
				{
					HostPort: "8080",
				},
			},
		},
	}, nil, nil, "")
	require.NoError(t, err)

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	require.NoError(t, err)

	// Wait for Mountebank to be ready
	time.Sleep(5 * time.Second)

	return ctx, cli, resp.ID
}

func stopMountebankContainer(ctx context.Context, cli *client.Client, containerID string) {
	cli.ContainerStop(ctx, containerID, container.StopOptions{})
	cli.ContainerRemove(ctx, containerID, container.RemoveOptions{})
}

func getLastFourCharacters(t *testing.T, i int) string {
	t.Helper()

	s := strconv.Itoa(i)
	require.Equal(t, 16, len(s))
	return s[len(s)-4:]
}
