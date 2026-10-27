package travelcalendar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Only this reference is stored in trip content. Google details and photos are
// obtained live, authorized per lodging audience, and never persisted here.
type TripLodgingProperty struct {
	Provider string `json:"provider"`
	PlaceID  string `json:"placeId"`
}

type LodgingPhotoAttribution struct {
	DisplayName string `json:"displayName"`
	URI         string `json:"uri,omitempty"`
	PhotoURI    string `json:"photoUri,omitempty"`
}
type LodgingPropertyPhoto struct {
	Name          string                    `json:"-"`
	Width         int                       `json:"width"`
	Height        int                       `json:"height"`
	URL           string                    `json:"url"`
	GoogleMapsURI string                    `json:"googleMapsUri,omitempty"`
	Attributions  []LodgingPhotoAttribution `json:"authorAttributions"`
}
type LodgingProperty struct {
	Provider     string                    `json:"provider"`
	PlaceID      string                    `json:"placeId"`
	Name         string                    `json:"name"`
	Address      string                    `json:"address"`
	WebsiteURL   string                    `json:"websiteUrl,omitempty"`
	MapsURL      string                    `json:"mapsUrl,omitempty"`
	Rating       float64                   `json:"rating,omitempty"`
	RatingCount  int                       `json:"ratingCount,omitempty"`
	Types        []string                  `json:"types"`
	Photos       []LodgingPropertyPhoto    `json:"photos"`
	Attributions []LodgingPhotoAttribution `json:"attributions"`
}
type LodgingPropertyProvider interface {
	Search(context.Context, string) ([]LodgingProperty, error)
	Details(context.Context, string) (*LodgingProperty, error)
	Photo(context.Context, string) ([]byte, string, error)
}

// Providers without geographical search keep the existing text-only behavior.
type LodgingPropertyNearProvider interface {
	SearchNear(context.Context, string, float64, float64) ([]LodgingProperty, error)
}

type lodgingSearchLocation struct{ latitude, longitude float64 }

func (s *Service) lodgingSearchBias(c *gin.Context) (*lodgingSearchLocation, error) {
	viewer := s.viewer(c)
	var bias *lodgingSearchLocation
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		stop, err := s.lockedStop(tx, c.Param("id"))
		if err != nil {
			return err
		}
		if !s.canManageStop(tx, stop, viewer) {
			return errForbidden
		}
		ids := []uuid.UUID{}
		for _, destination := range stop.Draft.Destinations {
			if destination.ID != nil {
				ids = append(ids, *destination.ID)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		// Read coordinates from trusted catalog rows rather than request fields
		// or trip-local labels. Preserve the host's ordered destination choices.
		var locations []Destination
		if err := tx.Select("id, latitude, longitude").Where("id IN ?", ids).Find(&locations).Error; err != nil {
			return err
		}
		byID := map[uuid.UUID]Destination{}
		for _, location := range locations {
			byID[location.ID] = location
		}
		for _, id := range ids {
			location := byID[id]
			if location.Latitude == nil || location.Longitude == nil {
				continue
			}
			latitude, longitude := *location.Latitude, *location.Longitude
			if math.IsNaN(latitude) || math.IsInf(latitude, 0) || math.IsNaN(longitude) || math.IsInf(longitude, 0) || math.Abs(latitude) > 90 || math.Abs(longitude) > 180 {
				continue
			}
			bias = &lodgingSearchLocation{latitude: latitude, longitude: longitude}
			break
		}
		return nil
	})
	return bias, err
}

func (s *Service) ConfigureLodgingPropertyProvider(provider LodgingPropertyProvider) {
	s.lodgingPropertyProvider = provider
}

