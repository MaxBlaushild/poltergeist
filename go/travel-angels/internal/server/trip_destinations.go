package server

import (
	"context"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/travel-angels/internal/travelcalendar"
)

// Optional context methods keep the shared Maps client interface compatible
// with other apps while allowing live destination searches to be cancelled.
type destinationMapsClient interface {
	FindCandidatesByQueryContext(context.Context, string) ([]googlemaps.Candidate, error)
	FindPlaceByIDContext(context.Context, string) (*googlemaps.Place, error)
}

func googleDestinationSearch(maps googlemaps.Client) travelcalendar.DestinationSearchProvider {
	return func(ctx context.Context, query string) ([]travelcalendar.DestinationCandidate, error) {
		if maps == nil {
			return nil, nil
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var candidates []googlemaps.Candidate
		var err error
		contextMaps, cancellable := maps.(destinationMapsClient)
		if cancellable {
			candidates, err = contextMaps.FindCandidatesByQueryContext(ctx, query)
		} else {
			candidates, err = maps.FindCandidatesByQuery(query)
		}
		if err != nil {
			return nil, err
		}
		result := make([]travelcalendar.DestinationCandidate, 0, len(candidates))
		seen := map[string]bool{}
		for _, candidate := range candidates {
			if err := ctx.Err(); err != nil {
				return result, err
			}
			if len(result) >= 5 {
				break
			}
			id := strings.TrimSpace(candidate.PlaceID)
			if id == "" || strings.TrimSpace(candidate.Name) == "" || seen[id] {
				continue
			}
			seen[id] = true
			latitude, longitude := candidate.Geometry.Location.Lat, candidate.Geometry.Location.Lng
			place := travelcalendar.DestinationSnapshot{
				Name: strings.TrimSpace(candidate.Name), Kind: destinationKind(candidate.Types),
				FormattedAddress: candidate.FormattedAddress, Latitude: &latitude, Longitude: &longitude,
			}
			var details *googlemaps.Place
			if cancellable {
				details, err = contextMaps.FindPlaceByIDContext(ctx, id)
			} else {
				details, err = maps.FindPlaceByID(id)
			}
			if contextErr := ctx.Err(); contextErr != nil {
				return result, contextErr
			}
			if err == nil && details != nil {
				if len(details.Types) > 0 {
					place.Kind = destinationKind(details.Types)
				}
				// Country identity comes from structured components. See
				// https://developers.google.com/maps/documentation/places/web-service/reference/rest/v1/places#AddressComponent
				for _, component := range details.AddressComponents {
					for _, kind := range component.Types {
						if kind == "country" {
							code := strings.ToUpper(strings.TrimSpace(component.ShortText))
							if len(code) == 2 && code[0] >= 'A' && code[0] <= 'Z' && code[1] >= 'A' && code[1] <= 'Z' {
								place.CountryCode = code
								place.CountryName = strings.TrimSpace(component.LongText)
							}
						}
					}
				}
				// A large region or country may span several time zones; a point
				// within its geometry is not a reliable default for the entire area.
				if details.TimeZone != nil && (place.Kind == "city" || place.Kind == "place") {
					zone := strings.TrimSpace(details.TimeZone.ID)
					if _, zoneErr := time.LoadLocation(zone); zoneErr == nil && zone != "" && zone != "Local" {
						place.TimeZone = zone
					}
				}
			}
			result = append(result, travelcalendar.DestinationCandidate{DestinationSnapshot: place, Provider: "google", ProviderID: id})
		}
		return result, nil
	}
}

func destinationKind(types []string) string {
	for _, candidate := range []struct{ source, kind string }{
		{"country", "country"}, {"locality", "city"}, {"postal_town", "city"},
		{"administrative_area_level_1", "region"}, {"administrative_area_level_2", "region"},
	} {
		for _, kind := range types {
			if kind == candidate.source {
				return candidate.kind
			}
		}
	}
	return "place"
}
