package processors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/deep_priest"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	zoneKindPatternTileBucket = "crew-profile-icons"
	zoneKindPatternTileSize   = 256
)

type GenerateZoneKindPatternTileProcessor struct {
	dbClient         db.DbClient
	deepPriestClient deep_priest.DeepPriest
	awsClient        aws.AWSClient
}

func NewGenerateZoneKindPatternTileProcessor(
	dbClient db.DbClient,
	deepPriestClient deep_priest.DeepPriest,
	awsClient aws.AWSClient,
) GenerateZoneKindPatternTileProcessor {
	log.Println("Initializing GenerateZoneKindPatternTileProcessor")
	return GenerateZoneKindPatternTileProcessor{
		dbClient:         dbClient,
		deepPriestClient: deepPriestClient,
		awsClient:        awsClient,
	}
}

func (p *GenerateZoneKindPatternTileProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing generate zone kind pattern tile task: %v", task.Type())

	var payload jobs.GenerateZoneKindPatternTileTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if payload.ZoneKindID == uuid.Nil {
		return fmt.Errorf("zoneKindId is required")
	}

	zoneKind, err := p.dbClient.ZoneKind().FindByID(ctx, payload.ZoneKindID)
	if err != nil {
		return fmt.Errorf("failed to find zone kind: %w", err)
	}
	if zoneKind == nil {
		return fmt.Errorf("zone kind not found")
	}

	prompt := strings.TrimSpace(payload.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(zoneKind.PatternTilePrompt)
	}
	if prompt == "" {
		prompt = fmt.Sprintf(
			"Create a seamless repeating square fantasy map texture tile for the %s zone kind. Keep it subtle, tileable, and free of text or borders.",
			zoneKind.Name,
		)
	}

	zoneKind.PatternTilePrompt = prompt
	zoneKind.PatternTileGenerationStatus = models.ZoneKindPatternTileGenerationStatusInProgress
	zoneKind.PatternTileGenerationError = ""
	if err := p.dbClient.ZoneKind().Update(ctx, zoneKind); err != nil {
		return fmt.Errorf("failed to mark zone kind tile generation in progress: %w", err)
	}

	request := deep_priest.GenerateImageRequest{Prompt: prompt}
	deep_priest.ApplyGenerateImageDefaults(&request)

	imagePayload, err := p.deepPriestClient.GenerateImage(request)
	if err != nil {
		return p.markFailed(ctx, zoneKind, err)
	}

	imageBytes, err := decodeImagePayload(imagePayload)
	if err != nil {
		return p.markFailed(ctx, zoneKind, err)
	}

	tileBytes, err := prepareZoneKindPatternTile(imageBytes)
	if err != nil {
		return p.markFailed(ctx, zoneKind, err)
	}

	imageURL, err := p.uploadImage(zoneKind, tileBytes)
	if err != nil {
		return p.markFailed(ctx, zoneKind, err)
	}

	zoneKind.PatternTileURL = imageURL
	zoneKind.PatternTilePrompt = prompt
	zoneKind.PatternTileGenerationStatus = models.ZoneKindPatternTileGenerationStatusComplete
	zoneKind.PatternTileGenerationError = ""
	if err := p.dbClient.ZoneKind().Update(ctx, zoneKind); err != nil {
		return fmt.Errorf("failed to update zone kind tile url: %w", err)
	}

	return nil
}

func (p *GenerateZoneKindPatternTileProcessor) markFailed(
	ctx context.Context,
	zoneKind *models.ZoneKind,
	err error,
) error {
	if zoneKind != nil {
		zoneKind.PatternTileGenerationStatus = models.ZoneKindPatternTileGenerationStatusFailed
		zoneKind.PatternTileGenerationError = strings.TrimSpace(err.Error())
		if updateErr := p.dbClient.ZoneKind().Update(ctx, zoneKind); updateErr != nil {
			log.Printf("Failed to mark zone kind pattern tile generation failed: %v", updateErr)
		}
	}
	return err
}

func prepareZoneKindPatternTile(imageBytes []byte) ([]byte, error) {
	decoded, err := imaging.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode pattern tile image: %w", err)
	}

	tile := imaging.Fill(decoded, zoneKindPatternTileSize, zoneKindPatternTileSize, imaging.Center, imaging.Lanczos)
	out := imaging.Clone(tile)
	bounds := out.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := color.NRGBAModel.Convert(out.At(x, y)).(color.NRGBA)
			if rgba.A == 0 {
				continue
			}
			luma := (int(rgba.R)*299 + int(rgba.G)*587 + int(rgba.B)*114) / 1000
			darkness := 255 - luma
			if darkness < 18 {
				out.SetNRGBA(x, y, color.NRGBA{R: rgba.R, G: rgba.G, B: rgba.B, A: 0})
				continue
			}
			alpha := clampZoneKindPatternAlpha(max(42, min(190, darkness*2)))
			if int(rgba.A) < alpha {
				alpha = int(rgba.A)
			}
			out.SetNRGBA(x, y, color.NRGBA{R: rgba.R, G: rgba.G, B: rgba.B, A: uint8(alpha)})
		}
	}

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, out, imaging.PNG, imaging.PNGCompressionLevel(png.BestCompression)); err != nil {
		return nil, fmt.Errorf("failed to encode pattern tile image: %w", err)
	}
	return buf.Bytes(), nil
}

func clampZoneKindPatternAlpha(value int) int {
	switch {
	case value < 0:
		return 0
	case value > 255:
		return 255
	default:
		return value
	}
}

func (p *GenerateZoneKindPatternTileProcessor) uploadImage(
	zoneKind *models.ZoneKind,
	imageBytes []byte,
) (string, error) {
	key := fmt.Sprintf(
		"zone-kind-patterns/%s-%d.png",
		models.NormalizeZoneKind(zoneKind.Slug),
		time.Now().UnixNano(),
	)
	return p.awsClient.UploadImageToS3(zoneKindPatternTileBucket, key, imageBytes)
}
