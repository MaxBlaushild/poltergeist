package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/generate"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/geomhash"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/paramschema"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/procexec"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/set"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/stlbbox"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type configureRequest struct {
	ProductSlug string                 `json:"productSlug" binding:"required"`
	Params      map[string]interface{} `json:"params" binding:"required"`
	SessionID   string                 `json:"sessionId"`
}

type setParams struct {
	SleeveProfileID string `json:"sleeveProfileId"`
	BoxProfileID    string `json:"boxProfileId"`
	Color           string `json:"color"`
}

// resolveProduct loads a configurable product and its active schema,
// validating params against the schema along the way (R-4.4/R-4.5) — the
// bgi analog of reef's resolveModule. It stops short of resolving a
// generate.Module: bgi_parameter_schemas.generator_module is the sentinel
// "bgi_tray_set", not a real module slug (R-3.3) — resolveSet below is what
// routes through go/pkg/reef/set instead.
func (s *server) resolveProduct(c *gin.Context, req configureRequest) (*models.BgiProduct, *models.BgiParameterSchema, bool) {
	ctx := c.Request.Context()

	product, err := s.deps.DbClient.BgiProduct().FindBySlug(ctx, req.ProductSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return nil, nil, false
	}

	schema, err := s.deps.DbClient.BgiParameterSchema().FindActiveByProductID(ctx, product.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active parameter schema for this product"})
		return nil, nil, false
	}

	parsedSchema, err := paramschema.Parse(schema.Schema)
	if err != nil {
		internalError(c, "parse parameter schema", err)
		return nil, nil, false
	}
	if errs := paramschema.Validate(parsedSchema, req.Params); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid parameters", "details": errs})
		return nil, nil, false
	}

	return product, schema, true
}

// resolveSet loads the sleeve/box/manifest rows a set.Assemble call needs
// and runs it — this is R-3.3's structural divergence from reef's
// single-module pattern: many parts, not one. Assemble itself is cheap
// (analytical Module.Analyze() calls, no subprocess), safe to run
// synchronously in an HTTP handler.
func (s *server) resolveSet(ctx context.Context, product *models.BgiProduct, params map[string]interface{}) (set.Resolution, setParams, *models.BgiSleeveProfile, *models.BgiBoxProfile, error) {
	var sp setParams
	raw, err := json.Marshal(params)
	if err != nil {
		return set.Resolution{}, sp, nil, nil, fmt.Errorf("encode params: %w", err)
	}
	if err := json.Unmarshal(raw, &sp); err != nil {
		return set.Resolution{}, sp, nil, nil, fmt.Errorf("decode set params: %w", err)
	}

	sleeveID, err := uuid.Parse(sp.SleeveProfileID)
	if err != nil {
		return set.Resolution{}, sp, nil, nil, fmt.Errorf("invalid sleeveProfileId")
	}
	boxID, err := uuid.Parse(sp.BoxProfileID)
	if err != nil {
		return set.Resolution{}, sp, nil, nil, fmt.Errorf("invalid boxProfileId")
	}

	sleeveProfile, err := s.deps.DbClient.BgiSleeveProfile().FindByID(ctx, sleeveID)
	if err != nil {
		return set.Resolution{}, sp, nil, nil, fmt.Errorf("sleeve profile not found")
	}
	boxProfile, err := s.deps.DbClient.BgiBoxProfile().FindByID(ctx, boxID)
	if err != nil {
		return set.Resolution{}, sp, nil, nil, fmt.Errorf("box profile not found")
	}

	manifestRows, err := s.deps.DbClient.BgiComponentManifest().FindByGameID(ctx, product.GameID, nil)
	if err != nil {
		return set.Resolution{}, sp, nil, nil, fmt.Errorf("load component manifests: %w", err)
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

		tmpl, err := s.deps.DbClient.BgiTrayTemplate().FindByComponentType(ctx, m.ComponentType)
		if err != nil {
			return set.Resolution{}, sp, nil, nil, fmt.Errorf("load tray template for %s: %w", m.ComponentType, err)
		}
		if tmpl != nil {
			templates = append(templates, set.TrayTemplate{ID: tmpl.ID, ComponentType: tmpl.ComponentType, GeneratorModule: tmpl.GeneratorModule})
		}
	}

	resolution, err := set.Assemble(manifests, templates,
		set.BoxProfile{InteriorLengthMm: boxProfile.InteriorLengthMm, InteriorWidthMm: boxProfile.InteriorWidthMm, InteriorDepthMm: boxProfile.InteriorDepthMm},
		set.SleeveProfile{TotalCardThicknessMm: sleeveProfile.TotalCardThicknessMm()},
		sp.Color,
	)
	if err != nil {
		return set.Resolution{}, sp, sleeveProfile, boxProfile, fmt.Errorf("assemble set: %w", err)
	}
	return resolution, sp, sleeveProfile, boxProfile, nil
}

