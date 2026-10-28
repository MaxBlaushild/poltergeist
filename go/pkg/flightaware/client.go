// Package flightaware looks up published airline timetables in AeroAPI v4.
// It deliberately uses /schedules, not the live /flights endpoint whose future
// horizon is much shorter. Schedule timestamps are UTC; dates entered by users
// are interpreted in the departure airport's IANA timezone.
package flightaware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // Standalone containers need the same IANA data as core.
)

const apiBase = "https://aeroapi.flightaware.com/aeroapi"
const responseLimit = 2 << 20

var (
	ErrUnavailable    = errors.New("Flight search is temporarily unavailable. Enter the flight manually.")
	ErrAccess         = errors.New("Flight search is not configured for schedule access. Enter the flight manually.")
	ErrRateLimited    = errors.New("Flight search is busy. Try again shortly or enter the flight manually.")
	ErrInvalidData    = errors.New("This flight's schedule or airport timezone could not be verified. Enter the flight manually.")
	ErrNarrowSearch   = errors.New("Too many flights match. Add a flight number or departure and arrival airports.")
	ErrScheduleWindow = errors.New("Published schedules are available up to one year ahead and three months back. Enter this flight manually.")
	airportPattern    = regexp.MustCompile(`^[A-Z]{3,4}$`)
	icaoFlightPattern = regexp.MustCompile(`^([A-Z]{3})([0-9]{1,4}[A-Z]?)$`)
	iataFlightPattern = regexp.MustCompile(`^([A-Z0-9]{2})([0-9]{1,4}[A-Z]?)$`)
)

type Query struct{ FlightNumber, DepartureDate, Origin, Destination string }

type Flight struct {
	ProviderID, Airline, FlightNumber, OperatingAirline, OperatingFlightNumber string
	DepartureAirport, ArrivalAirport                                           string
	DepartureDate, DepartureTime, DepartureTimeZone                            string
	ArrivalDate, ArrivalTime, ArrivalTimeZone                                  string
}

type Client struct {
	key     string
	baseURL string
	http    *http.Client
	now     func() time.Time
}

// New returns nil when lookup is not configured; manual flight entry continues
// to work. Keys remain in server memory and are sent only in x-apikey headers.
func New(key string) *Client {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	return &Client{key: key, baseURL: apiBase, now: time.Now, http: &http.Client{
		Timeout:       8 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}
}

type scheduledFlight struct {
	Ident           string `json:"ident"`
	IdentICAO       string `json:"ident_icao"`
	IdentIATA       string `json:"ident_iata"`
	ActualIdent     string `json:"actual_ident"`
	ActualIdentICAO string `json:"actual_ident_icao"`
	ActualIdentIATA string `json:"actual_ident_iata"`
	Origin          string `json:"origin"`
	OriginICAO      string `json:"origin_icao"`
	OriginIATA      string `json:"origin_iata"`
	Destination     string `json:"destination"`
	DestinationICAO string `json:"destination_icao"`
	DestinationIATA string `json:"destination_iata"`
	ScheduledOut    string `json:"scheduled_out"`
	ScheduledIn     string `json:"scheduled_in"`
}

type airport struct {
	Code     string `json:"airport_code"`
	ICAO     string `json:"code_icao"`
	IATA     string `json:"code_iata"`
	Timezone string `json:"timezone"`
	location *time.Location
}

func normalized(raw string) string { return strings.ToUpper(strings.Join(strings.Fields(raw), "")) }

// SplitFlightNumber treats B6123 and 2K8373 as two-character IATA prefixes;
// splitting at the first digit would silently select the wrong carrier.
func SplitFlightNumber(raw string) (airline, number string, ok bool) {
	ident := normalized(raw)
	match := icaoFlightPattern.FindStringSubmatch(ident)
	if match == nil {
		match = iataFlightPattern.FindStringSubmatch(ident)
	}
	if match == nil || strings.Trim(match[1], "0123456789") == "" {
		return "", "", false
	}
	return match[1], match[2], true
}

func canonicalIdent(raw string) string {
	a, n, ok := SplitFlightNumber(raw)
	if !ok {
		return ""
	}
	n = strings.TrimLeft(n, "0")
	if n == "" {
		n = "0"
	}
	return a + n
}

func first(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func (c *Client) get(ctx context.Context, path string, params url.Values, target any) error {
	if c == nil || c.key == "" {
		return ErrUnavailable
	}
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return ErrUnavailable
	}
	req.Header.Set("x-apikey", c.key)
	req.Header.Set("Accept", "application/json")
	response, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return ErrUnavailable
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusPaymentRequired:
		return ErrAccess
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return ErrUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, responseLimit+1))
	if err != nil || len(data) > responseLimit {
		return ErrInvalidData
	}
	if err := json.Unmarshal(data, target); err != nil {
		return ErrInvalidData
	}
	return nil
}

