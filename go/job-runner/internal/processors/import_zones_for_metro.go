package processors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	defaultNominatimURL = "https://nominatim.openstreetmap.org/search"
	defaultOverpassURL  = "https://overpass-api.de/api/interpreter"
)

type ImportZonesForMetroProcessor struct {
	dbClient db.DbClient
	client   *http.Client
}

func NewImportZonesForMetroProcessor(dbClient db.DbClient) *ImportZonesForMetroProcessor {
	return &ImportZonesForMetroProcessor{
		dbClient: dbClient,
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (p *ImportZonesForMetroProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	payload := jobs.ImportZonesForMetroTaskPayload{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	importItem, err := p.dbClient.ZoneImport().FindByID(ctx, payload.ImportID)
	if err != nil {
		return err
	}
	if importItem == nil {
		return nil
	}

	importItem.Status = "in_progress"
	importItem.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneImport().Update(ctx, importItem); err != nil {
		return err
	}

	metroName := strings.TrimSpace(importItem.MetroName)
	if metroName == "" {
		msg := "metro name is required"
		importItem.Status = "failed"
		importItem.ErrorMessage = &msg
		importItem.UpdatedAt = time.Now()
		_ = p.dbClient.ZoneImport().Update(ctx, importItem)
		return nil
	}

	areaID, err := p.lookupMetroAreaID(ctx, metroName)
	if err != nil {
		msg := err.Error()
		importItem.Status = "failed"
		importItem.ErrorMessage = &msg
		importItem.UpdatedAt = time.Now()
		_ = p.dbClient.ZoneImport().Update(ctx, importItem)
		return err
	}

	neighborhoods, err := p.fetchNeighborhoods(ctx, areaID)
	if err != nil {
		msg := err.Error()
		importItem.Status = "failed"
		importItem.ErrorMessage = &msg
		importItem.UpdatedAt = time.Now()
		_ = p.dbClient.ZoneImport().Update(ctx, importItem)
		return err
	}

	existingZones, err := p.dbClient.Zone().FindAll(ctx)
	if err != nil {
		msg := err.Error()
		importItem.Status = "failed"
		importItem.ErrorMessage = &msg
		importItem.UpdatedAt = time.Now()
		_ = p.dbClient.ZoneImport().Update(ctx, importItem)
		return err
	}

	existingNames := map[string]struct{}{}
	for _, zone := range existingZones {
		if zone == nil {
			continue
		}
		nameKey := strings.ToLower(strings.TrimSpace(zone.Name))
		if nameKey != "" {
			existingNames[nameKey] = struct{}{}
		}
	}

	createdCount := 0
	seen := map[string]struct{}{}
	for _, neighborhood := range neighborhoods {
		name := strings.TrimSpace(neighborhood.Name)
		if name == "" {
			continue
		}
		nameKey := strings.ToLower(name)
		if _, ok := seen[nameKey]; ok {
			continue
		}
		if _, ok := existingNames[nameKey]; ok {
			continue
		}

		boundary := simplifyBoundary(neighborhood.Boundary)
		if len(boundary) < 3 {
			continue
		}

		centroidLat, centroidLon := polygonCentroid(boundary)
		radius := maxDistanceMeters(centroidLat, centroidLon, boundary)
		if radius < 250 {
			radius = 250
		}

		zone := &models.Zone{
			ID:           uuid.New(),
			Name:         name,
			Description:  fmt.Sprintf("Imported from OSM neighborhoods for %s", metroName),
			Latitude:     centroidLat,
			Longitude:    centroidLon,
			Radius:       radius,
			ZoneImportID: &importItem.ID,
		}

		if err := p.dbClient.Zone().Create(ctx, zone); err != nil {
			continue
		}

		boundaryCoords := make([][]float64, 0, len(boundary))
		for _, point := range boundary {
			boundaryCoords = append(boundaryCoords, []float64{point.Lat, point.Lon})
		}
		if err := p.dbClient.Zone().UpdateBoundary(ctx, zone.ID, boundaryCoords); err != nil {
			continue
		}

		createdCount++
		seen[nameKey] = struct{}{}
	}

	importItem.ZoneCount = createdCount
	importItem.UpdatedAt = time.Now()
	if createdCount == 0 {
		msg := "no zones were created"
		importItem.Status = "failed"
		importItem.ErrorMessage = &msg
	} else {
		importItem.Status = "completed"
		importItem.ErrorMessage = nil
	}
	return p.dbClient.ZoneImport().Update(ctx, importItem)
}

type nominatimResult struct {
	OsmID       int64  `json:"osm_id"`
	OsmType     string `json:"osm_type"`
	DisplayName string `json:"display_name"`
}

func (p *ImportZonesForMetroProcessor) lookupMetroAreaID(ctx context.Context, metroName string) (int64, error) {
	baseURL := strings.TrimSpace(os.Getenv("NOMINATIM_URL"))
	if baseURL == "" {
		baseURL = defaultNominatimURL
	}

	query := url.Values{}
	query.Set("q", metroName)
	query.Set("format", "jsonv2")
	query.Set("limit", "1")
	query.Set("countrycodes", "us")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?%s", baseURL, query.Encode()), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "poltergeist-zone-import/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("nominatim request failed: %s", strings.TrimSpace(string(body)))
	}

	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, err
	}
	if len(results) == 0 {
		return 0, fmt.Errorf("no OSM result found for %s", metroName)
	}

	result := results[0]
	switch strings.ToLower(result.OsmType) {
	case "relation":
		return 3600000000 + result.OsmID, nil
	case "way":
		return 2400000000 + result.OsmID, nil
	default:
		return 0, fmt.Errorf("unsupported OSM type %s for %s", result.OsmType, metroName)
	}
}

type overpassResponse struct {
	Elements []overpassElement `json:"elements"`
}

type overpassElement struct {
	Type     string            `json:"type"`
	ID       int64             `json:"id"`
	Tags     map[string]string `json:"tags"`
	Geometry []overpassCoord   `json:"geometry"`
	Members  []overpassMember  `json:"members"`
}

type overpassMember struct {
	Type     string          `json:"type"`
	Role     string          `json:"role"`
	Geometry []overpassCoord `json:"geometry"`
}

type overpassCoord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type neighborhoodBoundary struct {
	Name     string
	Boundary []overpassCoord
}

func (p *ImportZonesForMetroProcessor) fetchNeighborhoods(ctx context.Context, areaID int64) ([]neighborhoodBoundary, error) {
	baseURL := strings.TrimSpace(os.Getenv("OVERPASS_URL"))
	if baseURL == "" {
		baseURL = defaultOverpassURL
	}

	query := fmt.Sprintf(`[out:json][timeout:180];
area(%d)->.searchArea;
(
  relation["place"~"neighbourhood|suburb"](area.searchArea);
  relation["boundary"="administrative"]["admin_level"="10"](area.searchArea);
  way["place"~"neighbourhood|suburb"](area.searchArea);
  way["boundary"="administrative"]["admin_level"="10"](area.searchArea);
);
out body geom;`, areaID)

	form := url.Values{}
	form.Set("data", query)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "poltergeist-zone-import/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("overpass request failed: %s", strings.TrimSpace(string(body)))
	}

	var parsed overpassResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	out := make([]neighborhoodBoundary, 0, len(parsed.Elements))
	for _, element := range parsed.Elements {
		name := ""
		if element.Tags != nil {
			name = element.Tags["name"]
		}
		if name == "" {
			continue
		}

		var geometry []overpassCoord
		switch element.Type {
		case "way":
			geometry = element.Geometry
		case "relation":
			geometry = mergeRelationOuterMembers(element.Members)
			if len(geometry) == 0 && len(element.Geometry) > 2 {
				geometry = element.Geometry
			}
		}

		if len(geometry) < 3 {
			continue
		}

		out = append(out, neighborhoodBoundary{
			Name:     name,
			Boundary: geometry,
		})
	}

	return out, nil
}

