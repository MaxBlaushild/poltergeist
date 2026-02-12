package models

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/encoding/wkt"
	"gorm.io/gorm"
)

type QuestNode struct {
	ID                uuid.UUID            `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt         time.Time            `json:"createdAt"`
	UpdatedAt         time.Time            `json:"updatedAt"`
	QuestID           uuid.UUID            `json:"questId" gorm:"type:uuid"`
	OrderIndex        int                  `json:"orderIndex"`
	PointOfInterestID *uuid.UUID           `json:"pointOfInterestId" gorm:"type:uuid"`
	Polygon           string               `json:"polygon" gorm:"type:geometry(Polygon,4326)"`
	PolygonPoints     [][2]float64         `json:"polygonPoints" gorm:"-"`
	Challenges        []QuestNodeChallenge `json:"challenges" gorm:"foreignKey:QuestNodeID"`
	Children          []QuestNodeChild     `json:"children" gorm:"foreignKey:QuestNodeID"`
}

func (q *QuestNode) TableName() string {
	return "quest_nodes"
}

func (q *QuestNode) AfterFind(tx *gorm.DB) (err error) {
	q.PolygonPoints = q.decodePolygonPoints()
	return nil
}

// SetPolygonFromPoints sets the polygon geometry from [lng,lat] points.
func (q *QuestNode) SetPolygonFromPoints(points [][2]float64) {
	if len(points) == 0 {
		q.Polygon = ""
		q.PolygonPoints = nil
		return
	}
	coords := ""
	for i, p := range points {
		if i > 0 {
			coords += ","
		}
		coords += fmt.Sprintf("%f %f", p[0], p[1])
	}
	// Ensure closed ring by repeating first point if needed.
	if points[0][0] != points[len(points)-1][0] || points[0][1] != points[len(points)-1][1] {
		coords += fmt.Sprintf(",%f %f", points[0][0], points[0][1])
	}
	q.Polygon = fmt.Sprintf("SRID=4326;POLYGON((%s))", coords)
	q.PolygonPoints = points
}

func (q *QuestNode) decodePolygonPoints() [][2]float64 {
	if strings.TrimSpace(q.Polygon) == "" {
		return nil
	}

	var geom orb.Geometry
	trimmed := strings.TrimSpace(q.Polygon)
	if strings.HasPrefix(trimmed, "SRID=") || strings.HasPrefix(strings.ToUpper(trimmed), "POLYGON") {
		if idx := strings.Index(trimmed, ";"); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[idx+1:])
		}
		if g, err := wkt.Unmarshal(trimmed); err == nil {
			geom = g
		}
	} else {
		if bytes, err := hex.DecodeString(trimmed); err == nil {
			if g, err := wkb.Unmarshal(bytes); err == nil {
				geom = g
			}
		}
	}

	polygon, ok := geom.(orb.Polygon)
	if !ok || len(polygon) == 0 {
		return nil
	}

	ring := polygon[0]
	if len(ring) == 0 {
		return nil
	}

	points := make([][2]float64, 0, len(ring))
	for _, pt := range ring {
		points = append(points, [2]float64{pt[0], pt[1]})
	}
	return points
}