func validLodgingPlaceID(id string) bool {
	if len(id) == 0 || len(id) > 512 {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func (s *Service) registerLodgingMediaRoutes(g *gin.RouterGroup) {
	g.GET("/calendar/stops/:id/lodging-properties", s.searchLodgingProperties)
	g.GET("/calendar/stops/:id/lodging-properties/:placeId", s.getLodgingProperty)
	g.GET("/calendar/stops/:id/lodging-properties/:placeId/photos/:index", s.getLodgingPropertyPhoto)
	g.GET("/calendar/stops/:id/lodging/:entryId/property", s.getLodgingProperty)
	g.GET("/calendar/stops/:id/lodging/:entryId/property/photos/:index", s.getLodgingPropertyPhoto)
	g.POST("/calendar/stops/:id/lodging/:entryId/photos", s.uploadLodgingPhoto)
	g.GET("/calendar/stops/:id/lodging/:entryId/photos/:mediaId", s.getLodgingPhoto)
}

func (s *Service) authorizedLodging(tx *gorm.DB, c *gin.Context, viewer Viewer) (*Stop, *TripLodging, error) {
	stop, err := s.lockedStop(tx, c.Param("id"))
	if err != nil {
		return nil, nil, err
	}
	viewer = s.currentTripMediaViewer(tx, stop, c, viewer)
	if !s.canViewStop(tx, stop, viewer) {
		return nil, nil, errUnavailable
	}
	manager := s.canManageStop(tx, stop, viewer)
	content := stop.Published
	if c.Query("draft") == "true" {
		if !manager {
			return nil, nil, errUnavailable
		}
		content = &stop.Draft
	}
	if content == nil {
		return nil, nil, errUnavailable
	}
	for i := range content.Lodging {
		entry := &content.Lodging[i]
		if entry.ID == c.Param("entryId") && canReadLogistics(entry.Visibility, manager, s.tripAttendee(tx, stop, viewer)) {
			return stop, entry, nil
		}
	}
	return nil, nil, errUnavailable
}

func (s *Service) lodgingPropertyAccess(tx *gorm.DB, c *gin.Context, viewer Viewer) (*TripLodgingProperty, error) {
	if c.Param("entryId") != "" {
		_, entry, err := s.authorizedLodging(tx, c, viewer)
		if err != nil {
			return nil, err
		}
		if entry.Property == nil {
			return nil, errUnavailable
		}
		return entry.Property, nil
	}
	stop, err := s.lockedStop(tx, c.Param("id"))
	if err != nil {
		return nil, err
	}
	if !s.canManageStop(tx, stop, viewer) {
		return nil, errForbidden
	}
	if c.Param("placeId") == "" {
		return nil, nil
	}
	if !validLodgingPlaceID(c.Param("placeId")) {
		return nil, bad("Choose a valid property")
	}
	return &TripLodgingProperty{Provider: "google", PlaceID: c.Param("placeId")}, nil
}

func (s *Service) propertyAccess(c *gin.Context) (*TripLodgingProperty, error) {
	viewer := s.viewer(c)
	var reference *TripLodgingProperty
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		var err error
		reference, err = s.lodgingPropertyAccess(tx, c, viewer)
		return err
	})
	return reference, err
}

func (s *Service) finishPropertyRequest(c *gin.Context, expected *TripLodgingProperty, respond func()) {
	viewer := s.viewer(c)
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		current, err := s.lodgingPropertyAccess(tx, c, viewer)
		if err != nil {
			return err
		}
		if (expected == nil) != (current == nil) || expected != nil && *expected != *current {
			return errUnavailable
		}
		respond()
		return nil
	})
	if err != nil {
		respondError(c, err)
	}
}

