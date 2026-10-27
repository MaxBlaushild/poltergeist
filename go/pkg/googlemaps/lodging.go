package googlemaps

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// LodgingClient is optional so existing Maps implementations remain compatible.
// Responses are live provider content, not a catalog or a persistence model.
type LodgingClient interface {
	SearchLodgingContext(context.Context, string) ([]LodgingPlace, error)
	GetLodgingContext(context.Context, string) (*LodgingPlace, error)
	GetLodgingPhotoContext(context.Context, string) ([]byte, string, error)
}

// LodgingNearClient supplies a geographic preference without restricting the
// results or changing the original lodging interface.
type LodgingNearClient interface {
	SearchLodgingNearContext(context.Context, string, float64, float64) ([]LodgingPlace, error)
}

func ValidLodgingCoordinates(latitude, longitude float64) bool {
	return !math.IsNaN(latitude) && !math.IsNaN(longitude) &&
		!math.IsInf(latitude, 0) && !math.IsInf(longitude, 0) &&
		latitude >= -90 && latitude <= 90 && longitude >= -180 && longitude <= 180
}

type LodgingAuthorAttribution struct {
	DisplayName string `json:"displayName"`
	URI         string `json:"uri"`
	PhotoURI    string `json:"photoUri"`
}

type LodgingAttribution struct {
	Provider    string `json:"provider"`
	ProviderURI string `json:"providerUri"`
}

type LodgingPhoto struct {
	Name               string                     `json:"name"`
	Width              int                        `json:"widthPx"`
	Height             int                        `json:"heightPx"`
	GoogleMapsURI      string                     `json:"googleMapsUri"`
	AuthorAttributions []LodgingAuthorAttribution `json:"authorAttributions"`
}

type LodgingPlace struct {
	ID               string               `json:"id"`
	DisplayName      LocalizedText        `json:"displayName"`
	FormattedAddress string               `json:"formattedAddress"`
	Types            []string             `json:"types"`
	WebsiteURI       string               `json:"websiteUri"`
	GoogleMapsURI    string               `json:"googleMapsUri"`
	Rating           float64              `json:"rating"`
	UserRatingCount  int                  `json:"userRatingCount"`
	Photos           []LodgingPhoto       `json:"photos"`
	Attributions     []LodgingAttribution `json:"attributions"`
}

const lodgingFields = "id,displayName,formattedAddress,types,websiteUri,googleMapsUri,rating,userRatingCount,photos,attributions"
const maxLodgingJSONBytes = 2 << 20
const maxLodgingPhotoBytes = 8 << 20

