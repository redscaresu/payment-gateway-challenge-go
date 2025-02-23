package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestPostPaymentHandler_Integration(t *testing.T) {
	// Start Mountebank container
	ctx, cli, containerID := startMountebankContainer(t)
	defer stopMountebankContainer(ctx, cli, containerID)

	api := api.New()

	go func() {
		api.Run(ctx, ":8090")
	}()

	// Create the payment request
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

	// Create a new HTTP request for testing
	req, err := http.NewRequest("POST", "http://localhost:8090/api/payments", bytes.NewBuffer(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	// Check the HTTP status code in the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Decode the response body
	var response models.PostPaymentResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

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
