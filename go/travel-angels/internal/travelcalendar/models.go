// Package travelcalendar implements permission-aware travel calendars. Public
// handlers deliberately share the same authorization policy as account handlers.
package travelcalendar

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Calendar struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	OwnerID        uuid.UUID `json:"ownerId" gorm:"type:uuid;uniqueIndex"`
	Title          string    `json:"title"`
	Visibility     string    `json:"visibility"`
	ShareToken     string    `json:"shareToken" gorm:"uniqueIndex"`
	SharingEnabled bool      `json:"sharingEnabled"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (Calendar) TableName() string { return "tc_calendars" }

type StopContent struct {
	TimeZone      string `json:"timeZone"`
	Title         string `json:"title"`
	Destination   string `json:"destination"`
	StartDate     string `json:"startDate"`
	EndDate       string `json:"endDate"`
	DatePrecision string `json:"datePrecision"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	JoinOpen      bool   `json:"joinOpen"`
}

func (c *StopContent) validate() error {
	c.TimeZone = strings.TrimSpace(c.TimeZone)
	if c.TimeZone == "" {
		c.TimeZone = "UTC"
	}
	if _, err := time.LoadLocation(c.TimeZone); err != nil || c.TimeZone == "Local" {
		return errors.New("Choose a valid IANA destination time zone, such as Asia/Tokyo")
	}

	c.Title = strings.TrimSpace(c.Title)
	c.Destination = strings.TrimSpace(c.Destination)
	if c.Title == "" || c.Destination == "" || len(c.Title) > 200 || len(c.Destination) > 300 || len(c.Description) > 10000 {
		return errors.New("title and destination are required; content exceeds limits")
	}
	if c.DatePrecision == "" {
		c.DatePrecision = "tbd"
	}
	if c.Status == "" {
		c.Status = "tentative"
	}
	if c.Status != "tentative" && c.Status != "confirmed" && c.Status != "cancelled" {
		return errors.New("invalid planning status")
	}
	if c.DatePrecision != "fixed" && c.DatePrecision != "approximate" && c.DatePrecision != "tbd" {
		return errors.New("invalid date precision")
	}
	if c.DatePrecision == "tbd" {
		c.StartDate = ""
		c.EndDate = ""
		return nil
	}
	if !validDate(c.StartDate) || !validDate(c.EndDate) || c.EndDate < c.StartDate {
		return errors.New("dates must be ordered local dates in YYYY-MM-DD format")
	}
	return nil
}
func validDate(s string) bool {
	t, e := time.Parse("2006-01-02", s)
	return e == nil && t.Format("2006-01-02") == s
}

type Stop struct {
	ID             uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	CalendarID     uuid.UUID    `json:"calendarId" gorm:"type:uuid;index"`
	Draft          StopContent  `json:"draft" gorm:"serializer:json;type:jsonb"`
	Published      *StopContent `json:"-" gorm:"serializer:json;type:jsonb"`
	Visibility     string       `json:"visibility"`
	ShareToken     string       `json:"shareToken" gorm:"uniqueIndex"`
	SharingEnabled bool         `json:"sharingEnabled"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

func (Stop) TableName() string { return "tc_stops" }

type Grant struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	CalendarID  *uuid.UUID `json:"calendarId,omitempty" gorm:"type:uuid;index"`
	StopID      *uuid.UUID `json:"stopId,omitempty" gorm:"type:uuid;index"`
	UserID      *uuid.UUID `json:"userId,omitempty" gorm:"type:uuid;index"`
	RecipientID *uuid.UUID `json:"recipientId,omitempty" gorm:"type:uuid;index"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func (Grant) TableName() string { return "tc_grants" }

type Collaborator struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	CalendarID           *uuid.UUID `json:"calendarId,omitempty" gorm:"type:uuid;index"`
	StopID               *uuid.UUID `json:"stopId,omitempty" gorm:"type:uuid;index"`
	UserID               uuid.UUID  `json:"userId" gorm:"type:uuid;index"`
	CanManagePermissions bool       `json:"canManagePermissions"`
	Accepted             bool       `json:"accepted"`
	CreatedAt            time.Time  `json:"createdAt"`
}

func (Collaborator) TableName() string { return "tc_collaborators" }

type Participation struct {
	CalendarLinkTokenHash string    `json:"-"`
	StopLinkTokenHash     string    `json:"-"`
	ID                    uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	StopID                uuid.UUID `json:"stopId" gorm:"type:uuid;uniqueIndex:tc_participation_identity"`
	UserID                uuid.UUID `json:"userId" gorm:"type:uuid;uniqueIndex:tc_participation_identity"`
	Status                string    `json:"status"`
	StartDate             string    `json:"startDate"`
	EndDate               string    `json:"endDate"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

func (Participation) TableName() string { return "tc_participations" }

type Viewer struct{ UserID, RecipientID, CalendarLinkID, StopLinkID uuid.UUID }

type Service struct {
	db            *gorm.DB
	authenticate  func(*gin.Context) (*models.User, error)
	baseURL       string
	guestDelivery *GuestDeliveryConfig
}

func New(db *gorm.DB, authenticate func(*gin.Context) (*models.User, error), baseURL string) *Service {
	// Public requests can carry expired app credentials. Cache both successful
	// and failed verification for this request so falling back to guest browsing
	// does not contact the authentication service repeatedly.
	if authenticate != nil {
		resolve := authenticate
		type result struct {
			user *models.User
			err  error
		}
		authenticate = func(c *gin.Context) (*models.User, error) {
			if cached, ok := c.Get("travelcalendar.authentication"); ok {
				if r, ok := cached.(result); ok {
					return r.user, r.err
				}
			}
			user, err := resolve(c)
			c.Set("travelcalendar.authentication", result{user, err})
			return user, err
		}
	}
	return &Service{db: db, authenticate: authenticate, baseURL: strings.TrimRight(baseURL, "/")}
}
func newToken() string {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
func (s *Service) requireUser(c *gin.Context) (*models.User, bool) {
	if s.authenticate == nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "Claim or sign in to your account to continue"})
		return nil, false
	}
	u, e := s.authenticate(c)
	if e != nil || u == nil || u.ID == uuid.Nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "Claim or sign in to your account to continue"})
		return nil, false
	}
	return u, true
}
func (s *Service) viewer(c *gin.Context) Viewer {
	v := Viewer{}
	if s.authenticate != nil {
		if u, e := s.authenticate(c); e == nil && u != nil {
			v.UserID = u.ID
		}
	}
	v.RecipientID = s.guestRecipient(c)
	if t := c.Query("calendarToken"); t != "" {
		var cal Calendar
		if s.db.Where("share_token = ? AND sharing_enabled = ?", t, true).First(&cal).Error == nil {
			v.CalendarLinkID = cal.ID
		}
	}
	if t := c.Query("stopToken"); t != "" {
		var stop Stop
		if s.db.Where("share_token = ? AND sharing_enabled = ?", t, true).First(&stop).Error == nil {
			v.StopLinkID = stop.ID
		}
	}
	return v
}

func effectiveTimeZone(zone string) string {
	if zone == "" {
		return "UTC"
	}
	return zone
}
