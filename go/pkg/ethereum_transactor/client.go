package ethereum_transactor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/http"
)

type Client interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error)
}

type client struct {
	httpClient http.Client
}

func NewClient(baseURL string) Client {
	return &client{
		httpClient: http.NewClient(baseURL, http.ApplicationJson),
	}
}

type CreateTransactionRequest struct {
	To       *string `json:"to,omitempty"`
	Value    string  `json:"value"`
	Data     *string `json:"data,omitempty"`
	GasLimit *uint64 `json:"gasLimit,omitempty"`
	GasPrice *string `json:"gasPrice,omitempty"`
	Type     *string `json:"type,omitempty"`
}

type CreateTransactionResponse struct {
	ID     string `json:"id"`
	TxHash string `json:"txHash"`
	Status string `json:"status"`
	Nonce  uint64 `json:"nonce"`
}

func (c *client) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	body, err := c.httpClient.Post(ctx, "/ethereum-transactor/transactions", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	var resp CreateTransactionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}
