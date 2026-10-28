package travelcalendar

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func tripDocumentsFixture(t *testing.T) (*guestTestFixture, Stop) {
	t.Helper()
	f := newGuestTestFixture(t)
	_, stop := f.plans(t)
	stop.HostModel = true
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	return f, stop
}

func tripDocumentsPath(stop Stop) string {
	return "/calendar/stops/" + stop.ID.String() + "/documents"
}

func tripDocumentLink(t *testing.T, f *guestTestFixture, stop Stop, user uuid.UUID) map[string]interface{} {
	t.Helper()
	return calendarResponse(t, f, "POST", tripDocumentsPath(stop), gin.H{"title": "Shared itinerary", "link": "https://docs.example.test/itinerary"}, user, "", 201)["document"].(map[string]interface{})
}

func TestTripDocumentsAccessAndCurrentAttendance(t *testing.T) {
	f, stop := tripDocumentsFixture(t)
	attendee := flightTraveler(t, f, stop)
	document := tripDocumentLink(t, f, stop, attendee)
	base := tripDocumentsPath(stop)
	detail := base + "/" + document["id"].(string)
	viewer := tripAuthorityAccount(t, f)
	for _, user := range []uuid.UUID{uuid.Nil, viewer} {
		state := calendarResponse(t, f, "GET", base+"?stopToken="+stop.ShareToken, nil, user, "", 200)
		if state["canView"] != false || state["canContribute"] != false || len(state["documents"].([]interface{})) != 0 {
			t.Fatal("trip viewer received attendees' shared documents", state)
		}
		calendarResponse(t, f, "GET", detail+"?stopToken="+stop.ShareToken, nil, user, "", 404)
		calendarResponse(t, f, "GET", detail+"/comments?stopToken="+stop.ShareToken, nil, user, "", 404)
	}
	calendarResponse(t, f, "POST", base, gin.H{"title": "x", "link": "https://example.test"}, uuid.Nil, "", 401)
	calendarResponse(t, f, "POST", detail+"/comments?stopToken="+stop.ShareToken, gin.H{"body": "Guest comment"}, uuid.Nil, "", 401)
	calendarResponse(t, f, "POST", base+"?stopToken="+stop.ShareToken, gin.H{"title": "x", "link": "https://example.test"}, viewer, "", 404)
	for _, user := range []uuid.UUID{f.owner, attendee} {
		state := calendarResponse(t, f, "GET", base, nil, user, "", 200)
		if state["canView"] != true || state["canContribute"] != true || len(state["documents"].([]interface{})) != 1 {
			t.Fatal("host or attendee cannot read shared documents")
		}
		trip := calendarResponse(t, f, "GET", "/calendar/stops/"+stop.ID.String(), nil, user, "", 200)["stop"].(map[string]interface{})
		if trip["canViewDocuments"] != true || trip["canContributeDocuments"] != true {
			t.Fatal("trip capability flags disagree with document access")
		}
	}
	// A pending request without a reserved place is not an attendee.
	f.db.Model(&Participation{}).Where("stop_id = ? AND user_id = ?", stop.ID, attendee).Updates(map[string]interface{}{"status": "requested", "reserved": false})
	calendarResponse(t, f, "GET", detail, nil, attendee, "", 404)
	// A reserved place awaiting date reconfirmation retains private logistics access.
	f.db.Model(&Participation{}).Where("stop_id = ? AND user_id = ?", stop.ID, attendee).Updates(map[string]interface{}{"status": "needs_reconfirmation", "reserved": true})
	calendarResponse(t, f, "GET", detail, nil, attendee, "", 200)
	f.db.Model(&Participation{}).Where("stop_id = ? AND user_id = ?", stop.ID, attendee).Updates(map[string]interface{}{"status": "withdrawn", "reserved": false})
	calendarResponse(t, f, "GET", detail+"?stopToken="+stop.ShareToken, nil, attendee, "", 404)
	calendarResponse(t, f, "DELETE", detail+"?stopToken="+stop.ShareToken, nil, attendee, "", 404)
	// The host can still remove a former attendee's contribution.
	calendarResponse(t, f, "DELETE", detail, nil, f.owner, "", 200)
}

