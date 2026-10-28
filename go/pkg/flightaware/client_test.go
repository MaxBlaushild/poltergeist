package flightaware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-apikey") != "private-test-key" || strings.Contains(r.URL.String(), "private-test-key") {
			t.Errorf("wrong key transport")
		}
		handler(w, r)
	}))
	t.Cleanup(s.Close)
	c := New("private-test-key")
	c.baseURL = s.URL
	c.now = func() time.Time { return time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC) }
	return c
}

func testAirports(w http.ResponseWriter, r *http.Request) bool {
	var data string
	switch r.URL.Path {
	case "/airports/KJFK", "/airports/JFK":
		data = `{"airport_code":"KJFK","code_icao":"KJFK","code_iata":"JFK","timezone":"America/New_York"}`
	case "/airports/RJTT", "/airports/HND":
		data = `{"airport_code":"RJTT","code_icao":"RJTT","code_iata":"HND","timezone":"Asia/Tokyo"}`
	case "/airports/NZAA", "/airports/AKL":
		data = `{"airport_code":"NZAA","code_icao":"NZAA","code_iata":"AKL","timezone":"Pacific/Auckland"}`
	case "/airports/PHNL", "/airports/HNL":
		data = `{"airport_code":"PHNL","code_icao":"PHNL","code_iata":"HNL","timezone":"Pacific/Honolulu"}`
	default:
		return false
	}
	w.Write([]byte(data))
	return true
}

func japanFlight() scheduledFlight {
	return scheduledFlight{Ident: "ANA109", IdentICAO: "ANA109", IdentIATA: "NH109", Origin: "KJFK", OriginICAO: "KJFK", OriginIATA: "JFK", Destination: "RJTT", DestinationICAO: "RJTT", DestinationIATA: "HND", ScheduledOut: "2027-04-10T17:00:00Z", ScheduledIn: "2027-04-11T07:00:00Z"}
}
func writeFlights(w http.ResponseWriter, rows ...scheduledFlight) {
	if rows == nil {
		rows = []scheduledFlight{}
	}
	json.NewEncoder(w).Encode(map[string]any{"scheduled": rows, "links": nil})
}

func TestFutureSchedulesUseLocalDatesAndCodeshareIdentity(t *testing.T) {
	var schedules, airports atomic.Int32
	operating := japanFlight()
	marketing := operating
	marketing.Ident = "UAL7909"
	marketing.IdentICAO = "UAL7909"
	marketing.IdentIATA = "UA7909"
	marketing.ActualIdent = "ANA109"
	marketing.ActualIdentICAO = "ANA109"
	marketing.ActualIdentIATA = "NH109"
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if testAirports(w, r) {
			airports.Add(1)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/schedules/") {
			t.Errorf("unexpected endpoint: %s", r.URL.Path)
		}
		schedules.Add(1)
		if r.URL.Query().Get("max_pages") != "3" || r.URL.Query().Get("include_codeshares") != "true" {
			t.Errorf("missing bounded schedule parameters")
		}
		if !strings.Contains(r.URL.Path, "2027-04-10T04:00:00Z/2027-04-11T04:00:00Z") {
			t.Errorf("not departure-local day: %s", r.URL.Path)
		}
		writeFlights(w, marketing, operating)
	})
	flights, err := c.Search(context.Background(), Query{Origin: "jfk", Destination: "HND", DepartureDate: "2027-04-10"})
	if err != nil || len(flights) != 1 {
		t.Fatalf("result: %+v %v", flights, err)
	}
	f := flights[0]
	if f.FlightNumber != "NH109" || f.OperatingFlightNumber != "NH109" || f.DepartureTime != "13:00" || f.ArrivalDate != "2027-04-11" || f.ArrivalTime != "16:00" || f.DepartureTimeZone != "America/New_York" || f.ArrivalTimeZone != "Asia/Tokyo" {
		t.Fatalf("wrong local flight: %+v", f)
	}
	if schedules.Load() != 1 || airports.Load() != 2 {
		t.Fatalf("unbounded/repeated requests: %d/%d", schedules.Load(), airports.Load())
	}
}

