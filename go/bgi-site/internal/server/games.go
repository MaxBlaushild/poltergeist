package server

import (
	"log"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type gameListItemResponse struct {
	models.BgiGame
	ProductSlug string `json:"productSlug,omitempty"`
}

// GET /api/bgi/games (R-8.1)
func (s *server) listGames(c *gin.Context) {
	ctx := c.Request.Context()
	games, err := s.deps.DbClient.BgiGame().FindActive(ctx)
	if err != nil {
		internalError(c, "list games", err)
		return
	}

	out := make([]gameListItemResponse, 0, len(games))
	for _, g := range games {
		item := gameListItemResponse{BgiGame: g}
		if product, err := s.deps.DbClient.BgiProduct().FindByGameID(ctx, g.ID); err == nil {
			item.ProductSlug = product.Slug
		}
		out = append(out, item)
	}

	c.JSON(http.StatusOK, out)
}

type gameDetailResponse struct {
	models.BgiGame
	ProductSlug    string                        `json:"productSlug,omitempty"`
	Expansions     []models.BgiExpansion         `json:"expansions"`
	BoxProfiles    []models.BgiBoxProfile        `json:"boxProfiles"`
	SleeveProfiles []models.BgiSleeveProfile     `json:"sleeveProfiles"`
	Manifest       []models.BgiComponentManifest `json:"manifest"`
}

// GET /api/bgi/games/:slug (R-8.1) — game + expansions + boxes + sleeves, per
// the requirements doc's own endpoint description. SleeveProfiles is the
// global list (not game-scoped, since a customer's sleeve collection isn't
// game-specific), included here so the configurator can load everything for
// a game page in one request.
func (s *server) getGame(c *gin.Context) {
	ctx := c.Request.Context()
	game, err := s.deps.DbClient.BgiGame().FindBySlug(ctx, c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	resp := gameDetailResponse{BgiGame: *game}
	if product, err := s.deps.DbClient.BgiProduct().FindByGameID(ctx, game.ID); err == nil {
		resp.ProductSlug = product.Slug
	}
	if expansions, err := s.deps.DbClient.BgiExpansion().FindByGameID(ctx, game.ID); err == nil {
		resp.Expansions = expansions
	}
	if boxes, err := s.deps.DbClient.BgiBoxProfile().FindByGameID(ctx, game.ID); err == nil {
		resp.BoxProfiles = boxes
	}
	if sleeves, err := s.deps.DbClient.BgiSleeveProfile().FindAll(ctx); err == nil {
		resp.SleeveProfiles = sleeves
	}
	if manifest, err := s.deps.DbClient.BgiComponentManifest().FindByGameID(ctx, game.ID, nil); err == nil {
		resp.Manifest = manifest
	}

	c.JSON(http.StatusOK, resp)
}

// GET /api/bgi/games/:slug/schema (R-8.1, R-4.4/R-5.2 — the direct analog of
// reef's getProductSchema).
func (s *server) getGameSchema(c *gin.Context) {
	ctx := c.Request.Context()
	game, err := s.deps.DbClient.BgiGame().FindBySlug(ctx, c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	product, err := s.deps.DbClient.BgiProduct().FindByGameID(ctx, game.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no product for this game"})
		return
	}
	schema, err := s.deps.DbClient.BgiParameterSchema().FindActiveByProductID(ctx, product.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active parameter schema for this game"})
		return
	}
	c.Data(http.StatusOK, "application/json", schema.Schema)
}

type compatibilityResponse struct {
	BoxProfiles []models.BgiBoxProfile        `json:"boxProfiles"`
	Manifest    []models.BgiComponentManifest `json:"manifest"`
}

// GET /api/bgi/games/:slug/compatibility (R-8.4/R-2.5) — the data the
// compatibility/disclaimer page needs to state which specific figures are
// still pending physical verification, per game.
func (s *server) getGameCompatibility(c *gin.Context) {
	ctx := c.Request.Context()
	game, err := s.deps.DbClient.BgiGame().FindBySlug(ctx, c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	resp := compatibilityResponse{}
	if boxes, err := s.deps.DbClient.BgiBoxProfile().FindByGameID(ctx, game.ID); err == nil {
		resp.BoxProfiles = boxes
	}
	if manifest, err := s.deps.DbClient.BgiComponentManifest().FindByGameID(ctx, game.ID, nil); err == nil {
		resp.Manifest = manifest
	}
	c.JSON(http.StatusOK, resp)
}

// GET /api/bgi/sleeve-profiles (R-8.1) — deliberately unfiltered by verified
// (see bgiSleeveProfileHandle.FindAll's own comment): the frontend must show
// these with an honest caveat, not hide them.
func (s *server) listSleeveProfiles(c *gin.Context) {
	profiles, err := s.deps.DbClient.BgiSleeveProfile().FindAll(c.Request.Context())
	if err != nil {
		internalError(c, "list sleeve profiles", err)
		return
	}
	c.JSON(http.StatusOK, profiles)
}

func internalError(c *gin.Context, action string, err error) {
	log.Printf("[bgi] %s: %v", action, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": action + " failed"})
}
