package fulfillment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
)

// SlantAdapter (R-7.3) implements Adapter against Slant 3D's real, public V2
// API (base https://slant3dapi.com/v2/api — OpenAPI spec fetched live from
// https://slant3dapi.com/v2/api/openapi.json, since Slant doesn't publish a
// static reference). It is deliberately NOT wired up as the checkout-time
// default (server.fulfillmentAdapter still only resolves "manual" —
// REEF_FULFILLMENT_PROVIDER is unchanged): R-7.3's own gate ("do not
// integrate until sample parts have been ordered and inspected") hasn't
// been cleared. This adapter exists so an operator can send an individual
// paid order to Slant by hand from the print queue, not so every order
// silently starts routing there.
type SlantAdapter struct {
	AwsClient  aws.AWSClient
	HTTPClient *http.Client
	BaseURL    string
	APIKey     string
	PlatformID string
	Bucket     string
}

const defaultSlantBaseURL = "https://slant3dapi.com/v2/api"

func NewSlantAdapter(awsClient aws.AWSClient, apiKey, platformID, bucket string) *SlantAdapter {
	return &SlantAdapter{
		AwsClient:  awsClient,
		HTTPClient: http.DefaultClient,
		BaseURL:    defaultSlantBaseURL,
		APIKey:     apiKey,
		PlatformID: platformID,
		Bucket:     bucket,
	}
}

type slantAPIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (s *SlantAdapter) request(ctx context.Context, method, path string, body interface{}) (json.RawMessage, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("slant: marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, s.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("slant: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("slant: request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("slant: read response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("slant: %s %s returned %d: %s", method, path, resp.StatusCode, string(respBytes))
	}

	var parsed slantAPIResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return nil, fmt.Errorf("slant: decode response: %w", err)
	}
	if !parsed.Success {
		return nil, fmt.Errorf("slant: %s %s reported failure: %s", method, path, parsed.Message)
	}
	return parsed.Data, nil
}

type slantFilePlaceholder = json.RawMessage

type slantDirectUploadResponse struct {
	PresignedUrl    string              `json:"presignedUrl"`
	Key             string              `json:"key"`
	FilePlaceholder slantFilePlaceholder `json:"filePlaceholder"`
}

type slantFile struct {
	PublicFileServiceID string `json:"publicFileServiceId"`
}