func TestFlightNumberCodeshareAndLeadingZeroNormalization(t *testing.T) {
	row := japanFlight()
	row.Ident = "UAL7909"
	row.IdentICAO = "UAL7909"
	row.IdentIATA = "UA7909"
	row.ActualIdentICAO = "ANA109"
	row.ActualIdentIATA = "NH109"
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if testAirports(w, r) {
			return
		}
		if r.URL.Query().Get("airline") != "UA" || r.URL.Query().Get("flight_number") != "7909" {
			t.Errorf("bad airline/number split")
		}
		writeFlights(w, row, japanFlight())
	})
	flights, err := c.Search(context.Background(), Query{FlightNumber: "ua 7909", DepartureDate: "2027-04-10"})
	if err != nil || len(flights) != 1 || flights[0].FlightNumber != "UA7909" || flights[0].OperatingFlightNumber != "NH109" {
		t.Fatalf("wrong marketed flight: %+v %v", flights, err)
	}
	if canonicalIdent("NH0109") != "NH109" {
		t.Fatal("leading zero normalization")
	}
}

func TestInternationalDateLineArrivesPreviousLocalDay(t *testing.T) {
	row := scheduledFlight{Ident: "ANZ10", IdentICAO: "ANZ10", IdentIATA: "NZ10", OriginICAO: "NZAA", DestinationICAO: "PHNL", ScheduledOut: "2027-04-09T22:00:00Z", ScheduledIn: "2027-04-10T07:00:00Z"}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if testAirports(w, r) {
			return
		}
		writeFlights(w, row)
	})
	flights, err := c.Search(context.Background(), Query{FlightNumber: "NZ10", DepartureDate: "2027-04-10"})
	if err != nil || len(flights) != 1 {
		t.Fatalf("%+v %v", flights, err)
	}
	if flights[0].DepartureDate != "2027-04-10" || flights[0].ArrivalDate != "2027-04-09" || flights[0].ArrivalTime != "21:00" {
		t.Fatalf("date line conversion: %+v", flights[0])
	}
}

func TestDSTLocalDepartureWindowUsesCalendarDay(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if testAirports(w, r) {
			return
		}
		if !strings.Contains(r.URL.Path, "2027-03-14T05:00:00Z/2027-03-15T04:00:00Z") {
			t.Errorf("DST window: %s", r.URL.Path)
		}
		writeFlights(w)
	})
	_, err := c.Search(context.Background(), Query{Origin: "JFK", Destination: "HND", DepartureDate: "2027-03-14"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWrongLocalDateAndWrongNumberAreExcluded(t *testing.T) {
	row := japanFlight()
	row.ScheduledOut = "2027-04-10T02:00:00Z"
	row.ScheduledIn = "2027-04-10T16:00:00Z"
	unrelated := japanFlight()
	unrelated.Ident = "ANA110"
	unrelated.IdentICAO = "ANA110"
	unrelated.IdentIATA = "NH110"
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if testAirports(w, r) {
			return
		}
		writeFlights(w, row, unrelated)
	})
	flights, err := c.Search(context.Background(), Query{FlightNumber: "NH109", DepartureDate: "2027-04-10"})
	if err != nil || len(flights) != 0 {
		t.Fatalf("wrong date/ident returned: %+v %v", flights, err)
	}
}

func TestNoSchedulesAndMalformedProviderResponses(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       error
	}{{"empty", `{"scheduled":[],"links":null}`, nil}, {"missing", `{}`, ErrInvalidData}, {"null", `{"scheduled":null}`, ErrInvalidData}, {"badjson", `not json private-test-key`, ErrInvalidData}, {"pagination", `{"scheduled":[],"links":{"next":"https://untrusted.invalid/private-test-key"}}`, ErrNarrowSearch}, {"too-large", strings.Repeat("x", responseLimit+1), ErrInvalidData}} {
		t.Run(tc.name, func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(tc.body)) })
			_, err := c.Search(context.Background(), Query{FlightNumber: "NH109", DepartureDate: "2027-04-10"})
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v got %v", tc.want, err)
			}
		})
	}
}

func TestInvalidTimezoneAndArrivalAreNeverGuessed(t *testing.T) {
	for _, badZone := range []string{"", "Mars/City", "Local"} {
		t.Run("zone-"+badZone, func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/airports/") {
					json.NewEncoder(w).Encode(map[string]string{"airport_code": "KJFK", "timezone": badZone})
					return
				}
				writeFlights(w, japanFlight())
			})
			_, err := c.Search(context.Background(), Query{FlightNumber: "NH109", DepartureDate: "2027-04-10"})
			if !errors.Is(err, ErrInvalidData) {
				t.Fatal(err)
			}
		})
	}
	row := japanFlight()
	row.ScheduledIn = row.ScheduledOut
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { writeFlights(w, row) })
	if _, err := c.Search(context.Background(), Query{FlightNumber: "NH109", DepartureDate: "2027-04-10"}); !errors.Is(err, ErrInvalidData) {
		t.Fatal(err)
	}
}

