package travelcalendar

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func testTripPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(3, 2, color.RGBA{0, 0, 255, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
func TestTripMediaSanitizesMetadataAndOrientation(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 12, 8))
	var original bytes.Buffer
	if err := jpeg.Encode(&original, img, nil); err != nil {
		t.Fatal(err)
	}
	// EXIF little-endian IFD with one Orientation=6 record and private trailing
	// bytes. Pixel re-encoding must not retain any of the source metadata.
	tiff := []byte{'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 0x12, 1, 3, 0, 1, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0}
	payload := append([]byte("Exif\x00\x00"), tiff...)
	payload = append(payload, []byte("PRIVATE GPS COORDINATES")...)
	segment := []byte{0xff, 0xe1, 0, 0}
	binary.BigEndian.PutUint16(segment[2:], uint16(len(payload)+2))
	segment = append(segment, payload...)
	raw := append([]byte{}, original.Bytes()[:2]...)
	raw = append(raw, segment...)
	raw = append(raw, original.Bytes()[2:]...)
	if jpegExifOrientation(raw) != 6 {
		t.Fatal("orientation not read")
	}
	encoded, mime, width, height, err := sanitizeTripImage(raw)
	if err != nil {
		t.Fatal(err)
	}
	if mime != "image/jpeg" || width != 8 || height != 12 {
		t.Fatalf("incorrect oriented result: %s %dx%d", mime, width, height)
	}
	if bytes.Contains(encoded, []byte("Exif")) || bytes.Contains(encoded, []byte("PRIVATE GPS")) {
		t.Fatal("private image metadata survived")
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(encoded))
	if err != nil || config.Width != 8 || config.Height != 12 {
		t.Fatal("encoded dimensions did not rotate")
	}
	for _, malformed := range [][]byte{nil, []byte("not an image"), bytes.Repeat([]byte{0}, maxTripImageBytes+1)} {
		if _, _, _, _, err := sanitizeTripImage(malformed); err == nil {
			t.Fatal("invalid or oversized image accepted")
		}
	}
}
func TestTripMediaRejectsDecompressionSizeBeforeDecode(t *testing.T) {
	raw := testTripPNG(t)
	binary.BigEndian.PutUint32(raw[16:20], 50_000)
	binary.BigEndian.PutUint32(raw[20:24], 50_000)
	binary.BigEndian.PutUint32(raw[29:33], crc32.ChecksumIEEE(raw[12:29]))
	if _, _, _, _, err := sanitizeTripImage(raw); err == nil {
		t.Fatal("decompression-sized image accepted")
	}
}
func TestTripHeroCropValidation(t *testing.T) {
	hero := TripHero{MediaID: uuid.New(), AltText: "Lisbon rooftops", FocalX: .5, FocalY: .5, CropWidth: 1, CropHeight: 1}
	if err := hero.validate(); err != nil {
		t.Fatal(err)
	}
	hero.CropX = .8
	hero.CropWidth = .4
	if hero.validate() == nil {
		t.Fatal("crop outside image accepted")
	}
	hero.CropX = 0
	hero.CropWidth = 1
	hero.AltText = ""
	if hero.validate() == nil {
		t.Fatal("missing accessible description accepted")
	}
}
func testUploadTripHero(t *testing.T, f *guestTestFixture, stop Stop, revision int, user uuid.UUID, expected int) map[string]interface{} {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("image", "photo.png")
	if err != nil {
		t.Fatal(err)
	}
	file.Write(testTripPNG(t))
	writer.WriteField("revision", strconv.Itoa(revision))
	writer.WriteField("altText", "Rooftops in Lisbon")
	writer.WriteField("focalX", "0.4")
	writer.WriteField("focalY", "0.6")
	writer.Close()
	req := httptest.NewRequest("POST", "/travel-angels/calendar/stops/"+stop.ID.String()+"/hero", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if user != uuid.Nil {
		req.Header.Set("X-Test-User", user.String())
	}
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, req)
	if recorder.Code != expected {
		t.Fatalf("upload wanted %d got %d: %s", expected, recorder.Code, recorder.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}
func testGetTripHero(t *testing.T, f *guestTestFixture, path string, user uuid.UUID, expected int, guestToken ...string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	if user != uuid.Nil {
		req.Header.Set("X-Test-User", user.String())
	}
	if len(guestToken) > 0 {
		req.Header.Set(guestHeader, guestToken[0])
	}
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, req)
	if recorder.Code != expected {
		t.Fatalf("GET %s wanted %d got %d: %s", path, expected, recorder.Code, recorder.Body.String())
	}
	return recorder
}
func TestTripHeroHTTPFollowsPublishedAttachmentAndCurrentAccess(t *testing.T) {
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	stop.HostModel = true
	stop.Revision = 1
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	stranger := uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?)", stranger).Error; err != nil {
		t.Fatal(err)
	}
	testUploadTripHero(t, f, stop, 1, uuid.Nil, 401)
	testUploadTripHero(t, f, stop, 1, stranger, 403)
	uploaded := testUploadTripHero(t, f, stop, 1, f.owner, 200)["stop"].(map[string]interface{})
	draft := uploaded["draft"].(map[string]interface{})
	hero := draft["hero"].(map[string]interface{})
	path := hero["url"].(string)
	testUploadTripHero(t, f, stop, 1, f.owner, 409)
	ownerImage := testGetTripHero(t, f, path, f.owner, 200)
	if ownerImage.Header().Get("Cache-Control") != "private, no-store, max-age=0" || ownerImage.Header().Get("Content-Type") != "image/png" {
		t.Fatal("hero must not be publicly cached")
	}
	testGetTripHero(t, f, path+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	published := calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": uploaded["revision"]}, f.owner, "", 200)["stop"].(map[string]interface{})
	testGetTripHero(t, f, path, uuid.Nil, 404)
	testGetTripHero(t, f, path+"?stopToken="+stop.ShareToken, uuid.Nil, 200)
	replacement := testUploadTripHero(t, f, stop, int(published["revision"].(float64)), f.owner, 200)["stop"].(map[string]interface{})
	newPath := replacement["draft"].(map[string]interface{})["hero"].(map[string]interface{})["url"].(string)
	testGetTripHero(t, f, path+"?stopToken="+stop.ShareToken, uuid.Nil, 200)
	testGetTripHero(t, f, newPath+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": replacement["revision"]}, f.owner, "", 200)
	testGetTripHero(t, f, path+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	testGetTripHero(t, f, newPath+"?stopToken="+stop.ShareToken, uuid.Nil, 200)
	// A raw image URL grants nothing after publication is withdrawn or access
	// narrowed, even if it was successfully loaded in an earlier request.
	if err := f.db.Model(&Stop{}).Where("id = ?", stop.ID).Update("visibility", "private").Error; err != nil {
		t.Fatal(err)
	}
	testGetTripHero(t, f, newPath+"?stopToken="+stop.ShareToken, uuid.Nil, 404)
	testGetTripHero(t, f, newPath, f.owner, 200)
	guestToken, recipient := f.verify(t, "image-viewer@example.com")
	guestGrant := Grant{ID: uuid.New(), StopID: &stop.ID, RecipientID: &recipient}
	if err := f.db.Create(&guestGrant).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Model(&Stop{}).Where("id = ?", stop.ID).Update("visibility", "specific").Error; err != nil {
		t.Fatal(err)
	}
	testGetTripHero(t, f, newPath, uuid.Nil, 200, guestToken)
	if err := f.db.Delete(&guestGrant).Error; err != nil {
		t.Fatal(err)
	}
	testGetTripHero(t, f, newPath, uuid.Nil, 404, guestToken)
	otherStop := Stop{ID: uuid.New(), CalendarID: stop.CalendarID, Draft: stop.Draft, Visibility: "link", SharingEnabled: true, ShareToken: newToken()}
	if err := f.db.Create(&otherStop).Error; err != nil {
		t.Fatal(err)
	}
	foreign := TripHero{MediaID: uuid.MustParse(hero["mediaId"].(string)), AltText: "Foreign media", CropWidth: 1, CropHeight: 1}
	content := stop.Draft
	content.Hero = &foreign
	if err := f.service.validateTripContentReferences(f.db, &otherStop, &content); err == nil {
		t.Fatal("another trip attached protected media")
	}
}

func TestTripHeroHTTPRetainsOnlyCurrentParticipantCapabilities(t *testing.T) {
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	uploaded := testUploadTripHero(t, f, stop, stop.Revision, f.owner, 200)["stop"].(map[string]interface{})
	path := uploaded["draft"].(map[string]interface{})["hero"].(map[string]interface{})["url"].(string)
	calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": uploaded["revision"]}, f.owner, "", 200)
	friend := uuid.New()
	if err := f.db.Exec("INSERT INTO users(id) VALUES (?)", friend).Error; err != nil {
		t.Fatal(err)
	}
	requested := calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/participation", gin.H{"status": "requested", "startDate": stop.Draft.StartDate, "endDate": stop.Draft.EndDate, "stopToken": stop.ShareToken, "rosterConsentVersion": stop.RosterVersion}, friend, "", 200)
	if err := f.db.First(&stop, "id = ?", stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "PUT", "/calendar/stops/"+stop.ID.String()+"/participations/"+requested["id"].(string), gin.H{"status": "joined", "revision": stop.Revision}, f.owner, "", 200)
	testGetTripHero(t, f, path, friend, 200)
	if err := f.db.First(&stop, "id = ?", stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/share-link", gin.H{"revision": stop.Revision, "enabled": true, "rotate": true, "confirmRemoval": true}, f.owner, "", 200)
	testGetTripHero(t, f, path, friend, 404)
	// A separately granted current audience still authorizes the image; the old
	// saved link itself must never fabricate a new capability.
	grant := Grant{ID: uuid.New(), StopID: &stop.ID, UserID: &friend}
	if err := f.db.Create(&grant).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Model(&Stop{}).Where("id = ?", stop.ID).Update("visibility", "specific").Error; err != nil {
		t.Fatal(err)
	}
	testGetTripHero(t, f, path, friend, 200)
}
