package travelcalendar

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

func TestVisibilityReplacesCalendarAudience(t *testing.T) {
	tests := []struct {
		name string
		p    viewPolicy
		want bool
	}{
		{"private override excludes calendar invitee", viewPolicy{Published: true, Visibility: "private", CalendarVisibility: "specific", CalendarGrant: true}, false},
		{"restricted override excludes public calendar", viewPolicy{Published: true, Visibility: "specific", CalendarVisibility: "link", HasEnabledLink: true}, false},
		{"restricted override excludes calendar grant", viewPolicy{Published: true, Visibility: "specific", CalendarVisibility: "specific", CalendarGrant: true}, false},
		{"restricted stop grant succeeds", viewPolicy{Published: true, Visibility: "specific", CalendarVisibility: "private", StopGrant: true}, true},
		{"inherited restriction ignores stop grant", viewPolicy{Published: true, Visibility: "inherit", CalendarVisibility: "specific", StopGrant: true}, false},
		{"inherited calendar grant succeeds", viewPolicy{Published: true, Visibility: "inherit", CalendarVisibility: "specific", CalendarGrant: true}, true},
		{"public stop replaces private calendar", viewPolicy{Published: true, Visibility: "link", CalendarVisibility: "private", HasEnabledLink: true}, true},
		{"public still requires a current capability", viewPolicy{Published: true, Visibility: "link", CalendarVisibility: "link"}, false},
		{"draft hidden even from invitee", viewPolicy{Visibility: "specific", StopGrant: true}, false},
		{"draft hidden even with link", viewPolicy{Visibility: "link", HasEnabledLink: true}, false},
		{"manager can read draft private stop", viewPolicy{Manager: true, Visibility: "private"}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := allowsView(test.p); got != test.want {
				t.Fatalf("allowsView = %v, want %v", got, test.want)
			}
		})
	}
}
func TestClaimRequiredBeforeParticipationDatabaseAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := New(nil, func(*gin.Context) (*models.User, error) { return nil, errors.New("expired app token") }, "")
	r := gin.New()
	s.RegisterRoutes(r)
	for _, path := range []string{"/travel-angels/calendar/stops/any/participation", "/travel-angels/calendar/stops", "/travel-angels/calendar/stops/any/publish"} {
		request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"status":"requested"}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Travel-Guest-Token", "a-guest-token-is-not-an-account")
		response := httptest.NewRecorder()
		r.ServeHTTP(response, request)
		if response.Code != 401 {
			t.Fatalf("%s returned %d, want 401", path, response.Code)
		}
		if !strings.Contains(response.Body.String(), "Claim or sign in") {
			t.Fatalf("missing claim guidance: %s", response.Body.String())
		}
	}
}
func TestStopDateValidationUsesLocalDateStrings(t *testing.T) {
	valid := StopContent{Title: "Tokyo", Destination: "Japan", DatePrecision: "fixed", StartDate: "2028-02-29", EndDate: "2028-03-02", Status: "tentative"}
	if err := valid.validate(); err != nil {
		t.Fatal(err)
	}
	for _, dates := range [][2]string{{"2027-02-29", "2027-03-02"}, {"2028-03-02", "2028-02-29"}, {"2028-02-29T00:00:00Z", "2028-03-02"}} {
		c := valid
		c.StartDate = dates[0]
		c.EndDate = dates[1]
		if c.validate() == nil {
			t.Fatalf("accepted invalid local dates %v", dates)
		}
	}
	tbd := StopContent{Title: "Somewhere", Destination: "Europe", DatePrecision: "tbd", StartDate: "2028-01-01", EndDate: "2028-01-31"}
	if err := tbd.validate(); err != nil {
		t.Fatal(err)
	}
	if tbd.StartDate != "" || tbd.EndDate != "" {
		t.Fatal("TBD must not imply precise dates")
	}
}
func TestParticipationEligibilityAndDateSubsets(t *testing.T) {
	fixed := StopContent{DatePrecision: "fixed", StartDate: "2028-06-01", EndDate: "2028-06-30", Status: "confirmed", JoinOpen: true}
	tests := []struct {
		name    string
		content *StopContent
		input   participationInput
		want    bool
	}{
		{"partial dates", &fixed, participationInput{Status: "requested", StartDate: "2028-06-05", EndDate: "2028-06-10"}, true},
		{"whole stop defaults", &fixed, participationInput{Status: "requested"}, true},
		{"outside stop", &fixed, participationInput{Status: "requested", StartDate: "2028-05-31", EndDate: "2028-06-10"}, false},
		{"cannot self approve", &fixed, participationInput{Status: "joined"}, false},
		{"unpublished", nil, participationInput{Status: "interested"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateParticipation(test.content, &test.input, "2028-01-01")
			if (err == nil) != test.want {
				t.Fatalf("error=%v, want valid=%v", err, test.want)
			}
		})
	}
	tbd := StopContent{DatePrecision: "tbd", Status: "tentative", JoinOpen: true}
	interest := participationInput{Status: "interested"}
	if err := validateParticipation(&tbd, &interest, "2028-01-01"); err != nil {
		t.Fatal(err)
	}
	request := participationInput{Status: "requested"}
	if validateParticipation(&tbd, &request, "2028-01-01") == nil {
		t.Fatal("TBD request must require fixed dates")
	}
	for _, c := range []StopContent{{Status: "cancelled", JoinOpen: true}, {Status: "confirmed", JoinOpen: false}, {Status: "confirmed", JoinOpen: true, EndDate: "2027-12-31"}} {
		input := participationInput{Status: "interested"}
		if validateParticipation(&c, &input, "2028-01-01") == nil {
			t.Fatal("accepted closed, cancelled, or ended stop")
		}
	}
}
func TestReconfirmationRequiresNewRequestBeforeApproval(t *testing.T) {
	for _, status := range []string{"interested", "needs_reconfirmation", "withdrawn", "removed", "cancelled", "declined"} {
		if validateReview(status, "joined") == nil {
			t.Fatalf("approved %s without request", status)
		}
	}
	if err := validateReview("requested", "joined"); err != nil {
		t.Fatal(err)
	}
	if err := validateReview("removed", "withdrawn"); err != nil {
		t.Fatal(err)
	}
	if validateReview("requested", "withdrawn") == nil {
		t.Fatal("host may not impersonate participant withdrawal")
	}
}
func TestMaterialChangesExcludeDraftCopyAndDescription(t *testing.T) {
	before := StopContent{Destination: "Tokyo", StartDate: "2028-01-01", EndDate: "2028-01-07", DatePrecision: "fixed"}
	after := before
	after.Description = "Packing details"
	after.TimeZone = "UTC" // Explicit default does not change legacy UTC plans.
	if materialChange(&before, &after) {
		t.Fatal("description alone is not material")
	}
	after.Destination = "Kyoto"
	if !materialChange(&before, &after) {
		t.Fatal("destination must require reconfirmation")
	}
	after = before
	after.StartDate = "2028-01-02"
	if !materialChange(&before, &after) {
		t.Fatal("date change must require reconfirmation")
	}
}

