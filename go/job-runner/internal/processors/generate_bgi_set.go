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
	"github.com/MaxBlaushild/poltergeist/pkg/reef/set"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/slice"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/stlbbox"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/validate"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/datatypes"
)

// resolvedTrayRecord is the persisted shape of one bgi_set_resolutions.resolved_trays
// entry — includes GeometryHash (unlike set.ResolvedTray, which doesn't know
// about geometry hashing) so a config_hash cache hit can skip straight to
// per-tray slice-result lookups without re-running set.Assemble.
type resolvedTrayRecord struct {
	TrayTemplateID  uuid.UUID              `json:"trayTemplateId"`
	GeneratorModule string                 `json:"generatorModule"`
	Params          map[string]interface{} `json:"params"`
	Quantity        int                    `json:"quantity"`
	GeometryHash    string                 `json:"geometryHash"`
	HeightMm        float64                `json:"heightMm"`
}

// GenerateBgiSetProcessor is bgi-site's analog of GenerateReefFullProcessor,
// generalized from one part to N (R-3.3): resolve which trays a
// configuration needs (set.Assemble, cached by config_hash), then
// render/slice/validate/price each one (cached by geometry_hash, reusing
// the exact same per-part pipeline reef's processor already established).
type GenerateBgiSetProcessor struct {
	dbClient  db.DbClient
	awsClient aws.AWSClient
	cfg       config.PublicConfig
	slice     sliceFunc
}

func NewGenerateBgiSetProcessor(dbClient db.DbClient, awsClient aws.AWSClient, cfg config.PublicConfig) *GenerateBgiSetProcessor {
	return &GenerateBgiSetProcessor{
		dbClient:  dbClient,
		awsClient: awsClient,
		cfg:       cfg,
		slice:     slice.Slice,
	}
}

func (p *GenerateBgiSetProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload jobs.GenerateBgiSetTaskPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("generate_bgi_set: unmarshal payload: %w", err)
	}
	log.Printf("[bgi] processing set generation for configuration %s (job %s)", payload.ConfigurationID, payload.JobID)

	if err := p.dbClient.BgiGenerationJob().IncrementAttempts(ctx, payload.JobID); err != nil {
		log.Printf("[bgi] failed to mark job %s running: %v", payload.JobID, err)
	}

	if err := p.process(ctx, payload); err != nil {
		log.Printf("[bgi] job %s failed: %v", payload.JobID, err)
		if statusErr := p.dbClient.BgiGenerationJob().UpdateStatus(ctx, payload.JobID, models.BgiGenerationJobStatusFailed, err.Error()); statusErr != nil {
			log.Printf("[bgi] additionally failed to record job failure: %v", statusErr)
		}
		return err
	}

	return p.dbClient.BgiGenerationJob().UpdateStatus(ctx, payload.JobID, models.BgiGenerationJobStatusCompleted, "")
}

