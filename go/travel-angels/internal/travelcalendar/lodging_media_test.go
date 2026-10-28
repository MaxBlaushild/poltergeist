package travelcalendar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type testLodgingProvider struct {
	searches, details, photos int
	beforeReturn              func()
	failed                    bool
	bounded                   bool
	propertyPhotos            []LodgingPropertyPhoto
	lastPhotoName             string
	searchQuery               string
}

func (p *testLodgingProvider) property(id string) *LodgingProperty {
	property := &LodgingProperty{Provider: "google", PlaceID: id, Name: "Live hotel " + id, Address: "Live public property address", Rating: 4.8, RatingCount: 123, Photos: []LodgingPropertyPhoto{{Name: "places/" + id + "/photos/live-name", Width: 1200, Height: 800, GoogleMapsURI: "https://maps.google.com/photo", Attributions: []LodgingPhotoAttribution{{DisplayName: "Photographer", URI: "https://maps.google.com/contributor"}}}}}
	if p.propertyPhotos != nil {
		property.Photos = append([]LodgingPropertyPhoto{}, p.propertyPhotos...)
	}
	return property
}
func (p *testLodgingProvider) Search(ctx context.Context, query string) ([]LodgingProperty, error) {
	p.searches++
	p.searchQuery = query
	_, p.bounded = ctx.Deadline()
	if p.failed {
		return nil, errors.New("provider unavailable")
	}
	if p.beforeReturn != nil {
		p.beforeReturn()
	}
	return []LodgingProperty{*p.property("place-a")}, nil
}

type testNearbyLodgingProvider struct {
	testLodgingProvider
	nearCalls           int
	latitude, longitude float64
}

func (p *testNearbyLodgingProvider) SearchNear(ctx context.Context, query string, latitude, longitude float64) ([]LodgingProperty, error) {
	p.nearCalls++
	p.searchQuery = query
	p.latitude, p.longitude = latitude, longitude
	_, p.bounded = ctx.Deadline()
	return []LodgingProperty{*p.property("place-a")}, nil
}

func TestLodgingPropertySearchUsesFirstTrustedDraftLocationWithFallback(t *testing.T) {
	f, stop, _ := lodgingFixture(t)
	tokyoLatitude, tokyoLongitude := 35.6762, 139.6503
	kyotoLatitude, kyotoLongitude := 35.0116, 135.7681
	tokyo := Destination{ID: uuid.New(), Name: "Tokyo", Kind: "city", Provider: "test", ProviderID: "tokyo", Latitude: &tokyoLatitude, Longitude: &tokyoLongitude, Aliases: []string{}, Metadata: map[string]any{}, Theme: map[string]any{}}
	kyoto := Destination{ID: uuid.New(), Name: "Kyoto", Kind: "city", Provider: "test", ProviderID: "kyoto", Latitude: &kyotoLatitude, Longitude: &kyotoLongitude, Aliases: []string{}, Metadata: map[string]any{}, Theme: map[string]any{}}
	if err := f.db.Create(&[]Destination{tokyo, kyoto}).Error; err != nil {
		t.Fatal(err)
	}
	fakeLatitude, fakeLongitude := 1.0, 2.0
	custom := DestinationSnapshot{Name: "A custom location", Latitude: &fakeLatitude, Longitude: &fakeLongitude}
	countryID := DestinationCountryID("JP")
	country := DestinationSnapshot{ID: &countryID, Name: "Japan"}
	tokyoSnapshot := tokyo.snapshot()
	tokyoSnapshot.Latitude, tokyoSnapshot.Longitude = &fakeLatitude, &fakeLongitude
	stop.Draft.Destinations = []DestinationSnapshot{custom, country, tokyoSnapshot, kyoto.snapshot()}
	stop.Published.Destinations = []DestinationSnapshot{kyoto.snapshot()}
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	p := &testNearbyLodgingProvider{}
	f.service.ConfigureLodgingPropertyProvider(p)
	path := "/calendar/stops/" + stop.ID.String() + "/lodging-properties?q=hotel+spa&latitude=1&longitude=2"
	calendarResponse(t, f, "GET", path, nil, f.owner, "", 200)
	if p.nearCalls != 1 || p.searches != 0 || p.latitude != tokyoLatitude || p.longitude != tokyoLongitude || p.searchQuery != "hotel spa" || !p.bounded {
		t.Fatalf("search must use first trusted draft coordinates without rewriting query: %+v", p)
	}
	stop.Draft.Destinations = []DestinationSnapshot{kyoto.snapshot(), tokyo.snapshot()}
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "GET", path, nil, f.owner, "", 200)
	if p.nearCalls != 2 || p.latitude != kyotoLatitude || p.longitude != kyotoLongitude {
		t.Fatal("destination order was ignored")
	}
	stop.Draft.Destinations = []DestinationSnapshot{custom, country}
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "GET", path, nil, f.owner, "", 200)
	if p.nearCalls != 2 || p.searches != 1 || p.searchQuery != "hotel spa" {
		t.Fatal("missing catalog coordinates did not fall back to text search")
	}
	stop.Draft.Destinations = []DestinationSnapshot{tokyo.snapshot()}
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	legacy := &testLodgingProvider{}
	f.service.ConfigureLodgingPropertyProvider(legacy)
	calendarResponse(t, f, "GET", path, nil, f.owner, "", 200)
	if legacy.searches != 1 {
		t.Fatal("provider without optional nearby search stopped working")
	}
}
func (p *testLodgingProvider) Details(ctx context.Context, id string) (*LodgingProperty, error) {
	p.details++
	_, p.bounded = ctx.Deadline()
	if p.failed {
		return nil, errors.New("provider unavailable")
	}
	if p.beforeReturn != nil {
		p.beforeReturn()
	}
	return p.property(id), nil
}
func (p *testLodgingProvider) Photo(ctx context.Context, name string) ([]byte, string, error) {
	p.photos++
	p.lastPhotoName = name
	_, p.bounded = ctx.Deadline()
	if p.failed {
		return nil, "", errors.New("provider unavailable")
	}
	if p.beforeReturn != nil {
		p.beforeReturn()
	}
	return []byte("provider-photo-bytes"), "image/jpeg", nil
}

