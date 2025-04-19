package googlemaps

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
