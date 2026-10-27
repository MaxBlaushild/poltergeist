package travelcalendar

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func destinationRows(t *testing.T, f *guestTestFixture, user uuid.UUID, query string) []interface{} {
	t.Helper()
	return calendarResponse(t, f, "GET", "/destinations?q="+url.QueryEscape(query), nil, user, "", 200)["destinations"].([]interface{})
}

func destinationRow(rows []interface{}, name string) map[string]interface{} {
	for _, row := range rows {
		d := row.(map[string]interface{})
		if d["name"] == name {
			return d
		}
	}
	return nil
}

func destinationBody(destinations interface{}) gin.H {
	return gin.H{"title": "A year abroad", "destinations": destinations, "datePrecision": "fixed", "startDate": "2032-12-20", "endDate": "2033-01-10", "joinOpen": true, "status": "confirmed"}
}

func destinationCreate(t *testing.T, f *guestTestFixture, user uuid.UUID, codes ...string) map[string]interface{} {
	t.Helper()
	selected := []gin.H{}
	for _, code := range codes {
		selected = append(selected, gin.H{"id": DestinationCountryID(code)})
	}
	return calendarResponse(t, f, "POST", "/calendar/stops", destinationBody(selected), user, "", 201)["stop"].(map[string]interface{})
}

func TestDestinationCountryCatalogOfflineAliasesAndAuthentication(t *testing.T) {
	f := newGuestTestFixture(t)
	var count int64
	if err := f.db.Model(&Destination{}).Where("kind = ?", "country").Count(&count).Error; err != nil || count != 249 {
		t.Fatalf("country catalog: %d %v", count, err)
	}
	calls := 0
	f.service.ConfigureDestinationSearch(func(context.Context, string) ([]DestinationCandidate, error) {
		calls++
		return nil, errors.New("offline")
	})
	for _, query := range []string{"ja", "JP", "日本"} {
		japan := destinationRow(destinationRows(t, f, f.owner, query), "Japan")
		if japan == nil || japan["id"] != DestinationCountryID("JP").String() || japan["countryCode"] != "JP" || japan["timeZone"] != "Asia/Tokyo" {
			t.Fatalf("Japan missing for %q: %#v", query, japan)
		}
		if japan["recent"] != false || japan["frequency"] != float64(0) {
			t.Fatal("catalog search invented usage")
		}
	}
	if destinationRow(destinationRows(t, f, f.owner, "UK"), "Britain (UK)") == nil && destinationRow(destinationRows(t, f, f.owner, "UK"), "United Kingdom") == nil {
		t.Fatal("UK alias missing")
	}
	if calls != 0 {
		t.Fatal("offline catalog result invoked external provider")
	}
	if len(destinationRows(t, f, f.owner, "%")) != 0 {
		t.Fatal("LIKE wildcard was not escaped")
	}
	calendarResponse(t, f, "GET", "/destinations?q=ja", nil, uuid.Nil, "", 401)
	calendarResponse(t, f, "GET", "/destinations?q="+strings.Repeat("a", 101), nil, f.owner, "", 400)
}

func TestDestinationProviderNonPrefixPersistenceAndEnrichment(t *testing.T) {
	f := newGuestTestFixture(t)
	calls := 0
	f.service.ConfigureDestinationSearch(func(ctx context.Context, q string) ([]DestinationCandidate, error) {
		calls++
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("provider request lacks deadline")
		}
		return []DestinationCandidate{{Provider: "test-places", ProviderID: "tokyo", DestinationSnapshot: DestinationSnapshot{Name: "Tokyo", Kind: "city"}}}, nil
	})
	first := destinationRow(destinationRows(t, f, f.owner, "Tokyo Japan"), "Tokyo")
	if first == nil {
		t.Fatal("non-prefix provider answer omitted")
	}
	id := first["id"].(string)
	if destinationRow(destinationRows(t, f, f.owner, "tokyo"), "Tokyo")["id"] != id || calls != 1 {
		t.Fatal("catalog did not retain stable provider identity")
	}
	lat, lng := 35.6762, 139.6503
	f.service.ConfigureDestinationSearch(func(context.Context, string) ([]DestinationCandidate, error) {
		return []DestinationCandidate{{Provider: "test-places", ProviderID: "tokyo", DestinationSnapshot: DestinationSnapshot{Name: "Tokyo renamed by provider", Kind: "city", CountryCode: "JP", Latitude: &lat, Longitude: &lng, TimeZone: "Asia/Tokyo", Metadata: map[string]any{"source": "test"}}}}, nil
	})
	enriched := destinationRow(destinationRows(t, f, f.owner, "Japan capital"), "Tokyo")
	if enriched == nil || enriched["id"] != id || enriched["countryCode"] != "JP" || enriched["countryName"] != "Japan" || enriched["parentId"] != DestinationCountryID("JP").String() || enriched["latitude"] != lat {
		t.Fatalf("provider enrichment invalid: %#v", enriched)
	}
	f.service.ConfigureDestinationSearch(func(context.Context, string) ([]DestinationCandidate, error) { return nil, errors.New("unavailable") })
	outage := calendarResponse(t, f, "GET", "/destinations?q=a+provider+outage", nil, f.owner, "", 200)
	if len(outage["destinations"].([]interface{})) != 0 || outage["lookupUnavailable"] != true {
		t.Fatal("unexpected provider outage results")
	}
	if destinationRow(destinationRows(t, f, f.owner, "tokyo"), "Tokyo") == nil {
		t.Fatal("provider outage lost stored place")
	}
}