func TestTripDocumentsOwnLibrarySnapshotAndPublishIndependence(t *testing.T) {
	f, stop := tripDocumentsFixture(t)
	attendee := flightTraveler(t, f, stop)
	if err := f.db.Exec(`CREATE TABLE documents (
 id uuid PRIMARY KEY, user_id uuid NOT NULL, title text NOT NULL, provider text NOT NULL DEFAULT 'internal',
 link text, content text, created_at timestamptz NOT NULL DEFAULT now()
 ); CREATE TABLE trending_destinations (id uuid PRIMARY KEY);`).Error; err != nil {
		t.Fatal(err)
	}
	_, source, _, _ := runtime.Caller(0)
	migration, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../migrate/internal/migrations/000476_travel_library.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.db.Exec(string(migration)).Error; err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	if err = f.db.Exec("INSERT INTO documents(id,user_id,title,content,visibility,personal_note,format) VALUES (?,?,?,?,?,?,?)", id, f.owner, "Private source", "# Shared snapshot", "private", "PRIVATE PERSONAL NOTE", "itinerary").Error; err != nil {
		t.Fatal(err)
	}
	if err = f.db.Exec("INSERT INTO document_trip_references(document_id,user_id,trip_id) VALUES (?,?,?)", id, f.owner, stop.ID).Error; err != nil {
		t.Fatal(err)
	}
	base := tripDocumentsPath(stop)
	calendarResponse(t, f, "POST", base, gin.H{"sourceDocumentId": id}, attendee, "", 404)
	// Even explicit library sharing is not permission for the reader to reshare.
	f.db.Exec("INSERT INTO document_shares(document_id,user_id) VALUES (?,?)", id, attendee)
	f.db.Exec("UPDATE documents SET visibility='selected' WHERE id=?", id)
	calendarResponse(t, f, "POST", base, gin.H{"sourceDocumentId": id}, attendee, "", 404)
	f.db.Exec("UPDATE documents SET visibility='private' WHERE id=?", id)
	copy := calendarResponse(t, f, "POST", base, gin.H{"sourceDocumentId": id}, f.owner, "", 201)["document"].(map[string]interface{})
	copyJSON, _ := json.Marshal(copy)
	if copy["content"] != "# Shared snapshot" || copy["format"] != "itinerary" || strings.Contains(string(copyJSON), "PRIVATE PERSONAL NOTE") || strings.Contains(string(copyJSON), id.String()) {
		t.Fatal("snapshot copied private metadata or omitted content", copy)
	}
	var visibility, content string
	f.db.Table("documents").Select("visibility").Where("id=?", id).Scan(&visibility)
	f.db.Table("documents").Select("content").Where("id=?", id).Scan(&content)
	var references, shares int64
	f.db.Table("document_trip_references").Count(&references)
	f.db.Table("document_shares").Count(&shares)
	if visibility != "private" || content != "# Shared snapshot" || references != 1 || shares != 1 {
		t.Fatal("sharing a trip copy changed the source audience or personal bookmarks")
	}
	detail := base + "/" + copy["id"].(string)
	calendarResponse(t, f, "POST", detail+"/comments", gin.H{"body": "Let's take this train."}, attendee, "", 201)
	calendarResponse(t, f, "POST", "/calendar/stops/"+stop.ID.String()+"/publish", gin.H{"revision": stop.Revision}, f.owner, "", 200)
	f.db.Exec("DELETE FROM documents WHERE id=?", id)
	read := calendarResponse(t, f, "GET", detail, nil, attendee, "", 200)["document"].(map[string]interface{})
	if read["content"] != "# Shared snapshot" || read["commentCount"] != float64(1) {
		t.Fatal("host publication or source deletion overwrote independent shared copy")
	}
}

