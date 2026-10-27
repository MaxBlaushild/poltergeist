package googlemaps

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"
)

const lodgingFixture = `{"id":"hotel_1","displayName":{"text":"Casa Lisboa"},"formattedAddress":"Lisbon, Portugal","types":["hotel","lodging"],"websiteUri":"https://hotel.example","googleMapsUri":"https://maps.google.com/?cid=1","rating":4.6,"userRatingCount":812,"attributions":[{"provider":"Local provider","providerUri":"https://provider.example"}],"photos":[{"name":"places/hotel_1/photos/photo_1","widthPx":1600,"heightPx":1200,"authorAttributions":[{"displayName":"Traveler","uri":"//maps.google.com/contrib/1","photoUri":"//lh3.googleusercontent.com/avatar"}]},{"name":"places/other/photos/photo_2"},{"name":"https://malicious.example/photo"}]}`

func TestLodgingSearchUsesNewAPIAndNarrowFields(t *testing.T) {
	destinationTransport(t, func(r *http.Request) (*http.Response, error) {
		if r.Method != "POST" || r.URL.String() != "https://places.googleapis.com/v1/places:searchText" {
			t.Fatalf("wrong API: %s %s", r.Method, r.URL)
		}
		if r.Header.Get("X-Goog-Api-Key") != "secret-key" || strings.Contains(r.Header.Get("X-Goog-FieldMask"), "*") || !strings.Contains(r.Header.Get("X-Goog-FieldMask"), "places.photos") || !strings.Contains(r.Header.Get("X-Goog-FieldMask"), "places.attributions") {
			t.Fatal("missing credentials or explicit fields")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["textQuery"] != "Casa Lisboa" || body["includedType"] != "lodging" || body["strictTypeFiltering"] != true || body["pageSize"] != float64(10) {
			t.Fatalf("wrong search body: %v", body)
		}
		if _, ok := body["locationBias"]; ok {
			t.Fatal("ordinary search unexpectedly sent location bias")
		}
		return destinationResponse(r, 200, `{"places":[`+lodgingFixture+`,`+lodgingFixture+`]}`), nil
	})
	rows, err := (&client{apiKey: "secret-key"}).SearchLodgingContext(context.Background(), " Casa Lisboa ")
	if err != nil || len(rows) != 1 {
		t.Fatalf("search: %+v %v", rows, err)
	}
	assertLodgingFixture(t, &rows[0])
}

func TestLodgingSearchNearSendsCircleWithoutRestrictingResults(t *testing.T) {
	for _, point := range [][2]float64{{38.7223, -9.1393}, {0, 0}, {90, 180}, {-90, -180}} {
		destinationTransport(t, func(r *http.Request) (*http.Response, error) {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			bias, ok := body["locationBias"].(map[string]any)
			if !ok {
				t.Fatalf("missing bias: %v", body)
			}
			circle, ok := bias["circle"].(map[string]any)
			if !ok {
				t.Fatalf("missing circle: %v", bias)
			}
			center, ok := circle["center"].(map[string]any)
			if !ok {
				t.Fatalf("missing center: %v", circle)
			}
			if center["latitude"] != point[0] || center["longitude"] != point[1] || circle["radius"] != float64(50000) || body["textQuery"] != "Casa Lisboa" || body["includedType"] != "lodging" {
				t.Fatalf("wrong location bias: %v", body)
			}
			if _, ok := body["locationRestriction"]; ok {
				t.Fatal("bias unexpectedly became hard geographic restriction")
			}
			return destinationResponse(r, 200, `{"places":[`+lodgingFixture+`]}`), nil
		})
		rows, err := (&client{apiKey: "key"}).SearchLodgingNearContext(context.Background(), "Casa Lisboa", point[0], point[1])
		if err != nil || len(rows) != 1 {
			t.Fatalf("nearby search: %v %v", rows, err)
		}
	}
}

func TestLodgingSearchNearRejectsNonfiniteAndOutOfRangeCoordinates(t *testing.T) {
	destinationTransport(t, func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid coordinates made a provider request")
		return nil, nil
	})
	for _, point := range [][2]float64{{math.NaN(), 0}, {0, math.NaN()}, {math.Inf(1), 0}, {0, math.Inf(-1)}, {90.1, 0}, {-90.1, 0}, {0, 180.1}, {0, -180.1}} {
		if _, err := (&client{apiKey: "key"}).SearchLodgingNearContext(context.Background(), "hotel", point[0], point[1]); err == nil {
			t.Fatalf("invalid coordinates accepted: %v", point)
		}
	}
}

