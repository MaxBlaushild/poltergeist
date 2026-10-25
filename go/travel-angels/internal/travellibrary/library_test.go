package travellibrary

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/travel-angels/internal/travelcalendar"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type libraryFixture struct {
	db                      *gorm.DB
	router                  *gin.Engine
	service                 *Service
	owner, friend, stranger uuid.UUID
	legacy                  uuid.UUID
}

func libraryMigrationSQL(t *testing.T, name string) string {
	t.Helper()
	_, source, _, _ := runtime.Caller(0)
	data, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../migrate/internal/migrations", name))
	if err != nil {
		t.Fatal(err)
	}
	// Old baseline migrations require uuid-ossp in production. The isolated test
	// schema uses PostgreSQL's built-in equivalent, without installing extensions.
	return strings.ReplaceAll(string(data), "uuid_generate_v4()", "gen_random_uuid()")
}
func libraryMigration(t *testing.T, tx *gorm.DB, name string) {
	t.Helper()
	if err := tx.Exec(libraryMigrationSQL(t, name)).Error; err != nil {
		t.Fatalf("%s: %v", name, err)
	}
}
func newLibraryFixture(t *testing.T) *libraryFixture {
	t.Helper()
	dsn := os.Getenv("TC_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("Set TC_TEST_DATABASE_URL for PostgreSQL Library integration tests")
	}
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	pool.SetMaxOpenConns(1)
	schema := "ta_library_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := database.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Exec("SET search_path TO " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Exec("DROP SCHEMA " + schema + " CASCADE"); pool.Close() })
	for _, sql := range []string{
		"CREATE TABLE users (id uuid PRIMARY KEY, name text NOT NULL DEFAULT '', username text, profile_picture_url text NOT NULL DEFAULT '', email text, phone_number text)",
		"CREATE TABLE friends (id uuid PRIMARY KEY, first_user_id uuid NOT NULL REFERENCES users(id), second_user_id uuid NOT NULL REFERENCES users(id))",
	} {
		if err := database.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"000124_create_documents_and_document_tags.up.sql", "000130_add_document_locations.up.sql", "000131_create_trending_destinations.up.sql", "000473_travel_calendars.up.sql", "000474_travel_calendar_guests.up.sql", "000475_travel_trip_pages.up.sql"} {
		libraryMigration(t, database, name)
	}
	f := &libraryFixture{db: database, owner: uuid.New(), friend: uuid.New(), stranger: uuid.New(), legacy: uuid.New()}
	for i, person := range []uuid.UUID{f.owner, f.friend, f.stranger} {
		if err := database.Exec("INSERT INTO users(id,name,email,phone_number) VALUES (?, ?, 'secret@example.com', 'secret-phone')", person, []string{"Owner", "Friend", "Stranger"}[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := database.Exec("INSERT INTO documents(id,title,provider,user_id) VALUES (?, 'Legacy sheet', 'google_sheets', ?)", f.legacy, f.owner).Error; err != nil {
		t.Fatal(err)
	}
	libraryMigration(t, database, "000476_travel_library.up.sql")
	if err := database.Exec("INSERT INTO friends(id,first_user_id,second_user_id) VALUES (?, ?, ?)", uuid.New(), f.owner, f.friend).Error; err != nil {
		t.Fatal(err)
	}
	auth := func(c *gin.Context) (*models.User, error) {
		id, err := uuid.Parse(c.GetHeader("X-Test-User"))
		if err != nil {
			return nil, errors.New("not signed in")
		}
		return &models.User{ID: id}, nil
	}
	calendar := travelcalendar.New(database, auth, "https://travel.example.test")
	f.service = New(database, auth, calendar.LibraryTrip, calendar.LibraryTrips)
	gin.SetMode(gin.TestMode)
	f.router = gin.New()
	f.service.RegisterRoutes(f.router)
	f.router.GET("/trending", f.service.Trending)
	// Legacy list handlers delegate to these same scope methods.
	f.router.GET("/friends", func(c *gin.Context) { f.service.ListScope(c, "network") })
	return f
}
func (f *libraryFixture) call(t *testing.T, method, path string, user uuid.UUID, body interface{}, status int) interface{} {
	t.Helper()
	var raw []byte
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if user != uuid.Nil {
		req.Header.Set("X-Test-User", user.String())
	}
	response := httptest.NewRecorder()
	f.router.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("%s %s: got %d want %d: %s", method, path, response.Code, status, response.Body.String())
	}
	var result interface{}
	if response.Body.Len() > 0 {
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
	}
	if status == 200 && path != "/friends" && response.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal("missing private cache policy")
	}
	return result
}
func (f *libraryFixture) doc(t *testing.T, title, visibility string, owner uuid.UUID) models.Document {
	t.Helper()
	content := "SECRET CONTENT: " + title
	doc := models.Document{ID: uuid.New(), Title: title, UserID: owner, Provider: models.CloudDocumentProviderInternal, Visibility: visibility, Content: &content}
	if err := doc.NormalizeLibraryMetadata(); err != nil {
		t.Fatal(err)
	}
	if err := f.db.Omit("User", "DocumentTags", "DocumentLocations").Create(&doc).Error; err != nil {
		t.Fatal(err)
	}
	return doc
}
func (f *libraryFixture) trip(t *testing.T, owner uuid.UUID, title string) uuid.UUID {
	t.Helper()
	cal := travelcalendar.Calendar{ID: uuid.New(), OwnerID: owner, Title: "Calendar", Visibility: "private", ShareToken: uuid.NewString()}
	if err := f.db.Create(&cal).Error; err != nil {
		t.Fatal(err)
	}
	trip := travelcalendar.Stop{ID: uuid.New(), CalendarID: cal.ID, Visibility: "inherit", ShareToken: uuid.NewString(), Draft: travelcalendar.StopContent{Title: title, Destination: "Paris", DatePrecision: "tbd", TimeZone: "UTC", Status: "tentative"}}
	if err := f.db.Create(&trip).Error; err != nil {
		t.Fatal(err)
	}
	return trip.ID
}

func TestLibraryMetadataDefaultsAndSourceLinkSafety(t *testing.T) {
	for _, item := range []struct {
		provider models.CloudDocumentProvider
		link     string
		format   string
	}{
		{models.CloudDocumentProviderInternal, "", "guide"}, {models.CloudDocumentProviderGoogleSheets, "https://docs.google.com/spreadsheets/d/one", "spreadsheet"}, {models.CloudDocumentProviderInternal, "https://files.example.test/videos/one.mp4", "video"},
	} {
		doc := models.Document{Title: "  Itinerary  ", Provider: item.provider, Link: &item.link}
		if err := doc.NormalizeLibraryMetadata(); err != nil || doc.Visibility != "private" || doc.Format != item.format || doc.Title != "Itinerary" {
			t.Fatalf("wrong defaults: %+v %v", doc, err)
		}
	}
	for _, link := range []string{"javascript:alert(1)", "file:///etc/passwd", "https://user:pass@example.com/x", "data:text/html,hello"} {
		if models.ValidateDocumentSourceLink(&link) == nil {
			t.Fatalf("accepted %q", link)
		}
	}
	for _, format := range []string{"spreadsheet", "itinerary", "guide", "video"} {
		doc := models.Document{Title: strings.Repeat("旅", 500), Format: format}
		if err := doc.NormalizeLibraryMetadata(); err != nil {
			t.Fatal(err)
		}
	}
	doc := models.Document{Title: "valid", Format: "unsupported"}
	if doc.NormalizeLibraryMetadata() == nil {
		t.Fatal("invalid format accepted")
	}
}

func TestLibraryMigrationPreservesLegacyAudienceAndDefaultsNewRowsPrivate(t *testing.T) {
	f := newLibraryFixture(t)
	var legacy models.Document
	if err := f.db.First(&legacy, "id = ?", f.legacy).Error; err != nil {
		t.Fatal(err)
	}
	if legacy.Visibility != "network" || legacy.Format != "spreadsheet" {
		t.Fatalf("legacy changed: %+v", legacy)
	}
	newID := uuid.New()
	if err := f.db.Exec("INSERT INTO documents(id,title,provider,user_id) VALUES (?, 'New upload', 'internal', ?)", newID, f.owner).Error; err != nil {
		t.Fatal(err)
	}
	var fresh models.Document
	if err := f.db.First(&fresh, "id = ?", newID).Error; err != nil {
		t.Fatal(err)
	}
	if fresh.Visibility != "private" {
		t.Fatal("new SQL row was shared")
	}
	// Downgrades require an explicit audience decision before removing the ACL.
	if err := f.db.Exec("UPDATE documents SET visibility = 'network'").Error; err != nil {
		t.Fatal(err)
	}
	libraryMigration(t, f.db, "000476_travel_library.down.sql")
	libraryMigration(t, f.db, "000476_travel_library.up.sql")
	if err := f.db.First(&legacy, "id = ?", f.legacy).Error; err != nil {
		t.Fatal(err)
	}
}

func TestLibraryMigrationRejectsDowngradeWithRestrictedDocuments(t *testing.T) {
	for _, visibility := range []string{"private", "selected"} {
		t.Run(visibility, func(t *testing.T) {
			f := newLibraryFixture(t)
			doc := f.doc(t, "Restricted travel notes", visibility, f.owner)
			base := "/travel-angels/documents/" + doc.ID.String()
			if visibility == "selected" {
				f.call(t, "PUT", base+"/sharing", f.owner, gin.H{"visibility": "selected", "userIds": []uuid.UUID{f.stranger}}, 200)
			}
			trip := f.trip(t, f.owner, "Owner's private trip")
			f.call(t, "PUT", base+"/trips/"+trip.String(), f.owner, gin.H{"saved": true}, 200)

			err := f.db.Exec(libraryMigrationSQL(t, "000476_travel_library.down.sql")).Error
			if err == nil || !strings.Contains(err.Error(), "restricted documents remain") {
				t.Fatalf("downgrade did not reject %s documents with an actionable error: %v", visibility, err)
			}
			for _, column := range []string{"visibility", "format", "personal_note", "destination"} {
				if !f.db.Migrator().HasColumn(&models.Document{}, column) {
					t.Fatalf("rejected downgrade removed documents.%s", column)
				}
			}
			for _, table := range []string{"document_shares", "document_saves", "document_trip_references"} {
				var count int64
				if err := f.db.Table(table).Where("document_id = ?", doc.ID).Count(&count).Error; err != nil {
					t.Fatalf("rejected downgrade removed %s: %v", table, err)
				}
				want := int64(1)
				if table == "document_shares" && visibility == "private" {
					want = 0
				}
				if count != want {
					t.Fatalf("rejected downgrade changed %s: got %d want %d", table, count, want)
				}
			}
			row := f.call(t, "GET", base, f.owner, nil, 200).(map[string]interface{})
			if row["visibility"] != visibility || row["content"] != *doc.Content || row["isSaved"] != true || len(row["savedTripIds"].([]interface{})) != 1 {
				t.Fatalf("rejected downgrade changed the restricted document: %v", row)
			}
			f.call(t, "GET", base, f.friend, nil, 404)
			if visibility == "selected" {
				f.call(t, "GET", base, f.stranger, nil, 200)
			} else {
				f.call(t, "GET", base, f.stranger, nil, 404)
			}
		})
	}
}

func TestLibraryAudienceReplacementAndSafeAuthorDTO(t *testing.T) {
	f := newLibraryFixture(t)
	private := f.doc(t, "Private plans", "private", f.owner)
	network := f.doc(t, "Friends guide", "network", f.owner)
	base := "/travel-angels/documents/" + private.ID.String()
	f.call(t, "GET", base, uuid.Nil, nil, 401)
	f.call(t, "GET", base, f.friend, nil, 404)
	f.call(t, "GET", "/travel-angels/documents/"+network.ID.String(), f.stranger, nil, 404)
	rows := f.call(t, "GET", "/friends", f.friend, nil, 200).([]interface{})
	if len(rows) != 2 {
		t.Fatalf("private leaked or legacy missing: %v", rows)
	}
	encoded, _ := json.Marshal(rows)
	for _, secret := range []string{"Private plans", "secret@example.com", "secret-phone", "phoneNumber", "credits", "sharedWithUserIds"} {
		if bytes.Contains(encoded, []byte(secret)) {
			t.Fatalf("leaked %s", secret)
		}
	}
	result := f.call(t, "PUT", base+"/sharing", f.owner, gin.H{"visibility": "selected", "userIds": []uuid.UUID{f.stranger, f.stranger, f.owner}}, 200).(map[string]interface{})
	if len(result["sharedWithUserIds"].([]interface{})) != 1 {
		t.Fatal("grants not deduplicated")
	}
	f.call(t, "GET", base, f.stranger, nil, 200)
	f.call(t, "GET", base, f.friend, nil, 404)
	f.call(t, "PUT", base+"/sharing", f.stranger, gin.H{"visibility": "network"}, 403)
	f.call(t, "PUT", base+"/sharing", f.owner, gin.H{"visibility": "selected", "userIds": []uuid.UUID{uuid.New()}}, 400)
	f.call(t, "GET", base, f.stranger, nil, 200) // Invalid replacement is atomic.
	f.call(t, "PUT", base+"/sharing", f.owner, gin.H{"visibility": "network"}, 200)
	f.call(t, "GET", base, f.stranger, nil, 404)
	f.call(t, "GET", base, f.friend, nil, 200)
	if err := f.db.Exec("DELETE FROM friends").Error; err != nil {
		t.Fatal(err)
	}
	f.call(t, "GET", base, f.friend, nil, 404)
}

func TestLibrarySavedAndTripReferencesNeverGrantAccess(t *testing.T) {
	f := newLibraryFixture(t)
	doc := f.doc(t, "Shared guide", "network", f.owner)
	trip := f.trip(t, f.friend, "Friend trip")
	otherTrip := f.trip(t, f.stranger, "Hidden trip")
	base := "/travel-angels/documents/" + doc.ID.String()
	result := f.call(t, "PUT", base+"/trips/"+trip.String(), f.friend, gin.H{"saved": true}, 200).(map[string]interface{})
	if result["isSaved"] != true || len(result["savedTripIds"].([]interface{})) != 1 {
		t.Fatal("trip reference did not bookmark")
	}
	f.call(t, "PUT", base+"/trips/"+otherTrip.String(), f.friend, gin.H{"saved": true}, 404)
	choices := f.call(t, "GET", base+"/trips", f.friend, nil, 200).([]interface{})
	if len(choices) != 1 || choices[0].(map[string]interface{})["isSaved"] != true {
		t.Fatalf("wrong trip choices: %v", choices)
	}
	owner := f.call(t, "GET", base, f.owner, nil, 200).(map[string]interface{})
	if owner["isSaved"] != false || len(owner["savedTripIds"].([]interface{})) != 0 {
		t.Fatal("another user's bookmarks leaked")
	}
	f.call(t, "PUT", base+"/saved", f.friend, gin.H{"saved": false}, 200)
	tripPath := "/travel-angels/documents/library?scope=saved&tripId=" + trip.String()
	if len(f.call(t, "GET", tripPath, f.friend, nil, 200).([]interface{})) != 1 {
		t.Fatal("unsave incorrectly deleted trip reference")
	}
	f.call(t, "GET", tripPath, f.stranger, nil, 404)
	f.call(t, "PUT", base+"/sharing", f.owner, gin.H{"visibility": "private"}, 200)
	f.call(t, "GET", base, f.friend, nil, 404)
	if len(f.call(t, "GET", tripPath, f.friend, nil, 200).([]interface{})) != 0 {
		t.Fatal("private document leaked through trip")
	}
	if len(f.call(t, "GET", "/travel-angels/documents/library?scope=saved", f.friend, nil, 200).([]interface{})) != 0 {
		t.Fatal("revoked bookmark still readable")
	}
	f.call(t, "PUT", base+"/saved", f.friend, gin.H{"saved": true}, 404)
	f.call(t, "PUT", base+"/sharing", f.owner, gin.H{"visibility": "network"}, 200)
	f.call(t, "PUT", base+"/saved", f.friend, gin.H{"saved": true}, 200)
	if err := f.db.Delete(&models.Document{}, "id = ?", doc.ID).Error; err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"document_shares", "document_saves", "document_trip_references"} {
		var count int64
		if err := f.db.Table(table).Where("document_id = ?", doc.ID).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s did not cascade", table)
		}
	}
}

