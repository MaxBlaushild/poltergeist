package server

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/flightaware"
	"github.com/MaxBlaushild/poltergeist/travel-angels/internal/travelcalendar"
)

type flightScheduleStub struct {
	query   flightaware.Query
	flights []flightaware.Flight
	err     error
}

func (s *flightScheduleStub) Search(_ context.Context, q flightaware.Query) ([]flightaware.Flight, error) {
	s.query = q
	return s.flights, s.err
}

func TestFlightAwareAdapterPreservesScheduleAndOperatingIdentity(t *testing.T) {
	stub := &flightScheduleStub{flights: []flightaware.Flight{{
		ProviderID: "schedule-operating-flight", Airline: "UA", FlightNumber: "UA7909", OperatingAirline: "NH", OperatingFlightNumber: "NH109",
		DepartureAirport: "JFK", ArrivalAirport: "HND", DepartureDate: "2027-04-10", DepartureTime: "13:00", DepartureTimeZone: "America/New_York",
		ArrivalDate: "2027-04-11", ArrivalTime: "16:00", ArrivalTimeZone: "Asia/Tokyo",
	}}}
	query := travelcalendar.FlightSearchQuery{FlightNumber: "UA7909", DepartureDate: "2027-04-10", Origin: "JFK", Destination: "HND"}
	legs, err := (&flightAwareAdapter{client: stub}).Search(context.Background(), query)
	if err != nil || len(legs) != 1 {
		t.Fatalf("%+v %v", legs, err)
	}
	want := travelcalendar.TripFlightLeg{Provider: "flightaware", ProviderID: "schedule-operating-flight", Airline: "UA", FlightNumber: "UA7909", OperatingAirline: "NH", OperatingFlightNumber: "NH109", DepartureAirport: "JFK", ArrivalAirport: "HND", DepartureDate: "2027-04-10", DepartureTime: "13:00", DepartureTimeZone: "America/New_York", ArrivalDate: "2027-04-11", ArrivalTime: "16:00", ArrivalTimeZone: "Asia/Tokyo"}
	if !reflect.DeepEqual(legs[0], want) {
		t.Fatalf("mapped %+v", legs[0])
	}
	if stub.query.FlightNumber != query.FlightNumber || stub.query.DepartureDate != query.DepartureDate || stub.query.Origin != query.Origin || stub.query.Destination != query.Destination {
		t.Fatalf("query lost: %+v", stub.query)
	}
}

func TestFlightAwareAdapterUnconfiguredEmptyAndFailure(t *testing.T) {
	if flightAwareSearchProvider(" \n ") != nil {
		t.Fatal("empty env configured paid provider")
	}
	if flightAwareSearchProvider("test-key") == nil {
		t.Fatal("configured provider missing")
	}
	stub := &flightScheduleStub{}
	legs, err := (&flightAwareAdapter{client: stub}).Search(context.Background(), travelcalendar.FlightSearchQuery{})
	if err != nil || legs == nil || len(legs) != 0 {
		t.Fatalf("empty result %+v %v", legs, err)
	}
	stub.err = flightaware.ErrRateLimited
	if _, err := (&flightAwareAdapter{client: stub}).Search(context.Background(), travelcalendar.FlightSearchQuery{}); !errors.Is(err, flightaware.ErrRateLimited) {
		t.Fatal(err)
	}
}

func TestFlightAwareAdapterExposesOnlyKnownSafeMessages(t *testing.T) {
	stub := &flightScheduleStub{err: flightaware.ErrNarrowSearch}
	_, err := (&flightAwareAdapter{client: stub}).Search(context.Background(), travelcalendar.FlightSearchQuery{})
	var safe interface{ PublicMessage() string }
	if !errors.As(err, &safe) || safe.PublicMessage() != flightaware.ErrNarrowSearch.Error() {
		t.Fatal("known static guidance not exposed")
	}
	stub.err = errors.New("private key or upstream data")
	_, err = (&flightAwareAdapter{client: stub}).Search(context.Background(), travelcalendar.FlightSearchQuery{})
	if errors.As(err, &safe) {
		t.Fatal("untrusted error marked public")
	}
}