func (s *server) renderConfig() generate.RenderConfig {
	return generate.RenderConfig{
		OpenSCADBin: s.deps.Config.Public.OpenSCADBin,
		BaseTempDir: os.TempDir(),
		Timeout:     time.Duration(s.deps.Config.Public.PreviewTimeoutSec) * time.Second,
		MemoryMB:    s.deps.Config.Public.SubprocessMemoryMB,
	}
}

type previewResponse struct {
	AssembledHeightMm     float64  `json:"assembledHeightMm"`
	BoxInteriorDepthMm    float64  `json:"boxInteriorDepthMm"`
	FitsBox               bool     `json:"fitsBox"`
	BoxVerified           bool     `json:"boxVerified"`
	DepthIsPlaceholder    bool     `json:"depthIsPlaceholder"`
	UnassembledComponents []string `json:"unassembledComponents,omitempty"`
	FirstTrayPreviewURL   string   `json:"firstTrayPreviewUrl,omitempty"`
	TrayCount             int      `json:"trayCount"`
}

// POST /api/bgi/configure/preview (R-8.1, R-3.4/R-5.3). Synchronous, like
// reef's own preview — the fit indicator (assembledHeightMm/fitsBox) is
// pure arithmetic from set.Assemble, available without rendering anything;
// only the *first* resolved tray is actually rendered to a mesh, since a
// full multi-tray render on every keystroke would be far too slow for
// R-2.6-style live preview (full renders happen only at configure/validate
// time, asynchronously).
func (s *server) configurePreview(c *gin.Context) {
	var req configureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = c.ClientIP()
	}
	if !s.limiter.allow(sessionID) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many preview requests, slow down"})
		return
	}

	product, _, ok := s.resolveProduct(c, req)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	resolution, _, sleeveProfile, boxProfile, err := s.resolveSet(ctx, product, req.Params)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	resp := previewResponse{
		AssembledHeightMm:     resolution.AssembledHeightMm,
		BoxInteriorDepthMm:    boxProfile.InteriorDepthMm,
		FitsBox:               resolution.FitsBox,
		BoxVerified:           boxProfile.Verified && sleeveProfile.Verified,
		DepthIsPlaceholder:    boxProfile.DepthIsPlaceholder,
		UnassembledComponents: resolution.UnassembledComponents,
		TrayCount:             len(resolution.ResolvedTrays),
	}

	if len(resolution.ResolvedTrays) == 0 {
		c.JSON(http.StatusOK, resp)
		return
	}

	firstTray := resolution.ResolvedTrays[0]
	previewURL, err := s.renderTrayPreview(ctx, firstTray.TrayTemplateID, firstTray.GeneratorModule, firstTray.Params)
	if err != nil {
		// The fit indicator is still useful even if the mesh render fails
		// (e.g. a transient OpenSCAD error) — don't fail the whole preview.
		log.Printf("[bgi] render tray preview: %v", err)
	} else {
		resp.FirstTrayPreviewURL = previewURL
	}

	c.JSON(http.StatusOK, resp)
}

