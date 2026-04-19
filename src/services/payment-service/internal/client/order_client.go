package client

import (
	"context"
	"fmt"
	"net/http"
	"github.com/vishalss1/CartGO/pkg/auth"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/google/uuid"
)

type OrderClient interface {
	ConfirmOrder(ctx context.Context, orderID uuid.UUID) error
}

type HttpOrderClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHttpOrderClient(baseURL string) *HttpOrderClient {
	return &HttpOrderClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *HttpOrderClient) ConfirmOrder(ctx context.Context, orderID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/orders/%s/confirm-after-payment", c.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set(auth.HeaderUserRole, "SERVICE_PAYMENT")
	util.TraceRequest(ctx, req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("order confirmation failed with status: %d", resp.StatusCode)
	}

	return nil
}
