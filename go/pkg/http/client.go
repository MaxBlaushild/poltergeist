package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	ApplicationJson = "application/json"
)

type client struct {
	baseUrl     string
	contentType string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Client interface {
	Get(ctx context.Context, uri string) ([]byte, error)
	Post(ctx context.Context, uri string, body interface{}) ([]byte, error)
}

func NewClient(baseUrl string, contentType string) Client {
	return &client{
		baseUrl:     baseUrl,
		contentType: contentType,
	}
}

func (c *client) Get(ctx context.Context, uri string) ([]byte, error) {
	resp, err := http.Get(c.baseUrl + uri)
	if err != nil {
		return nil, fmt.Errorf(
			"error calling %s%s: %s",
			c.baseUrl,
			uri,
			err.Error(),
		)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading body from request to %s%s: %s",
			c.baseUrl,
			uri,
			err.Error(),
		)
	}

	// handle error status codes
	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err = json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf(
				"error unwrapping error response from %s%s: %s",
				c.baseUrl,
				uri,
				err.Error(),
			)
		}

		return nil, fmt.Errorf(
			"error returned from %s%s: %s",
			c.baseUrl,
			uri,
			errorResp.Error,
		)
	}

	return body, nil
}

func (c *client) Post(ctx context.Context, uri string, body interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.baseUrl+uri, c.contentType, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf(
			"error calling %s%s: %s",
			c.baseUrl,
			uri,
			err.Error(),
		)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading body from request to %s%s: %s",
			c.baseUrl,
			uri,
			err.Error(),
		)
	}

	// handle error status codes
	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err = json.Unmarshal(respBody, &errorResp); err != nil {
			return nil, fmt.Errorf(
				"error unwrapping error response from %s%s: %s",
				c.baseUrl,
				uri,
				err.Error(),
			)
		}

		return nil, fmt.Errorf(
			"error returned from %s%s: %s",
			c.baseUrl,
			uri,
			errorResp.Error,
		)
	}

	return respBody, nil
}