func TestDestinationProviderRetainsCompletedResultsBeforeBudgetTimeout(t *testing.T) {
	f := newGuestTestFixture(t)
	f.service.ConfigureDestinationSearch(func(context.Context, string) ([]DestinationCandidate, error) {
		return []DestinationCandidate{{Provider: "test", ProviderID: "completed", DestinationSnapshot: DestinationSnapshot{Name: "Completed place", Kind: "place"}}}, context.DeadlineExceeded
	})
	response := calendarResponse(t, f, "GET", "/destinations?q=provider+completed", nil, f.owner, "", 200)
	if response["lookupUnavailable"] == true || destinationRow(response["destinations"].([]interface{}), "Completed place") == nil {
		t.Fatal("provider budget discarded usable completed result")
	}
}

func TestDestinationPrefixLimitPreservesExactAndPersonalMatches(t *testing.T) {
	f := newGuestTestFixture(t)
	rows := []Destination{}
	for i := 0; i < 205; i++ {
		rows = append(rows, Destination{ID: uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", i+1)), Name: fmt.Sprintf("Tokyo landmark %03d", i), Kind: "place", Provider: "ranking", ProviderID: fmt.Sprint(i), Aliases: []string{}, Metadata: map[string]any{}, Theme: map[string]any{}})
	}
	exact := uuid.MustParse("ffffffff-ffff-ffff-ffff-fffffffffffe")
	favorite := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	rows = append(rows, Destination{ID: exact, Name: "Tokyo", Kind: "city", Provider: "ranking", ProviderID: "exact", Aliases: []string{}, Metadata: map[string]any{}, Theme: map[string]any{}}, Destination{ID: favorite, Name: "Tokyo Z favorite", Kind: "place", Provider: "ranking", ProviderID: "favorite", Aliases: []string{}, Metadata: map[string]any{}, Theme: map[string]any{}})
	if err := f.db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Exec("INSERT INTO tc_destination_terms(destination_id,term) SELECT id,lower(name) FROM tc_destinations WHERE provider = ?", "ranking").Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "POST", "/calendar/stops", destinationBody([]gin.H{{"id": favorite}}), f.owner, "", 201)
	results := destinationRows(t, f, f.owner, "tokyo")
	if results[0].(map[string]interface{})["id"] != exact.String() || results[1].(map[string]interface{})["id"] != favorite.String() {
		t.Fatalf("bounded prefix search lost exact or personal result: %#v", results[:2])
	}
}

func TestTripDestinationsHydrateCanonicalIDsAndKeepDraftPrivate(t *testing.T) {
	f := newGuestTestFixture(t)
	created := destinationCreate(t, f, f.owner, "JP", "FR")
	id := created["id"].(string)
	path := "/calendar/stops/" + id
	if created["destination"] != "Japan, France" {
		t.Fatalf("legacy summary: %#v", created["destination"])
	}
	if err := f.db.Model(&Stop{}).Where("id = ?", id).Update("visibility", "link").Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "POST", path+"/publish", gin.H{"revision": 1}, f.owner, "", 200)
	body := destinationBody([]gin.H{{"id": DestinationCountryID("JP"), "name": "Forged private label"}, {"id": DestinationCountryID("IT")}})
	body["revision"] = tripAuthorityRevision(t, f, id)
	updated := calendarResponse(t, f, "PUT", path, body, f.owner, "", 200)["stop"].(map[string]interface{})
	if updated["destination"] != "Japan, France" || updated["draft"].(map[string]interface{})["destination"] != "Japan, Italy" {
		t.Fatalf("draft overwrote published content: %#v", updated)
	}
	guest := calendarResponse(t, f, "GET", "/shared/stops/"+created["shareToken"].(string), nil, uuid.Nil, "", 200)["stop"].(map[string]interface{})
	if guest["draft"] != nil || guest["destination"] != "Japan, France" || guest["destinations"].([]interface{})[1].(map[string]interface{})["countryCode"] != "FR" {
		t.Fatal("unpublished destinations leaked to guest")
	}
	preview := calendarResponse(t, f, "POST", path+"/preview", gin.H{"audience": "attendee"}, f.owner, "", 200)["stop"].(map[string]interface{})
	if preview["destinations"].([]interface{})[1].(map[string]interface{})["countryCode"] != "IT" {
		t.Fatal("manager preview lost draft destinations")
	}
	// An old client changing only copy cannot erase the canonical ordered list.
	body = gin.H{"revision": tripAuthorityRevision(t, f, id), "title": "Updated title", "destination": "Japan, Italy", "datePrecision": "fixed", "startDate": "2032-12-20", "endDate": "2033-01-10", "joinOpen": true}
	legacy := calendarResponse(t, f, "PUT", path, body, f.owner, "", 200)["stop"].(map[string]interface{})["draft"].(map[string]interface{})
	if len(legacy["destinations"].([]interface{})) != 2 || legacy["destinations"].([]interface{})[0].(map[string]interface{})["id"] != DestinationCountryID("JP").String() {
		t.Fatal("old client erased canonical choices")
	}
	// Explicit legacy label changes remain one private, trip-scoped label.
	body["revision"] = tripAuthorityRevision(t, f, id)
	body["destination"] = "Our private family cabin"
	custom := calendarResponse(t, f, "PUT", path, body, f.owner, "", 200)["stop"].(map[string]interface{})["draft"].(map[string]interface{})
	if len(custom["destinations"].([]interface{})) != 1 || custom["destinations"].([]interface{})[0].(map[string]interface{})["id"] != nil {
		t.Fatal("custom label became global canonical record")
	}
	if len(destinationRows(t, f, f.owner, "Our private family cabin")) != 0 {
		t.Fatal("custom destination leaked to search catalog")
	}
}

