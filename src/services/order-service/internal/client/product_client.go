package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/vishalss1/CartGO/pkg/auth"
	"github.com/vishalss1/CartGO/pkg/util"
	"github.com/vishalss1/CartGO/services/order-service/internal/middleware"
)

type Product struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Price    float64   `json:"price"`
}

type ProductClient interface {
	GetProduct(ctx context.Context, productID uuid.UUID) (*Product, error)
}

type HttpProductClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHttpProductClient(baseURL string) *HttpProductClient {
	return &HttpProductClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *HttpProductClient) GetProduct(ctx context.Context, productID uuid.UUID) (*Product, error) {
	url := fmt.Sprintf("%s/api/v1/products/%s", c.baseURL, productID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(auth.HeaderUserRole, "SERVICE_ORDER")
	userID := middleware.GetUserID(ctx)
	if userID != uuid.Nil {
		req.Header.Set(auth.HeaderUserID, userID.String())
	}
	util.TraceRequest(ctx, req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("item no longer exists")
		}
		return nil, fmt.Errorf("failed to fetch product: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool    `json:"success"`
		Data    Product `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}
