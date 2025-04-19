package googlemaps

type Candidate struct {
	PlaceID          string `json:"place_id"`
	Name             string `json:"name"`
	FormattedAddress string `json:"formatted_address"`
	Geometry         struct {
		Location struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"location"`
		Viewport struct {
			Northeast struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"northeast"`
			Southwest struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"southwest"`
		} `json:"viewport"`
	} `json:"geometry"`
	Types  []string `json:"types"`
	Photos []struct {
		Height           int      `json:"height"`
		Width            int      `json:"width"`
		PhotoReference   string   `json:"photo_reference"`
		HTMLAttributions []string `json:"html_attributions"`
	} `json:"photos"`
	OpeningHours struct {
		OpenNow     bool     `json:"open_now"`
		WeekdayText []string `json:"weekday_text"`
	} `json:"opening_hours"`
}