func TestPastParticipationUsesDestinationMidnight(t *testing.T) {
	now := time.Date(2030, 5, 11, 0, 30, 0, 0, time.UTC)
	for _, test := range []struct {
		zone, today string
		eligible    bool
	}{
		{"Asia/Tokyo", "2030-05-11", false},
		{"America/Los_Angeles", "2030-05-10", true},
	} {
		t.Run(test.zone, func(t *testing.T) {
			content := StopContent{TimeZone: test.zone, DatePrecision: "fixed", StartDate: "2030-05-10", EndDate: "2030-05-10", Status: "confirmed", JoinOpen: true}
			today := todayForStop(&content, now)
			if today != test.today {
				t.Fatalf("today=%s, want %s", today, test.today)
			}
			input := participationInput{Status: "requested"}
			err := validateParticipation(&content, &input, today)
			if (err == nil) != test.eligible {
				t.Fatalf("error=%v, expected eligible=%v", err, test.eligible)
			}
		})
	}
	invalid := StopContent{Title: "Meet", Destination: "Tokyo", TimeZone: "Not/A_Timezone", DatePrecision: "tbd"}
	if invalid.validate() == nil {
		t.Fatal("invalid zone accepted")
	}
	unspecified := StopContent{Title: "Meet", Destination: "Tokyo", DatePrecision: "tbd"}
	if err := unspecified.validate(); err != nil {
		t.Fatal(err)
	}
	if unspecified.TimeZone != "UTC" {
		t.Fatal("unspecified timezone must explicitly default to UTC")
	}
}