func TestTripDocumentCommentsOwnershipScopePaginationAndCancellation(t *testing.T) {
	f, stop := tripDocumentsFixture(t)
	a, b := flightTraveler(t, f, stop), flightTraveler(t, f, stop)
	document := tripDocumentLink(t, f, stop, a)
	detail := tripDocumentsPath(stop) + "/" + document["id"].(string)
	comments := detail + "/comments"
	created := calendarResponse(t, f, "POST", comments, gin.H{"body": "A useful suggestion", "authorId": b}, a, "", 201)["comment"].(map[string]interface{})
	if created["author"].(map[string]interface{})["id"] != a.String() {
		t.Fatal("comment author spoofed")
	}
	commentPath := comments + "/" + created["id"].(string)
	calendarResponse(t, f, "PUT", commentPath, gin.H{"body": "edited"}, b, "", 403)
	calendarResponse(t, f, "PUT", commentPath, gin.H{"body": "host edit"}, f.owner, "", 403)
	calendarResponse(t, f, "DELETE", commentPath, nil, b, "", 403)
	calendarResponse(t, f, "DELETE", detail, nil, b, "", 403)
	calendarResponse(t, f, "PUT", commentPath, gin.H{"body": "Author edit"}, a, "", 200)
	other := tripDocumentLink(t, f, stop, f.owner)
	calendarResponse(t, f, "PUT", tripDocumentsPath(stop)+"/"+other["id"].(string)+"/comments/"+created["id"].(string), gin.H{"body": "wrong document"}, a, "", 404)
	anotherStop := stop
	anotherStop.ID, anotherStop.ShareToken = uuid.New(), newToken()
	if err := f.db.Create(&anotherStop).Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "GET", tripDocumentsPath(anotherStop)+"/"+document["id"].(string), nil, f.owner, "", 404)
	calendarResponse(t, f, "GET", tripDocumentsPath(anotherStop)+"/"+document["id"].(string)+"/comments", nil, f.owner, "", 404)
	for _, body := range []string{"", strings.Repeat("a", 4001)} {
		calendarResponse(t, f, "POST", comments, gin.H{"body": body}, a, "", 400)
	}
	second := calendarResponse(t, f, "POST", comments, gin.H{"body": "Second"}, b, "", 201)["comment"].(map[string]interface{})
	third := calendarResponse(t, f, "POST", comments, gin.H{"body": "Third"}, b, "", 201)["comment"].(map[string]interface{})
	page := calendarResponse(t, f, "GET", comments+"?limit=2", nil, a, "", 200)
	rows := page["comments"].([]interface{})
	if page["hasMore"] != true || page["nextBefore"] != second["id"] || len(rows) != 2 || rows[0].(map[string]interface{})["id"] != second["id"] || rows[1].(map[string]interface{})["id"] != third["id"] {
		t.Fatal("comment pagination order/cursor incorrect", page)
	}
	older := calendarResponse(t, f, "GET", comments+"?limit=2&before="+page["nextBefore"].(string), nil, a, "", 200)
	if older["hasMore"] != false || len(older["comments"].([]interface{})) != 1 {
		t.Fatal("older comments missing")
	}
	stop.Published.Status = "cancelled"
	f.db.Save(&stop)
	read := calendarResponse(t, f, "GET", detail, nil, a, "", 200)["document"].(map[string]interface{})
	if read["canContribute"] != false || read["canDelete"] != true {
		t.Fatal("cancelled trip lost delete or retained add capability")
	}
	calendarResponse(t, f, "POST", tripDocumentsPath(stop), gin.H{"title": "cancelled", "link": "https://example.test"}, a, "", 403)
	calendarResponse(t, f, "POST", comments, gin.H{"body": "new"}, a, "", 403)
	calendarResponse(t, f, "PUT", commentPath, gin.H{"body": "edit"}, a, "", 403)
	calendarResponse(t, f, "DELETE", commentPath, nil, f.owner, "", 200)
	calendarResponse(t, f, "DELETE", detail, nil, a, "", 200)
	var remaining int64
	f.db.Model(&TripDocumentComment{}).Where("document_id=?", document["id"]).Count(&remaining)
	if remaining != 0 {
		t.Fatal("document removal retained orphaned comments")
	}
	calendarResponse(t, f, "GET", comments, nil, f.owner, "", 404)
}

