package processors

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
)

const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO8G6ioAAAAASUVORK5CYII="

func TestDownloadThumbnailSourceFromURL(t *testing.T) {
	expected, err := base64.StdEncoding.DecodeString(tinyPNGBase64)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(expected)
	}))
	defer server.Close()

	got, err := downloadThumbnailSource(server.URL)
	if err != nil {
		t.Fatalf("downloadThumbnailSource() error = %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("downloadThumbnailSource() returned empty bytes")
	}
	if string(got) != string(expected) {
		t.Fatalf("downloadThumbnailSource() bytes mismatch")
	}
}

func TestDownloadThumbnailSourceFromBase64(t *testing.T) {
	got, err := downloadThumbnailSource(tinyPNGBase64)
	if err != nil {
		t.Fatalf("downloadThumbnailSource() error = %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("downloadThumbnailSource() returned empty bytes")
	}
}

func TestDownloadThumbnailSourceFromDataURI(t *testing.T) {
	dataURI := "data:image/png;base64," + tinyPNGBase64
	got, err := downloadThumbnailSource(dataURI)
	if err != nil {
		t.Fatalf("downloadThumbnailSource() error = %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("downloadThumbnailSource() returned empty bytes")
	}
}

func TestDownloadThumbnailSourceFromJSONPayload(t *testing.T) {
	payload := "{\"data\":[{\"b64_json\":\"" + tinyPNGBase64 + "\"}]}"
	got, err := downloadThumbnailSource(payload)
	if err != nil {
		t.Fatalf("downloadThumbnailSource() error = %v", err)
	}
	if len(got) == 0 {
		t.Fatalf("downloadThumbnailSource() returned empty bytes")
	}
}
