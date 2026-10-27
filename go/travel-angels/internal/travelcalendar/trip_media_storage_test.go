package travelcalendar

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type testTripImageStorage struct {
	objects              map[string][]byte
	puts, reads, deletes int
	failPut, failRead    bool
	bounded              bool
	onRead               func()
}

func (s *testTripImageStorage) PutObject(ctx context.Context, bucket, key, contentType string, data []byte) error {
	s.puts++
	_, s.bounded = ctx.Deadline()
	if s.objects == nil {
		s.objects = map[string][]byte{}
	}
	s.objects[bucket+"/"+key] = append([]byte(nil), data...)
	if s.failPut {
		return errors.New("simulated uncertain upload failure")
	}
	return nil
}
func (s *testTripImageStorage) ReadObject(ctx context.Context, bucket, key string, limit int64) ([]byte, error) {
	s.reads++
	if s.onRead != nil {
		s.onRead()
	}
	_, s.bounded = ctx.Deadline()
	if s.failRead {
		return nil, errors.New("simulated read failure")
	}
	data, ok := s.objects[bucket+"/"+key]
	if !ok || int64(len(data)) > limit {
		return nil, errors.New("missing or oversized object")
	}
	return data, nil
}
func (s *testTripImageStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	s.deletes++
	_, s.bounded = ctx.Deadline()
	delete(s.objects, bucket+"/"+key)
	return nil
}

func storedHeroPath(stop map[string]interface{}) string {
	return stop["draft"].(map[string]interface{})["hero"].(map[string]interface{})["url"].(string)
}