// renderTrayPreview renders (or serves from cache) a single tray's preview
// mesh — same render→bbox→upload→cache shape as reef's configurePreview,
// looped by go/bgi-site to cover multiple trays across the whole slice
// (only the first is rendered live; configure/validate's async job renders
// every resolved tray at full detail).
func (s *server) renderTrayPreview(ctx context.Context, trayTemplateID uuid.UUID, generatorModule string, params map[string]interface{}) (string, error) {
	module, err := generate.Get(generatorModule)
	if err != nil {
		return "", err
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	renderCfg := s.renderConfig()
	openscadVersion, err := generate.Version(ctx, renderCfg)
	if err != nil {
		return "", err
	}

	hash, err := geomhash.Hash(generatorModule, module.Version(), openscadVersion, paramsJSON)
	if err != nil {
		return "", err
	}

	existing, err := s.deps.DbClient.BgiTraySliceResult().FindByGeometryHash(ctx, hash)
	if err != nil {
		return "", err
	}
	if existing != nil && existing.PreviewKey != "" {
		return s.previewURL(existing.PreviewKey), nil
	}

	scad, err := module.SCAD(params, generate.Preview)
	if err != nil {
		return "", err
	}
	renderResult, err := generate.Render(ctx, renderCfg, scad, openscadVersion)
	if err != nil {
		return "", fmt.Errorf("preview generation failed: %w", err)
	}
	defer procexec.Cleanup(renderResult.WorkDir)

	box, err := stlbbox.FromFile(renderResult.STLPath)
	if err != nil {
		return "", err
	}
	stlBytes, err := os.ReadFile(renderResult.STLPath)
	if err != nil {
		return "", err
	}
	previewKey := fmt.Sprintf("bgi/preview/%s.stl", hash)
	if _, err := s.deps.AwsClient.UploadImageToS3(s.deps.Config.Public.S3Bucket, previewKey, stlBytes); err != nil {
		return "", err
	}

	plateFits := box.MaxDimensionMm() <= s.deps.Config.Public.MaxBboxMm
	bboxJSON, _ := json.Marshal(map[string]float64{"xMm": box.XMm(), "yMm": box.YMm(), "zMm": box.ZMm()})

	if existing != nil {
		existing.PreviewKey = previewKey
		existing.BboxMm = datatypes.JSON(bboxJSON)
		existing.PlateFits = &plateFits
		if err := s.deps.DbClient.BgiTraySliceResult().Upsert(ctx, existing); err != nil {
			return "", err
		}
	} else {
		if err := s.deps.DbClient.BgiTraySliceResult().Create(ctx, &models.BgiTraySliceResult{
			GeometryHash:   hash,
			TrayTemplateID: trayTemplateID,
			Status:         models.BgiTraySliceStatusPending,
			PreviewKey:     previewKey,
			BboxMm:         datatypes.JSON(bboxJSON),
			PlateFits:      &plateFits,
			Warnings:       datatypes.JSON([]byte(`[]`)),
		}); err != nil {
			return "", err
		}
	}

	return s.previewURL(previewKey), nil
}

func (s *server) previewURL(key string) string {
	return "https://" + s.deps.Config.Public.S3Bucket + ".s3.amazonaws.com/" + key
}

type configureValidateResponse struct {
	ConfigurationID string `json:"configurationId"`
	Status          string `json:"status"`
}

// POST /api/bgi/configure/validate (R-8.1, R-5.1). The add-to-cart gate —
// creates a pending configuration and enqueues the full-set generation job
// (every resolved tray rendered/sliced/validated/priced); the client polls
// GET /configurations/:id until status leaves "pending".
func (s *server) configureValidate(c *gin.Context) {
	var req configureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, _, ok := s.resolveProduct(c, req)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	paramsJSON, err := json.Marshal(req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}

	cfg, err := s.deps.DbClient.BgiConfiguration().Create(ctx, &models.BgiConfiguration{
		ProductID: product.ID,
		Params:    datatypes.JSON(paramsJSON),
		Status:    models.BgiConfigurationStatusPending,
		SessionID: req.SessionID,
	})
	if err != nil {
		internalError(c, "create configuration", err)
		return
	}

	job, err := s.deps.DbClient.BgiGenerationJob().Create(ctx, &models.BgiGenerationJob{
		ConfigurationID: cfg.ID,
		Kind:            models.BgiGenerationJobKindFullSet,
		Status:          models.BgiGenerationJobStatusQueued,
	})
	if err != nil {
		internalError(c, "create generation job", err)
		return
	}

	payload, err := json.Marshal(jobs.GenerateBgiSetTaskPayload{ConfigurationID: cfg.ID, JobID: job.ID})
	if err != nil {
		internalError(c, "encode job payload", err)
		return
	}
	if err := s.deps.JobsClient.QueueJob(ctx, jobs.Job{Type: jobs.GenerateBgiSetTaskType, Payload: payload}); err != nil {
		internalError(c, "enqueue generation job", err)
		return
	}

	c.JSON(http.StatusAccepted, configureValidateResponse{
		ConfigurationID: cfg.ID.String(),
		Status:          cfg.Status,
	})
}

type resolvedTrayRecord struct {
	TrayTemplateID  uuid.UUID              `json:"trayTemplateId"`
	GeneratorModule string                 `json:"generatorModule"`
	Params          map[string]interface{} `json:"params"`
	Quantity        int                    `json:"quantity"`
	GeometryHash    string                 `json:"geometryHash"`
	HeightMm        float64                `json:"heightMm"`
}

type trayResponse struct {
	STLUrl   string  `json:"stlUrl,omitempty"`
	HeightMm float64 `json:"heightMm"`
	Quantity int     `json:"quantity"`
}

type configurationResponse struct {
	models.BgiConfiguration
	ProductSlug string         `json:"productSlug,omitempty"`
	Trays       []trayResponse `json:"trays,omitempty"`
}

// GET /api/bgi/configurations/:id (R-8.1) — what the client polls after
// configure/validate. Once resolved, Trays carries every tray in the set
// (not just one, unlike reef's single-part equivalent) so the frontend's
// StackedStlViewer can render the whole assembled set.
func (s *server) getConfiguration(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration id"})
		return
	}
	ctx := c.Request.Context()
	cfg, err := s.deps.DbClient.BgiConfiguration().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "configuration not found"})
		return
	}

	resp := configurationResponse{BgiConfiguration: *cfg}
	if product, err := s.deps.DbClient.BgiProduct().FindByID(ctx, cfg.ProductID); err == nil {
		resp.ProductSlug = product.Slug
	}

	if cfg.ConfigHash != nil {
		if resolution, err := s.deps.DbClient.BgiSetResolution().FindByConfigHash(ctx, *cfg.ConfigHash); err == nil && resolution != nil {
			var trays []resolvedTrayRecord
			if err := json.Unmarshal(resolution.ResolvedTrays, &trays); err == nil {
				for _, t := range trays {
					tr := trayResponse{HeightMm: t.HeightMm, Quantity: t.Quantity}
					if sliceResult, err := s.deps.DbClient.BgiTraySliceResult().FindByGeometryHash(ctx, t.GeometryHash); err == nil && sliceResult != nil {
						if sliceResult.STLKey != "" {
							tr.STLUrl = s.previewURL(sliceResult.STLKey)
						} else if sliceResult.PreviewKey != "" {
							tr.STLUrl = s.previewURL(sliceResult.PreviewKey)
						}
					}
					resp.Trays = append(resp.Trays, tr)
				}
			}
		}
	}

	c.JSON(http.StatusOK, resp)
}