func (c *Client) getAirport(ctx context.Context, code string, cache map[string]airport) (airport, error) {
	code = normalized(code)
	if !airportPattern.MatchString(code) {
		return airport{}, ErrInvalidData
	}
	if a, ok := cache[code]; ok {
		return a, nil
	}
	// A number-only lookup can cover multiple airports. Bound paid enrichment
	// calls rather than following an unbounded timetable or pagination link.
	unique := map[string]bool{}
	for _, a := range cache {
		unique[a.Code] = true
	}
	if len(unique) >= 8 {
		return airport{}, ErrNarrowSearch
	}
	var a airport
	if err := c.get(ctx, "/airports/"+code, nil, &a); err != nil {
		return airport{}, err
	}
	a.Code, a.ICAO, a.IATA = normalized(a.Code), normalized(a.ICAO), normalized(a.IATA)
	if code != a.Code && code != a.ICAO && code != a.IATA {
		return airport{}, ErrInvalidData
	}
	a.Timezone = strings.TrimPrefix(strings.TrimSpace(a.Timezone), ":")
	if a.Timezone == "" || a.Timezone == "Local" {
		return airport{}, ErrInvalidData
	}
	location, err := time.LoadLocation(a.Timezone)
	if err != nil {
		return airport{}, ErrInvalidData
	}
	a.location = location
	if !airportPattern.MatchString(a.Code) {
		a.Code = first(a.ICAO, a.IATA)
	}
	for _, alias := range []string{code, a.Code, a.ICAO, a.IATA} {
		if airportPattern.MatchString(alias) {
			cache[alias] = a
		}
	}
	return a, nil
}