func TestDestinationUsageIsPersonalUniqueAndTransactional(t *testing.T) {
	f := newGuestTestFixture(t)
	other := tripAuthorityAccount(t, f)
	first := destinationCreate(t, f, f.owner, "JP")
	destinationCreate(t, f, f.owner, "JP")
	row := destinationRow(destinationRows(t, f, f.owner, "ja"), "Japan")
	if row["frequency"] != float64(2) || row["recent"] != true {
		t.Fatalf("missing own frequency: %#v", row)
	}
	stranger := destinationRow(destinationRows(t, f, other, "ja"), "Japan")
	if stranger["frequency"] != float64(0) || stranger["recent"] != false {
		t.Fatal("other account's usage exposed")
	}
	if destinationRow(destinationRows(t, f, other, ""), "Japan") != nil {
		t.Fatal("other user's recent selection ranked in empty search")
	}
	if destinationRow(destinationRows(t, f, f.owner, ""), "Japan") == nil {
		t.Fatal("own recent selection missing")
	}
	var before destinationUsage
	if err := f.db.Where("user_id = ? AND stop_id = ?", f.owner, first["id"]).First(&before).Error; err != nil {
		t.Fatal(err)
	}
	body := destinationBody([]gin.H{{"id": DestinationCountryID("JP")}})
	body["revision"] = 1
	calendarResponse(t, f, "PUT", "/calendar/stops/"+first["id"].(string), body, f.owner, "", 200)
	var after destinationUsage
	f.db.Where("user_id = ? AND stop_id = ?", f.owner, first["id"]).First(&after)
	if !before.LastSelectedAt.Equal(after.LastSelectedAt) || destinationRow(destinationRows(t, f, f.owner, "ja"), "Japan")["frequency"] != float64(2) {
		t.Fatal("unchanged saves or reads inflated usage")
	}
	invalid := destinationBody([]gin.H{{"id": DestinationCountryID("FR")}})
	invalid["title"] = ""
	calendarResponse(t, f, "POST", "/calendar/stops", invalid, f.owner, "", 400)
	if destinationRow(destinationRows(t, f, f.owner, "france"), "France")["frequency"] != float64(0) {
		t.Fatal("failed save recorded usage")
	}
	calendarResponse(t, f, "DELETE", "/calendar/stops/"+first["id"].(string), gin.H{"revision": tripAuthorityRevision(t, f, first["id"].(string)), "confirm": true}, f.owner, "", 200)
	if destinationRow(destinationRows(t, f, f.owner, "ja"), "Japan")["frequency"] != float64(1) {
		t.Fatal("deleted trip retained usage reference")
	}
}