func (p *GenerateBgiSetProcessor) process(ctx context.Context, payload jobs.GenerateBgiSetTaskPayload) error {
	cfgRow, err := p.dbClient.BgiConfiguration().FindByID(ctx, payload.ConfigurationID)
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	product, err := p.dbClient.BgiProduct().FindByID(ctx, cfgRow.ProductID)
	if err != nil {
		return fmt.Errorf("load product: %w", err)
	}

	var params struct {
		SleeveProfileID string `json:"sleeveProfileId"`
		BoxProfileID    string `json:"boxProfileId"`
		Color           string `json:"color"`
	}
	if err := json.Unmarshal(cfgRow.Params, &params); err != nil {
		return fmt.Errorf("decode params: %w", err)
	}

	sleeveID, err := uuid.Parse(params.SleeveProfileID)
	if err != nil {
		return fmt.Errorf("invalid sleeveProfileId: %w", err)
	}
	boxID, err := uuid.Parse(params.BoxProfileID)
	if err != nil {
		return fmt.Errorf("invalid boxProfileId: %w", err)
	}

	sleeveProfile, err := p.dbClient.BgiSleeveProfile().FindByID(ctx, sleeveID)
	if err != nil {
		return fmt.Errorf("load sleeve profile: %w", err)
	}
	boxProfile, err := p.dbClient.BgiBoxProfile().FindByID(ctx, boxID)
	if err != nil {
		return fmt.Errorf("load box profile: %w", err)
	}

	renderCfg := generate.RenderConfig{
		OpenSCADBin: p.cfg.BgiOpenSCADBin,
		BaseTempDir: os.TempDir(),
		Timeout:     time.Duration(p.cfg.BgiSubprocessTimeoutSec) * time.Second,
		MemoryMB:    p.cfg.BgiSubprocessMemoryMB,
	}
	openscadVersion, err := generate.Version(ctx, renderCfg)
	if err != nil {
		return fmt.Errorf("resolve openscad version: %w", err)
	}

	// R-4.3: config_hash caches "which trays, how many, do they fit" —
	// independent of and cheaper than geometry_hash's per-tray render+slice
	// cache below. A hit here skips set.Assemble entirely.
	configHash := set.ConfigHash(product.Slug, nil, sleeveProfile.ClassKey, boxProfile.Slug, params.Color)

	var trays []resolvedTrayRecord
	var unassembled []string
	var assembledHeightMm float64
	var fitsBox bool

	existingResolution, err := p.dbClient.BgiSetResolution().FindByConfigHash(ctx, configHash)
	if err != nil {
		return fmt.Errorf("check set_resolution cache: %w", err)
	}
	if existingResolution != nil {
		log.Printf("[bgi] config_hash %s already resolved, reusing cached recipe (no re-assembly)", configHash)
		if err := json.Unmarshal(existingResolution.ResolvedTrays, &trays); err != nil {
			return fmt.Errorf("decode cached resolved_trays: %w", err)
		}
		if err := json.Unmarshal(existingResolution.UnassembledComponents, &unassembled); err != nil {
			return fmt.Errorf("decode cached unassembled_components: %w", err)
		}
		if existingResolution.AssembledHeightMm != nil {
			assembledHeightMm = *existingResolution.AssembledHeightMm
		}
		if existingResolution.FitsBox != nil {
			fitsBox = *existingResolution.FitsBox
		}
	} else {
		manifestRows, err := p.dbClient.BgiComponentManifest().FindByGameID(ctx, product.GameID, nil)
		if err != nil {
			return fmt.Errorf("load component manifests: %w", err)
		}
		manifests := make([]set.ComponentManifest, 0, len(manifestRows))
		templates := make([]set.TrayTemplate, 0, len(manifestRows))
		for _, m := range manifestRows {
			cm := set.ComponentManifest{ComponentType: m.ComponentType, Count: m.Count}
			if m.CardWidthMm != nil {
				cm.CardWidthMm = *m.CardWidthMm
			}
			if m.CardHeightMm != nil {
				cm.CardHeightMm = *m.CardHeightMm
			}
			manifests = append(manifests, cm)

			tmpl, err := p.dbClient.BgiTrayTemplate().FindByComponentType(ctx, m.ComponentType)
			if err != nil {
				return fmt.Errorf("load tray template for %s: %w", m.ComponentType, err)
			}
			if tmpl != nil {
				templates = append(templates, set.TrayTemplate{ID: tmpl.ID, ComponentType: tmpl.ComponentType, GeneratorModule: tmpl.GeneratorModule})
			}
		}

		resolution, err := set.Assemble(manifests, templates,
			set.BoxProfile{InteriorLengthMm: boxProfile.InteriorLengthMm, InteriorWidthMm: boxProfile.InteriorWidthMm, InteriorDepthMm: boxProfile.InteriorDepthMm},
			set.SleeveProfile{TotalCardThicknessMm: sleeveProfile.TotalCardThicknessMm()},
			params.Color,
		)
		if err != nil {
			return fmt.Errorf("assemble set: %w", err)
		}

		for _, rt := range resolution.ResolvedTrays {
			paramsJSON, err := json.Marshal(rt.Params)
			if err != nil {
				return fmt.Errorf("encode tray params: %w", err)
			}
			module, err := generate.Get(rt.GeneratorModule)
			if err != nil {
				return fmt.Errorf("resolve module %q: %w", rt.GeneratorModule, err)
			}
			geometryHash, err := geomhash.Hash(rt.GeneratorModule, module.Version(), openscadVersion, paramsJSON)
			if err != nil {
				return fmt.Errorf("hash tray geometry: %w", err)
			}
			trays = append(trays, resolvedTrayRecord{
				TrayTemplateID:  rt.TrayTemplateID,
				GeneratorModule: rt.GeneratorModule,
				Params:          rt.Params,
				Quantity:        rt.Quantity,
				GeometryHash:    geometryHash,
				HeightMm:        rt.HeightMm,
			})
		}
		unassembled = resolution.UnassembledComponents
		assembledHeightMm = resolution.AssembledHeightMm
		fitsBox = resolution.FitsBox

		traysJSON, err := json.Marshal(trays)
		if err != nil {
			return fmt.Errorf("encode resolved trays: %w", err)
		}
		unassembledJSON, err := json.Marshal(unassembled)
		if err != nil {
			return fmt.Errorf("encode unassembled components: %w", err)
		}
		expansionIDsJSON := datatypes.JSON([]byte(`[]`))
		if err := p.dbClient.BgiSetResolution().Create(ctx, &models.BgiSetResolution{
			ConfigHash:            configHash,
			ProductID:             product.ID,
			GameID:                product.GameID,
			ExpansionIDs:          expansionIDsJSON,
			SleeveProfileID:       sleeveID,
			BoxProfileID:          boxID,
			ResolvedTrays:         datatypes.JSON(traysJSON),
			UnassembledComponents: datatypes.JSON(unassembledJSON),
			AssembledHeightMm:     &assembledHeightMm,
			FitsBox:               &fitsBox,
		}); err != nil {
			return fmt.Errorf("persist set resolution: %w", err)
		}
	}

	// R-6.2 rule 2 (the lid-won't-close rule) — checked before any render,
	// exactly like FragRack.ValidateParams catches an impossible geometry
	// before paying for a render+slice cycle.
	if !fitsBox {
		return p.reject(ctx, cfgRow, configHash, fmt.Sprintf(
			"This sleeve class needs the assembled trays to stand %.1fmm tall, which won't fit this box's %.1fmm interior depth (%s). Try a thinner sleeve class or a deeper box.",
			assembledHeightMm, boxProfile.InteriorDepthMm, verificationCaveat(boxProfile.Verified),
		))
	}

	var totalPriceCents int64
	var totalSetPrintTimeS int64
	for i := range trays {
		sliceRow, err := p.resolveTraySlice(ctx, trays[i], openscadVersion)
		if err != nil {
			return err
		}
		if sliceRow.Status == models.BgiTraySliceStatusRejected {
			// A single tray failing a platform rule rejects the whole set
			// (R-6.2 rule 1) — can't partially fulfill a tray set.
			return p.reject(ctx, cfgRow, configHash, sliceRow.RejectionReason)
		}
		if sliceRow.PriceCents != nil {
			totalPriceCents += *sliceRow.PriceCents * int64(trays[i].Quantity)
		}
		if sliceRow.PrintTimeS != nil {
			totalSetPrintTimeS += *sliceRow.PrintTimeS * int64(trays[i].Quantity)
		}
	}

	if p.cfg.BgiMaxSetPrintTimeS > 0 && totalSetPrintTimeS > p.cfg.BgiMaxSetPrintTimeS {
		return p.reject(ctx, cfgRow, configHash, fmt.Sprintf(
			"Estimated total print time for this set is %.1f hours, over the %.1f hour lead-time ceiling. Try a thinner sleeve class or fewer trays.",
			float64(totalSetPrintTimeS)/3600, float64(p.cfg.BgiMaxSetPrintTimeS)/3600,
		))
	}

	totalPriceCents += p.cfg.BgiSetAssemblyFeeCents

	cfgRow.ConfigHash = &configHash
	cfgRow.Status = models.BgiConfigurationStatusValid
	cfgRow.RejectionReason = ""
	cfgRow.PriceCents = &totalPriceCents
	if err := p.dbClient.BgiConfiguration().Update(ctx, cfgRow); err != nil {
		return fmt.Errorf("update configuration: %w", err)
	}
	return nil
}