var lodgingPlaceID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,512}$`)
var lodgingPhotoName = regexp.MustCompile(`^places/[A-Za-z0-9_-]{1,512}/photos/[A-Za-z0-9_=-]+$`)

func ValidLodgingPlaceID(id string) bool { return lodgingPlaceID.MatchString(id) }
func ValidLodgingPhotoName(name string) bool {
	return len(name) <= 2584 && lodgingPhotoName.MatchString(name)
}

func (c *client) lodgingJSON(ctx context.Context, method, path, fields string, payload any, result any) error {
	if strings.TrimSpace(c.apiKey) == "" {
		return errors.New("Google Places is not configured")
	}
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return errors.New("invalid Google Places request")
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, "https://places.googleapis.com/v1/"+path, body)
	if err != nil {
		return errors.New("invalid Google Places request")
	}
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if fields != "" {
		req.Header.Set("X-Goog-FieldMask", fields)
	}
	// Never forward credentials to an API redirect or include provider bodies in
	// returned errors/logs. Photos use a second, credential-free request below.
	httpClient := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("Google Places request failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Google Places returned HTTP %d", response.StatusCode)
	}
	data, err := readLodgingBytes(response, maxLodgingJSONBytes)
	if err != nil {
		return err
	}
	if json.Unmarshal(data, result) != nil {
		return errors.New("Google Places returned invalid data")
	}
	return nil
}

func (c *client) SearchLodgingContext(ctx context.Context, query string) ([]LodgingPlace, error) {
	return c.searchLodging(ctx, query, nil)
}

func (c *client) SearchLodgingNearContext(ctx context.Context, query string, latitude, longitude float64) ([]LodgingPlace, error) {
	if !ValidLodgingCoordinates(latitude, longitude) {
		return nil, errors.New("invalid property search coordinates")
	}
	return c.searchLodging(ctx, query, map[string]any{
		"circle": map[string]any{
			"center": map[string]float64{"latitude": latitude, "longitude": longitude},
			"radius": 50000,
		},
	})
}

func (c *client) searchLodging(ctx context.Context, query string, locationBias map[string]any) ([]LodgingPlace, error) {
	query = strings.TrimSpace(query)
	if query == "" || utf8.RuneCountInString(query) > 500 {
		return nil, errors.New("enter a property search of up to 500 characters")
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var response struct {
		Places []LodgingPlace `json:"places"`
	}
	fields := "places." + strings.ReplaceAll(lodgingFields, ",", ",places.")
	payload := map[string]any{
		"textQuery": query, "includedType": "lodging", "strictTypeFiltering": true, "pageSize": 10,
	}
	if locationBias != nil {
		payload["locationBias"] = locationBias
	}
	err := c.lodgingJSON(ctx, http.MethodPost, "places:searchText", fields, payload, &response)
	if err != nil {
		return nil, err
	}
	result := make([]LodgingPlace, 0, len(response.Places))
	seen := map[string]bool{}
	for _, place := range response.Places {
		if !ValidLodgingPlaceID(place.ID) || strings.TrimSpace(place.DisplayName.Text) == "" || seen[place.ID] {
			continue
		}
		seen[place.ID] = true
		place.Photos = validLodgingPhotos(place.ID, place.Photos)
		result = append(result, place)
		if len(result) == 10 {
			break
		}
	}
	return result, nil
}

func (c *client) GetLodgingContext(ctx context.Context, id string) (*LodgingPlace, error) {
	if !ValidLodgingPlaceID(id) {
		return nil, errors.New("invalid Google place ID")
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var place LodgingPlace
	if err := c.lodgingJSON(ctx, http.MethodGet, "places/"+id, lodgingFields, nil, &place); err != nil {
		return nil, err
	}
	if place.ID != id || strings.TrimSpace(place.DisplayName.Text) == "" {
		return nil, errors.New("Google Places returned an unavailable property")
	}
	place.Photos = validLodgingPhotos(place.ID, place.Photos)
	return &place, nil
}

func validLodgingPhotos(placeID string, photos []LodgingPhoto) []LodgingPhoto {
	result := make([]LodgingPhoto, 0, 5)
	seen := map[string]bool{}
	for _, photo := range photos {
		if !ValidLodgingPhotoName(photo.Name) || !strings.HasPrefix(photo.Name, "places/"+placeID+"/photos/") || seen[photo.Name] {
			continue
		}
		seen[photo.Name] = true
		result = append(result, photo)
		if len(result) == 5 {
			break
		}
	}
	return result
}

func googleLodgingImageURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Fragment != "" || (u.Port() != "" && u.Port() != "443") {
		return nil, errors.New("Google Places returned an invalid photo URL")
	}
	host := strings.ToLower(u.Hostname())
	if host != "googleusercontent.com" && !strings.HasSuffix(host, ".googleusercontent.com") {
		return nil, errors.New("Google Places returned an invalid photo host")
	}
	return u, nil
}

func readLodgingBytes(response *http.Response, limit int64) ([]byte, error) {
	if response.ContentLength > limit {
		return nil, errors.New("Google Places response exceeds size limit")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, errors.New("unable to read Google Places response")
	}
	if int64(len(data)) > limit {
		return nil, errors.New("Google Places response exceeds size limit")
	}
	return data, nil
}

func (c *client) GetLodgingPhotoContext(ctx context.Context, name string) ([]byte, string, error) {
	if !ValidLodgingPhotoName(name) {
		return nil, "", errors.New("invalid Google photo resource")
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var media struct {
		PhotoURI string `json:"photoUri"`
	}
	// Resolve fresh photo references from Places. Do not persist names or bytes.
	if err := c.lodgingJSON(ctx, http.MethodGet, name+"/media?maxWidthPx=1600&maxHeightPx=1600&skipHttpRedirect=true", "", nil, &media); err != nil {
		return nil, "", err
	}
	u, err := googleLodgingImageURL(media.PhotoURI)
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, "", errors.New("invalid Google photo request")
	}
	req.Header.Set("Accept", "image/jpeg,image/png,image/webp")
	httpClient := &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(next *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("too many Google photo redirects")
		}
		if _, err := googleLodgingImageURL(next.URL.String()); err != nil {
			return err
		}
		next.Header.Del("X-Goog-Api-Key")
		next.Header.Del("Authorization")
		return nil
	}}
	response, err := httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, "", ctx.Err()
		}
		return nil, "", errors.New("Google photo request failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("Google photo returned HTTP %d", response.StatusCode)
	}
	data, err := readLodgingBytes(response, maxLodgingPhotoBytes)
	if err != nil {
		return nil, "", err
	}
	contentType := http.DetectContentType(data)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		return nil, "", errors.New("Google photo response is not a supported image")
	}
	return data, contentType, nil
}
