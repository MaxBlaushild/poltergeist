package models

import "testing"

func TestZoneGetPolygonPrefersBoundaryPoints(t *testing.T) {
	zone := Zone{
		Name:      "Test Zone",
		Latitude:  5,
		Longitude: 5,
		Boundary:  "POLYGON((0 0, 10 0, 5 10, 0 0))",
		Points: []Point{
			{Latitude: 0, Longitude: 10},
			{Latitude: 10, Longitude: 10},
			{Latitude: 0, Longitude: 0},
			{Latitude: 10, Longitude: 0},
		},
	}

	polygon := zone.GetPolygon()
	if len(polygon) == 0 {
		t.Fatal("expected polygon")
	}
	if len(polygon[0]) != 5 {
		t.Fatalf("expected points-based polygon with 5 ring entries, got %d", len(polygon[0]))
	}
}

func TestZoneIsPointInBoundaryOrdersPointsByAngle(t *testing.T) {
	zone := Zone{
		Name:      "Scrambled",
		Latitude:  5,
		Longitude: 5,
		Points: []Point{
			{Latitude: 0, Longitude: 10},
			{Latitude: 10, Longitude: 0},
			{Latitude: 0, Longitude: 0},
			{Latitude: 10, Longitude: 10},
		},
	}

	if !zone.IsPointInBoundary(5, 5) {
		t.Fatal("expected center point to be inside reordered polygon")
	}

	boundary := zone.GetBoundary()
	if len(boundary) != 5 {
		t.Fatalf("expected boundary coords from chosen polygon, got %d", len(boundary))
	}
}
