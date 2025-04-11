package googlemaps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	baseURL = "https://maps.googleapis.com/maps/api/place/nearbysearch/json"

	TypeGroceryStore = "grocery_or_supermarket"
)

type client struct {
	apiKey string
}

type Client interface {
	FindPlaces(query PlaceQuery) ([]Place, error)
}

func NewClient(apiKey string) Client {
	return &client{apiKey: apiKey}
}

type Place struct {
	Name     string `json:"name"`
	Vicinity string `json:"vicinity"`
}

type GooglePlacesResponse struct {
	Results []Place `json:"results"`
}

type PlaceQuery struct {
	Lat      float64
	Long     float64
	Category string
	Radius   int
}

func (c *client) FindPlaces(query PlaceQuery) ([]Place, error) {
	params := url.Values{}
	params.Add("location", fmt.Sprintf("%f,%f", query.Lat, query.Long))
	params.Add("radius", fmt.Sprintf("%d", query.Radius))
	params.Add("type", query.Category)
	params.Add("key", c.apiKey)

	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data GooglePlacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Results, nil
}