func TestDestinationInvalidIDsListsAndCoordinatesDoNotMutateTrip(t *testing.T) {
	f := newGuestTestFixture(t)
	created := destinationCreate(t, f, f.owner, "JP")
	id := created["id"].(string)
	tooMany := []gin.H{}
	for i := 0; i < 13; i++ {
		tooMany = append(tooMany, gin.H{"name": "Custom " + strings.Repeat("x", i), "kind": "place"})
	}
	for _, input := range []interface{}{[]gin.H{}, []gin.H{{"id": uuid.New()}}, []gin.H{{"id": DestinationCountryID("JP")}, {"id": DestinationCountryID("JP")}}, tooMany, []gin.H{{"name": "Invalid", "latitude": 100, "longitude": 1}}, []gin.H{{"name": "Invalid", "latitude": 1}}, []gin.H{{"name": "Invalid", "kind": "continent"}}} {
		body := destinationBody(input)
		body["revision"] = 1
		calendarResponse(t, f, "PUT", "/calendar/stops/"+id, body, f.owner, "", 400)
	}
	if tripAuthorityRevision(t, f, id) != 1 {
		t.Fatal("failed validation changed trip")
	}
	var stop Stop
	f.db.First(&stop, "id = ?", id)
	if len(stop.Draft.Destinations) != 1 || stop.Draft.Destinations[0].Name != "Japan" {
		t.Fatal("failed validation replaced destinations")
	}
}

func TestDestinationMaterialChangesCompareOrderedIdentity(t *testing.T) {
	jp, fr := DestinationCountryID("JP"), DestinationCountryID("FR")
	before := StopContent{Destination: "Japan, France", Destinations: []DestinationSnapshot{{ID: &jp, Name: "Japan", Kind: "country"}, {ID: &fr, Name: "France", Kind: "country"}}}
	after := before
	after.Destinations = []DestinationSnapshot{{ID: &jp, Name: "Updated catalog label", Kind: "country"}, {ID: &fr, Name: "France", Kind: "country"}}
	if materialChange(&before, &after) {
		t.Fatal("same-ID label refresh marked material")
	}
	after.Destinations[0], after.Destinations[1] = after.Destinations[1], after.Destinations[0]
	if !materialChange(&before, &after) {
		t.Fatal("ordered route change not material")
	}
	after.Destinations = []DestinationSnapshot{{ID: &jp, Name: "Japan", Kind: "country"}, {ID: &jp, Name: "France", Kind: "country"}}
	if !materialChange(&before, &after) {
		t.Fatal("same-summary identity change ignored")
	}
	legacy := StopContent{Destination: "Our cabin"}
	modern := StopContent{Destination: "Our cabin", Destinations: []DestinationSnapshot{{Name: "Our cabin", Kind: "place"}}}
	if materialChange(&legacy, &modern) {
		t.Fatal("legacy representation migration alone marked material")
	}
}

func TestDestinationRouteChangeRequiresReconfirmationOnlyWhenPublished(t *testing.T) {
	f := newGuestTestFixture(t)
	created := destinationCreate(t, f, f.owner, "JP", "FR")
	id := created["id"].(string)
	path := "/calendar/stops/" + id
	calendarResponse(t, f, "POST", path+"/publish", gin.H{"revision": 1}, f.owner, "", 200)
	body := destinationBody([]gin.H{{"id": DestinationCountryID("FR")}, {"id": DestinationCountryID("JP")}})
	body["revision"] = tripAuthorityRevision(t, f, id)
	calendarResponse(t, f, "PUT", path, body, f.owner, "", 200)
	var attendance Participation
	f.db.Where("stop_id = ? AND user_id = ?", id, f.owner).First(&attendance)
	if attendance.Status != "joined" {
		t.Fatal("draft route edit changed attendance")
	}
	calendarResponse(t, f, "POST", path+"/publish", gin.H{"revision": tripAuthorityRevision(t, f, id)}, f.owner, "", 200)
	f.db.Where("stop_id = ? AND user_id = ?", id, f.owner).First(&attendance)
	if attendance.Status != "needs_reconfirmation" || !attendance.Reserved {
		t.Fatalf("published reordered route did not preserve reserved reconfirmation: %#v", attendance)
	}
}

