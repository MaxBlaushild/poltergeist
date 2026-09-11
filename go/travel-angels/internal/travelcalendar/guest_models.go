package travelcalendar

import (
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/google/uuid"
)

// GuestDeliveryConfig deliberately receives a sender so tests never contact a
// mail provider. BaseURL is the frontend origin used for browser links.
type GuestDeliveryConfig struct {
	Sender  email.EmailClient
	BaseURL string
}

func (s *Service) ConfigureDelivery(config GuestDeliveryConfig) { s.guestDelivery = &config }

type GuestRecipient struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Email                string     `json:"email"`
	Name                 string     `json:"name"`
	UserID               *uuid.UUID `gorm:"type:uuid" json:"userId,omitempty"`
	VerifiedAt           *time.Time `json:"-"`
	ParticipationNotices bool       `json:"participationNotices"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

func (GuestRecipient) TableName() string { return "tc_guest_recipients" }

type guestChallenge struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	RecipientID uuid.UUID `gorm:"type:uuid"`
	CodeHash    string
	Attempts    int
	ExpiresAt   time.Time
	ConsumedAt  *time.Time
	CreatedAt   time.Time
}

func (guestChallenge) TableName() string { return "tc_guest_challenges" }

type guestToken struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	RecipientID uuid.UUID `gorm:"type:uuid"`
	TokenHash   string
	Purpose     string
	VerifiedAt  time.Time
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

func (guestToken) TableName() string { return "tc_guest_tokens" }

type calendarInvitation struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	RecipientID uuid.UUID  `gorm:"type:uuid" json:"recipientId"`
	CalendarID  *uuid.UUID `gorm:"type:uuid" json:"calendarId,omitempty"`
	StopID      *uuid.UUID `gorm:"type:uuid" json:"stopId,omitempty"`
	SenderID    uuid.UUID  `gorm:"type:uuid" json:"senderId"`
	Status      string     `json:"status"`
	LastError   string     `json:"lastError,omitempty"`
	SentAt      *time.Time `json:"sentAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func (calendarInvitation) TableName() string { return "tc_calendar_invitations" }

type calendarSubscription struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	RecipientID           uuid.UUID  `gorm:"type:uuid" json:"-"`
	CalendarID            *uuid.UUID `gorm:"type:uuid" json:"calendarId,omitempty"`
	StopID                *uuid.UUID `gorm:"type:uuid" json:"stopId,omitempty"`
	Status                string     `json:"status"`
	CalendarLinkTokenHash string     `json:"-"`
	StopLinkTokenHash     string     `json:"-"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

func (calendarSubscription) TableName() string { return "tc_calendar_subscriptions" }

type subscriptionExclusion struct {
	RecipientID uuid.UUID `gorm:"type:uuid;primaryKey"`
	StopID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt   time.Time
}

func (subscriptionExclusion) TableName() string { return "tc_subscription_exclusions" }

type calendarBroadcast struct {
	ID         uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	SenderID   uuid.UUID   `gorm:"type:uuid" json:"senderId"`
	CalendarID *uuid.UUID  `gorm:"type:uuid" json:"calendarId,omitempty"`
	StopIDs    []uuid.UUID `gorm:"serializer:json;type:jsonb" json:"stopIds"`
	Message    string      `json:"message"`
	CreatedAt  time.Time   `json:"createdAt"`
}

func (calendarBroadcast) TableName() string { return "tc_calendar_broadcasts" }

type calendarDelivery struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	BroadcastID     *uuid.UUID `gorm:"type:uuid" json:"broadcastId,omitempty"`
	RecipientID     uuid.UUID  `gorm:"type:uuid" json:"recipientId"`
	StopID          *uuid.UUID `gorm:"type:uuid" json:"stopId,omitempty"`
	ParticipationID *uuid.UUID `gorm:"type:uuid" json:"participationId,omitempty"`
	Event           string     `json:"event"`
	DedupeKey       string     `json:"-"`
	Status          string     `json:"status"`
	Attempts        int        `json:"attempts"`
	LastError       string     `json:"lastError,omitempty"`
	SentAt          *time.Time `json:"sentAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (calendarDelivery) TableName() string { return "tc_calendar_deliveries" }
