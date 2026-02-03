package models

import (
  "fmt"
  "time"

  "github.com/google/uuid"
)

type QuestNode struct {
  ID               uuid.UUID   `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
  CreatedAt        time.Time   `json:"createdAt"`
  UpdatedAt        time.Time   `json:"updatedAt"`
  QuestID          uuid.UUID   `json:"questId" gorm:"type:uuid"`
  OrderIndex       int         `json:"orderIndex"`
  PointOfInterestID *uuid.UUID `json:"pointOfInterestId" gorm:"type:uuid"`
  Polygon          string      `json:"polygon" gorm:"type:geometry(Polygon,4326)"`
  Challenges       []QuestNodeChallenge `json:"challenges" gorm:"foreignKey:QuestNodeID"`
  Children         []QuestNodeChild `json:"children" gorm:"foreignKey:QuestNodeID"`
}

func (q *QuestNode) TableName() string {
  return "quest_nodes"
}

// SetPolygonFromPoints sets the polygon geometry from [lng,lat] points.
func (q *QuestNode) SetPolygonFromPoints(points [][2]float64) {
  if len(points) == 0 {
    q.Polygon = ""
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
}
