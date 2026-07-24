package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/MaxBlaushild/job-runner/internal/config"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/generate"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/geomhash"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/pricing"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/procexec"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/slice"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/stlbbox"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/validate"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/datatypes"
)

// sliceFunc matches slice.Slice's signature — injectable so tests can stand
// in a fake slicer without needing a real PrusaSlicer/OrcaSlicer binary
// installed, while still exercising the real OpenSCAD render + cache path.
type sliceFunc func(ctx context.Context, cfg slice.Config, stlPath string) (*slice.Result, error)

// GenerateReefFullProcessor is R-5.1's add-to-cart pipeline: generate full
// geometry, slice it, run R-5.2's six rejection rules, price it, and cache
// the result by geometry_hash (R-3.3) so an identical configuration never
// regenerates or re-slices.
type GenerateReefFullProcessor struct {
	dbClient  db.DbClient
	awsClient aws.AWSClient
	cfg       config.PublicConfig
	slice     sliceFunc
}

func NewGenerateReefFullProcessor(dbClient db.DbClient, awsClient aws.AWSClient, cfg config.PublicConfig) *GenerateReefFullProcessor {
	return &GenerateReefFullProcessor{
		dbClient:  dbClient,
		awsClient: awsClient,
		cfg:       cfg,
		slice:     slice.Slice,
	}
}

func (p *GenerateReefFullProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload jobs.GenerateReefFullTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("generate_reef_full: unmarshal payload: %w", err)
	}
	log.Printf("[reef] processing full generation for configuration %s (job %s)", payload.ConfigurationID, payload.JobID)

	if err := p.dbClient.ReefGenerationJob().IncrementAttempts(ctx, payload.JobID); err != nil {
		log.Printf("[reef] failed to mark job %s running: %v", payload.JobID, err)
	}

	if err := p.process(ctx, payload); err != nil {
		log.Printf("[reef] job %s failed: %v", payload.JobID, err)
		if statusErr := p.dbClient.ReefGenerationJob().UpdateStatus(ctx, payload.JobID, models.ReefGenerationJobStatusFailed, err.Error()); statusErr != nil {
			log.Printf("[reef] additionally failed to record job failure: %v", statusErr)
		}
		return err
	}

	return p.dbClient.ReefGenerationJob().UpdateStatus(ctx, payload.JobID, models.ReefGenerationJobStatusCompleted, "")
}

