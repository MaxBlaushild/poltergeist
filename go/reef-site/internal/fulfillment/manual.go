package fulfillment

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
)

// ManualAdapter is R-7.2's v1 default: write the STL files (by reference —
// they're already in S3 under their geometry_hash key) and a manifest CSV
// to object storage, then notify the operator by email. This lets the site
// ship before any fulfillment contract exists, and lets the operator print
// the first orders themselves to measure real COGS and reject rates
// (R-7.4).
type ManualAdapter struct {
	AwsClient     aws.AWSClient
	EmailClient   email.EmailClient
	Bucket        string
	OperatorEmail string
	FromAddress   string
}

func NewManualAdapter(awsClient aws.AWSClient, emailClient email.EmailClient, bucket, operatorEmail, fromAddress string) *ManualAdapter {
	return &ManualAdapter{
		AwsClient:     awsClient,
		EmailClient:   emailClient,
		Bucket:        bucket,
		OperatorEmail: operatorEmail,
		FromAddress:   fromAddress,
	}
}

func (m *ManualAdapter) SubmitOrder(ctx context.Context, o Order) (string, error) {
	manifestKey := fmt.Sprintf("reef/orders/%s/manifest.csv", o.OrderToken)
	manifestCSV, err := buildManifestCSV(o)
	if err != nil {
		return "", fmt.Errorf("fulfillment/manual: building manifest: %w", err)
	}
	manifestURL, err := m.AwsClient.UploadImageToS3(m.Bucket, manifestKey, manifestCSV)
	if err != nil {
		return "", fmt.Errorf("fulfillment/manual: uploading manifest: %w", err)
	}

	if err := m.notifyOperator(o, manifestURL); err != nil {
		// The manifest is already durably in object storage — a failed
		// notification shouldn't fail order submission outright, but it
		// must not be silent either. Callers should alert on this in logs.
		return "", fmt.Errorf("fulfillment/manual: notifying operator (manifest uploaded to %s): %w", manifestURL, err)
	}

	return "manual-" + o.OrderToken, nil
}

func (m *ManualAdapter) GetStatus(ctx context.Context, externalID string) (Status, error) {
	// Manual fulfillment progress (printed / shipped) is tracked by the
	// operator directly against reef_orders.fulfillment_status, not polled
	// from anywhere — there is nowhere else for that state to live for a
	// human-run print queue. Submission having succeeded is the only thing
	// this adapter itself can attest to.
	return StatusSubmitted, nil
}

func buildManifestCSV(o Order) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write([]string{"order_token", "customer_email", "ship_to", "product_slug", "variant", "quantity", "stl_key"}); err != nil {
		return nil, err
	}
	shipTo := fmt.Sprintf("%s | %s %s | %s, %s %s %s", o.ShippingName, o.ShippingLine1, o.ShippingLine2, o.ShippingCity, o.ShippingState, o.ShippingZip, o.ShippingCountry)
	for _, item := range o.Items {
		if err := w.Write([]string{
			o.OrderToken,
			o.CustomerEmail,
			shipTo,
			item.ProductSlug,
			item.VariantKey,
			strconv.Itoa(item.Quantity),
			item.STLKey,
		}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *ManualAdapter) notifyOperator(o Order, manifestURL string) error {
	var body bytes.Buffer
	fmt.Fprintf(&body, "New reef-site order %s\n\n", o.OrderToken)
	fmt.Fprintf(&body, "Customer: %s\n", o.CustomerEmail)
	fmt.Fprintf(&body, "Ship to: %s, %s, %s, %s %s, %s\n\n", o.ShippingName, o.ShippingLine1, o.ShippingCity, o.ShippingState, o.ShippingZip, o.ShippingCountry)
	fmt.Fprintf(&body, "Items:\n")
	for _, item := range o.Items {
		fmt.Fprintf(&body, "- %s x%d", item.ProductSlug, item.Quantity)
		if item.VariantKey != "" {
			fmt.Fprintf(&body, " (%s)", item.VariantKey)
		}
		if item.STLKey != "" {
			fmt.Fprintf(&body, " — STL: %s", item.STLKey)
		}
		fmt.Fprintf(&body, "\n")
	}
	fmt.Fprintf(&body, "\nManifest: %s\n", manifestURL)
	fmt.Fprintf(&body, "\nGenerated %s\n", time.Now().UTC().Format(time.RFC3339))

	return m.EmailClient.SendMail(email.Email{
		Subject:          "reef-site order " + o.OrderToken,
		Name:             "reef-site",
		Email:            m.OperatorEmail,
		PlainTextContent: body.String(),
		HtmlContent:      "<pre>" + body.String() + "</pre>",
	})
}