func (s *Service) propertyProviderAvailable(c *gin.Context) bool {
	if s.lodgingPropertyProvider != nil {
		return true
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Property search is temporarily unavailable. You can add lodging manually."})
	return false
}

func (s *Service) propertyPhotoURLs(c *gin.Context, property *LodgingProperty) {
	base := "/travel-angels/calendar/stops/" + c.Param("id") + "/lodging-properties/" + url.PathEscape(property.PlaceID)
	if c.Param("entryId") != "" {
		base = "/travel-angels/calendar/stops/" + c.Param("id") + "/lodging/" + c.Param("entryId") + "/property"
	}
	if len(property.Photos) > 5 {
		property.Photos = property.Photos[:5]
	}
	for i := range property.Photos {
		property.Photos[i].URL = base + "/photos/" + strconv.Itoa(i)
		query := url.Values{"photoVersion": []string{lodgingPhotoVersion(property.Photos[i].Name)}}
		if c.Query("draft") == "true" {
			query.Set("draft", "true")
		}
		property.Photos[i].URL += "?" + query.Encode()
	}
}

func lodgingPhotoVersion(name string) string {
	digest := sha256.Sum256([]byte(name))
	return hex.EncodeToString(digest[:])
}

func (s *Service) searchLodgingProperties(c *gin.Context) {
	bias, err := s.lodgingSearchBias(c)
	if err != nil {
		respondError(c, err)
		return
	}
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 {
		c.JSON(200, gin.H{"properties": []LodgingProperty{}})
		return
	}
	if len(query) > 300 {
		respondError(c, bad("Keep property searches under 300 characters"))
		return
	}
	if !s.propertyProviderAvailable(c) {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	var properties []LodgingProperty
	if nearby, ok := s.lodgingPropertyProvider.(LodgingPropertyNearProvider); ok && bias != nil {
		properties, err = nearby.SearchNear(ctx, query, bias.latitude, bias.longitude)
	} else {
		properties, err = s.lodgingPropertyProvider.Search(ctx, query)
	}
	cancel()
	if err != nil {
		c.JSON(503, gin.H{"error": "Property search is temporarily unavailable. You can add lodging manually."})
		return
	}
	if len(properties) > 10 {
		properties = properties[:10]
	}
	if properties == nil {
		properties = []LodgingProperty{}
	}
	for i := range properties {
		s.propertyPhotoURLs(c, &properties[i])
	}
	s.finishPropertyRequest(c, nil, func() { c.JSON(200, gin.H{"properties": properties}) })
}

func (s *Service) liveLodgingProperty(c *gin.Context, reference *TripLodgingProperty) (*LodgingProperty, error) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()
	property, err := s.lodgingPropertyProvider.Details(ctx, reference.PlaceID)
	if err != nil {
		return nil, err
	}
	if property == nil || property.PlaceID != reference.PlaceID || property.Provider != "google" {
		return nil, errUnavailable
	}
	return property, nil
}

func (s *Service) getLodgingProperty(c *gin.Context) {
	reference, err := s.propertyAccess(c)
	if err != nil {
		respondError(c, err)
		return
	}
	if !s.propertyProviderAvailable(c) {
		return
	}
	property, err := s.liveLodgingProperty(c, reference)
	if err != nil {
		c.JSON(503, gin.H{"error": "Property details are temporarily unavailable"})
		return
	}
	s.propertyPhotoURLs(c, property)
	s.finishPropertyRequest(c, reference, func() { c.JSON(200, gin.H{"property": property}) })
}

func (s *Service) getLodgingPropertyPhoto(c *gin.Context) {
	reference, err := s.propertyAccess(c)
	if err != nil {
		respondError(c, err)
		return
	}
	index, err := strconv.Atoi(c.Param("index"))
	if err != nil || index < 0 || index >= 5 {
		respondError(c, errUnavailable)
		return
	}
	if !s.propertyProviderAvailable(c) {
		return
	}
	property, err := s.liveLodgingProperty(c, reference)
	if err != nil {
		respondError(c, errUnavailable)
		return
	}
	// Bind bytes to the photo whose attribution the client saw, even if the
	// provider reorders or removes photos before this request. Resource names
	// remain short-lived provider data, never stored or exposed to the client.
	photoName := ""
	for _, photo := range property.Photos {
		if lodgingPhotoVersion(photo.Name) == c.Query("photoVersion") {
			photoName = photo.Name
			break
		}
	}
	if photoName == "" {
		respondError(c, errUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 12*time.Second)
	data, contentType, err := s.lodgingPropertyProvider.Photo(ctx, photoName)
	cancel()
	if err != nil || len(data) == 0 || len(data) > maxTripImageBytes || !oneOf(contentType, "image/jpeg", "image/png", "image/webp") {
		respondError(c, errUnavailable)
		return
	}
	s.finishPropertyRequest(c, reference, func() {
		c.Header("Cache-Control", "private, no-store, max-age=0")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Data(200, contentType, data)
	})
}
