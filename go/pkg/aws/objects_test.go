package aws

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func testObjectClient(t *testing.T, handler http.HandlerFunc) *client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	sess := session.Must(session.NewSession(&sdk.Config{
		Region: sdk.String("us-east-1"), Endpoint: sdk.String(server.URL),
		S3ForcePathStyle: sdk.Bool(true), MaxRetries: sdk.Int(0),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
	}))
	return &client{s3: s3.New(sess)}
}

func TestObjectStorageUploadAndDelete(t *testing.T) {
	methods := []string{}
	storage := testObjectClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		if r.URL.Path != "/private-images/trip-heroes/photo.png" {
			t.Errorf("unexpected storage path %s", r.URL.Path)
		}
		if r.Method == "PUT" {
			body, _ := io.ReadAll(r.Body)
			if string(body) != "pixels" || r.Header.Get("Content-Type") != "image/png" || r.Header.Get("Cache-Control") != "private, no-store" {
				t.Error("upload lost bytes or privacy/content headers")
			}
		}
		w.WriteHeader(200)
	})
	if err := storage.PutObject(context.Background(), "private-images", "trip-heroes/photo.png", "image/png", []byte("pixels")); err != nil {
		t.Fatal(err)
	}
	if err := storage.DeleteObject(context.Background(), "private-images", "trip-heroes/photo.png"); err != nil {
		t.Fatal(err)
	}
	if strings.Join(methods, ",") != "PUT,DELETE" {
		t.Fatalf("unexpected methods %v", methods)
	}
}

func TestObjectStorageReadLimit(t *testing.T) {
	for _, chunked := range []bool{false, true} {
		t.Run(map[bool]string{false: "content-length", true: "chunked"}[chunked], func(t *testing.T) {
			storage := testObjectClient(t, func(w http.ResponseWriter, r *http.Request) {
				if chunked {
					w.(http.Flusher).Flush()
				}
				io.WriteString(w, "pixels")
			})
			data, err := storage.ReadObject(context.Background(), "private-images", "photo.png", 6)
			if err != nil || string(data) != "pixels" {
				t.Fatalf("read failed: %q %v", data, err)
			}
			if _, err := storage.ReadObject(context.Background(), "private-images", "photo.png", 5); err == nil {
				t.Fatal("oversized object accepted")
			}
		})
	}
}

func TestObjectStorageHonorsCancelledContext(t *testing.T) {
	storage := testObjectClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("cancelled request reached storage") })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := storage.PutObject(ctx, "private-images", "photo.png", "image/png", []byte("pixels")); err == nil {
		t.Fatal("cancelled upload succeeded")
	}
	if _, err := storage.ReadObject(ctx, "private-images", "photo.png", 6); err == nil {
		t.Fatal("cancelled read succeeded")
	}
	if err := storage.DeleteObject(ctx, "private-images", "photo.png"); err == nil {
		t.Fatal("cancelled delete succeeded")
	}
}