func (c *Client) Search(ctx context.Context, query Query) ([]Flight, error) {
	if c == nil {
		return nil, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	query.FlightNumber = normalized(query.FlightNumber)
	query.Origin = normalized(query.Origin)
	query.Destination = normalized(query.Destination)
	day, err := time.Parse("2006-01-02", query.DepartureDate)
	if err != nil || day.Format("2006-01-02") != query.DepartureDate {
		return nil, errors.New("Enter a departure date in YYYY-MM-DD format.")
	}
	if day.Before(c.now().UTC().AddDate(0, -3, -1)) || day.After(c.now().UTC().AddDate(1, 0, 1)) {
		return nil, ErrScheduleWindow
	}
	params := url.Values{"max_pages": {"3"}, "include_codeshares": {"true"}}
	if query.FlightNumber != "" {
		airline, number, ok := SplitFlightNumber(query.FlightNumber)
		if !ok {
			return nil, errors.New("Enter a flight number such as UA123 or B6123.")
		}
		params.Set("airline", airline)
		// AeroAPI's numeric filter excludes a possible operational suffix;
		// preserve that suffix in the exact identifier match below.
		numeric, _ := strconv.Atoi(strings.TrimRight(number, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
		params.Set("flight_number", strconv.Itoa(numeric))
	} else if query.Origin == "" || query.Destination == "" {
		return nil, errors.New("Enter a flight number or both departure and arrival airports.")
	}
	for key, value := range map[string]string{"origin": query.Origin, "destination": query.Destination} {
		if value != "" {
			if !airportPattern.MatchString(value) {
				return nil, errors.New("Enter a three- or four-letter airport code.")
			}
			params.Set(key, value)
		}
	}
	cache := map[string]airport{}
	// Covers every local departure date, including UTC+14 and UTC-12, when
	// the departure airport is not yet known. Exact dates are filtered below.
	start, end := day.Add(-14*time.Hour), day.Add(36*time.Hour)
	if query.Origin != "" {
		origin, err := c.getAirport(ctx, query.Origin, cache)
		if err != nil {
			return nil, err
		}
		localDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, origin.location)
		start, end = localDay.UTC(), localDay.AddDate(0, 0, 1).UTC()
	}
	var response struct {
		Scheduled *[]scheduledFlight `json:"scheduled"`
		Links     *struct {
			Next string `json:"next"`
		} `json:"links"`
	}
	path := "/schedules/" + start.Format(time.RFC3339) + "/" + end.Format(time.RFC3339)
	if err := c.get(ctx, path, params, &response); err != nil {
		return nil, err
	}
	if response.Scheduled == nil {
		return nil, ErrInvalidData
	}
	if response.Links != nil && response.Links.Next != "" {
		return nil, ErrNarrowSearch
	}
	rows := *response.Scheduled
	if len(rows) > 45 {
		return nil, ErrNarrowSearch
	}
	flights := []Flight{}
	seen := map[string]int{}
	invalid := 0
	for _, row := range rows {
		flight, matched, err := c.mapFlight(ctx, query, row, cache)
		if err != nil {
			if errors.Is(err, ErrInvalidData) {
				invalid++
				continue
			}
			return nil, err
		}
		if !matched {
			continue
		}
		if index, exists := seen[flight.ProviderID]; exists {
			// Prefer the requested marketing number, otherwise the operating
			// carrier's own row, when codeshares describe the same airplane.
			old := flights[index]
			if (canonicalIdent(flight.FlightNumber) == canonicalIdent(query.FlightNumber) && canonicalIdent(old.FlightNumber) != canonicalIdent(query.FlightNumber)) || (query.FlightNumber == "" && flight.FlightNumber == flight.OperatingFlightNumber) {
				flights[index] = flight
			}
			continue
		}
		seen[flight.ProviderID] = len(flights)
		flights = append(flights, flight)
	}
	if len(flights) == 0 && invalid > 0 {
		return nil, ErrInvalidData
	}
	sort.Slice(flights, func(i, j int) bool {
		a, b := flights[i], flights[j]
		return a.DepartureDate+a.DepartureTime+a.DepartureAirport+a.FlightNumber < b.DepartureDate+b.DepartureTime+b.DepartureAirport+b.FlightNumber
	})
	if len(flights) > 20 {
		return nil, ErrNarrowSearch
	}
	return flights, nil
}

func (c *Client) mapFlight(ctx context.Context, query Query, row scheduledFlight, cache map[string]airport) (Flight, bool, error) {
	var flight Flight
	idents := []string{row.Ident, row.IdentICAO, row.IdentIATA, row.ActualIdent, row.ActualIdentICAO, row.ActualIdentIATA}
	if query.FlightNumber != "" {
		match := false
		for _, ident := range idents {
			if canonicalIdent(ident) != "" && canonicalIdent(ident) == canonicalIdent(query.FlightNumber) {
				match = true
			}
		}
		if !match {
			return flight, false, nil
		}
	}
	originCode := normalized(first(row.OriginICAO, row.Origin, row.OriginIATA))
	destinationCode := normalized(first(row.DestinationICAO, row.Destination, row.DestinationIATA))
	departure, err := time.Parse(time.RFC3339, row.ScheduledOut)
	if err != nil {
		return flight, false, ErrInvalidData
	}
	arrival, err := time.Parse(time.RFC3339, row.ScheduledIn)
	if err != nil || !arrival.After(departure) || arrival.Sub(departure) > 48*time.Hour {
		return flight, false, ErrInvalidData
	}
	origin, err := c.getAirport(ctx, originCode, cache)
	if err != nil {
		return flight, false, err
	}
	if query.Origin != "" && query.Origin != origin.Code && query.Origin != origin.IATA && query.Origin != origin.ICAO {
		return flight, false, nil
	}
	localDeparture := departure.In(origin.location)
	if localDeparture.Format("2006-01-02") != query.DepartureDate {
		return flight, false, nil
	}
	destination, err := c.getAirport(ctx, destinationCode, cache)
	if err != nil {
		return flight, false, err
	}
	if query.Destination != "" && query.Destination != destination.Code && query.Destination != destination.IATA && query.Destination != destination.ICAO {
		return flight, false, nil
	}
	localArrival := arrival.In(destination.location)
	marketing := canonicalIdent(first(row.IdentIATA, row.IdentICAO, row.Ident))
	operating := canonicalIdent(first(row.ActualIdentIATA, row.ActualIdentICAO, row.ActualIdent, row.IdentIATA, row.IdentICAO, row.Ident))
	operatingCanonical := canonicalIdent(first(row.ActualIdentICAO, row.ActualIdent, row.ActualIdentIATA, row.IdentICAO, row.Ident, row.IdentIATA))
	if marketing == "" || operating == "" || operatingCanonical == "" {
		return flight, false, ErrInvalidData
	}
	if query.FlightNumber != "" {
		marketing = canonicalIdent(query.FlightNumber)
	}
	airline, _, _ := SplitFlightNumber(marketing)
	operatingAirline, _, _ := SplitFlightNumber(operating)
	identity := strings.Join([]string{operatingCanonical, first(origin.ICAO, origin.Code), first(destination.ICAO, destination.Code), strconv.FormatInt(departure.Unix(), 10)}, "|")
	digest := sha256.Sum256([]byte(identity))
	flight = Flight{
		ProviderID: "schedule-" + hex.EncodeToString(digest[:]), Airline: airline, FlightNumber: marketing,
		OperatingAirline: operatingAirline, OperatingFlightNumber: operating,
		DepartureAirport: first(origin.IATA, origin.Code), ArrivalAirport: first(destination.IATA, destination.Code),
		DepartureDate: localDeparture.Format("2006-01-02"), DepartureTime: localDeparture.Format("15:04"), DepartureTimeZone: origin.Timezone,
		ArrivalDate: localArrival.Format("2006-01-02"), ArrivalTime: localArrival.Format("15:04"), ArrivalTimeZone: destination.Timezone,
	}
	return flight, true, nil
}