func (p *GenerateReefFullProcessor) process(ctx context.Context, payload jobs.GenerateReefFullTaskPayload) error {
	cfgRow, err := p.dbClient.ReefConfiguration().FindByID(ctx, payload.ConfigurationID)
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	product, err := p.dbClient.ReefProduct().FindByID(ctx, cfgRow.ProductID)
	if err != nil {
		return fmt.Errorf("load product: %w", err)
	}

	schema, err := p.dbClient.ReefParameterSchema().FindActiveByProductID(ctx, product.ID)
	if err != nil {
		return fmt.Errorf("load parameter schema: %w", err)
	}

	module, err := generate.Get(schema.GeneratorModule)
	if err != nil {
		return err
	}

	var params map[string]interface{}
	if err := json.Unmarshal(cfgRow.Params, &params); err != nil {
		return fmt.Errorf("decode params: %w", err)
	}

	renderCfg := generate.RenderConfig{
		OpenSCADBin: p.cfg.ReefOpenSCADBin,
		BaseTempDir: os.TempDir(),
		Timeout:     time.Duration(p.cfg.ReefSubprocessTimeoutSec) * time.Second,
		MemoryMB:    p.cfg.ReefSubprocessMemoryMB,
	}

	openscadVersion, err := generate.Version(ctx, renderCfg)
	if err != nil {
		return fmt.Errorf("resolve openscad version: %w", err)
	}

	hash, err := geomhash.Hash(product.Slug, module.Version(), openscadVersion, json.RawMessage(cfgRow.Params))
	if err != nil {
		return fmt.Errorf("hash geometry: %w", err)
	}

	// R-3.3: identical inputs must never regenerate or re-slice.
	existing, err := p.dbClient.ReefSliceResult().FindByGeometryHash(ctx, hash)
	if err != nil {
		return fmt.Errorf("check slice_result cache: %w", err)
	}
	if existing != nil && existing.Status != models.ReefSliceStatusPending {
		log.Printf("[reef] geometry_hash %s already resolved, serving from cache (no regeneration/re-slice)", hash)
		return p.applyResult(ctx, cfgRow, product.Slug, hash, existing)
	}

	scad, err := module.SCAD(params, generate.Full)
	if err != nil {
		return fmt.Errorf("render scad source: %w", err)
	}

	renderResult, err := generate.Render(ctx, renderCfg, scad, openscadVersion)
	if err != nil {
		return fmt.Errorf("openscad render: %w", err)
	}
	defer procexec.Cleanup(renderResult.WorkDir)

	box, err := stlbbox.FromFile(renderResult.STLPath)
	if err != nil {
		return fmt.Errorf("compute bounding box: %w", err)
	}

	analysis, err := module.Analyze(params)
	if err != nil {
		return fmt.Errorf("analyze geometry: %w", err)
	}

	sliceCfg := slice.Config{
		SlicerBin:   p.cfg.ReefSlicerBin,
		BaseTempDir: os.TempDir(),
		Timeout:     time.Duration(p.cfg.ReefSubprocessTimeoutSec) * time.Second,
		MemoryMB:    p.cfg.ReefSubprocessMemoryMB,
	}
	sliceResult, err := p.slice(ctx, sliceCfg, renderResult.STLPath)
	if err != nil {
		return fmt.Errorf("slice: %w", err)
	}
	if sliceResult.GCodePath != "" {
		defer procexec.Cleanup(filepath.Dir(sliceResult.GCodePath))
	}

	meta := validate.Metadata{
		BboxMaxDimensionMm: box.MaxDimensionMm(),
		SupportRequired:    sliceResult.SupportRequired,
		MinWallMm:          analysis.MinWallMm,
		PrintTimeS:         sliceResult.PrintTimeS,
		WeightG:            sliceResult.WeightG,
		SealedVoid:         analysis.SealedVoid,
		DrainPathMm:        analysis.DrainPathMm,
		HasInternalCavity:  analysis.HasInternalCavity,
	}
	thresholds := validate.Thresholds{
		MaxBboxMm:      p.cfg.ReefMaxBboxMm,
		MinWallMm:      p.cfg.ReefMinWallMm,
		MaxPrintTimeS:  p.cfg.ReefMaxPrintTimeS,
		MaxWeightG:     p.cfg.ReefMaxWeightG,
		MinDrainPathMm: p.cfg.ReefMinDrainPathMm,
	}
	rejection := validate.Validate(meta, thresholds)

	bboxJSON, err := json.Marshal(map[string]float64{"xMm": box.XMm(), "yMm": box.YMm(), "zMm": box.ZMm()})
	if err != nil {
		return fmt.Errorf("encode bbox: %w", err)
	}

	weightG := sliceResult.WeightG
	printTimeS := sliceResult.PrintTimeS
	minWallMm := analysis.MinWallMm
	sealedVoid := analysis.SealedVoid
	supportRequired := sliceResult.SupportRequired
	plateFits := box.MaxDimensionMm() <= thresholds.MaxBboxMm

	sliceRow := &models.ReefSliceResult{
		GeometryHash:    hash,
		ProductID:       product.ID,
		WeightG:         &weightG,
		PrintTimeS:      &printTimeS,
		BboxMm:          datatypes.JSON(bboxJSON),
		PlateFits:       &plateFits,
		SupportRequired: &supportRequired,
		MinWallMm:       &minWallMm,
		SealedVoid:      &sealedVoid,
		Warnings:        datatypes.JSON([]byte(`[]`)),
		SlicerVersion:   sliceResult.SlicerVersion,
		OpenSCADVersion: openscadVersion,
	}

	if rejection != nil {
		sliceRow.Status = models.ReefSliceStatusRejected
		sliceRow.RejectionRule = string(rejection.Rule)
		sliceRow.RejectionReason = rejection.Reason
	} else {
		priceCents := pricing.Price(sliceResult.WeightG, sliceResult.PrintTimeS, pricing.Rates{
			SetupFeeCents:             p.cfg.ReefPriceSetupFeeCents,
			MaterialRateCentsPerGram:  p.cfg.ReefPriceMaterialRateCentsPerGram,
			MachineRateCentsPerMinute: p.cfg.ReefPriceMachineRateCentsPerMinute,
			FulfillmentFeeCents:       p.cfg.ReefPriceFulfillmentFeeCents,
			MarginMultiplier:          p.cfg.ReefPriceMarginMultiplier,
		})

		stlBytes, err := os.ReadFile(renderResult.STLPath)
		if err != nil {
			return fmt.Errorf("read stl: %w", err)
		}
		stlKey := fmt.Sprintf("reef/stl/%s.stl", hash)
		if _, err := p.awsClient.UploadImageToS3(p.cfg.ReefS3Bucket, stlKey, stlBytes); err != nil {
			return fmt.Errorf("upload stl to s3: %w", err)
		}

		sliceRow.Status = models.ReefSliceStatusValid
		sliceRow.STLKey = stlKey
		sliceRow.PriceCents = &priceCents
	}

	if err := p.dbClient.ReefSliceResult().Create(ctx, sliceRow); err != nil {
		return fmt.Errorf("persist slice result: %w", err)
	}
	// Create is a no-op on conflict (a concurrent request may have won the
	// race to populate the cache first) — re-read so we apply whichever row
	// actually won, not necessarily the one we just computed.
	winner, err := p.dbClient.ReefSliceResult().FindByGeometryHash(ctx, hash)
	if err != nil || winner == nil {
		winner = sliceRow
	}

	return p.applyResult(ctx, cfgRow, product.Slug, hash, winner)
}

