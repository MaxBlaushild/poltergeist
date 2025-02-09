package mapbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type client struct {
	apiKey string
}

type Client interface {
	GetPlaces(ctx context.Context, address string) ([]Place, error)
}

func NewClient(apiKey string) Client {
	return &client{apiKey: apiKey}
}

func (c *client) GetPlaces(ctx context.Context, address string) ([]Place, error) {
	// Create the request URL
	baseURL := "https://api.mapbox.com/geocoding/v5/mapbox.places/"
	url := fmt.Sprintf("%s%s.json?access_token=%s", baseURL, address, c.apiKey)

	// Make the HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var geocodeResponse GeocodeResponse
	if err := json.Unmarshal(body, &geocodeResponse); err != nil {
		return nil, err
	}

	// Check if we have any results
	if len(geocodeResponse.Features) == 0 {
		return nil, nil
	}

	// Convert each feature into a Place
	places := make([]Place, len(geocodeResponse.Features))
	for i, feature := range geocodeResponse.Features {
		places[i] = Place{
			Name: feature.PlaceName,
			LatLong: LatLong{
				Lat: feature.Geometry.Coordinates[1],
				Lng: feature.Geometry.Coordinates[0],
			},
		}
	}

	return places, nil
}