func simplifyBoundary(points []overpassCoord) []overpassCoord {
	if len(points) < 3 {
		return nil
	}
	out := make([]overpassCoord, 0, len(points))
	for _, point := range points {
		if len(out) > 0 {
			last := out[len(out)-1]
			if last.Lat == point.Lat && last.Lon == point.Lon {
				continue
			}
		}
		out = append(out, point)
	}
	if len(out) > 2 {
		first := out[0]
		last := out[len(out)-1]
		if first.Lat == last.Lat && first.Lon == last.Lon {
			out = out[:len(out)-1]
		}
	}
	if len(out) < 3 {
		return nil
	}
	return out
}

func mergeRelationOuterMembers(members []overpassMember) []overpassCoord {
	segments := make([][]overpassCoord, 0)
	for _, member := range members {
		if member.Role != "outer" {
			continue
		}
		if len(member.Geometry) < 2 {
			continue
		}
		segment := make([]overpassCoord, len(member.Geometry))
		copy(segment, member.Geometry)
		segments = append(segments, segment)
	}
	if len(segments) == 0 {
		return nil
	}

	used := make([]bool, len(segments))
	rings := make([][]overpassCoord, 0)

	for i := range segments {
		if used[i] {
			continue
		}
		ring := make([]overpassCoord, len(segments[i]))
		copy(ring, segments[i])
		used[i] = true

		for {
			extended := false
			for j := range segments {
				if used[j] {
					continue
				}
				segment := segments[j]
				if len(segment) < 2 {
					used[j] = true
					continue
				}
				first := segment[0]
				last := segment[len(segment)-1]

				switch {
				case coordsEqual(ring[len(ring)-1], first):
					ring = append(ring, segment[1:]...)
					used[j] = true
					extended = true
				case coordsEqual(ring[len(ring)-1], last):
					reversed := reverseCoords(segment)
					ring = append(ring, reversed[1:]...)
					used[j] = true
					extended = true
				case coordsEqual(ring[0], last):
					ring = append(segment[:len(segment)-1], ring...)
					used[j] = true
					extended = true
				case coordsEqual(ring[0], first):
					reversed := reverseCoords(segment)
					ring = append(reversed[:len(reversed)-1], ring...)
					used[j] = true
					extended = true
				}

				if extended {
					break
				}
			}
			if !extended {
				break
			}
			if coordsEqual(ring[0], ring[len(ring)-1]) {
				break
			}
		}

		rings = append(rings, ring)
	}

	if len(rings) == 0 {
		return nil
	}

	longest := rings[0]
	for _, ring := range rings[1:] {
		if len(ring) > len(longest) {
			longest = ring
		}
	}
	return longest
}