func tripDocumentUpload(t *testing.T, f *guestTestFixture, stop Stop, user uuid.UUID, filename string, data []byte, expected int) map[string]interface{} {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	file.Write(data)
	writer.WriteField("title", "Original file")
	writer.Close()
	request := httptest.NewRequest("POST", "/travel-angels"+tripDocumentsPath(stop)+"/upload?stopToken="+stop.ShareToken, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if user != uuid.Nil {
		request.Header.Set("X-Test-User", user.String())
	}
	response := httptest.NewRecorder()
	f.router.ServeHTTP(response, request)
	if response.Code != expected {
		t.Fatalf("upload expected %d, got %d: %s", expected, response.Code, response.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func tripDocumentDownload(t *testing.T, f *guestTestFixture, url string, user uuid.UUID, expected int) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest("GET", url, nil)
	if user != uuid.Nil {
		request.Header.Set("X-Test-User", user.String())
	}
	response := httptest.NewRecorder()
	f.router.ServeHTTP(response, request)
	if response.Code != expected {
		t.Fatalf("download expected %d, got %d: %s", expected, response.Code, response.Body.String())
	}
	return response
}

func TestTripDocumentFilesPrivateStorageRevocationAndCleanup(t *testing.T) {
	f, stop := tripDocumentsFixture(t)
	a := flightTraveler(t, f, stop)
	storage := &testTripImageStorage{}
	f.service.ConfigureTripImageStorage(storage, "private-trip-media")
	data := []byte("Train itinerary\nTokyo to Kyoto\n")
	tripDocumentUpload(t, f, stop, uuid.Nil, "plan.txt", data, 401)
	stranger := tripAuthorityAccount(t, f)
	tripDocumentUpload(t, f, stop, stranger, "plan.txt", data, 404)
	if storage.puts != 0 {
		t.Fatal("unauthorized upload reached storage")
	}
	created := tripDocumentUpload(t, f, stop, a, "plan.txt", data, 201)["document"].(map[string]interface{})
	url := created["fileUrl"].(string)
	serialized, _ := json.Marshal(created)
	if strings.Contains(string(serialized), "private-trip-media") || strings.Contains(string(serialized), "trip-documents/") || !storage.bounded {
		t.Fatal("storage details leaked or IO unbounded")
	}
	for _, user := range []uuid.UUID{uuid.Nil, stranger} {
		tripDocumentDownload(t, f, url+"?stopToken="+stop.ShareToken, user, 404)
	}
	if storage.reads != 0 {
		t.Fatal("unauthorized download reached storage")
	}
	read := tripDocumentDownload(t, f, url, a, 200)
	if !bytes.Equal(read.Body.Bytes(), data) || read.Header().Get("X-Content-Type-Options") != "nosniff" || !strings.HasPrefix(read.Header().Get("Content-Disposition"), "attachment;") || !strings.Contains(read.Header().Get("Cache-Control"), "no-store") {
		t.Fatal("original bytes or private download headers incorrect")
	}
	storage.onRead = func() {
		if err := f.db.Model(&Stop{}).Where("id=?", stop.ID).Update("share_token", newToken()).Error; err != nil {
			t.Fatal(err)
		}
	}
	denied := tripDocumentDownload(t, f, url+"?stopToken="+stop.ShareToken, a, 404)
	if bytes.Contains(denied.Body.Bytes(), data) {
		t.Fatal("revoked storage read leaked protected bytes")
	}
	storage.onRead = nil
	// Failed/uncertain writes are deleted and never become listable rows.
	storage.failPut = true
	tripDocumentUpload(t, f, stop, f.owner, "plan.txt", data, 500)
	var count int64
	f.db.Model(&TripDocument{}).Count(&count)
	if count != 1 || len(storage.objects) != 1 || storage.deletes != 1 {
		t.Fatal("failed upload left a document row or object")
	}
	storage.failPut = false
	calendarResponse(t, f, "DELETE", tripDocumentsPath(stop)+"/"+created["id"].(string), nil, f.owner, "", 200)
	if len(storage.objects) != 0 {
		t.Fatal("removing document left original object")
	}
	tripDocumentUpload(t, f, stop, f.owner, "plan.txt", data, 201)
	calendarResponse(t, f, "DELETE", "/calendar/stops/"+stop.ID.String(), gin.H{"confirm": true, "revision": stop.Revision}, f.owner, "", 200)
	if len(storage.objects) != 0 {
		t.Fatal("trip deletion left original document object")
	}
}

func TestTripDocumentFileValidation(t *testing.T) {
	for _, test := range []struct {
		name string
		data []byte
	}{
		{"active.html", []byte("<html>private</html>")},
		{"disguised.txt", []byte("<html>private</html>")},
		{"wrong.pdf", []byte("not a pdf")},
		{"wrong.png", []byte("not an image")},
		{"binary.csv", []byte{0xff, 0, 1}},
		{"big.txt", bytes.Repeat([]byte("x"), maxTripDocumentBytes+1)},
		{"empty.txt", nil},
	} {
		if _, _, err := validateTripDocumentFile(test.name, test.data); err == nil {
			t.Fatalf("accepted %s", test.name)
		}
	}
	for _, name := range []string{"itinerary.txt", "budget.csv", "plan.PDF"} {
		data := []byte("Tokyo,Kyoto\n20,30")
		if strings.HasSuffix(name, ".PDF") {
			data = []byte("%PDF-1.4\n%%EOF")
		}
		if _, _, err := validateTripDocumentFile(name, data); err != nil {
			t.Fatal(name, err)
		}
	}
	archive := func(macro bool) []byte {
		var body bytes.Buffer
		w := zip.NewWriter(&body)
		entry, _ := w.Create("[Content_Types].xml")
		entry.Write([]byte(`<Types><Override ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/></Types>`))
		entry, _ = w.Create("xl/workbook.xml")
		entry.Write([]byte("<workbook/>"))
		if macro {
			entry, _ = w.Create("xl/vbaProject.bin")
			entry.Write([]byte("macro"))
		}
		w.Close()
		return body.Bytes()
	}
	if contentType, format, err := validateTripDocumentFile("budget.xlsx", archive(false)); err != nil || !strings.Contains(contentType, "spreadsheetml") || format != "spreadsheet" {
		t.Fatal("valid spreadsheet rejected")
	}
	if _, _, err := validateTripDocumentFile("budget.xlsx", archive(true)); err == nil {
		t.Fatal("macro-enabled file disguised as XLSX accepted")
	}
	if got := tripDocumentFilename("../secret\r\nname.txt"); got != "secretname.txt" {
		t.Fatal("filename path/header normalization failed", got)
	}
	for _, link := range []string{"javascript:alert(1)", "file:///tmp/private", "https://user:password@example.test", "https://example.test\r\nHeader: invalid"} {
		if validateTripDocumentLink(link) == nil {
			t.Fatal("unsafe source link accepted")
		}
	}
}

func TestTripDocumentsCannotBorrowUnclaimedGuestGrant(t *testing.T) {
	f, stop := tripDocumentsFixture(t)
	attendee := flightTraveler(t, f, stop)
	document := tripDocumentLink(t, f, stop, f.owner)
	guestToken, recipientID := f.verify(t, "separate-reader@example.test")
	stop.Visibility = "specific"
	if err := f.db.Save(&stop).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Create(&Grant{ID: uuid.New(), StopID: &stop.ID, RecipientID: &recipientID}).Error; err != nil {
		t.Fatal(err)
	}
	base := tripDocumentsPath(stop)
	// Guest verification can view the trip, but cannot lend that access to an
	// unrelated claimed account which still has an old participation record.
	state := calendarResponse(t, f, "GET", base, nil, attendee, guestToken, 200)
	if state["canView"] != false || len(state["documents"].([]interface{})) != 0 {
		t.Fatal("unrelated guest grant unlocked attendee documents")
	}
	calendarResponse(t, f, "GET", base+"/"+document["id"].(string), nil, attendee, guestToken, 404)
	calendarResponse(t, f, "POST", base, gin.H{"title": "x", "link": "https://example.test"}, attendee, guestToken, 404)
	// Linking the recipient to this claimed account is a real account grant.
	if err := f.db.Model(&GuestRecipient{}).Where("id=?", recipientID).Update("user_id", attendee).Error; err != nil {
		t.Fatal(err)
	}
	calendarResponse(t, f, "GET", base+"/"+document["id"].(string), nil, attendee, "", 200)
}

type tripDocumentRevokingStorage struct {
	testTripImageStorage
	onPut func()
}

func (s *tripDocumentRevokingStorage) PutObject(ctx context.Context, bucket, key, contentType string, data []byte) error {
	err := s.testTripImageStorage.PutObject(ctx, bucket, key, contentType, data)
	if s.onPut != nil {
		s.onPut()
	}
	return err
}

func TestTripDocumentUploadRechecksAttendanceAfterStorage(t *testing.T) {
	f, stop := tripDocumentsFixture(t)
	attendee := flightTraveler(t, f, stop)
	storage := &tripDocumentRevokingStorage{onPut: func() {
		if err := f.db.Model(&Participation{}).Where("stop_id=? AND user_id=?", stop.ID, attendee).Update("status", "removed").Error; err != nil {
			t.Fatal(err)
		}
	}}
	f.service.ConfigureTripImageStorage(storage, "private-trip-media")
	tripDocumentUpload(t, f, stop, attendee, "plan.txt", []byte("Travel plans"), 404)
	var count int64
	f.db.Model(&TripDocument{}).Count(&count)
	if count != 0 || len(storage.objects) != 0 || storage.deletes != 1 {
		t.Fatal("revoked upload became a shared document or retained its object")
	}
}

func TestTripDocumentMigrationPreservesStoredFileReferences(t *testing.T) {
	f, stop := tripDocumentsFixture(t)
	attendee := flightTraveler(t, f, stop)
	storage := &testTripImageStorage{}
	f.service.ConfigureTripImageStorage(storage, "private-trip-media")
	document := tripDocumentUpload(t, f, stop, attendee, "plan.txt", []byte("Travel plans"), 201)["document"].(map[string]interface{})
	// Match the existing media uploader restriction: account deletion may not
	// bypass the API's private-object cleanup by cascading contributed files.
	if err := f.db.Exec("DELETE FROM users WHERE id=?", attendee).Error; err == nil {
		t.Fatal("account deletion orphaned a contributed private original")
	}
	_, source, _, _ := runtime.Caller(0)
	down, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../../migrate/internal/migrations/000481_travel_trip_documents.down.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.db.Exec(string(down)).Error; err == nil || !strings.Contains(err.Error(), "Remove uploaded trip documents") {
		t.Fatal("rollback did not preserve private object references", err)
	}
	calendarResponse(t, f, "GET", tripDocumentsPath(stop)+"/"+document["id"].(string), nil, attendee, "", 200)
	calendarResponse(t, f, "DELETE", tripDocumentsPath(stop)+"/"+document["id"].(string), nil, attendee, "", 200)
	if len(storage.objects) != 0 {
		t.Fatal("API deletion left an object before rollback")
	}
	if err = f.db.Exec(string(down)).Error; err != nil {
		t.Fatal("clean rollback failed", err)
	}
}
