package fulfillment

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/email"
)

type fakeAWSClient struct {
	uploads map[string][]byte
	failOn  string
}

func newFakeAWSClient() *fakeAWSClient {
	return &fakeAWSClient{uploads: map[string][]byte{}}
}

func (f *fakeAWSClient) UploadImageToS3(bucket, key string, image []byte) (string, error) {
	if f.failOn != "" && key == f.failOn {
		return "", errUploadFailed
	}
	f.uploads[key] = image
	return "https://" + bucket + ".s3.amazonaws.com/" + key, nil
}
func (f *fakeAWSClient) DeleteObjectFromS3(bucket, key string) error { return nil }
func (f *fakeAWSClient) GetObjectLastModified(bucket, key string) (*time.Time, error) {
	return nil, nil
}
func (f *fakeAWSClient) GeneratePresignedURL(bucket, key string, expiry time.Duration) (string, error) {
	return "", nil
}
func (f *fakeAWSClient) GeneratePresignedUploadURL(bucket, key string, expiry time.Duration) (string, error) {
	return "", nil
}
func (f *fakeAWSClient) GeneratePresignedUploadURLWithContentType(bucket, key, contentType string, expiry time.Duration) (string, error) {
	return "", nil
}

type errString string

func (e errString) Error() string { return string(e) }

const errUploadFailed = errString("upload failed")

type fakeEmailClient struct {
	sent []email.Email
}

func (f *fakeEmailClient) SendMail(e email.Email) error {
	f.sent = append(f.sent, e)
	return nil
}

func testOrder() Order {
	return Order{
		OrderToken:      "abc123",
		CustomerEmail:   "customer@example.com",
		ShippingName:    "Jane Reefer",
		ShippingLine1:   "123 Coral Ave",
		ShippingCity:    "Miami",
		ShippingState:   "FL",
		ShippingZip:     "33101",
		ShippingCountry: "US",
		Items: []OrderItem{
			{ProductSlug: "magnetic-frag-rack", Quantity: 1, STLKey: "reef/stl/hash123.stl"},
			{ProductSlug: "feeding-ring", VariantKey: "small", Quantity: 2},
		},
	}
}

// R-7.2/R-10 acceptance: "An order placed under ManualAdapter delivers STL
// files and a manifest to the operator's inbox." STL files are already in
// S3 by geometry_hash key by the time an order is placed (R-3.3's cache);
// this asserts the manifest referencing them is uploaded and the operator
// is emailed with that manifest link and full order detail.
func TestManualAdapter_SubmitOrder_UploadsManifestAndNotifiesOperator(t *testing.T) {
	awsClient := newFakeAWSClient()
	emailClient := &fakeEmailClient{}
	adapter := NewManualAdapter(awsClient, emailClient, "reef-bucket", "operator@example.com", "reef@example.com")

	externalID, err := adapter.SubmitOrder(context.Background(), testOrder())
	if err != nil {
		t.Fatalf("SubmitOrder: %v", err)
	}
	if externalID == "" {
		t.Fatal("expected a non-empty external ID")
	}

	manifestKey := "reef/orders/abc123/manifest.csv"
	manifest, ok := awsClient.uploads[manifestKey]
	if !ok {
		t.Fatalf("expected a manifest uploaded to %s, got keys %v", manifestKey, keysOf(awsClient.uploads))
	}
	manifestStr := string(manifest)
	if !strings.Contains(manifestStr, "magnetic-frag-rack") || !strings.Contains(manifestStr, "feeding-ring") {
		t.Fatalf("manifest missing expected products:\n%s", manifestStr)
	}
	if !strings.Contains(manifestStr, "reef/stl/hash123.stl") {
		t.Fatalf("manifest missing STL key reference:\n%s", manifestStr)
	}
	// quantity column for the feeding-ring row should be 2
	if !strings.Contains(manifestStr, ",2,") {
		t.Fatalf("manifest missing expected quantity:\n%s", manifestStr)
	}

	if len(emailClient.sent) != 1 {
		t.Fatalf("expected exactly one operator email, got %d", len(emailClient.sent))
	}
	sent := emailClient.sent[0]
	if sent.Email != "operator@example.com" {
		t.Fatalf("email sent to %q, want operator@example.com", sent.Email)
	}
	if !strings.Contains(sent.PlainTextContent, "abc123") {
		t.Fatalf("email body missing order token:\n%s", sent.PlainTextContent)
	}
	if !strings.Contains(sent.PlainTextContent, manifestKey) {
		t.Fatalf("email body missing manifest reference:\n%s", sent.PlainTextContent)
	}
}

func TestManualAdapter_SubmitOrder_FailsIfManifestUploadFails(t *testing.T) {
	awsClient := newFakeAWSClient()
	awsClient.failOn = "reef/orders/abc123/manifest.csv"
	emailClient := &fakeEmailClient{}
	adapter := NewManualAdapter(awsClient, emailClient, "reef-bucket", "operator@example.com", "reef@example.com")

	if _, err := adapter.SubmitOrder(context.Background(), testOrder()); err == nil {
		t.Fatal("expected an error when the manifest upload fails")
	}
	if len(emailClient.sent) != 0 {
		t.Fatal("must not notify the operator if the manifest never made it to storage")
	}
}

func TestManualAdapter_GetStatus_AlwaysSubmitted(t *testing.T) {
	adapter := NewManualAdapter(nil, nil, "", "", "")
	status, err := adapter.GetStatus(context.Background(), "manual-abc123")
	if err != nil {
		t.Fatal(err)
	}
	if status != StatusSubmitted {
		t.Fatalf("status = %s, want %s", status, StatusSubmitted)
	}
}

func keysOf(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
