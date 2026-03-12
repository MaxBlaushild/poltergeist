package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/hibiken/asynq"
	"github.com/paulmach/orb"
)

const zoneFlavorPromptTemplate = `
You are writing zone flavor text for a fantasy MMORPG-style location-based game.

Zone:
- name: %s
- current description: %s

Geometry context:
%s

Return JSON only:
{
  "description": "2-4 vivid sentences, about 55-110 words"
}

Rules:
- Write in the same adventurous, mystical voice as a fantasy MMO map or quest journal.
- Base the fiction on the shape, scale, and positioning cues implied by the coordinates.
- Treat the real-world geometry as inspiration for in-world terrain, districts, routes, choke points, shorelines, courtyards, or edges.
- Do not mention GPS, coordinates, polygons, latitude, longitude, OpenStreetMap, apps, or modern map tooling.
- Do not mention brands, real businesses, or modern infrastructure by name.
- Keep it useful as a zone description players might read in the UI.
- Avoid second-person instructions and avoid explicit quest hooks; this is setting flavor, not a task prompt.
`

type zoneFlavorGenerationResponse struct {
	Description string `json:"description"`
}

type GenerateZoneFlavorProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
}

func NewGenerateZoneFlavorProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
) GenerateZoneFlavorProcessor {
	log.Println("Initializing GenerateZoneFlavorProcessor")
	return GenerateZoneFlavorProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
	}
}

func (p *GenerateZoneFlavorProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate zone flavor task: %v", task.Type())

	var payload jobs.GenerateZoneFlavorTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	job, err := p.dbClient.ZoneFlavorGenerationJob().FindByID(ctx, payload.JobID)
	if err != nil {
		return err
	}
	if job == nil {
		log.Printf("Zone flavor generation job %s not found", payload.JobID)
		return nil
	}

	job.Status = models.ZoneFlavorGenerationStatusInProgress
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneFlavorGenerationJob().Update(ctx, job); err != nil {
		return err
	}

	if err := p.generateFlavor(ctx, job); err != nil {
		return p.failZoneFlavorGenerationJob(ctx, job, err)
	}

	return nil
}

func (p *GenerateZoneFlavorProcessor) generateFlavor(ctx context.Context, job *models.ZoneFlavorGenerationJob) error {
	zone, err := p.dbClient.Zone().FindByID(ctx, job.ZoneID)
	if err != nil {
		return fmt.Errorf("failed to load zone: %w", err)
	}
	if zone == nil {
		return fmt.Errorf("zone not found")
	}

	zoneName := strings.TrimSpace(zone.Name)
	if zoneName == "" {
		zoneName = "Unknown Zone"
	}
	currentDescription := strings.TrimSpace(zone.Description)
	if currentDescription == "" {
		currentDescription = "none"
	}

	prompt := fmt.Sprintf(
		zoneFlavorPromptTemplate,
		zoneName,
		currentDescription,
		buildZoneFlavorGeometrySummary(*zone),
	)
	answer, err := p.deepPriestClient.PetitionTheFount(&deep_priest.Question{Question: prompt})
	if err != nil {
		return fmt.Errorf("failed to generate zone flavor: %w", err)
	}

	var generated zoneFlavorGenerationResponse
	if err := json.Unmarshal([]byte(extractGeneratedJSONObject(answer.Answer)), &generated); err != nil {
		return fmt.Errorf("failed to parse generated zone flavor payload: %w", err)
	}

	description := sanitizeZoneFlavorDescription(generated.Description)
	if description == "" {
		return fmt.Errorf("generated zone flavor was empty")
	}

	if err := p.dbClient.Zone().UpdateNameAndDescription(ctx, zone.ID, zone.Name, description); err != nil {
		return fmt.Errorf("failed to update zone description: %w", err)
	}

	job.Status = models.ZoneFlavorGenerationStatusCompleted
	job.GeneratedDescription = &description
	job.ErrorMessage = nil
	job.UpdatedAt = time.Now()
	if err := p.dbClient.ZoneFlavorGenerationJob().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update zone flavor generation job: %w", err)
	}

	return nil
}

func buildZoneFlavorGeometrySummary(zone models.Zone) string {
	lines := []string{
		fmt.Sprintf("- zone center: latitude %.6f, longitude %.6f", zone.Latitude, zone.Longitude),
	}

	polygon := zone.GetPolygon()
	if len(polygon) == 0 || len(polygon[0]) == 0 {
		lines = append(lines, "- boundary: unavailable")
		return strings.Join(lines, "\n")
	}

	ring := polygon[0]
	bounds := polygon.Bound()
	lines = append(lines,
		fmt.Sprintf("- boundary vertices: %d", len(ring)),
		fmt.Sprintf(
			"- bounding box: south %.6f, west %.6f, north %.6f, east %.6f",
			bounds.Min.Y(),
			bounds.Min.X(),
			bounds.Max.Y(),
			bounds.Max.X(),
		),
		fmt.Sprintf("- sampled outline: %s", strings.Join(sampleZoneRing(ring, 12), " | ")),
	)

	return strings.Join(lines, "\n")
}

func sampleZoneRing(ring orb.Ring, maxPoints int) []string {
	if len(ring) == 0 {
		return nil
	}
	if maxPoints <= 0 {
		maxPoints = 1
	}
	limit := len(ring)
	if limit > 1 && ring[0].Equal(ring[len(ring)-1]) {
		limit--
	}
	if limit <= 0 {
		limit = len(ring)
	}
	if limit <= maxPoints {
		out := make([]string, 0, limit)
		for i := 0; i < limit; i++ {
			out = append(out, fmt.Sprintf("(%.6f, %.6f)", ring[i].Y(), ring[i].X()))
		}
		return out
	}

	out := make([]string, 0, maxPoints)
	for i := 0; i < maxPoints; i++ {
		index := int(math.Round(float64(i) * float64(limit-1) / float64(maxPoints-1)))
		if index < 0 {
			index = 0
		}
		if index >= limit {
			index = limit - 1
		}
		out = append(out, fmt.Sprintf("(%.6f, %.6f)", ring[index].Y(), ring[index].X()))
	}
	return out
}

func sanitizeZoneFlavorDescription(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	text = strings.Join(strings.Fields(text), " ")
	return text
}

func (p *GenerateZoneFlavorProcessor) failZoneFlavorGenerationJob(
	ctx context.Context,
	job *models.ZoneFlavorGenerationJob,
	err error,
) error {
	msg := err.Error()
	job.Status = models.ZoneFlavorGenerationStatusFailed
	job.ErrorMessage = &msg
	job.UpdatedAt = time.Now()
	if updateErr := p.dbClient.ZoneFlavorGenerationJob().Update(ctx, job); updateErr != nil {
		log.Printf("Failed to mark zone flavor generation job %s as failed: %v", job.ID, updateErr)
	}
	return err
}
