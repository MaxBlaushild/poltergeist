package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/googledrive"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// These transport tests capture the repository command and deliberately inject
// a write failure. Actual SQL/response/privacy behavior is covered by the Library
// PostgreSQL suite; this checks legacy routes before any persistence can occur.
type documentCommandCapture struct {
	db.DocumentHandle
	existing *models.Document
	created  *models.Document
	updated  *models.Document
}

func (h *documentCommandCapture) Create(_ context.Context, document *models.Document, _ []uuid.UUID, _ []string) (*models.Document, error) {
	copy := *document
	h.created = &copy
	return nil, errors.New("injected write failure")
}
func (h *documentCommandCapture) FindByID(_ context.Context, _ uuid.UUID) (*models.Document, error) {
	copy := *h.existing
	return &copy, nil
}
func (h *documentCommandCapture) Update(_ context.Context, document *models.Document) error {
	copy := *document
	h.updated = &copy
	return errors.New("injected write failure")
}

type documentCaptureDB struct {
	db.DbClient
	documents *documentCommandCapture
}

func (d documentCaptureDB) Document() db.DocumentHandle { return d.documents }

type documentDrive struct {
	googledrive.Client
	exported bool
}

func (d *documentDrive) GetFile(_ context.Context, _, _ string) (*googledrive.File, error) {
	return &googledrive.File{Name: "Original sheet", MimeType: "application/vnd.google-apps.spreadsheet", WebViewLink: "https://docs.google.com/spreadsheets/d/one"}, nil
}
func (d *documentDrive) ExportFileAsHTML(_ context.Context, _, _ string) ([]byte, error) {
	d.exported = true
	return []byte("<h1>Travel budget</h1><p>Train fares</p>"), nil
}
func documentCommandRequest(t *testing.T, handler gin.HandlerFunc, user uuid.UUID, body interface{}, id uuid.UUID, status int) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/:id", func(c *gin.Context) { c.Set("user", &models.User{ID: user}); handler(c) })
	req := httptest.NewRequest("POST", "/"+id.String(), strings.NewReader(string(data)))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("got %d want %d: %s", response.Code, status, response.Body.String())
	}
}
func TestDocumentCreateRejectsImplicitSharingAndUnsafeSourceURLs(t *testing.T) {
	owner := uuid.New()
	for _, provider := range []string{"internal", "google_docs", "google_sheets"} {
		capture := &documentCommandCapture{}
		s := &server{dbClient: documentCaptureDB{documents: capture}}
		documentCommandRequest(t, s.CreateDocument, owner, gin.H{"title": "My plan", "provider": provider, "format": "itinerary", "visibility": "network", "userId": uuid.New(), "personalNote": "  My recommendation  ", "destination": "  Paris  "}, uuid.New(), 500)
		if capture.created == nil || capture.created.Visibility != "private" || capture.created.UserID != owner || capture.created.Format != "itinerary" || capture.created.PersonalNote != "My recommendation" || capture.created.Destination != "Paris" {
			t.Fatalf("unsafe create command: %+v", capture.created)
		}
	}
	capture := &documentCommandCapture{}
	s := &server{dbClient: documentCaptureDB{documents: capture}}
	documentCommandRequest(t, s.CreateDocument, owner, gin.H{"title": "Bad source", "provider": "internal", "link": "javascript:alert(1)"}, uuid.New(), 400)
	if capture.created != nil {
		t.Fatal("unsafe URL reached persistence")
	}
	documentCommandRequest(t, s.CreateDocument, owner, gin.H{"title": "Bad format", "provider": "internal", "format": "unsupported"}, uuid.New(), 400)
	if capture.created != nil {
		t.Fatal("unsupported format reached persistence")
	}
}
func TestDriveReferenceAndSnapshotBothDefaultPrivateAndKeepMetadata(t *testing.T) {
	for _, kind := range []string{"reference", "import"} {
		capture := &documentCommandCapture{}
		drive := &documentDrive{}
		s := &server{dbClient: documentCaptureDB{documents: capture}, googleDriveClient: drive}
		documentCommandRequest(t, s.ImportGoogleDriveDocument, uuid.New(), gin.H{"fileId": "one", "importType": kind, "title": "Custom budget", "personalNote": "Useful for trains", "destination": "Japan", "visibility": "network"}, uuid.New(), 500)
		doc := capture.created
		if doc == nil || doc.Title != "Custom budget" || doc.Visibility != "private" || doc.Format != "spreadsheet" || doc.PersonalNote != "Useful for trains" || doc.Destination != "Japan" || doc.Link == nil {
			t.Fatalf("wrong import command: %+v", doc)
		}
		if kind == "reference" && (doc.Content != nil || drive.exported) {
			t.Fatal("reference unexpectedly copied source")
		}
		if kind == "import" && (doc.Content == nil || !strings.Contains(*doc.Content, "Train fares") || !drive.exported) {
			t.Fatal("snapshot content missing")
		}
	}
}
func TestDocumentMetadataEditRequiresOwnerAndCannotChangeAudience(t *testing.T) {
	owner, other, id := uuid.New(), uuid.New(), uuid.New()
	capture := &documentCommandCapture{existing: &models.Document{ID: id, UserID: owner, Title: "Original", Visibility: "selected", Format: "guide"}}
	s := &server{dbClient: documentCaptureDB{documents: capture}}
	changes := gin.H{"title": "Revised", "format": "itinerary", "personalNote": "A note", "destination": "Rome", "visibility": "network"}
	documentCommandRequest(t, s.UpdateDocument, other, changes, id, 403)
	if capture.updated != nil {
		t.Fatal("nonowner edit reached persistence")
	}
	documentCommandRequest(t, s.UpdateDocument, owner, changes, id, 500)
	if capture.updated == nil || capture.updated.Title != "Revised" || capture.updated.Visibility != "selected" || capture.updated.Format != "itinerary" || capture.updated.PersonalNote != "A note" || capture.updated.Destination != "Rome" {
		t.Fatalf("wrong update command: %+v", capture.updated)
	}
}
