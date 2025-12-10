package googlemaps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type RankPreference string

const (
	RankPreferenceDistance   RankPreference = "DISTANCE"
	RankPreferencePopularity RankPreference = "POPULARITY"
)

const (
	baseURL         = "https://places.googleapis.com/v1/places:searchNearby"
	placeDetailsURL = "https://places.googleapis.com/v1/places"
	placeSearchURL  = "https://maps.googleapis.com/maps/api/place/findplacefromtext/json"
)

type client struct {
	apiKey string
}

type Client interface {
	FindPlaces(query PlaceQuery) ([]Place, error)
	FindPlaceByID(id string) (*Place, error)
	FindCandidatesByQuery(query string) ([]Candidate, error)
}

func NewClient(apiKey string) Client {
	return &client{apiKey: apiKey}
}

type CandidateResponse struct {
	Candidates   []Candidate `json:"candidates"`
	Status       string      `json:"status"`
	ErrorMessage string      `json:"error_message,omitempty"`
}

type GooglePlacesResponse struct {
	Places []Place `json:"places"`
}

type PlaceQuery struct {
	Lat            float64
	Long           float64
	Radius         float64
	MaxResultCount int32
	IncludedTypes  []PlaceType
	ExcludedTypes  []PlaceType
	RankPreference RankPreference
}

type searchNearbyRequest struct {
	LocationRestriction struct {
		Circle struct {
			Center struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"center"`
			Radius float64 `json:"radius"`
		} `json:"circle"`
	} `json:"locationRestriction"`
	IncludedTypes  []PlaceType    `json:"includedTypes,omitempty"`
	ExcludedTypes  []PlaceType    `json:"excludedTypes,omitempty"`
	MaxResultCount int32          `json:"maxResultCount,omitempty"`
	RankPreference RankPreference `json:"rankPreference,omitempty"`
}

func (c *client) FindCandidatesByQuery(query string) ([]Candidate, error) {
	// Create URL with query parameters
	url := fmt.Sprintf("%s?input=%s&inputtype=textquery&fields=place_id,name,formatted_address,geometry,types,photos,opening_hours&key=%s",
		placeSearchURL,
		url.QueryEscape(query),
		c.apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Log the raw response for debugging
	log.Printf("Google Places API response status: %d", resp.StatusCode)
	log.Printf("Google Places API response body: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google places API returned status %d: %s", resp.StatusCode, string(body))
	}

	var data CandidateResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding response: %w, body: %s", err, string(body))
	}

	// Check API response status
	if data.Status != "OK" && data.Status != "ZERO_RESULTS" {
		return nil, fmt.Errorf("google places API error: status=%s, error_message=%s", data.Status, data.ErrorMessage)
	}

	if data.Status == "ZERO_RESULTS" {
		log.Printf("Google Places API returned ZERO_RESULTS for query: %s", query)
		return []Candidate{}, nil
	}

	log.Printf("Google Places API returned %d candidates for query: %s", len(data.Candidates), query)
	return data.Candidates, nil
}

func (c *client) FindPlaceByID(id string) (*Place, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", placeDetailsURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask", "*")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var place Place
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&place); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &place, nil
}

func (c *client) FindPlaces(query PlaceQuery) ([]Place, error) {
	reqBody := searchNearbyRequest{}
	reqBody.LocationRestriction.Circle.Center.Latitude = query.Lat
	reqBody.LocationRestriction.Circle.Center.Longitude = query.Long
	reqBody.LocationRestriction.Circle.Radius = query.Radius
	reqBody.IncludedTypes = query.IncludedTypes
	reqBody.ExcludedTypes = query.ExcludedTypes
	reqBody.MaxResultCount = query.MaxResultCount
	reqBody.RankPreference = RankPreferenceDistance

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	log.Printf("Request body: %s", string(jsonBody))

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask", "*")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var data GooglePlacesResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(data.Places) == 0 {
		fmt.Println(string(body))
		fmt.Printf("%+v", query)
		fmt.Printf("%+v", reqBody)
		fmt.Println("--------------------------------")
	}

	return data.Places, nil
}