func (p *GenerateReefFullProcessor) applyResult(ctx context.Context, cfgRow *models.ReefConfiguration, productSlug, hash string, sliceRow *models.ReefSliceResult) error {
	cfgRow.GeometryHash = &hash
	if sliceRow.Status == models.ReefSliceStatusValid {
		cfgRow.Status = models.ReefConfigurationStatusValid
		cfgRow.PriceCents = sliceRow.PriceCents
		cfgRow.RejectionReason = ""
	} else {
		cfgRow.Status = models.ReefConfigurationStatusRejected
		cfgRow.RejectionReason = sliceRow.RejectionReason
	}
	if err := p.dbClient.ReefConfiguration().Update(ctx, cfgRow); err != nil {
		return fmt.Errorf("update configuration: %w", err)
	}

	if sliceRow.Status == models.ReefSliceStatusRejected {
		// R-5.4: every rejection writes product, parameters, and rule
		// triggered.
		metadata, _ := json.Marshal(map[string]interface{}{
			"params": json.RawMessage(cfgRow.Params),
			"reason": sliceRow.RejectionReason,
		})
		configurationID := cfgRow.ID
		if err := p.dbClient.ReefEvent().Create(ctx, &models.ReefEvent{
			ID:              uuid.New(),
			EventType:       models.ReefEventValidationRejected,
			ProductSlug:     productSlug,
			ConfigurationID: &configurationID,
			Rule:            sliceRow.RejectionRule,
			Metadata:        datatypes.JSON(metadata),
		}); err != nil {
			log.Printf("[reef] failed to record validation_rejected event: %v", err)
		}
	}

	return nil
}
