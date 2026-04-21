package models

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/encoding/wkt"
	"gorm.io/gorm"
)

type Challenge struct {
	ID                   uuid.UUID                   `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt            time.Time                   `json:"createdAt"`
	UpdatedAt            time.Time                   `json:"updatedAt"`
	ZoneID               uuid.UUID                   `json:"zoneId" gorm:"column:zone_id"`
	ZoneKind             string                      `json:"zoneKind,omitempty" gorm:"column:zone_kind"`
	Zone                 Zone                        `json:"zone"`
	PointOfInterestID    *uuid.UUID                  `json:"pointOfInterestId,omitempty" gorm:"column:point_of_interest_id;type:uuid"`
	PointOfInterest      *PointOfInterest            `json:"pointOfInterest,omitempty" gorm:"foreignKey:PointOfInterestID"`
	Latitude             float64                     `json:"latitude"`
	Longitude            float64                     `json:"longitude"`
	Geometry             string                      `json:"geometry" gorm:"type:geometry(Point,4326)"`
	Polygon              *string                     `json:"-" gorm:"column:polygon;type:geometry(Polygon,4326)"`
	PolygonPoints        [][2]float64                `json:"polygonPoints,omitempty" gorm:"-"`
	Question             string                      `json:"question"`
	Description          string                      `json:"description"`
	RequiredStoryFlags   StringArray                 `json:"requiredStoryFlags" gorm:"column:required_story_flags;type:jsonb;default:'[]'"`
	ImageURL             string                      `json:"imageUrl" gorm:"column:image_url"`
	ThumbnailURL         string                      `json:"thumbnailUrl" gorm:"column:thumbnail_url"`
	ScaleWithUserLevel   bool                        `json:"scaleWithUserLevel" gorm:"column:scale_with_user_level"`
	RecurringChallengeID *uuid.UUID                  `json:"recurringChallengeId,omitempty" gorm:"column:recurring_challenge_id;type:uuid"`
	RecurrenceFrequency  *string                     `json:"recurrenceFrequency,omitempty" gorm:"column:recurrence_frequency"`
	NextRecurrenceAt     *time.Time                  `json:"nextRecurrenceAt,omitempty" gorm:"column:next_recurrence_at"`
	RetiredAt            *time.Time                  `json:"retiredAt,omitempty" gorm:"column:retired_at"`
	RewardMode           RewardMode                  `json:"rewardMode" gorm:"column:reward_mode"`
	RandomRewardSize     RandomRewardSize            `json:"randomRewardSize" gorm:"column:random_reward_size"`
	RewardExperience     int                         `json:"rewardExperience" gorm:"column:reward_experience"`
	Reward               int                         `json:"reward"`
	MaterialRewards      BaseMaterialRewards         `json:"materialRewards" gorm:"column:material_rewards_json;type:jsonb;default:'[]'"`
	InventoryItemID      *int                        `json:"inventoryItemId" gorm:"column:inventory_item_id"`
	ItemChoiceRewards    []ChallengeItemChoiceReward `json:"itemChoiceRewards" gorm:"foreignKey:ChallengeID"`
	SubmissionType       QuestNodeSubmissionType     `json:"submissionType" gorm:"type:text;default:photo"`
	Difficulty           int                         `json:"difficulty" gorm:"default:0"`
	StatTags             StringArray                 `json:"statTags,omitempty" gorm:"type:jsonb"`
	Proficiency          *string                     `json:"proficiency,omitempty"`
}

func (c *Challenge) TableName() string {
	return "challenges"
}

func (c *Challenge) AfterFind(tx *gorm.DB) error {
	c.PolygonPoints = decodeChallengePolygonPoints(c.Polygon)
	return nil
}

func (c *Challenge) BeforeSave(tx *gorm.DB) error {
	if err := c.SyncLocationGeometry(); err != nil {
		return err
	}
	return nil
}

func (c *Challenge) SyncLocationGeometry() error {
	if len(c.PolygonPoints) >= 3 {
		return c.SetPolygonPoints(c.PolygonPoints)
	}
	c.Polygon = nil
	c.PolygonPoints = nil
	return c.SetGeometry(c.Latitude, c.Longitude)
}

func (c *Challenge) SetGeometry(latitude, longitude float64) error {
	c.Geometry = fmt.Sprintf("SRID=4326;POINT(%f %f)", longitude, latitude)
	return nil
}

func (c *Challenge) SetPolygonPoints(points [][2]float64) error {
	normalized, centerLat, centerLng, polygonWKT, err := encodeChallengePolygon(points)
	if err != nil {
		return err
	}
	c.PolygonPoints = normalized
	c.Latitude = centerLat
	c.Longitude = centerLng
	c.Polygon = &polygonWKT
	return c.SetGeometry(centerLat, centerLng)
}

func (c *Challenge) HasPolygon() bool {
	return len(c.PolygonPoints) >= 3
}

func (c *Challenge) ContainsPoint(latitude, longitude float64) bool {
	if len(c.PolygonPoints) < 3 {
		return false
	}
	inside := false
	for i, j := 0, len(c.PolygonPoints)-1; i < len(c.PolygonPoints); j, i = i, i+1 {
		xi := c.PolygonPoints[i][0]
		yi := c.PolygonPoints[i][1]
		xj := c.PolygonPoints[j][0]
		yj := c.PolygonPoints[j][1]
		intersects := ((yi > latitude) != (yj > latitude)) &&
			(longitude < (xj-xi)*(latitude-yi)/(yj-yi+0.0)+xi)
		if intersects {
			inside = !inside
		}
	}
	return inside
}

func encodeChallengePolygon(points [][2]float64) ([][2]float64, float64, float64, string, error) {
	if len(points) < 3 {
		return nil, 0, 0, "", fmt.Errorf("polygonPoints must include at least 3 points")
	}

	normalized := make([][2]float64, 0, len(points))
	var sumLat float64
	var sumLng float64
	for _, point := range points {
		lng := point[0]
		lat := point[1]
		if math.IsNaN(lat) || math.IsInf(lat, 0) || lat < -90 || lat > 90 {
			return nil, 0, 0, "", fmt.Errorf("polygonPoints contain an invalid latitude")
		}
		if math.IsNaN(lng) || math.IsInf(lng, 0) || lng < -180 || lng > 180 {
			return nil, 0, 0, "", fmt.Errorf("polygonPoints contain an invalid longitude")
		}
		normalized = append(normalized, [2]float64{lng, lat})
		sumLat += lat
		sumLng += lng
	}

	ring := make(orb.Ring, 0, len(normalized)+1)
	for _, point := range normalized {
		ring = append(ring, orb.Point{point[0], point[1]})
	}
	if ring[0] != ring[len(ring)-1] {
		ring = append(ring, ring[0])
	}
	polygon := orb.Polygon{ring}
	geometry := wkt.MarshalString(polygon)

	centerLat := sumLat / float64(len(normalized))
	centerLng := sumLng / float64(len(normalized))
	return normalized, centerLat, centerLng, "SRID=4326;" + geometry, nil
}

func decodeChallengePolygonPoints(raw *string) [][2]float64 {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}

	var geom orb.Geometry
	if strings.HasPrefix(trimmed, "SRID=") || strings.HasPrefix(strings.ToUpper(trimmed), "POLYGON") {
		if idx := strings.Index(trimmed, ";"); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[idx+1:])
		}
		if g, err := wkt.Unmarshal(trimmed); err == nil {
			geom = g
		}
	} else {
		if strings.HasPrefix(trimmed, "\\x") || strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
			trimmed = trimmed[2:]
		}
	}
	if geom == nil {
		if bytes, err := hex.DecodeString(trimmed); err == nil {
			if g, err := wkb.Unmarshal(bytes); err == nil {
				geom = g
			}
		}
	}

	polygon, ok := geom.(orb.Polygon)
	if !ok || len(polygon) == 0 || len(polygon[0]) < 4 {
		return nil
	}
	ring := polygon[0]
	points := make([][2]float64, 0, len(ring))
	lastIndex := len(ring)
	if ring[0] == ring[len(ring)-1] {
		lastIndex--
	}
	for i := 0; i < lastIndex; i++ {
		points = append(points, [2]float64{ring[i][0], ring[i][1]})
	}
	if len(points) < 3 {
		return nil
	}
	return points
}
