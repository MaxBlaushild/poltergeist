package server

import (
	"context"
	"errors"

	"github.com/MaxBlaushild/poltergeist/pkg/flightaware"
	"github.com/MaxBlaushild/poltergeist/travel-angels/internal/travelcalendar"
)

type flightScheduleClient interface {
	Search(context.Context, flightaware.Query) ([]flightaware.Flight, error)
}

type flightAwareAdapter struct{ client flightScheduleClient }

// Only known, static provider messages may cross the HTTP boundary. Transport
// errors and future provider implementations retain the generic route error.
type flightAwareSearchError struct{ error }

func (e flightAwareSearchError) PublicMessage() string { return e.error.Error() }
func (e flightAwareSearchError) Unwrap() error         { return e.error }

func flightAwareSearchProvider(key string) travelcalendar.FlightSearchProvider {
	client := flightaware.New(key)
	if client == nil {
		return nil
	}
	return &flightAwareAdapter{client: client}
}

func (p *flightAwareAdapter) Search(ctx context.Context, query travelcalendar.FlightSearchQuery) ([]travelcalendar.TripFlightLeg, error) {
	flights, err := p.client.Search(ctx, flightaware.Query{
		FlightNumber: query.FlightNumber, DepartureDate: query.DepartureDate,
		Origin: query.Origin, Destination: query.Destination,
	})
	if err != nil {
		for _, safe := range []error{flightaware.ErrUnavailable, flightaware.ErrAccess, flightaware.ErrRateLimited, flightaware.ErrInvalidData, flightaware.ErrNarrowSearch, flightaware.ErrScheduleWindow} {
			if errors.Is(err, safe) {
				return nil, flightAwareSearchError{safe}
			}
		}
		return nil, err
	}
	legs := make([]travelcalendar.TripFlightLeg, 0, len(flights))
	for _, flight := range flights {
		legs = append(legs, travelcalendar.TripFlightLeg{
			Provider: "flightaware", ProviderID: flight.ProviderID,
			Airline: flight.Airline, FlightNumber: flight.FlightNumber,
			OperatingAirline: flight.OperatingAirline, OperatingFlightNumber: flight.OperatingFlightNumber,
			DepartureAirport: flight.DepartureAirport, ArrivalAirport: flight.ArrivalAirport,
			DepartureDate: flight.DepartureDate, DepartureTime: flight.DepartureTime, DepartureTimeZone: flight.DepartureTimeZone,
			ArrivalDate: flight.ArrivalDate, ArrivalTime: flight.ArrivalTime, ArrivalTimeZone: flight.ArrivalTimeZone,
		})
	}
	return legs, nil
}