func assertLodgingFixture(t *testing.T, got *LodgingPlace) {
	t.Helper()
	if got.ID != "hotel_1" || got.DisplayName.Text != "Casa Lisboa" || got.FormattedAddress != "Lisbon, Portugal" || got.Rating != 4.6 || got.UserRatingCount != 812 || got.WebsiteURI != "https://hotel.example" || len(got.Attributions) != 1 || got.Attributions[0].Provider != "Local provider" {
		t.Fatalf("details lost property data: %+v", got)
	}
	if len(got.Photos) != 1 || len(got.Photos[0].AuthorAttributions) != 1 || got.Photos[0].AuthorAttributions[0].DisplayName != "Traveler" {
		t.Fatalf("photos or required attribution incorrect: %+v", got.Photos)
	}
}

func TestLodgingDetailsValidateReturnedIdentityAndFreshPhotos(t *testing.T) {
	responseBody := lodgingFixture
	destinationTransport(t, func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" || r.URL.Path != "/v1/places/hotel_1" || r.Header.Get("X-Goog-FieldMask") != lodgingFields {
			t.Fatalf("details request = %v", r)
		}
		return destinationResponse(r, 200, responseBody), nil
	})
	c := &client{apiKey: "test-key"}
	place, err := c.GetLodgingContext(context.Background(), "hotel_1")
	if err != nil {
		t.Fatal(err)
	}
	assertLodgingFixture(t, place)
	responseBody = strings.Replace(lodgingFixture, `"id":"hotel_1"`, `"id":"other"`, 1)
	if _, err := c.GetLodgingContext(context.Background(), "hotel_1"); err == nil {
		t.Fatal("mismatched place response accepted")
	}
}

func TestLodgingRejectsInvalidResourcesWithoutRequests(t *testing.T) {
	destinationTransport(t, func(*http.Request) (*http.Response, error) { t.Fatal("invalid input made a request"); return nil, nil })
	c := &client{apiKey: "test-key"}
	for _, id := range []string{"", "../hotel", "hotel?key=stolen", "https://evil.example", strings.Repeat("x", 513)} {
		if _, err := c.GetLodgingContext(context.Background(), id); err == nil {
			t.Fatalf("invalid place ID accepted: %q", id)
		}
	}
	for _, name := range []string{"", "places/a/photos/../secret", "places/a/photos/abc?x=y", "https://evil.example/abc", "places/a/photos/b/media"} {
		if _, _, err := c.GetLodgingPhotoContext(context.Background(), name); err == nil {
			t.Fatalf("invalid photo name accepted: %q", name)
		}
	}
	if _, err := (&client{}).GetLodgingContext(context.Background(), "hotel_1"); err == nil {
		t.Fatal("missing key accepted")
	}
}

func TestLodgingPhotoUsesCredentialFreeAllowedImageHost(t *testing.T) {
	var calls int
	image := []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82}
	destinationTransport(t, func(r *http.Request) (*http.Response, error) {
		calls++
		switch calls {
		case 1:
			if r.URL.Host != "places.googleapis.com" || r.URL.Path != "/v1/places/hotel_1/photos/photo_1/media" || r.URL.Query().Get("skipHttpRedirect") != "true" || r.Header.Get("X-Goog-Api-Key") != "secret-key" {
				t.Fatalf("photo resolve request = %v", r.URL)
			}
			return destinationResponse(r, 200, `{"photoUri":"https://lh3.googleusercontent.com/photo"}`), nil
		case 2:
			if r.Header.Get("X-Goog-Api-Key") != "" || r.Header.Get("Authorization") != "" {
				t.Fatal("credentials leaked to image host")
			}
			response := destinationResponse(r, 302, "")
			response.Header.Set("Location", "https://lh4.googleusercontent.com/photo")
			return response, nil
		default:
			if calls != 3 || r.URL.Host != "lh4.googleusercontent.com" || r.Header.Get("X-Goog-Api-Key") != "" {
				t.Fatal("unexpected image redirect")
			}
			return destinationResponse(r, 200, string(image)), nil
		}
	})
	got, mime, err := (&client{apiKey: "secret-key"}).GetLodgingPhotoContext(context.Background(), "places/hotel_1/photos/photo_1")
	if err != nil || mime != "image/png" || !bytes.Equal(got, image) || calls != 3 {
		t.Fatalf("photo = %v %s %v (calls=%d)", got, mime, err, calls)
	}
}

