package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/auth"
	"github.com/vishalss1/CartGO/pkg/util"
)

type CreateDeliveryRequest struct {
	OrderID         uuid.UUID `json:"order_id"`
	DeliveryAddress string    `json:"delivery_address"`
}

type DeliveryClient interface {
	CreateDelivery(ctx context.Context, orderID uuid.UUID, address string) error
}

type HttpDeliveryClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHttpDeliveryClient(baseURL string) *HttpDeliveryClient {
	return &HttpDeliveryClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *HttpDeliveryClient) CreateDelivery(ctx context.Context, orderID uuid.UUID, address string) error {
	url := fmt.Sprintf("%s/api/v1/deliveries", c.baseURL)

	reqBody := CreateDeliveryRequest{
		OrderID:         orderID,
		DeliveryAddress: address,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserRole, "SERVICE_ORDER")
	util.TraceRequest(ctx, req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create delivery: status %d", resp.StatusCode)
	}

	return nil
}