func TestDestinationNormalizationDoesNotMutatePublishedSlice(t *testing.T) {
	f := newGuestTestFixture(t)
	jp := DestinationCountryID("JP")
	published := StopContent{Destination: "Prior label", Destinations: []DestinationSnapshot{{ID: &jp, Name: "Prior label", Kind: "country"}}}
	draft := published
	if err := f.service.prepareDestinations(f.db, &draft, nil); err != nil {
		t.Fatal(err)
	}
	if published.Destinations[0].Name != "Prior label" || draft.Destinations[0].Name != "Japan" {
		t.Fatal("canonical hydration mutated separately published snapshot")
	}
}

func TestDestinationLegacyThreeHundredCharactersAndVisibleSearch(t *testing.T) {
	s := &Service{}
	before := StopContent{Title: "Legacy", Destination: strings.Repeat("旅", 300), DatePrecision: "tbd"}
	after := before
	if err := s.prepareDestinations(nil, &after, &before); err != nil {
		t.Fatal(err)
	}
	if err := after.validate(); err != nil {
		t.Fatal(err)
	}
	if after.Destination != before.Destination || materialChange(&before, &after) {
		t.Fatal("legacy label lost during unrelated update")
	}
	city := StopContent{Destination: "Tokyo", Destinations: []DestinationSnapshot{{Name: "Tokyo", Kind: "city", CountryName: "Japan", CountryCode: "JP", Aliases: []string{"東京"}}}}
	for _, query := range []string{"tokyo", "japan", "jp", "東京"} {
		if !matchesDestinationQuery(&city, query) {
			t.Fatalf("visible location field not searchable: %q", query)
		}
	}
	if matchesDestinationQuery(&city, "hidden draft") {
		t.Fatal("search matched unrelated content")
	}
}

func TestDestinationRollbackPreservesProviderTripAsEditableCustomSnapshots(t *testing.T) {
	f := newGuestTestFixture(t)
	f.service.ConfigureDestinationSearch(func(context.Context, string) ([]DestinationCandidate, error) {
		return []DestinationCandidate{{Provider: "rollback", ProviderID: "city", DestinationSnapshot: DestinationSnapshot{Name: "Provider city", Kind: "city", CountryCode: "JP"}}}, nil
	})
	city := destinationRow(destinationRows(t, f, f.owner, "provider city"), "Provider city")
	created := calendarResponse(t, f, "POST", "/calendar/stops", destinationBody([]gin.H{{"id": city["id"]}, {"id": DestinationCountryID("JP")}}), f.owner, "", 201)["stop"].(map[string]interface{})
	id := created["id"].(string)
	path := "/calendar/stops/" + id
	calendarResponse(t, f, "POST", path+"/publish", gin.H{"revision": 1}, f.owner, "", 200)
	_, source, _, _ := runtime.Caller(0)
	for _, migration := range []string{"000477_travel_destinations.down.sql", "000477_travel_destinations.up.sql"} {
		data, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../migrate/internal/migrations", migration))
		if err != nil {
			t.Fatal(err)
		}
		if err = f.db.Exec(string(data)).Error; err != nil {
			t.Fatal(err)
		}
	}
	var stop Stop
	if err := f.db.First(&stop, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	for _, content := range []*StopContent{&stop.Draft, stop.Published} {
		if content == nil || len(content.Destinations) != 2 || content.Destinations[0].Name != "Provider city" || content.Destinations[1].Name != "Japan" || content.Destinations[0].ID != nil || content.Destinations[1].ID != nil || content.Destination != "Provider city, Japan" {
			t.Fatalf("rollback lost trip route or retained dangling IDs: %#v", content)
		}
	}
	body := destinationBody(stop.Draft.Destinations)
	body["revision"] = stop.Revision
	calendarResponse(t, f, "PUT", path, body, f.owner, "", 200)
	calendarResponse(t, f, "POST", path+"/publish", gin.H{"revision": tripAuthorityRevision(t, f, id)}, f.owner, "", 200)
}