func reverseCoords(points []overpassCoord) []overpassCoord {
	if len(points) == 0 {
		return nil
	}
	out := make([]overpassCoord, len(points))
	for i := range points {
		out[len(points)-1-i] = points[i]
	}
	return out
}

func coordsEqual(a overpassCoord, b overpassCoord) bool {
	return round6(a.Lat) == round6(b.Lat) && round6(a.Lon) == round6(b.Lon)
}

func round6(val float64) float64 {
	return math.Round(val*1e6) / 1e6
}

func polygonCentroid(points []overpassCoord) (float64, float64) {
	if len(points) == 0 {
		return 0, 0
	}
	var sumLat, sumLon float64
	for _, point := range points {
		sumLat += point.Lat
		sumLon += point.Lon
	}
	count := float64(len(points))
	return sumLat / count, sumLon / count
}

func maxDistanceMeters(lat float64, lon float64, points []overpassCoord) float64 {
	var maxDistance float64
	for _, point := range points {
		distance := haversineMeters(lat, lon, point.Lat, point.Lon)
		if distance > maxDistance {
			maxDistance = distance
		}
	}
	return maxDistance
}

func haversineMeters(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	const earthRadius = 6371000.0
	rad := func(deg float64) float64 {
		return deg * math.Pi / 180
	}

	dLat := rad(lat2 - lat1)
	dLon := rad(lon2 - lon1)
	lat1Rad := rad(lat1)
	lat2Rad := rad(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}
