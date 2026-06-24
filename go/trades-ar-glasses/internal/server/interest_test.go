package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeLeadStore struct {
	leads map[string]models.TradesARGlassesLead
}

func (f *fakeLeadStore) CreateOrGetByEmail(ctx context.Context, lead *models.TradesARGlassesLead) (bool, error) {
	if existing, ok := f.leads[lead.Email]; ok {
		*lead = existing
		return false, nil
	}
	if lead.ID == uuid.Nil {
		lead.ID = uuid.New()
	}
	if lead.CreatedAt.IsZero() {
		lead.CreatedAt = time.Now()
	}
	lead.UpdatedAt = time.Now()
	f.leads[lead.Email] = *lead
	return true, nil
}

func (f *fakeLeadStore) ListRecent(ctx context.Context, limit int) ([]models.TradesARGlassesLead, error) {
	return nil, nil
}

func TestCreateInterestLead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeLeadStore{leads: map[string]models.TradesARGlassesLead{}}
	router := gin.New()
	newServerFromLeadStore(store).SetupRoutes(router)

	body := bytes.NewBufferString(`{"email":"CrewLead@Example.COM","trade":"HVAC","crewSize":"11-50","source":"hero"}`)
	req := httptest.NewRequest(http.MethodPost, "/trades-ar-glasses/interest", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "unit-test")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("response JSON: %v", err)
	}
	if response["created"] != true {
		t.Fatalf("expected created=true, got %#v", response["created"])
	}

	lead := store.leads["crewlead@example.com"]
	if lead.Email != "crewlead@example.com" {
		t.Fatalf("expected normalized email, got %q", lead.Email)
	}
	if lead.Trade != "HVAC" || lead.CrewSize != "11-50" || lead.Source != "hero" || lead.UserAgent != "unit-test" {
		t.Fatalf("unexpected saved lead: %#v", lead)
	}
}

func TestCreateInterestLeadDuplicate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeLeadStore{leads: map[string]models.TradesARGlassesLead{}}
	router := gin.New()
	newServerFromLeadStore(store).SetupRoutes(router)

	for i, expectedStatus := range []int{http.StatusCreated, http.StatusOK} {
		body := bytes.NewBufferString(`{"email":"crewlead@example.com"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/interest", body)
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		if recorder.Code != expectedStatus {
			t.Fatalf("request %d: expected status %d, got %d: %s", i+1, expectedStatus, recorder.Code, recorder.Body.String())
		}
	}
}

func TestCreateInterestLeadRejectsInvalidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeLeadStore{leads: map[string]models.TradesARGlassesLead{}}
	router := gin.New()
	newServerFromLeadStore(store).SetupRoutes(router)

	body := bytes.NewBufferString(`{"email":"not-an-email"}`)
	req := httptest.NewRequest(http.MethodPost, "/trades-ar-glasses/interest", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}