func TestLibraryTrendingCountsOnlyCurrentlyReadableDocuments(t *testing.T) {
	f := newLibraryFixture(t)
	visible := f.doc(t, "Visible", "network", f.owner)
	private := f.doc(t, "Hidden", "private", f.owner)
	unrelated := f.doc(t, "Unrelated network", "network", f.stranger)
	for _, item := range []struct {
		doc   models.Document
		place string
	}{{visible, "Paris"}, {private, "Secret Island"}, {unrelated, "Secret Island"}} {
		loc := models.DocumentLocation{ID: uuid.New(), DocumentID: item.doc.ID, PlaceID: item.place, Name: item.place, FormattedAddress: item.place, LocationType: models.LocationTypeCity, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := f.db.Create(&loc).Error; err != nil {
			t.Fatal(err)
		}
	}
	result := f.call(t, "GET", "/trending", f.friend, nil, 200).(map[string]interface{})
	cities := result["cities"].([]interface{})
	if len(cities) != 1 || cities[0].(map[string]interface{})["name"] != "Paris" || cities[0].(map[string]interface{})["documentCount"] != float64(1) {
		t.Fatalf("aggregate privacy failed: %v", result)
	}
	f.call(t, "PUT", "/travel-angels/documents/"+visible.ID.String()+"/sharing", f.owner, gin.H{"visibility": "private"}, 200)
	result = f.call(t, "GET", "/trending", f.friend, nil, 200).(map[string]interface{})
	if len(result["cities"].([]interface{})) != 0 {
		t.Fatal("revoked destination remained in cache")
	}
}

func TestLibraryTripAccessRevocationHidesReferenceAndTripDeletionCascades(t *testing.T) {
	f := newLibraryFixture(t)
	doc := f.doc(t, "Shared guide", "network", f.owner)
	trip := f.trip(t, f.stranger, "Private host trip")
	host := travelcalendar.Collaborator{ID: uuid.New(), StopID: &trip, UserID: f.friend, Accepted: true, IsHost: true}
	if err := f.db.Create(&host).Error; err != nil {
		t.Fatal(err)
	}
	base := "/travel-angels/documents/" + doc.ID.String()
	f.call(t, "PUT", base+"/trips/"+trip.String(), f.friend, gin.H{"saved": true}, 200)
	if err := f.db.Delete(&host).Error; err != nil {
		t.Fatal(err)
	}
	row := f.call(t, "GET", base, f.friend, nil, 200).(map[string]interface{})
	if len(row["savedTripIds"].([]interface{})) != 0 || row["isSaved"] != true {
		t.Fatal("removed host kept trip access or lost independent bookmark")
	}
	f.call(t, "GET", "/travel-angels/documents/library?scope=saved&tripId="+trip.String(), f.friend, nil, 404)
	choices := f.call(t, "GET", base+"/trips", f.friend, nil, 200).([]interface{})
	if len(choices) != 0 {
		t.Fatal("restricted trip remained in picker")
	}
	if err := f.db.Delete(&travelcalendar.Stop{}, "id = ?", trip).Error; err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := f.db.Model(&tripReference{}).Where("trip_id = ?", trip).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("trip deletion did not cascade personal reference")
	}
}