func TestProviderErrorsAreRedactedAndRedirectsDoNotSendKey(t *testing.T) {
	for _, tc := range []struct {
		status int
		want   error
	}{{401, ErrAccess}, {403, ErrAccess}, {402, ErrAccess}, {429, ErrRateLimited}, {500, ErrUnavailable}, {302, ErrUnavailable}} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			var hits atomic.Int32
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				w.Header().Set("Location", "/leaked-key")
				w.WriteHeader(tc.status)
				w.Write([]byte(`private-test-key upstream stacktrace`))
			})
			_, err := c.Search(context.Background(), Query{FlightNumber: "NH109", DepartureDate: "2027-04-10"})
			if !errors.Is(err, tc.want) || strings.Contains(err.Error(), "private-test-key") || hits.Load() != 1 {
				t.Fatalf("unsafe error/redirect: %v (%d)", err, hits.Load())
			}
		})
	}
}

func TestSplitFlightNumberAndInputValidation(t *testing.T) {
	for _, tc := range []struct{ input, airline, number string }{{"B6 123", "B6", "123"}, {"2K8373", "2K", "8373"}, {"D84331", "D8", "4331"}, {"UAL123", "UAL", "123"}, {"UA123A", "UA", "123A"}} {
		a, n, ok := SplitFlightNumber(tc.input)
		if !ok || a != tc.airline || n != tc.number {
			t.Errorf("split %s => %s/%s", tc.input, a, n)
		}
	}
	for _, input := range []string{"1234", "../../UA123", "U1", "UNITED123", "UA12345"} {
		if _, _, ok := SplitFlightNumber(input); ok {
			t.Errorf("accepted %q", input)
		}
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Fatal("invalid input made paid request") })
	for _, query := range []Query{{FlightNumber: "UA123", DepartureDate: "2027-02-30"}, {FlightNumber: "UA123", DepartureDate: "2028-04-10"}, {FlightNumber: "1234", DepartureDate: "2027-04-10"}, {Origin: "JFK", DepartureDate: "2027-04-10"}, {Origin: "../", Destination: "HND", DepartureDate: "2027-04-10"}} {
		if _, err := c.Search(context.Background(), query); err == nil {
			t.Errorf("accepted %+v", query)
		}
	}
	if New("  ") != nil {
		t.Fatal("empty key configured provider")
	}
}

func TestCancellationAndNoClientLeak(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Fatal("cancelled context made request") })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Search(ctx, Query{FlightNumber: "NH109", DepartureDate: "2027-04-10"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestOperatingIdentityIncludesAirportsAndScheduledInstant(t *testing.T) {
	first := japanFlight()
	second := first
	second.ScheduledOut = "2027-04-10T18:00:00Z"
	second.ScheduledIn = "2027-04-11T08:00:00Z"
	third := first
	third.Destination, third.DestinationICAO, third.DestinationIATA = "PHNL", "PHNL", "HNL"
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if testAirports(w, r) {
			return
		}
		writeFlights(w, first, second, third)
	})
	flights, err := c.Search(context.Background(), Query{FlightNumber: "NH109", DepartureDate: "2027-04-10"})
	if err != nil || len(flights) != 3 {
		t.Fatalf("different scheduled operations merged: %+v %v", flights, err)
	}
	ids := map[string]bool{}
	for _, flight := range flights {
		ids[flight.ProviderID] = true
	}
	if len(ids) != 3 {
		t.Fatal("operating identity lost airport or departure timestamp")
	}
}

func TestLookupCapsPaidAirportEnrichment(t *testing.T) {
	var calls atomic.Int32
	rows := []scheduledFlight{}
	for _, airport := range []string{"KAAA", "KAAB", "KAAC", "KAAD", "KAAE", "KAAF", "KAAG", "KAAH", "KAAI"} {
		row := japanFlight()
		row.Origin, row.OriginICAO, row.OriginIATA = airport, airport, ""
		rows = append(rows, row)
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/airports/") {
			calls.Add(1)
			code := strings.TrimPrefix(r.URL.Path, "/airports/")
			json.NewEncoder(w).Encode(map[string]string{"airport_code": code, "code_icao": code, "timezone": "UTC"})
			return
		}
		writeFlights(w, rows...)
	})
	_, err := c.Search(context.Background(), Query{FlightNumber: "NH109", DepartureDate: "2027-04-10"})
	if !errors.Is(err, ErrNarrowSearch) || calls.Load() != 8 {
		t.Fatalf("unbounded enrichment: %d %v", calls.Load(), err)
	}
}
