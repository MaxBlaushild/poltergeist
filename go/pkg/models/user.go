package models

import (
	"time"

	"github.com/google/uuid"
)

const LoadingProfilePictureUrl = "https://crew-profile-icons.s3.us-east-1.amazonaws.com/loading-image.gif"

type User struct {
	ID                    uuid.UUID  `db:"id" gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt             time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updatedAt"`
	Name                  string     `json:"name"`
	PhoneNumber           string     `json:"phoneNumber" gorm:"unique"`
	Active                bool       `json:"active"`
	Profile               *SonarUser `json:"profile" gorm:"foreignKey:ID"`
	ProfilePictureUrl     string     `json:"profilePictureUrl"`
	HasCustomizedPortrait bool       `json:"hasCustomizedPortrait" gorm:"-"`
	HasSeenTutorial       bool       `json:"hasSeenTutorial" gorm:"default:false"`
	Party                 *Party     `json:"party" gorm:"foreignKey:ID"`
	PartyID               *uuid.UUID `json:"partyId" gorm:"type:uuid;default:null"`
	Username              *string    `json:"username" gorm:"unique"`
	IsActive              *bool      `json:"isActive" gorm:"-"`
	Gold                  int        `json:"gold"`
	Credits               int        `json:"credits"`
	DateOfBirth           *time.Time `json:"dateOfBirth" db:"date_of_birth"`
	Gender                *string    `json:"gender" db:"gender"`
	Latitude              *float64   `json:"latitude" db:"latitude"`
	Longitude             *float64   `json:"longitude" db:"longitude"`
	LocationAddress       *string    `json:"locationAddress" db:"location_address"`
	Bio                   *string    `json:"bio" db:"bio"`
	Category              *string    `json:"category" db:"category"`
	AgeRange              *string    `json:"ageRange" db:"age_range"`

	// Email/PasswordHash support email+password login (reef-site) alongside
	// the original phone+SMS flow every other domain uses — both are
	// optional so neither login method requires the other's field.
	Email        *string `json:"email" gorm:"unique"`
	PasswordHash *string `json:"-" gorm:"column:password_hash"`

	// GoogleID is Google's stable per-account "sub" claim — a user who
	// signs in with Google after already having an email/password account
	// gets this column linked onto their existing row (matched by email)
	// rather than a second duplicate account.
	GoogleID *string `json:"-" gorm:"column:google_id;unique"`
}
