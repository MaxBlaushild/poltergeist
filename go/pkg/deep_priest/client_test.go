package deep_priest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPetitionTheFount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/consult" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"answer":"ok"}`))
	}))
	defer server.Close()

	client := &deepPriest{
		baseURL:       server.URL,
		consultClient: &http.Client{Timeout: time.Second},
	}

	answer, err := client.PetitionTheFount(&Question{Question: "hello"})
	if err != nil {
		t.Fatalf("PetitionTheFount returned error: %v", err)
	}
	if answer == nil || answer.Answer != "ok" {
		t.Fatalf("unexpected answer: %#v", answer)
	}
}

func TestPetitionTheFountTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"answer":"late"}`))
	}))
	defer server.Close()

	client := &deepPriest{
		baseURL:       server.URL,
		consultClient: &http.Client{Timeout: 10 * time.Millisecond},
	}

	_, err := client.PetitionTheFount(&Question{Question: "hello"})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}
