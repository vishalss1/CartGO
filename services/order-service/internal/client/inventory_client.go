package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/vishalss1/CartGO/pkg/auth"
	"github.com/google/uuid"
)

type InventoryClient interface {
	Reserve(ctx context.Context, productID uuid.UUID, orderID uuid.UUID, quantity int) error
	Release(ctx context.Context, orderID uuid.UUID) error
	Commit(ctx context.Context, orderID uuid.UUID) error
}

type HttpInventoryClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHttpInventoryClient(baseURL string) *HttpInventoryClient {
	return &HttpInventoryClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type reserveRequest struct {
	OrderID  uuid.UUID `json:"order_id"`
	Quantity int       `json:"quantity"`
}

type idempotentRequest struct {
	OrderID uuid.UUID `json:"order_id"`
}

func (c *HttpInventoryClient) Reserve(ctx context.Context, productID uuid.UUID, orderID uuid.UUID, quantity int) error {
	url := fmt.Sprintf("%s/api/v1/inventory/%s/reserve", c.baseURL, productID)
	
	reqBody := reserveRequest{
		OrderID:  orderID,
		Quantity: quantity,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Role", "SERVICE_ORDER") // Internal service role

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("inventory reservation failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *HttpInventoryClient) Release(ctx context.Context, orderID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/inventory/release", c.baseURL)
	
	reqBody := idempotentRequest{OrderID: orderID}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserRole, "SERVICE_ORDER")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("inventory release failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (c *HttpInventoryClient) Commit(ctx context.Context, orderID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/inventory/commit", c.baseURL)
	
	reqBody := idempotentRequest{OrderID: orderID}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HeaderUserRole, "SERVICE_ORDER")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("inventory commit failed with status: %d", resp.StatusCode)
	}

	return nil
}
