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

type PaymentClient interface {
	ProcessPayment(ctx context.Context, orderID uuid.UUID, amount float64) error
}

type HttpPaymentClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHttpPaymentClient(baseURL string) *HttpPaymentClient {
	return &HttpPaymentClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type paymentProcessRequest struct {
	OrderID uuid.UUID `json:"order_id"`
	Amount  float64   `json:"amount"`
}

func (c *HttpPaymentClient) ProcessPayment(ctx context.Context, orderID uuid.UUID, amount float64) error {
	url := fmt.Sprintf("%s/api/v1/payments", c.baseURL)
	
	reqBody := paymentProcessRequest{
		OrderID: orderID,
		Amount:  amount,
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("payment processing failed with status: %d", resp.StatusCode)
	}

	return nil
}
