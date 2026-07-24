package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/generate"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/geomhash"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/procexec"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/stlbbox"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/paramschema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type configureRequest struct {
	ProductSlug string                 `json:"productSlug" binding:"required"`
	Params      map[string]interface{} `json:"params" binding:"required"`
	SessionID   string                 `json:"sessionId"`
}

type bboxResponse struct {
	XMm float64 `json:"xMm"`
	YMm float64 `json:"yMm"`
	ZMm float64 `json:"zMm"`
}

type previewResponse struct {
	GeometryHash string       `json:"geometryHash"`
	PreviewURL   string       `json:"previewUrl"`
	BboxMm       bboxResponse `json:"bboxMm"`
	PlateFits    bool         `json:"plateFits"`
	Cached       bool         `json:"cached"`
}

// resolveModule loads a configurable product, its active schema, and the
// matching generate.Module, validating params against the schema along the
// way (R-4.4/R-4.5). It's shared by the preview and validate handlers so
// they can never drift on what counts as valid input.
func (s *server) resolveModule(c *gin.Context, req configureRequest) (*models.ReefProduct, *models.ReefParameterSchema, generate.Module, bool) {
	ctx := c.Request.Context()

	product, err := s.deps.DbClient.ReefProduct().FindBySlug(ctx, req.ProductSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return nil, nil, nil, false
	}
	if product.Kind != models.ReefProductKindConfigurable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product is not configurable"})
		return nil, nil, nil, false
	}

	schema, err := s.deps.DbClient.ReefParameterSchema().FindActiveByProductID(ctx, product.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active parameter schema for this product"})
		return nil, nil, nil, false
	}

	parsedSchema, err := paramschema.Parse(schema.Schema)
	if err != nil {
		internalError(c, "parse parameter schema", err)
		return nil, nil, nil, false
	}
	if errs := paramschema.Validate(parsedSchema, req.Params); len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid parameters", "details": errs})
		return nil, nil, nil, false
	}

	module, err := generate.Get(schema.GeneratorModule)
	if err != nil {
		internalError(c, "resolve generator module", err)
		return nil, nil, nil, false
	}

	return product, schema, module, true
}

func (s *server) renderConfig() generate.RenderConfig {
	return generate.RenderConfig{
		OpenSCADBin: s.deps.Config.Public.OpenSCADBin,
		BaseTempDir: os.TempDir(),
		Timeout:     time.Duration(s.deps.Config.Public.PreviewTimeoutSec) * time.Second,
		MemoryMB:    s.deps.Config.Public.SubprocessMemoryMB,
	}
}

// POST /api/reef/configure/preview (R-8.1, R-2.6). Synchronous by design —
// R-2.10's explicit carve-out is that generation "must not block an HTTP
// request beyond the preview path." Rate-limited per session and cached by
// geometry_hash so dragging a slider doesn't re-render on every tick.
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

	product, _, module, ok := s.resolveModule(c, req)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	paramsJSON, err := json.Marshal(req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}

	renderCfg := s.renderConfig()
	openscadVersion, err := generate.Version(ctx, renderCfg)
	if err != nil {
		internalError(c, "resolve openscad version", err)
		return
	}

	hash, err := geomhash.Hash(product.Slug, module.Version(), openscadVersion, paramsJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := s.deps.DbClient.ReefSliceResult().FindByGeometryHash(ctx, hash)
	if err != nil {
		internalError(c, "check preview cache", err)
		return
	}
	if existing != nil && existing.PreviewKey != "" {
		c.JSON(http.StatusOK, previewResponse{
			GeometryHash: hash,
			PreviewURL:   s.previewURL(existing.PreviewKey),
			BboxMm:       decodeBbox(existing.BboxMm),
			PlateFits:    boolValue(existing.PlateFits),
			Cached:       true,
		})
		return
	}

	scad, err := module.SCAD(req.Params, generate.Preview)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	renderResult, err := generate.Render(ctx, renderCfg, scad, openscadVersion)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "preview generation failed: " + err.Error()})
		return
	}
	defer procexec.Cleanup(renderResult.WorkDir)

	box, err := stlbbox.FromFile(renderResult.STLPath)
	if err != nil {
		internalError(c, "compute preview bounding box", err)
		return
	}

	stlBytes, err := os.ReadFile(renderResult.STLPath)
	if err != nil {
		internalError(c, "read preview stl", err)
		return
	}
	previewKey := fmt.Sprintf("reef/preview/%s.stl", hash)
	if _, err := s.deps.AwsClient.UploadImageToS3(s.deps.Config.Public.S3Bucket, previewKey, stlBytes); err != nil {
		internalError(c, "upload preview stl", err)
		return
	}

	plateFits := box.MaxDimensionMm() <= s.deps.Config.Public.MaxBboxMm
	bboxJSON, _ := json.Marshal(map[string]float64{"xMm": box.XMm(), "yMm": box.YMm(), "zMm": box.ZMm()})

	if err := s.upsertPreviewSliceResult(ctx, existing, hash, product.ID, previewKey, datatypes.JSON(bboxJSON), plateFits); err != nil {
		internalError(c, "persist preview cache entry", err)
		return
	}

	c.JSON(http.StatusOK, previewResponse{
		GeometryHash: hash,
		PreviewURL:   s.previewURL(previewKey),
		BboxMm:       decodeBbox(bboxJSON),
		PlateFits:    plateFits,
		Cached:       false,
	})
}