// uploadFile mirrors Slant's documented 3-step upload: request a presigned
// S3 URL (POST /files/direct-upload), PUT the actual bytes there, then
// confirm (POST /files/confirm-upload) so Slant slices/analyzes it. The STL
// itself is pulled from our own S3 bucket via a presigned GET rather than a
// new AWSClient method, since GeneratePresignedURL already exists for
// exactly this.
func (s *SlantAdapter) uploadFile(ctx context.Context, stlKey string) (string, error) {
	downloadURL, err := s.AwsClient.GeneratePresignedURL(s.Bucket, stlKey, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("slant: presign our own stl for download: %w", err)
	}

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("slant: build stl download request: %w", err)
	}
	getResp, err := s.HTTPClient.Do(getReq)
	if err != nil {
		return "", fmt.Errorf("slant: download our own stl: %w", err)
	}
	defer getResp.Body.Close()
	stlBytes, err := io.ReadAll(getResp.Body)
	if err != nil {
		return "", fmt.Errorf("slant: read our own stl: %w", err)
	}
	if getResp.StatusCode >= 300 {
		return "", fmt.Errorf("slant: downloading our own stl returned %d", getResp.StatusCode)
	}

	uploadData, err := s.request(ctx, http.MethodPost, "/files/direct-upload", map[string]string{
		"name":       stlKey,
		"platformId": s.PlatformID,
	})
	if err != nil {
		return "", fmt.Errorf("slant: request direct-upload url: %w", err)
	}
	var uploadResp slantDirectUploadResponse
	if err := json.Unmarshal(uploadData, &uploadResp); err != nil {
		return "", fmt.Errorf("slant: decode direct-upload response: %w", err)
	}

	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadResp.PresignedUrl, bytes.NewReader(stlBytes))
	if err != nil {
		return "", fmt.Errorf("slant: build presigned put request: %w", err)
	}
	putResp, err := s.HTTPClient.Do(putReq)
	if err != nil {
		return "", fmt.Errorf("slant: upload stl to presigned url: %w", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode >= 300 {
		putBody, _ := io.ReadAll(putResp.Body)
		return "", fmt.Errorf("slant: presigned put returned %d: %s", putResp.StatusCode, string(putBody))
	}

	confirmData, err := s.request(ctx, http.MethodPost, "/files/confirm-upload", map[string]json.RawMessage{
		"filePlaceholder": uploadResp.FilePlaceholder,
	})
	if err != nil {
		return "", fmt.Errorf("slant: confirm upload: %w", err)
	}
	var file slantFile
	if err := json.Unmarshal(confirmData, &file); err != nil {
		return "", fmt.Errorf("slant: decode confirmed file: %w", err)
	}
	if file.PublicFileServiceID == "" {
		return "", fmt.Errorf("slant: confirm-upload returned no publicFileServiceId")
	}
	return file.PublicFileServiceID, nil
}

type slantAddress struct {
	Name    string `json:"name"`
	Line1   string `json:"line1"`
	Line2   string `json:"line2,omitempty"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

type slantOrderItem struct {
	Type                string `json:"type"`
	Quantity            int    `json:"quantity"`
	PublicFileServiceID string `json:"publicFileServiceId,omitempty"`
}

type slantDraftOrderRequest struct {
	PlatformID string `json:"platformId"`
	OwnerID    string `json:"ownerId"`
	Customer   struct {
		Details struct {
			Email   string       `json:"email"`
			Address slantAddress `json:"address"`
		} `json:"details"`
	} `json:"customer"`
	Items []slantOrderItem `json:"items"`
}

type slantOrder struct {
	PublicID string `json:"publicId"`
	Status   string `json:"status"`
}

type slantOrderEnvelope struct {
	Order slantOrder `json:"order"`
}

// SubmitOrder drafts then immediately processes a Slant order for every
// PRINT-able item (one with an STLKey — non-configurable/fixed-SKU items
// have no generated file to send and are skipped). Country defaults to
// "US" to match Address's own documented default when we don't have a
// country on file.
func (s *SlantAdapter) SubmitOrder(ctx context.Context, o Order) (string, error) {
	items := make([]slantOrderItem, 0, len(o.Items))
	for _, item := range o.Items {
		if item.STLKey == "" {
			continue
		}
		fileID, err := s.uploadFile(ctx, item.STLKey)
		if err != nil {
			return "", fmt.Errorf("slant: upload %s: %w", item.STLKey, err)
		}
		items = append(items, slantOrderItem{Type: "PRINT", Quantity: item.Quantity, PublicFileServiceID: fileID})
	}
	if len(items) == 0 {
		return "", fmt.Errorf("slant: order %s has no printable (STL) items", o.OrderToken)
	}

	country := o.ShippingCountry
	if country == "" {
		country = "US"
	}

	draft := slantDraftOrderRequest{PlatformID: s.PlatformID, OwnerID: o.OrderToken, Items: items}
	draft.Customer.Details.Email = o.CustomerEmail
	draft.Customer.Details.Address = slantAddress{
		Name:    o.ShippingName,
		Line1:   o.ShippingLine1,
		Line2:   o.ShippingLine2,
		City:    o.ShippingCity,
		State:   o.ShippingState,
		Zip:     o.ShippingZip,
		Country: country,
	}

	draftData, err := s.request(ctx, http.MethodPost, "/orders", draft)
	if err != nil {
		return "", fmt.Errorf("slant: draft order: %w", err)
	}
	var draftEnvelope slantOrderEnvelope
	if err := json.Unmarshal(draftData, &draftEnvelope); err != nil {
		return "", fmt.Errorf("slant: decode draft order: %w", err)
	}
	if draftEnvelope.Order.PublicID == "" {
		return "", fmt.Errorf("slant: draft order returned no publicId")
	}

	if _, err := s.request(ctx, http.MethodPost, "/orders/"+draftEnvelope.Order.PublicID, nil); err != nil {
		return "", fmt.Errorf("slant: process order %s (drafted, payment/production not started): %w", draftEnvelope.Order.PublicID, err)
	}

	return draftEnvelope.Order.PublicID, nil
}

// GetStatus polls GET /orders/{publicOrderId} — Slant's webhook support
// ("order.shipped" is the only event type as of this writing, per their own
// docs) is described as not yet fully implemented, so polling is the
// reliable path rather than waiting on a webhook that may never arrive.
func (s *SlantAdapter) GetStatus(ctx context.Context, externalID string) (Status, error) {
	data, err := s.request(ctx, http.MethodGet, "/orders/"+externalID, nil)
	if err != nil {
		return StatusUnknown, fmt.Errorf("slant: get order %s: %w", externalID, err)
	}
	var envelope slantOrderEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return StatusUnknown, fmt.Errorf("slant: decode order %s: %w", externalID, err)
	}

	switch envelope.Order.Status {
	case "DRAFT", "PROCESSING":
		return StatusSubmitted, nil
	case "SHIPPED", "DELIVERED":
		return StatusShipped, nil
	case "CANCELED":
		return Status("canceled"), nil
	default:
		return StatusUnknown, nil
	}
}