func TestLodgingPhotoRejectsUntrustedURLAndRedirect(t *testing.T) {
	for _, raw := range []string{"http://lh3.googleusercontent.com/p", "https://127.0.0.1/p", "https://googleusercontent.com.attacker.example/p", "https://evilgoogleusercontent.com/p", "https://user:pass@lh3.googleusercontent.com/p", "https://lh3.googleusercontent.com:8443/p", "https://lh3.googleusercontent.com/p#fragment"} {
		t.Run(raw, func(t *testing.T) {
			calls := 0
			destinationTransport(t, func(r *http.Request) (*http.Response, error) {
				calls++
				data, _ := json.Marshal(map[string]string{"photoUri": raw})
				return destinationResponse(r, 200, string(data)), nil
			})
			if _, _, err := (&client{apiKey: "key"}).GetLodgingPhotoContext(context.Background(), "places/a/photos/b"); err == nil || calls != 1 {
				t.Fatalf("unsafe image URL requested: %v calls=%d", err, calls)
			}
		})
	}
	t.Run("redirect", func(t *testing.T) {
		calls := 0
		destinationTransport(t, func(r *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				return destinationResponse(r, 200, `{"photoUri":"https://lh3.googleusercontent.com/p"}`), nil
			}
			if calls > 2 {
				t.Fatal("untrusted redirect followed")
			}
			response := destinationResponse(r, 302, "")
			response.Header.Set("Location", "https://127.0.0.1/private")
			return response, nil
		})
		if _, _, err := (&client{apiKey: "key"}).GetLodgingPhotoContext(context.Background(), "places/a/photos/b"); err == nil {
			t.Fatal("unsafe redirect accepted")
		}
	})
}

func TestLodgingPhotoBoundsAndRejectsHTML(t *testing.T) {
	for _, tc := range []struct {
		name   string
		data   []byte
		length int64
	}{
		{"HTML", []byte("<html>error</html>"), -1},
		{"declared oversized", nil, maxLodgingPhotoBytes + 1},
		{"stream oversized", bytes.Repeat([]byte("a"), maxLodgingPhotoBytes+1), -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			destinationTransport(t, func(r *http.Request) (*http.Response, error) {
				if r.URL.Host == "places.googleapis.com" {
					return destinationResponse(r, 200, `{"photoUri":"https://lh3.googleusercontent.com/p"}`), nil
				}
				response := destinationResponse(r, 200, "")
				response.ContentLength = tc.length
				response.Body = io.NopCloser(bytes.NewReader(tc.data))
				return response, nil
			})
			if _, _, err := (&client{apiKey: "key"}).GetLodgingPhotoContext(context.Background(), "places/a/photos/b"); err == nil {
				t.Fatal("invalid/oversized photo accepted")
			}
		})
	}
}

func TestLodgingProviderErrorsDoNotExposeResponseOrKeys(t *testing.T) {
	for _, status := range []int{302, 403, 429, 500} {
		destinationTransport(t, func(r *http.Request) (*http.Response, error) {
			response := destinationResponse(r, status, `{"error":"secret-key private upstream body"}`)
			response.Header.Set("Location", "https://attacker.example")
			return response, nil
		})
		_, err := (&client{apiKey: "secret-key"}).GetLodgingContext(context.Background(), "hotel_1")
		if err == nil || strings.Contains(err.Error(), "secret-key") || strings.Contains(err.Error(), "upstream body") {
			t.Fatalf("unsafe provider error: %v", err)
		}
	}
}

func TestLodgingRequestsHonorCancellation(t *testing.T) {
	for _, method := range []string{"search", "details", "photo"} {
		t.Run(method, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			destinationTransport(t, func(r *http.Request) (*http.Response, error) { cancel(); return nil, r.Context().Err() })
			c := &client{apiKey: "key"}
			var err error
			switch method {
			case "search":
				_, err = c.SearchLodgingContext(ctx, "Lisbon")
			case "details":
				_, err = c.GetLodgingContext(ctx, "hotel_1")
			case "photo":
				_, _, err = c.GetLodgingPhotoContext(ctx, "places/a/photos/b")
			}
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation lost: %v", err)
			}
		})
	}
}