func (s *server) upsertPreviewSliceResult(ctx context.Context, existing *models.ReefSliceResult, hash string, productID uuid.UUID, previewKey string, bboxMm datatypes.JSON, plateFits bool) error {
	if existing != nil {
		existing.PreviewKey = previewKey
		existing.BboxMm = bboxMm
		existing.PlateFits = &plateFits
		return s.deps.DbClient.ReefSliceResult().Update(ctx, existing)
	}
	return s.deps.DbClient.ReefSliceResult().Create(ctx, &models.ReefSliceResult{
		GeometryHash: hash,
		ProductID:    productID,
		Status:       models.ReefSliceStatusPending,
		PreviewKey:   previewKey,
		BboxMm:       bboxMm,
		PlateFits:    &plateFits,
		Warnings:     datatypes.JSON([]byte(`[]`)),
	})
}

func (s *server) previewURL(key string) string {
	return "https://" + s.deps.Config.Public.S3Bucket + ".s3.amazonaws.com/" + key
}

type configureValidateResponse struct {
	ConfigurationID string `json:"configurationId"`
	Status          string `json:"status"`
}

// POST /api/reef/configure/validate (R-8.1, R-5.1). This is the
// add-to-cart gate: nothing enters a cart without a passing server-side
// slice. Generation is async (R-2.10) — the client polls
// GET /configurations/:id until status leaves "pending".
func (s *server) configureValidate(c *gin.Context) {
	var req configureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, _, _, ok := s.resolveModule(c, req)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	paramsJSON, err := json.Marshal(req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}

	cfg, err := s.deps.DbClient.ReefConfiguration().Create(ctx, &models.ReefConfiguration{
		ProductID: product.ID,
		Params:    datatypes.JSON(paramsJSON),
		Status:    models.ReefConfigurationStatusPending,
		SessionID: req.SessionID,
	})
	if err != nil {
		internalError(c, "create configuration", err)
		return
	}

	job, err := s.deps.DbClient.ReefGenerationJob().Create(ctx, &models.ReefGenerationJob{
		ConfigurationID: cfg.ID,
		Kind:            models.ReefGenerationJobKindFull,
		Status:          models.ReefGenerationJobStatusQueued,
	})
	if err != nil {
		internalError(c, "create generation job", err)
		return
	}

	payload, err := json.Marshal(jobs.GenerateReefFullTaskPayload{ConfigurationID: cfg.ID, JobID: job.ID})
	if err != nil {
		internalError(c, "encode job payload", err)
		return
	}
	if err := s.deps.JobsClient.QueueJob(ctx, jobs.Job{Type: jobs.GenerateReefFullTaskType, Payload: payload}); err != nil {
		internalError(c, "enqueue generation job", err)
		return
	}

	c.JSON(http.StatusAccepted, configureValidateResponse{
		ConfigurationID: cfg.ID.String(),
		Status:          cfg.Status,
	})
}

// GET /api/reef/configurations/:id (R-8.1) — what the client polls after
// configure/validate.
func (s *server) getConfiguration(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration id"})
		return
	}
	cfg, err := s.deps.DbClient.ReefConfiguration().FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "configuration not found"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func decodeBbox(raw datatypes.JSON) bboxResponse {
	var b bboxResponse
	_ = json.Unmarshal(raw, &b)
	return b
}

func boolValue(b *bool) bool {
	return b != nil && *b
}