func TestTripHeroObjectStorageKeepsPermissionsAndPublishedVersion(t *testing.T) {
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	storage := &testTripImageStorage{}
	f.service.ConfigureTripImageStorage(storage, "private-trip-images")
	stranger := uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?)", stranger).Error; err != nil {
		t.Fatal(err)
	}
	testUploadTripHero(t, f, stop, stop.Revision, stranger, 403)
	if storage.puts != 0 {
		t.Fatal("unauthorized upload reached storage")
	}
	uploaded := testUploadTripHero(t, f, stop, stop.Revision, f.owner, 200)["stop"].(map[string]interface{})
	path := storedHeroPath(uploaded)
	var media TripMedia
	if err := f.db.Where("stop_id = ?", stop.ID).First(&media).Error; err != nil {
		t.Fatal(err)
	}
	if len(media.Data) != 0 || media.StorageBucket != "private-trip-images" || !strings.HasPrefix(media.StorageKey, "trip-heroes/"+stop.ID.String()+"/") || !storage.bounded {
		t.Fatalf("expected bounded private object storage, got %+v", media)
	}
	if strings.Contains(path, "amazonaws") || strings.Contains(path, media.StorageKey) {
		t.Fatal("object key escaped through API")
	}
	testUploadTripHero(t, f, stop, stop.Revision, f.owner, 409)
	if storage.puts != 1 {
		t.Fatal("stale revision uploaded an object")
	}
	owner := testGetTripHero(t, f, path, f.owner, 200)
	if !bytes.Equal(owner.Body.Bytes(), storage.objects[media.StorageBucket+"/"+media.StorageKey]) {
		t.Fatal("stored pixels did not round trip")
	}
	reads := storage.reads
	testGetTripHero(t, f, path+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	if storage.reads != reads {
		t.Fatal("unauthorized draft read reached storage")
	}
	published := calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": uploaded["revision"]}, f.owner, "", 200)["stop"].(map[string]interface{})
	replacement := testUploadTripHero(t, f, stop, int(published["revision"].(float64)), f.owner, 200)["stop"].(map[string]interface{})
	replacementPath := storedHeroPath(replacement)
	testGetTripHero(t, f, path+"?stopToken="+stop.ShareToken, uuid.Nil, 200)
	testGetTripHero(t, f, replacementPath+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	newest := testUploadTripHero(t, f, stop, int(replacement["revision"].(float64)), f.owner, 200)["stop"].(map[string]interface{})
	if len(storage.objects) != 2 || storage.deletes != 1 {
		t.Fatal("superseded draft object was not removed while preserving published image")
	}
	testGetTripHero(t, f, replacementPath, f.owner, 404)
	calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": newest["revision"]}, f.owner, "", 200)
	if len(storage.objects) != 1 {
		t.Fatal("publication retained the previous published object")
	}
	testGetTripHero(t, f, path+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	newPath := storedHeroPath(newest)
	testGetTripHero(t, f, newPath+"?stopToken="+stop.ShareToken, uuid.Nil, 200)
	if err := f.db.Model(&Stop{}).Where("id = ?", stop.ID).Update("visibility", "private").Error; err != nil {
		t.Fatal(err)
	}
	reads = storage.reads
	testGetTripHero(t, f, newPath+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	if storage.reads != reads {
		t.Fatal("revoked audience still reached storage")
	}
	storage.failRead = true
	testGetTripHero(t, f, newPath, f.owner, 500)
}

func TestTripHeroRemovalAndTripDeletionCleanStoredObjects(t *testing.T) {
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	storage := &testTripImageStorage{}
	f.service.ConfigureTripImageStorage(storage, "private-trip-images")
	first := testUploadTripHero(t, f, stop, stop.Revision, f.owner, 200)["stop"].(map[string]interface{})
	published := calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": first["revision"]}, f.owner, "", 200)["stop"].(map[string]interface{})
	replacement := testUploadTripHero(t, f, stop, int(published["revision"].(float64)), f.owner, 200)["stop"].(map[string]interface{})
	draft := replacement["draft"].(map[string]interface{})
	draft["hero"] = nil
	draft["revision"] = replacement["revision"]
	removed := calendarResponse(t, f, "PUT", "/calendar/stops/"+stop.ID.String(), draft, f.owner, "", 200)["stop"].(map[string]interface{})
	if len(storage.objects) != 1 {
		t.Fatal("removing draft cover must preserve only the published object")
	}
	published = calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": removed["revision"]}, f.owner, "", 200)["stop"].(map[string]interface{})
	if len(storage.objects) != 0 {
		t.Fatal("publishing removal retained the old cover object")
	}
	final := testUploadTripHero(t, f, stop, int(published["revision"].(float64)), f.owner, 200)["stop"].(map[string]interface{})
	calendarResponse(t, f, "DELETE", "/calendar/stops/"+stop.ID.String(), gin.H{"revision": final["revision"], "confirm": true}, f.owner, "", 200)
	if len(storage.objects) != 0 {
		t.Fatal("trip deletion orphaned its cover object")
	}
}

func TestTripHeroRechecksAccessAfterStorageReadWithoutHoldingCalendarLock(t *testing.T) {
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	storage := &testTripImageStorage{}
	f.service.ConfigureTripImageStorage(storage, "private-trip-images")
	uploaded := testUploadTripHero(t, f, stop, stop.Revision, f.owner, 200)["stop"].(map[string]interface{})
	calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": uploaded["revision"]}, f.owner, "", 200)
	storage.onRead = func() {
		// A single DB connection detects accidental transaction locks during S3.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := f.db.WithContext(ctx).Model(&Stop{}).Where("id = ?", stop.ID).Update("visibility", "private").Error; err != nil {
			t.Fatal(err)
		}
	}
	testGetTripHero(t, f, storedHeroPath(uploaded)+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	guestToken, recipient := f.verify(t, "hero-reader@example.com")
	grant := Grant{ID: uuid.New(), StopID: &stop.ID, RecipientID: &recipient}
	if err := f.db.Create(&grant).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Model(&Stop{}).Where("id = ?", stop.ID).Update("visibility", "specific").Error; err != nil {
		t.Fatal(err)
	}
	storage.onRead = func() {
		if err := f.db.Table("tc_guest_tokens").Where("token_hash = ?", guestDigest(guestToken)).Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
			t.Fatal(err)
		}
	}
	testGetTripHero(t, f, storedHeroPath(uploaded), uuid.Nil, 404, guestToken)
}

func TestTripHeroStorageFailurePreservesCurrentHeroAndCleansNewObject(t *testing.T) {
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	// Images saved before S3 was enabled continue to work with the new service.
	legacy := testUploadTripHero(t, f, stop, stop.Revision, f.owner, 200)["stop"].(map[string]interface{})
	storage := &testTripImageStorage{failPut: true}
	f.service.ConfigureTripImageStorage(storage, "private-trip-images")
	testGetTripHero(t, f, storedHeroPath(legacy), f.owner, 200)
	if storage.reads != 0 {
		t.Fatal("legacy image incorrectly fetched from S3")
	}
	revision := int(legacy["revision"].(float64))
	testUploadTripHero(t, f, stop, revision, f.owner, 500)
	if storage.deletes != 1 || len(storage.objects) != 0 {
		t.Fatal("uncertain failed upload object was not cleaned up")
	}
	var saved Stop
	if err := f.db.First(&saved, "id = ?", stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.Revision != revision || saved.Draft.Hero.URL != storedHeroPath(legacy) {
		t.Fatal("failed upload changed the trip")
	}
	// Also roll back the uploaded object if the database write fails afterwards.
	storage.failPut = false
	if err := f.db.Exec(`CREATE FUNCTION reject_trip_media() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'simulated database failure'; END $$; CREATE TRIGGER reject_trip_media BEFORE INSERT ON tc_trip_media FOR EACH ROW EXECUTE FUNCTION reject_trip_media();`).Error; err != nil {
		t.Fatal(err)
	}
	testUploadTripHero(t, f, stop, revision, f.owner, 500)
	if storage.deletes != 2 || len(storage.objects) != 0 {
		t.Fatal("database failure leaked uploaded object")
	}
	testGetTripHero(t, f, storedHeroPath(legacy), f.owner, 200)
}

func TestTripHeroStorageMigrationRollbackPreservesImages(t *testing.T) {
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	storage := &testTripImageStorage{}
	f.service.ConfigureTripImageStorage(storage, "private-trip-images")
	uploaded := testUploadTripHero(t, f, stop, stop.Revision, f.owner, 200)["stop"].(map[string]interface{})
	_, source, _, _ := runtime.Caller(0)
	down, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../migrate/internal/migrations/000478_travel_trip_media_storage.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.db.Exec(string(down)).Error; err == nil {
		t.Fatal("rollback discarded external image references")
	}
	testGetTripHero(t, f, storedHeroPath(uploaded), f.owner, 200)
}