func lodgingFixture(t *testing.T) (*guestTestFixture, Stop, string) {
	t.Helper()
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	stop.HostModel = true
	id := uuid.NewString()
	stop.Draft.Lodging = []TripLodging{{ID: id, Name: "Our stay", Type: "hotel", DatesTbd: true, TimeZone: "UTC", Kind: "group", Status: "considering", Visibility: "viewers", Property: &TripLodgingProperty{Provider: "google", PlaceID: "place-a"}, PrivateDetails: &TripLodgingPrivateDetails{Visibility: "hosts", Address: "Host-only instructions address"}}}
	encoded, _ := json.Marshal(stop.Draft)
	var published StopContent
	json.Unmarshal(encoded, &published)
	stop.Published = &published
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	return f, stop, id
}

func uploadLodgingTestPhoto(t *testing.T, f *guestTestFixture, stop Stop, entryID string, revision int, user uuid.UUID, status int) map[string]interface{} {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	image, _ := w.CreateFormFile("image", "room.png")
	image.Write(testTripPNG(t))
	w.WriteField("altText", "The sunny living room")
	w.WriteField("revision", strconv.Itoa(revision))
	w.Close()
	req := httptest.NewRequest("POST", "/travel-angels/calendar/stops/"+stop.ID.String()+"/lodging/"+entryID+"/photos", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if user != uuid.Nil {
		req.Header.Set("X-Test-User", user.String())
	}
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, req)
	if recorder.Code != status {
		t.Fatalf("lodging upload wanted %d, got %d: %s", status, recorder.Code, recorder.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestLodgingPropertiesAreLiveProtectedAndNeverPersisted(t *testing.T) {
	f, stop, id := lodgingFixture(t)
	p := &testLodgingProvider{}
	f.service.ConfigureLodgingPropertyProvider(p)
	base := "/calendar/stops/" + stop.ID.String()
	calendarResponse(t, f, "GET", base+"/lodging-properties?q=hotel", nil, uuid.Nil, "", 403)
	if p.searches != 0 {
		t.Fatal("unauthorized search contacted provider")
	}
	searched := calendarResponse(t, f, "GET", base+"/lodging-properties?q=hotel", nil, f.owner, "", 200)
	encoded, _ := json.Marshal(searched)
	if bytes.Contains(encoded, []byte("live-name")) || !bytes.Contains(encoded, []byte("authorAttributions")) || !bytes.Contains(encoded, []byte("googleMapsUri")) {
		t.Fatal("photo attribution or protected resource mapping missing")
	}
	preview := calendarResponse(t, f, "GET", base+"/lodging-properties/place-a", nil, f.owner, "", 200)
	if preview["property"].(map[string]interface{})["name"] != "Live hotel place-a" || !p.bounded {
		t.Fatal("live provider detail missing")
	}
	path := base + "/lodging/" + id + "/property"
	public := calendarResponse(t, f, "GET", path+"?stopToken="+stop.ShareToken, nil, uuid.Nil, "", 200)["property"].(map[string]interface{})
	if public["address"] != "Live public property address" {
		t.Fatal("public property identity should follow lodging audience")
	}
	calendarResponse(t, f, "GET", path+"?draft=true&stopToken="+stop.ShareToken, nil, uuid.Nil, "", 404)
	stop.Draft.Lodging[0].Property = &TripLodgingProperty{Provider: "google", PlaceID: "place-b"}
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	draft := calendarResponse(t, f, "GET", path+"?draft=true", nil, f.owner, "", 200)["property"].(map[string]interface{})
	public = calendarResponse(t, f, "GET", path+"?stopToken="+stop.ShareToken, nil, uuid.Nil, "", 200)["property"].(map[string]interface{})
	if draft["placeId"] != "place-b" || public["placeId"] != "place-a" {
		t.Fatal("draft property leaked into published lodging")
	}
	var saved Stop
	if err := f.db.First(&saved, "id = ?", stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	stored, _ := json.Marshal(saved)
	if bytes.Contains(stored, []byte("Live hotel")) || bytes.Contains(stored, []byte("live-name")) {
		t.Fatal("provider details were persisted")
	}
	photoURL := public["photos"].([]interface{})[0].(map[string]interface{})["url"].(string)
	photo := testGetTripHero(t, f, photoURL+"&stopToken="+stop.ShareToken, uuid.Nil, 200)
	if photo.Body.String() != "provider-photo-bytes" || p.photos != 1 {
		t.Fatal("provider photo was not fetched live")
	}
	p.failed = true
	calendarResponse(t, f, "GET", path+"?draft=true", nil, f.owner, "", 503)
}

func TestLodgingProviderPhotosStayBoundToDisplayedAttribution(t *testing.T) {
	f, stop, id := lodgingFixture(t)
	p := &testLodgingProvider{}
	f.service.ConfigureLodgingPropertyProvider(p)
	path := "/calendar/stops/" + stop.ID.String() + "/lodging/" + id + "/property"
	property := calendarResponse(t, f, "GET", path+"?stopToken="+stop.ShareToken, nil, uuid.Nil, "", 200)["property"].(map[string]interface{})
	photoURL := property["photos"].([]interface{})[0].(map[string]interface{})["url"].(string)
	original := p.property("place-a").Photos[0]
	rotated := original
	rotated.Name = "places/place-a/photos/rotated-name"
	different := original
	different.Name = "places/place-a/photos/new-first"
	different.GoogleMapsURI = "https://maps.google.com/another-photo"
	p.propertyPhotos = []LodgingPropertyPhoto{different, rotated}
	testGetTripHero(t, f, photoURL+"&stopToken="+stop.ShareToken, uuid.Nil, 200)
	if p.lastPhotoName != rotated.Name {
		t.Fatal("rotated/reordered photo did not resolve to the current resource matching displayed attribution")
	}
	changedAttribution := rotated
	changedAttribution.Attributions = []LodgingPhotoAttribution{{DisplayName: "Different photographer", URI: "https://maps.google.com/different-contributor"}}
	p.propertyPhotos = []LodgingPropertyPhoto{changedAttribution}
	testGetTripHero(t, f, photoURL+"&stopToken="+stop.ShareToken, uuid.Nil, 404)
	p.propertyPhotos = []LodgingPropertyPhoto{different}
	testGetTripHero(t, f, photoURL+"&stopToken="+stop.ShareToken, uuid.Nil, 404)
	if p.photos != 1 {
		t.Fatal("removed photo or changed attribution was silently replaced")
	}
}

func TestLodgingProviderPhotosWithoutMapsLinkRequireSameResourceAndAttribution(t *testing.T) {
	f, stop, id := lodgingFixture(t)
	p := &testLodgingProvider{}
	original := p.property("place-a").Photos[0]
	original.GoogleMapsURI = ""
	p.propertyPhotos = []LodgingPropertyPhoto{original}
	f.service.ConfigureLodgingPropertyProvider(p)
	path := "/calendar/stops/" + stop.ID.String() + "/lodging/" + id + "/property"
	property := calendarResponse(t, f, "GET", path+"?stopToken="+stop.ShareToken, nil, uuid.Nil, "", 200)["property"].(map[string]interface{})
	photoURL := property["photos"].([]interface{})[0].(map[string]interface{})["url"].(string)
	different := original
	different.Name = "places/place-a/photos/different-resource"
	p.propertyPhotos = []LodgingPropertyPhoto{different, original}
	testGetTripHero(t, f, photoURL+"&stopToken="+stop.ShareToken, uuid.Nil, 200)
	if p.lastPhotoName != original.Name {
		t.Fatal("photo without Maps link was replaced by array position")
	}
	p.propertyPhotos = []LodgingPropertyPhoto{different}
	testGetTripHero(t, f, photoURL+"&stopToken="+stop.ShareToken, uuid.Nil, 404)
	changedAttribution := original
	changedAttribution.Attributions = []LodgingPhotoAttribution{{DisplayName: "Different photographer"}}
	p.propertyPhotos = []LodgingPropertyPhoto{changedAttribution}
	testGetTripHero(t, f, photoURL+"&stopToken="+stop.ShareToken, uuid.Nil, 404)
	if p.photos != 1 {
		t.Fatal("photo without Maps link accepted an unbound resource or attribution")
	}
}

func TestLodgingPropertyAccessRechecksAudienceAndAttachmentAfterProviderCall(t *testing.T) {
	f, stop, id := lodgingFixture(t)
	p := &testLodgingProvider{}
	f.service.ConfigureLodgingPropertyProvider(p)
	path := "/calendar/stops/" + stop.ID.String() + "/lodging/" + id + "/property"
	p.beforeReturn = func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		stop.Published.Lodging[0].Visibility = "hosts"
		if err := f.db.WithContext(ctx).Save(&stop).Error; err != nil {
			t.Fatal("provider fetch held DB lock", err)
		}
	}
	calendarResponse(t, f, "GET", path+"?stopToken="+stop.ShareToken, nil, uuid.Nil, "", 404)
	before := p.details
	calendarResponse(t, f, "GET", path+"?stopToken="+stop.ShareToken, nil, uuid.Nil, "", 404)
	if p.details != before {
		t.Fatal("restricted lodging called provider")
	}
	p.beforeReturn = func() {
		stop.Draft.Lodging[0].Property = &TripLodgingProperty{Provider: "google", PlaceID: "place-b"}
		if err := f.db.Save(&stop).Error; err != nil {
			t.Fatal(err)
		}
	}
	calendarResponse(t, f, "GET", path+"?draft=true", nil, f.owner, "", 404)
}

func TestLodgingUploadsRespectScopeDraftsLimitsAndLifecycle(t *testing.T) {
	f, stop, id := lodgingFixture(t)
	storage := &testTripImageStorage{}
	f.service.ConfigureTripImageStorage(storage, "private-trip-images")
	stranger := uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?)", stranger).Error; err != nil {
		t.Fatal(err)
	}
	uploadLodgingTestPhoto(t, f, stop, id, stop.Revision, uuid.Nil, 401)
	uploadLodgingTestPhoto(t, f, stop, id, stop.Revision, stranger, 403)
	uploadLodgingTestPhoto(t, f, stop, uuid.NewString(), stop.Revision, f.owner, 400)
	if storage.puts != 0 {
		t.Fatal("invalid lodging upload reached storage")
	}
	result := uploadLodgingTestPhoto(t, f, stop, id, stop.Revision, f.owner, 200)["stop"].(map[string]interface{})
	photo := result["draft"].(map[string]interface{})["lodging"].([]interface{})[0].(map[string]interface{})["photos"].([]interface{})[0].(map[string]interface{})
	path := photo["url"].(string)
	mediaID := uuid.MustParse(photo["mediaId"].(string))
	var media TripMedia
	if err := f.db.First(&media, "id = ?", mediaID).Error; err != nil {
		t.Fatal(err)
	}
	if media.LodgingID == nil || media.LodgingID.String() != id || !strings.HasPrefix(media.StorageKey, "trip-lodging/") || len(media.Data) != 0 {
		t.Fatal("lodging image missing storage scope")
	}
	testGetTripHero(t, f, path+"?draft=true", f.owner, 200)
	testGetTripHero(t, f, path+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	testGetTripHero(t, f, tripHeroURL(stop.ID, mediaID), f.owner, 404)
	uploadLodgingTestPhoto(t, f, stop, id, stop.Revision, f.owner, 409)
	published := calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": result["revision"]}, f.owner, "", 200)["stop"].(map[string]interface{})
	testGetTripHero(t, f, path+"?stopToken="+stop.ShareToken, uuid.Nil, 200)
	revision := int(published["revision"].(float64))
	for i := 1; i < 6; i++ {
		result = uploadLodgingTestPhoto(t, f, stop, id, revision, f.owner, 200)["stop"].(map[string]interface{})
		revision = int(result["revision"].(float64))
	}
	uploadLodgingTestPhoto(t, f, stop, id, revision, f.owner, 400)
	if len(storage.objects) != 6 {
		t.Fatal("photo limit or retained versions incorrect")
	}
	result = testUploadTripHero(t, f, stop, revision, f.owner, 200)["stop"].(map[string]interface{})
	revision = int(result["revision"].(float64))
	if len(storage.objects) != 7 {
		t.Fatal("hero upload pruned lodging photos")
	}
	draft := result["draft"].(map[string]interface{})
	draft["lodging"].([]interface{})[0].(map[string]interface{})["photos"] = []interface{}{}
	draft["revision"] = revision
	removed := calendarResponse(t, f, "PUT", "/calendar/stops/"+stop.ID.String(), draft, f.owner, "", 200)["stop"].(map[string]interface{})
	if len(storage.objects) != 2 {
		t.Fatal("draft removal must retain published photo and trip hero")
	}
	published = calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": removed["revision"]}, f.owner, "", 200)["stop"].(map[string]interface{})
	if len(storage.objects) != 1 {
		t.Fatal("published photo removal must preserve the hero")
	}
	result = uploadLodgingTestPhoto(t, f, stop, id, int(published["revision"].(float64)), f.owner, 200)["stop"].(map[string]interface{})
	calendarResponse(t, f, "DELETE", "/calendar/stops/"+stop.ID.String(), gin.H{"revision": result["revision"], "confirm": true}, f.owner, "", 200)
	if len(storage.objects) != 0 {
		t.Fatal("trip deletion leaked lodging/hero media")
	}
}

func TestLodgingPhotoRechecksGuestAndAudienceAfterStorageRead(t *testing.T) {
	f, stop, id := lodgingFixture(t)
	storage := &testTripImageStorage{}
	f.service.ConfigureTripImageStorage(storage, "private-trip-images")
	result := uploadLodgingTestPhoto(t, f, stop, id, stop.Revision, f.owner, 200)["stop"].(map[string]interface{})
	calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": result["revision"]}, f.owner, "", 200)
	if err := f.db.First(&stop, "id = ?", stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	path := stop.Draft.Lodging[0].Photos[0].URL
	storage.onRead = func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		stop.Published.Lodging[0].Visibility = "hosts"
		if err := f.db.WithContext(ctx).Save(&stop).Error; err != nil {
			t.Fatal("S3 fetch held DB lock", err)
		}
	}
	testGetTripHero(t, f, path+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	storage.onRead = nil
	stop.Published.Lodging[0].Visibility = "viewers"
	stop.Visibility = "specific"
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	token, recipient := f.verify(t, "lodging-viewer@example.com")
	if err := f.db.Create(&Grant{ID: uuid.New(), StopID: &stop.ID, RecipientID: &recipient}).Error; err != nil {
		t.Fatal(err)
	}
	storage.onRead = func() {
		if err := f.db.Table("tc_guest_tokens").Where("token_hash = ?", guestDigest(token)).Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
			t.Fatal(err)
		}
	}
	testGetTripHero(t, f, path, uuid.Nil, 404, token)
}

func TestLodgingMediaCannotBeAttachedAcrossScopes(t *testing.T) {
	f, stop, id := lodgingFixture(t)
	result := uploadLodgingTestPhoto(t, f, stop, id, stop.Revision, f.owner, 200)["stop"].(map[string]interface{})
	if err := f.db.First(&stop, "id = ?", stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	photo := stop.Draft.Lodging[0].Photos[0]
	content := stop.Draft
	content.Hero = &TripHero{MediaID: photo.MediaID, AltText: "Not a cover", CropWidth: 1, CropHeight: 1}
	if f.service.validateTripContentReferences(f.db, &stop, &content) == nil {
		t.Fatal("lodging photo became hero")
	}
	content.Hero = nil
	content.Lodging = append([]TripLodging(nil), stop.Draft.Lodging...)
	content.Lodging[0].ID = uuid.NewString()
	if f.service.validateTripContentReferences(f.db, &stop, &content) == nil {
		t.Fatal("lodging photo moved to another entry")
	}
	storage := &testTripImageStorage{failPut: true}
	f.service.ConfigureTripImageStorage(storage, "private-trip-images")
	uploadLodgingTestPhoto(t, f, stop, id, int(result["revision"].(float64)), f.owner, 500)
	if len(storage.objects) != 0 {
		t.Fatal("failed photo upload orphaned object")
	}
	if err := f.db.First(&stop, "id = ?", stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	if len(stop.Draft.Lodging[0].Photos) != 1 {
		t.Fatal("failed upload changed saved photo list")
	}
}
