package models

import (
	"time"

	"github.com/google/uuid"
)

const LoadingProfilePictureUrl = "https://crew-profile-icons.s3.us-east-1.amazonaws.com/loading-image.gif"

type User struct {
	ID                uuid.UUID  `db:"id" gorm:"type:uuid;default:uuid_generate_v4()" json:"id"`
	CreatedAt         time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updatedAt"`
	Name              string     `json:"name"`
	PhoneNumber       string     `json:"phoneNumber" gorm:"unique"`
	Active            bool       `json:"active"`
	Profile           *SonarUser `json:"profile" gorm:"foreignKey:ID"`
	ProfilePictureUrl string     `json:"profilePictureUrl"`
	HasSeenTutorial   bool       `json:"hasSeenTutorial" gorm:"default:false"`
	Party             *Party     `json:"party" gorm:"foreignKey:ID"`
	PartyID           *uuid.UUID `json:"partyId" gorm:"type:uuid;default:null"`
	Username          *string    `json:"username" gorm:"unique"`
	IsActive          *bool      `json:"isActive" gorm:"-"`
	Gold              int        `json:"gold"`
	Credits           int        `json:"credits"`
	DateOfBirth       *time.Time `json:"dateOfBirth" db:"date_of_birth"`
	Gender            *string    `json:"gender" db:"gender"`
	Latitude          *float64   `json:"latitude" db:"latitude"`
	Longitude         *float64   `json:"longitude" db:"longitude"`
	LocationAddress   *string    `json:"locationAddress" db:"location_address"`
	Bio               *string    `json:"bio" db:"bio"`
	Category          *string    `json:"category" db:"category"`
	AgeRange          *string    `json:"ageRange" db:"age_range"`
}
