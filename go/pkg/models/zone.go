package models

import (
	"crypto/rand"
	"encoding/binary"
	"log"
	"math"
	mathrand "math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/paulmach/orb"
	"gorm.io/gorm"
)

// init seeds the random number generator once when the package is loaded
func init() {
	seed := time.Now().UnixNano()
	log.Printf("Seeding math/rand with Unix nano time: %d", seed)
	mathrand.Seed(seed)
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Zone struct {
	ID             uuid.UUID    `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	Latitude       float64      `json:"latitude"`
	Longitude      float64      `json:"longitude"`
	Radius         float64      `json:"radius"`
	Boundary       string       `json:"boundary"`
	BoundaryCoords []Location   `json:"boundaryCoords" gorm:"-"`
	Polygon        *orb.Polygon `json:"polygon" gorm:"-"`
	Points         []Point      `json:"points" gorm:"many2many:boundary_points;"`
}

func (z *Zone) AfterFind(tx *gorm.DB) (err error) {
	z.BoundaryCoords = z.GetBoundary()
	return z.LoadPoints(tx)
}

func (z *Zone) LoadPoints(db *gorm.DB) error {
	points := []Point{}
	if err := db.Model(z).Association("Points").Find(&points); err != nil {
		return err
	}
	if len(points) == 0 {
		return nil
	}

	// Calculate centroid of all points
	var sumX, sumY float64
	for _, point := range points {
		sumX += point.Longitude
		sumY += point.Latitude
	}
	centroidX := sumX / float64(len(points))
	centroidY := sumY / float64(len(points))

	// Sort points by angle around centroid
	sort.Slice(points, func(i, j int) bool {
		angleI := math.Atan2(points[i].Latitude-centroidY, points[i].Longitude-centroidX)
		angleJ := math.Atan2(points[j].Latitude-centroidY, points[j].Longitude-centroidX)
		return angleI < angleJ
	})

	z.Points = points

	return nil
}

func (z *Zone) GetPolygon() orb.Polygon {
	if z.Polygon != nil {
		return *z.Polygon
	}

	if len(z.Points) == 0 {
		return nil
	}

	// Create a ring from the sorted points
	ring := make(orb.Ring, len(z.Points))
	for i, point := range z.Points {
		ring[i] = orb.Point{point.Longitude, point.Latitude}
	}

	// Close the ring by adding the first point at the end if needed
	if len(ring) > 0 && !ring[0].Equal(ring[len(ring)-1]) {
		ring = append(ring, ring[0])
	}

	// Create polygon from ring
	p := orb.Polygon{ring}
	z.Polygon = &p

	return p
}

func (z *Zone) GetRandomPoint() orb.Point {
	polygon := z.GetPolygon()
	if polygon == nil {
		log.Printf("GetRandomPoint: polygon is nil, returning empty point")
		return orb.Point{}
	}

	// Get the bounds of the polygon
	bounds := polygon.Bound()
	if bounds.IsEmpty() {
		log.Printf("GetRandomPoint: polygon bounds are empty, returning empty point")
		return orb.Point{}
	}

	// Create a truly random seed using crypto/rand
	var seed int64
	err := binary.Read(rand.Reader, binary.BigEndian, &seed)
	if err != nil {
		log.Printf("Error generating random seed: %v", err)
		return orb.Point{}
	}
	log.Printf("Generated random seed: %d", seed)
	r := mathrand.New(mathrand.NewSource(seed))

	// Try up to 1000 times to find a point inside the polygon
	for i := 0; i < 1000; i++ {
		// Generate a random point within the bounds
		// Note: X is longitude, Y is latitude in geographic coordinates
		lng := bounds.Min.X() + (bounds.Max.X()-bounds.Min.X())*r.Float64()
		lat := bounds.Min.Y() + (bounds.Max.Y()-bounds.Min.Y())*r.Float64()

		log.Printf("Attempt %d: Generated random point lat=%f, lng=%f", i+1, lat, lng)

		// Check if the point is inside the polygon using ray casting algorithm
		if z.IsPointInBoundary(lat, lng) {
			log.Printf("Found valid point inside boundary on attempt %d", i+1)
			return orb.Point{lng, lat}
		}
	}

	// If we couldn't find a point after 1000 tries, return the centroid
	log.Printf("Failed to find point inside boundary after 1000 attempts, falling back to centroid")
	return calculateCentroid(polygon)
}

// isPointInPolygon uses the ray casting algorithm to determine if a point is inside a polygon
func (z *Zone) IsPointInBoundary(lat float64, lng float64) bool {
	point := orb.Point{lng, lat}

	polygon := z.GetPolygon()
	if polygon == nil {
		return false
	}

	if len(polygon) == 0 {
		return false
	}

	// Get the outer ring
	ring := polygon[0]
	if len(ring) < 3 {
		return false
	}

	inside := false
	for i, j := 0, len(ring)-1; i < len(ring); j, i = i, i+1 {
		if ((ring[i].Y() > point.Y()) != (ring[j].Y() > point.Y())) &&
			(point.X() < (ring[j].X()-ring[i].X())*(point.Y()-ring[i].Y())/(ring[j].Y()-ring[i].Y())+ring[i].X()) {
			inside = !inside
		}
	}

	return inside
}

// calculateCentroid calculates the centroid of a polygon
func calculateCentroid(polygon orb.Polygon) orb.Point {
	if len(polygon) == 0 {
		log.Printf("calculateCentroid: empty polygon, returning empty point")
		return orb.Point{}
	}

	// Get the outer ring
	ring := polygon[0]
	if len(ring) < 3 {
		log.Printf("calculateCentroid: polygon has fewer than 3 points, returning empty point")
		return orb.Point{}
	}

	var area, x, y float64
	for i, j := 0, len(ring)-1; i < len(ring); j, i = i, i+1 {
		xi, yi := ring[i].X(), ring[i].Y()
		xj, yj := ring[j].X(), ring[j].Y()
		common := xi*yj - xj*yi
		area += common
		x += (xi + xj) * common
		y += (yi + yj) * common
	}

	area *= 0.5
	if area == 0 {
		log.Printf("calculateCentroid: polygon area is 0, returning empty point")
		return orb.Point{}
	}

	centroid := orb.Point{x / (6 * area), y / (6 * area)}
	log.Printf("Calculated centroid: %v", centroid)
	return centroid
}

func (z *Zone) GetBoundary() []Location {
	polygon := z.GetPolygon()
	if polygon == nil {
		return nil
	}

	points := make([]Location, 0)
	outerRing := polygon[0]
	for _, coord := range outerRing {
		points = append(points, Location{
			Latitude:  coord.Y(),
			Longitude: coord.X(),
		})
	}

	return points
}
