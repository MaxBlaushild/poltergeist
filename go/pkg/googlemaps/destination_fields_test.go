package googlemaps

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type destinationTestTransport func(*http.Request) (*http.Response, error)

func (f destinationTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func destinationTransport(t *testing.T, transport destinationTestTransport) {
	t.Helper()
	before := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = before })
}

func destinationResponse(r *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header), Request: r}
}

func TestDestinationDetailsDecodeStructuredGeography(t *testing.T) {
	destinationTransport(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "places.googleapis.com" || !strings.HasSuffix(r.URL.Path, "/tokyo") || r.Header.Get("X-Goog-Api-Key") != "test-key" {
			t.Fatalf("unexpected details request: %v", r.URL)
		}
		return destinationResponse(r, 200, `{"id":"tokyo","types":["locality"],"addressComponents":[{"longText":"Japan","shortText":"JP","types":["country","political"]}],"timeZone":{"id":"Asia/Tokyo"}}`), nil
	})
	got, err := (&client{apiKey: "test-key"}).FindPlaceByIDContext(context.Background(), "tokyo")
	if err != nil || got == nil || len(got.AddressComponents) != 1 || got.AddressComponents[0].LongText != "Japan" || got.AddressComponents[0].ShortText != "JP" || len(got.AddressComponents[0].Types) != 2 || got.TimeZone == nil || got.TimeZone.ID != "Asia/Tokyo" {
		t.Fatalf("structured geography did not decode: %+v, %v", got, err)
	}
}

func TestDestinationSearchEscapesQueryAndDecodesCandidates(t *testing.T) {
	const query = "São Paulo & nearby 日本"
	destinationTransport(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("input") != query || r.URL.Query().Get("key") != "test-key" {
			t.Fatalf("search query changed in transit: %v", r.URL)
		}
		return destinationResponse(r, 200, `{"status":"OK","candidates":[{"place_id":"sao","name":"São Paulo","types":["locality"],"geometry":{"location":{"lat":-23.55,"lng":-46.63}}}]}`), nil
	})
	got, err := (&client{apiKey: "test-key"}).FindCandidatesByQueryContext(context.Background(), query)
	if err != nil || len(got) != 1 || got[0].PlaceID != "sao" || got[0].Geometry.Location.Lat != -23.55 {
		t.Fatalf("search = %+v, %v", got, err)
	}
}

func TestDestinationRequestsRejectHTTPFailures(t *testing.T) {
	for _, method := range []string{"search", "details"} {
		t.Run(method, func(t *testing.T) {
			destinationTransport(t, func(r *http.Request) (*http.Response, error) {
				// A superficially valid body must not turn an HTTP failure into success.
				return destinationResponse(r, 503, `{"status":"OK","id":"tokyo","candidates":[]}`), nil
			})
			c := &client{apiKey: "test-key"}
			var err error
			if method == "search" {
				_, err = c.FindCandidatesByQueryContext(context.Background(), "Tokyo")
			} else {
				_, err = c.FindPlaceByIDContext(context.Background(), "tokyo")
			}
			if err == nil || !strings.Contains(err.Error(), "503") {
				t.Fatalf("HTTP failure was accepted: %v", err)
			}
		})
	}
}

func TestDestinationSearchDistinguishesEmptyFromProviderFailure(t *testing.T) {
	for _, status := range []string{"ZERO_RESULTS", "REQUEST_DENIED"} {
		t.Run(status, func(t *testing.T) {
			destinationTransport(t, func(r *http.Request) (*http.Response, error) {
				return destinationResponse(r, 200, `{"status":"`+status+`","error_message":"fixture"}`), nil
			})
			rows, err := (&client{}).FindCandidatesByQueryContext(context.Background(), "query")
			if status == "ZERO_RESULTS" && (err != nil || len(rows) != 0) {
				t.Fatalf("empty result = %v, %v", rows, err)
			}
			if status == "REQUEST_DENIED" && err == nil {
				t.Fatal("provider failure was reported as empty success")
			}
		})
	}
}

func TestDestinationRequestsCarryCancellation(t *testing.T) {
	for _, method := range []string{"search", "details"} {
		t.Run(method, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			destinationTransport(t, func(r *http.Request) (*http.Response, error) {
				cancel()
				if !errors.Is(r.Context().Err(), context.Canceled) {
					t.Fatal("outbound request lost caller cancellation")
				}
				return nil, r.Context().Err()
			})
			c := &client{}
			var err error
			if method == "search" {
				_, err = c.FindCandidatesByQueryContext(ctx, "Tokyo")
			} else {
				_, err = c.FindPlaceByIDContext(ctx, "tokyo")
			}
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation lost through error wrapping: %v", err)
			}
		})
	}
}
