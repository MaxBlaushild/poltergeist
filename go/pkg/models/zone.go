package models

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"log"
	mathrand "math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
	"github.com/paulmach/orb/encoding/wkt"
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
	InternalTags   StringArray  `json:"internalTags" gorm:"column:internal_tags;type:jsonb"`
	Latitude       float64      `json:"latitude"`
	Longitude      float64      `json:"longitude"`
	ZoneImportID   *uuid.UUID   `json:"zoneImportId" gorm:"type:uuid"`
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
	if err := db.Table("points").
		Joins("JOIN boundary_points ON boundary_points.point_id = points.id").
		Where("boundary_points.zone_id = ?", z.ID).
		Order("boundary_points.created_at ASC").
		Find(&points).Error; err != nil {
		return err
	}
	if len(points) == 0 {
		return nil
	}

	z.Points = points

	return nil
}

func (z *Zone) GetPolygon() orb.Polygon {
	if z.Polygon != nil {
		return *z.Polygon
	}

	centerLat, centerLng, hasCenter := z.centerForPolygon()

	var polygonFromPoints orb.Polygon
	if len(z.Points) > 0 {
		// Create a ring from the sorted points
		ring := make(orb.Ring, len(z.Points))
		for i, point := range z.Points {
			ring[i] = orb.Point{point.Longitude, point.Latitude}
		}

		// Close the ring by adding the first point at the end if needed
		if len(ring) > 0 && !ring[0].Equal(ring[len(ring)-1]) {
			ring = append(ring, ring[0])
		}
		polygonFromPoints = orb.Polygon{ring}
	}

	if len(polygonFromPoints) > 0 {
		polygonFromPoints = z.normalizePolygon(polygonFromPoints, "boundary points", centerLat, centerLng, hasCenter)
	}

	polygonFromBoundary := z.decodeBoundaryPolygon()
	if len(polygonFromBoundary) > 0 {
		polygonFromBoundary = z.normalizePolygon(polygonFromBoundary, "boundary geometry", centerLat, centerLng, hasCenter)
	}
	if len(polygonFromBoundary) == 0 {
		log.Printf("Zone %s (%s): boundary geometry empty or failed to decode", z.ID, z.Name)
	}
	if len(polygonFromPoints) == 0 {
		log.Printf("Zone %s (%s): boundary points empty or missing", z.ID, z.Name)
	}

	var chosen orb.Polygon
	if hasCenter {
		if isPolygonUsable(polygonFromBoundary, centerLat, centerLng) {
			log.Printf("Zone %s (%s): using boundary geometry", z.ID, z.Name)
			chosen = polygonFromBoundary
		} else if isPolygonUsable(polygonFromPoints, centerLat, centerLng) {
			log.Printf("Zone %s (%s): boundary geometry unusable; using boundary points", z.ID, z.Name)
			chosen = polygonFromPoints
		}
	}

	if len(chosen) == 0 {
		if isPolygonUsable(polygonFromPoints, 0, 0) {
			log.Printf("Zone %s (%s): using boundary points without center check", z.ID, z.Name)
			chosen = polygonFromPoints
		} else if isPolygonUsable(polygonFromBoundary, 0, 0) {
			log.Printf("Zone %s (%s): using boundary geometry without center check", z.ID, z.Name)
			chosen = polygonFromBoundary
		}
	}

	if len(chosen) == 0 {
		log.Printf("Zone %s (%s): no usable polygon found", z.ID, z.Name)
		return nil
	}

	z.Polygon = &chosen
	return chosen
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
	log.Printf(
		"GetRandomPoint: bounds for zone %s (%s) min=(%f,%f) max=(%f,%f)",
		z.ID,
		z.Name,
		bounds.Min.Y(),
		bounds.Min.X(),
		bounds.Max.Y(),
		bounds.Max.X(),
	)

	// Create a new random generator with crypto/rand seed for each call
	// This ensures true randomness across different program runs
	var seed int64
	if err := binary.Read(rand.Reader, binary.BigEndian, &seed); err != nil {
		log.Printf("Error generating crypto random seed: %v, falling back to time-based seed", err)
		seed = time.Now().UnixNano()
	}
	rng := mathrand.New(mathrand.NewSource(seed))

	// Try up to 5000 times to find a point inside the polygon
	maxAttempts := 5000
	for i := 0; i < maxAttempts; i++ {
		// Generate a random point within the bounds using crypto-seeded RNG
		// Note: X is longitude, Y is latitude in geographic coordinates
		lng := bounds.Min.X() + (bounds.Max.X()-bounds.Min.X())*rng.Float64()
		lat := bounds.Min.Y() + (bounds.Max.Y()-bounds.Min.Y())*rng.Float64()

		// Check if the point is inside the polygon using ray casting algorithm
		if z.IsPointInBoundary(lat, lng) {
			log.Printf("Found valid random point inside boundary on attempt %d (lat=%f, lng=%f)", i+1, lat, lng)
			return orb.Point{lng, lat}
		}
	}

	// If we couldn't find a point after many tries, use centroid with random offset
	// This ensures we don't always return the exact same point
	log.Printf("Failed to find point inside boundary after %d attempts, using centroid with random offset", maxAttempts)
	centroid := calculateCentroid(polygon)

	// Add small random offset to centroid (within 10% of bounds size)
	offsetLng := (bounds.Max.X() - bounds.Min.X()) * 0.1 * (rng.Float64() - 0.5)
	offsetLat := (bounds.Max.Y() - bounds.Min.Y()) * 0.1 * (rng.Float64() - 0.5)

	offsetPoint := orb.Point{centroid.X() + offsetLng, centroid.Y() + offsetLat}

	// Check if offset point is valid, otherwise use centroid
	if z.IsPointInBoundary(offsetPoint.Y(), offsetPoint.X()) {
		log.Printf("Using offset centroid point (lat=%f, lng=%f)", offsetPoint.Y(), offsetPoint.X())
		return offsetPoint
	}

	log.Printf("Using exact centroid as fallback (lat=%f, lng=%f)", centroid.Y(), centroid.X())
	return centroid
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
	centerLat, centerLng, hasCenter := z.centerForPolygon()
	polygon := z.decodeBoundaryPolygon()
	if len(polygon) > 0 {
		polygon = z.normalizePolygon(polygon, "boundary geometry", centerLat, centerLng, hasCenter)
	}
	if len(polygon) == 0 {
		polygon = z.GetPolygon()
	}
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

func (z *Zone) decodeBoundaryPolygon() orb.Polygon {
	raw := strings.TrimSpace(z.Boundary)
	if raw == "" {
		log.Printf("Zone %s (%s): boundary string empty", z.ID, z.Name)
		return nil
	}

	var geom orb.Geometry
	trimmed := raw
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
	if !ok || len(polygon) == 0 {
		log.Printf("Zone %s (%s): decoded boundary is not a polygon", z.ID, z.Name)
		return nil
	}
	return polygon
}

func (z *Zone) centerForPolygon() (float64, float64, bool) {
	if z.Latitude != 0 || z.Longitude != 0 {
		return z.Latitude, z.Longitude, true
	}
	if len(z.Points) == 0 {
		return 0, 0, false
	}

	var sumLat float64
	var sumLng float64
	for _, point := range z.Points {
		sumLat += point.Latitude
		sumLng += point.Longitude
	}

	return sumLat / float64(len(z.Points)), sumLng / float64(len(z.Points)), true
}

func (z *Zone) normalizePolygon(polygon orb.Polygon, source string, centerLat float64, centerLng float64, hasCenter bool) orb.Polygon {
	if len(polygon) == 0 {
		return polygon
	}
	if !hasCenter {
		return polygon
	}
	if polygonContainsPoint(polygon, centerLat, centerLng) {
		return polygon
	}

	swapped := swapPolygon(polygon)
	if polygonContainsPoint(swapped, centerLat, centerLng) {
		log.Printf(
			"Zone %s (%s): %s polygon appears lat/lng swapped; using swapped coordinates",
			z.ID,
			z.Name,
			source,
		)
		return swapped
	}

	return polygon
}

func swapPolygon(polygon orb.Polygon) orb.Polygon {
	swapped := make(orb.Polygon, len(polygon))
	for i, ring := range polygon {
		swappedRing := make(orb.Ring, len(ring))
		for j, point := range ring {
			swappedRing[j] = orb.Point{point.Y(), point.X()}
		}
		swapped[i] = swappedRing
	}
	return swapped
}

func isPolygonUsable(polygon orb.Polygon, lat float64, lng float64) bool {
	if len(polygon) == 0 {
		return false
	}
	if polygon.Bound().IsEmpty() {
		return false
	}
	if lat == 0 && lng == 0 {
		return true
	}
	return polygonContainsPoint(polygon, lat, lng)
}

func polygonContainsPoint(polygon orb.Polygon, lat float64, lng float64) bool {
	if len(polygon) == 0 {
		return false
	}
	ring := polygon[0]
	if len(ring) < 3 {
		return false
	}

	point := orb.Point{lng, lat}
	inside := false
	for i, j := 0, len(ring)-1; i < len(ring); j, i = i, i+1 {
		if ((ring[i].Y() > point.Y()) != (ring[j].Y() > point.Y())) &&
			(point.X() < (ring[j].X()-ring[i].X())*(point.Y()-ring[i].Y())/(ring[j].Y()-ring[i].Y())+ring[i].X()) {
			inside = !inside
		}
	}
	return inside
}
