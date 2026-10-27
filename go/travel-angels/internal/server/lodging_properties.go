package server

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/travel-angels/internal/travelcalendar"
)

type googleLodgingAdapter struct{ client googlemaps.LodgingClient }

func googleLodgingPropertyProvider(maps googlemaps.Client) travelcalendar.LodgingPropertyProvider {
	provider, ok := maps.(googlemaps.LodgingClient)
	if !ok {
		return nil
	}
	return &googleLodgingAdapter{client: provider}
}

func (p *googleLodgingAdapter) Search(ctx context.Context, query string) ([]travelcalendar.LodgingProperty, error) {
	places, err := p.client.SearchLodgingContext(ctx, query)
	return lodgingPropertyRows(places, err)
}

func (p *googleLodgingAdapter) SearchNear(ctx context.Context, query string, latitude, longitude float64) ([]travelcalendar.LodgingProperty, error) {
	if !googlemaps.ValidLodgingCoordinates(latitude, longitude) {
		return nil, errors.New("invalid property search coordinates")
	}
	if nearby, ok := p.client.(googlemaps.LodgingNearClient); ok {
		places, err := nearby.SearchLodgingNearContext(ctx, query, latitude, longitude)
		return lodgingPropertyRows(places, err)
	}
	return p.Search(ctx, query)
}

func lodgingPropertyRows(places []googlemaps.LodgingPlace, err error) ([]travelcalendar.LodgingProperty, error) {
	if err != nil {
		return nil, err
	}
	rows := make([]travelcalendar.LodgingProperty, 0, len(places))
	for _, place := range places {
		rows = append(rows, lodgingPropertyDTO(place))
	}
	return rows, nil
}

func (p *googleLodgingAdapter) Details(ctx context.Context, id string) (*travelcalendar.LodgingProperty, error) {
	place, err := p.client.GetLodgingContext(ctx, id)
	if err != nil {
		return nil, err
	}
	if place == nil || place.ID != id {
		return nil, errors.New("this property is unavailable")
	}
	row := lodgingPropertyDTO(*place)
	return &row, nil
}

func (p *googleLodgingAdapter) Photo(ctx context.Context, name string) ([]byte, string, error) {
	return p.client.GetLodgingPhotoContext(ctx, name)
}

// Attribution links are supplied by Google, sometimes as scheme-relative URLs.
// Keep them clickable without accepting executable schemes or credentials.
func lodgingSourceURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		return ""
	}
	return u.String()
}

func lodgingPropertyDTO(place googlemaps.LodgingPlace) travelcalendar.LodgingProperty {
	row := travelcalendar.LodgingProperty{
		Provider: "google", PlaceID: place.ID, Name: place.DisplayName.Text,
		Address: place.FormattedAddress, WebsiteURL: lodgingSourceURL(place.WebsiteURI), MapsURL: lodgingSourceURL(place.GoogleMapsURI),
		Rating: place.Rating, RatingCount: place.UserRatingCount, Types: append([]string{}, place.Types...),
		Photos: []travelcalendar.LodgingPropertyPhoto{}, Attributions: []travelcalendar.LodgingPhotoAttribution{},
	}
	for _, attribution := range place.Attributions {
		row.Attributions = append(row.Attributions, travelcalendar.LodgingPhotoAttribution{DisplayName: attribution.Provider, URI: lodgingSourceURL(attribution.ProviderURI)})
	}
	for _, photo := range place.Photos {
		if !googlemaps.ValidLodgingPhotoName(photo.Name) || !strings.HasPrefix(photo.Name, "places/"+place.ID+"/photos/") {
			continue
		}
		mapped := travelcalendar.LodgingPropertyPhoto{
			Name: photo.Name, Width: photo.Width, Height: photo.Height,
			GoogleMapsURI: lodgingSourceURL(photo.GoogleMapsURI), Attributions: []travelcalendar.LodgingPhotoAttribution{},
		}
		for _, attribution := range photo.AuthorAttributions {
			mapped.Attributions = append(mapped.Attributions, travelcalendar.LodgingPhotoAttribution{
				DisplayName: attribution.DisplayName, URI: lodgingSourceURL(attribution.URI), PhotoURI: lodgingSourceURL(attribution.PhotoURI),
			})
		}
		// The route layer inserts authorized API URLs; provider resource names
		// and image download URLs are never client-facing photo URLs.
		row.Photos = append(row.Photos, mapped)
		if len(row.Photos) == 5 {
			break
		}
	}
	return row
}
