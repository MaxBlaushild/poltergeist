package models

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PointOfInterest struct {
	ID                        uuid.UUID                  `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt                 time.Time                  `json:"createdAt"`
	UpdatedAt                 time.Time                  `json:"updatedAt"`
	Name                      string                     `json:"name"`
	OriginalName              string                     `json:"originalName"`
	Clue                      string                     `json:"clue"`
	Lat                       string                     `json:"lat"`
	Lng                       string                     `json:"lng"`
	ImageUrl                  string                     `json:"imageURL"`
	Description               string                     `json:"description"`
	PointOfInterestChallenges []PointOfInterestChallenge `json:"pointOfInterestChallenges"`
	Geometry                  string                     `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Tags                      []Tag                      `json:"tags" gorm:"many2many:tag_entities;joinForeignKey:point_of_interest_id;joinReferences:tag_id"`
	GoogleMapsPlaceID         *string                    `json:"googleMapsPlaceId"`
	LastUsedInQuestAt         *time.Time                 `json:"lastUsedInQuestAt,omitempty"`
}

func (p *PointOfInterest) TableName() string {
	return "points_of_interest"
}

func (p *PointOfInterest) BeforeSave(tx *gorm.DB) error {
	if p.Lat != "" && p.Lng != "" {
		if err := p.SetGeometry(p.Lat, p.Lng); err != nil {
			return err
		}
	}
	return nil
}

func (p *PointOfInterest) SetGeometry(lat string, lng string) error {
	floatLat, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return err
	}
	floatLng, err := strconv.ParseFloat(lng, 64)
	if err != nil {
		return err
	}

	// Create WKT (Well-Known Text) format: 'SRID=4326;POINT(lng lat)'
	p.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", floatLng, floatLat)
	return nil
}
