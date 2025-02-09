package mapbox

type LatLong struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Place struct {
	Name    string  `json:"name"`
	LatLong LatLong `json:"latLong"`
}

type GeocodeResponse struct {
	Type     string    `json:"type"`
	Query    []string  `json:"query"`
	Features []Feature `json:"features"`
}

type Feature struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Geometry   Geometry `json:"geometry"`
	Properties struct {
		Accuracy   string `json:"accuracy"`
		Foursquare string `json:"foursquare"`
	} `json:"properties"`
	PlaceName string    `json:"place_name"`
	Text      string    `json:"text"`
	PlaceType []string  `json:"place_type"`
	Context   []Context `json:"context"`
}

type Geometry struct {
	Coordinates []float64 `json:"coordinates"`
}

type Context struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Wikidata string `json:"wikidata"`
}
