package googlemaps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL = "https://places.googleapis.com/v1/places:searchNearby"

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
	Name                     string        `json:"name,omitempty"`
	ID                       string        `json:"id,omitempty"`
	DisplayName              LocalizedText `json:"displayName,omitempty"`
	Types                    []string      `json:"types,omitempty"`
	PrimaryType              string        `json:"primaryType,omitempty"`
	PrimaryTypeDisplayName   LocalizedText `json:"primaryTypeDisplayName,omitempty"`
	NationalPhoneNumber      string        `json:"nationalPhoneNumber,omitempty"`
	InternationalPhoneNumber string        `json:"internationalPhoneNumber,omitempty"`
	FormattedAddress         string        `json:"formattedAddress,omitempty"`
	Location                 struct {
		Latitude  float64 `json:"latitude,omitempty"`
		Longitude float64 `json:"longitude,omitempty"`
	} `json:"location,omitempty"`
	Rating                float64       `json:"rating,omitempty"`
	UtcOffsetMinutes      *int32        `json:"utcOffsetMinutes,omitempty"`
	BusinessStatus        string        `json:"businessStatus,omitempty"`
	PriceLevel            string        `json:"priceLevel,omitempty"`
	UserRatingCount       *int32        `json:"userRatingCount,omitempty"`
	Takeout               *bool         `json:"takeout,omitempty"`
	Delivery              *bool         `json:"delivery,omitempty"`
	DineIn                *bool         `json:"dineIn,omitempty"`
	CurbsidePickup        *bool         `json:"curbsidePickup,omitempty"`
	Reservable            *bool         `json:"reservable,omitempty"`
	ServesBreakfast       *bool         `json:"servesBreakfast,omitempty"`
	ServesLunch           *bool         `json:"servesLunch,omitempty"`
	ServesDinner          *bool         `json:"servesDinner,omitempty"`
	ServesBeer            *bool         `json:"servesBeer,omitempty"`
	ServesWine            *bool         `json:"servesWine,omitempty"`
	ServesBrunch          *bool         `json:"servesBrunch,omitempty"`
	ServesVegetarianFood  *bool         `json:"servesVegetarianFood,omitempty"`
	EditorialSummary      LocalizedText `json:"editorialSummary,omitempty"`
	OutdoorSeating        *bool         `json:"outdoorSeating,omitempty"`
	LiveMusic             *bool         `json:"liveMusic,omitempty"`
	MenuForChildren       *bool         `json:"menuForChildren,omitempty"`
	ServesCocktails       *bool         `json:"servesCocktails,omitempty"`
	ServesDessert         *bool         `json:"servesDessert,omitempty"`
	ServesCoffee          *bool         `json:"servesCoffee,omitempty"`
	GoodForChildren       *bool         `json:"goodForChildren,omitempty"`
	AllowsDogs            *bool         `json:"allowsDogs,omitempty"`
	Restroom              *bool         `json:"restroom,omitempty"`
	GoodForGroups         *bool         `json:"goodForGroups,omitempty"`
	GoodForWatchingSports *bool         `json:"goodForWatchingSports,omitempty"`
	PlaceId               string        `json:"placeId,omitempty"`
}

type Photo struct {
	Uri           string `json:"uri,omitempty"`
	PhotoUri      string `json:"photoUri,omitempty"`
	HeightPx      int    `json:"heightPx,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
	GoogleMapsUri string `json:"googleMapsUri,omitempty"`
}

type Review struct {
	Text                   string `json:"text,omitempty"`
	LanguageCode           string `json:"languageCode,omitempty"`
	OverviewFlagContentUri string `json:"overviewFlagContentUri,omitempty"`
}

type LocalizedText struct {
	Text         string `json:"text,omitempty"`
	LanguageCode string `json:"languageCode,omitempty"`
}

// End of Selection

type GooglePlacesResponse struct {
	Places []Place `json:"places"`
}

type PlaceQuery struct {
	Lat            float64
	Long           float64
	Category       string
	Radius         float64
	MaxResultCount int32
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
	IncludedTypes  []string `json:"includedTypes,omitempty"`
	MaxResultCount int32    `json:"maxResultCount,omitempty"`
}

func (c *client) FindPlaces(query PlaceQuery) ([]Place, error) {
	reqBody := searchNearbyRequest{}
	reqBody.LocationRestriction.Circle.Center.Latitude = query.Lat
	reqBody.LocationRestriction.Circle.Center.Longitude = query.Long
	reqBody.LocationRestriction.Circle.Radius = query.Radius
	reqBody.IncludedTypes = []string{query.Category}
	reqBody.MaxResultCount = query.MaxResultCount

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

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

	return data.Places, nil
}