func verificationCaveat(verified bool) string {
	if verified {
		return "verified"
	}
	return "unverified — pending physical measurement"
}

func (p *GenerateBgiSetProcessor) reject(ctx context.Context, cfgRow *models.BgiConfiguration, configHash, reason string) error {
	cfgRow.ConfigHash = &configHash
	cfgRow.Status = models.BgiConfigurationStatusRejected
	cfgRow.RejectionReason = reason
	if err := p.dbClient.BgiConfiguration().Update(ctx, cfgRow); err != nil {
		return fmt.Errorf("update rejected configuration: %w", err)
	}
	return nil
}

// resolveTraySlice returns the geometry_hash's slice result, computing it
// (render, slice, validate, price, upload) only on a cache miss — the exact
// same per-part pipeline GenerateReefFullProcessor.process uses, looped per
// resolved tray instead of run once.
func (p *GenerateBgiSetProcessor) resolveTraySlice(ctx context.Context, tray resolvedTrayRecord, openscadVersion string) (*models.BgiTraySliceResult, error) {
	existing, err := p.dbClient.BgiTraySliceResult().FindByGeometryHash(ctx, tray.GeometryHash)
	if err != nil {
		return nil, fmt.Errorf("check tray slice_result cache: %w", err)
	}
	if existing != nil && existing.Status != models.BgiTraySliceStatusPending {
		log.Printf("[bgi] geometry_hash %s already resolved, serving from cache (no regeneration/re-slice)", tray.GeometryHash)
		return existing, nil
	}

	module, err := generate.Get(tray.GeneratorModule)
	if err != nil {
		return nil, err
	}

	renderCfg := generate.RenderConfig{
		OpenSCADBin: p.cfg.BgiOpenSCADBin,
		BaseTempDir: os.TempDir(),
		Timeout:     time.Duration(p.cfg.BgiSubprocessTimeoutSec) * time.Second,
		MemoryMB:    p.cfg.BgiSubprocessMemoryMB,
	}

	scad, err := module.SCAD(tray.Params, generate.Full)
	if err != nil {
		return nil, fmt.Errorf("render scad source for %s: %w", tray.GeometryHash, err)
	}
	renderResult, err := generate.Render(ctx, renderCfg, scad, openscadVersion)
	if err != nil {
		return nil, fmt.Errorf("openscad render for %s: %w", tray.GeometryHash, err)
	}
	defer procexec.Cleanup(renderResult.WorkDir)

	box, err := stlbbox.FromFile(renderResult.STLPath)
	if err != nil {
		return nil, fmt.Errorf("compute bounding box for %s: %w", tray.GeometryHash, err)
	}
	analysis, err := module.Analyze(tray.Params)
	if err != nil {
		return nil, fmt.Errorf("analyze geometry for %s: %w", tray.GeometryHash, err)
	}

	sliceCfg := slice.Config{
		SlicerBin:           p.cfg.BgiSlicerBin,
		BaseTempDir:         os.TempDir(),
		Timeout:             time.Duration(p.cfg.BgiSubprocessTimeoutSec) * time.Second,
		MemoryMB:            p.cfg.BgiSubprocessMemoryMB,
		FilamentDensityGCm3: p.cfg.BgiFilamentDensityGCm3,
	}
	sliceResult, err := p.slice(ctx, sliceCfg, renderResult.STLPath)
	if err != nil {
		return nil, fmt.Errorf("slice %s: %w", tray.GeometryHash, err)
	}
	if sliceResult.GCodePath != "" {
		defer procexec.Cleanup(filepath.Dir(sliceResult.GCodePath))
	}

	meta := validate.Metadata{
		BboxMaxDimensionMm:     box.MaxDimensionMm(),
		SupportMaterialPercent: sliceResult.SupportMaterialPercent,
		MinWallMm:              analysis.MinWallMm,
		PrintTimeS:             sliceResult.PrintTimeS,
		WeightG:                sliceResult.WeightG,
		SealedVoid:             analysis.SealedVoid,
		DrainPathMm:            analysis.DrainPathMm,
		HasInternalCavity:      analysis.HasInternalCavity,
	}
	thresholds := validate.Thresholds{
		MaxBboxMm:             p.cfg.BgiMaxBboxMm,
		MinWallMm:             p.cfg.BgiMinWallMm,
		MaxPrintTimeS:         p.cfg.BgiMaxPrintTimeS,
		MaxWeightG:            p.cfg.BgiMaxWeightG,
		MaxSupportMaterialPct: p.cfg.BgiMaxSupportMaterialPct,
		// Open-top wells have no cavity/buoyancy concern at all — see
		// go/bgi-site/PLATFORM_FINDINGS.md.
		SealedVoidRuleEnabled: false,
	}
	rejection := validate.Validate(meta, thresholds)

	bboxJSON, err := json.Marshal(map[string]float64{"xMm": box.XMm(), "yMm": box.YMm(), "zMm": box.ZMm()})
	if err != nil {
		return nil, fmt.Errorf("encode bbox: %w", err)
	}

	weightG := sliceResult.WeightG
	printTimeS := sliceResult.PrintTimeS
	minWallMm := analysis.MinWallMm
	sealedVoid := analysis.SealedVoid
	supportMaterialPercent := sliceResult.SupportMaterialPercent
	supportRequired := supportMaterialPercent > 0
	plateFits := box.MaxDimensionMm() <= thresholds.MaxBboxMm

	sliceRow := &models.BgiTraySliceResult{
		GeometryHash:           tray.GeometryHash,
		TrayTemplateID:         tray.TrayTemplateID,
		WeightG:                &weightG,
		PrintTimeS:             &printTimeS,
		BboxMm:                 datatypes.JSON(bboxJSON),
		PlateFits:              &plateFits,
		SupportRequired:        &supportRequired,
		SupportMaterialPercent: &supportMaterialPercent,
		MinWallMm:              &minWallMm,
		SealedVoid:             &sealedVoid,
		Warnings:               datatypes.JSON([]byte(`[]`)),
		SlicerVersion:          sliceResult.SlicerVersion,
		OpenSCADVersion:        openscadVersion,
	}

	if rejection != nil {
		sliceRow.Status = models.BgiTraySliceStatusRejected
		sliceRow.RejectionRule = string(rejection.Rule)
		sliceRow.RejectionReason = rejection.Reason
	} else {
		priceCents := pricing.Price(sliceResult.WeightG, sliceResult.PrintTimeS, pricing.Rates{
			SetupFeeCents:             p.cfg.BgiPriceSetupFeeCents,
			MaterialRateCentsPerGram:  p.cfg.BgiPriceMaterialRateCentsPerGram,
			MachineRateCentsPerMinute: p.cfg.BgiPriceMachineRateCentsPerMinute,
			FulfillmentFeeCents:       p.cfg.BgiPriceFulfillmentFeeCents,
			MarginMultiplier:          p.cfg.BgiPriceMarginMultiplier,
		})

		stlBytes, err := os.ReadFile(renderResult.STLPath)
		if err != nil {
			return nil, fmt.Errorf("read stl for %s: %w", tray.GeometryHash, err)
		}
		stlKey := fmt.Sprintf("bgi/stl/%s.stl", tray.GeometryHash)
		if _, err := p.awsClient.UploadImageToS3(p.cfg.BgiS3Bucket, stlKey, stlBytes); err != nil {
			return nil, fmt.Errorf("upload stl to s3 for %s: %w", tray.GeometryHash, err)
		}

		sliceRow.Status = models.BgiTraySliceStatusValid
		sliceRow.STLKey = stlKey
		sliceRow.PriceCents = &priceCents
	}

	// Same race-closing reasoning as GenerateReefFullProcessor.process's
	// own Upsert call: only an atomic upsert (not check-then-act) can
	// correctly beat a concurrent preview/full request landing on this
	// exact geometry_hash mid-render.
	if err := p.dbClient.BgiTraySliceResult().Upsert(ctx, sliceRow); err != nil {
		return nil, fmt.Errorf("persist tray slice result: %w", err)
	}
	return sliceRow, nil
}
