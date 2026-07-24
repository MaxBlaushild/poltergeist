package server

import (
	"log"
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type productResponse struct {
	models.ReefProduct
	Variants []models.ReefProductVariant `json:"variants,omitempty"`
}

// GET /api/reef/products (R-8.1)
func (s *server) listProducts(c *gin.Context) {
	products, err := s.deps.DbClient.ReefProduct().FindActive(c.Request.Context())
	if err != nil {
		internalError(c, "list products", err)
		return
	}

	out := make([]productResponse, 0, len(products))
	for _, p := range products {
		resp := productResponse{ReefProduct: p}
		if p.Kind == models.ReefProductKindFixed {
			variants, err := s.deps.DbClient.ReefProductVariant().FindByProductID(c.Request.Context(), p.ID)
			if err != nil {
				internalError(c, "load variants", err)
				return
			}
			resp.Variants = variants
		}
		out = append(out, resp)
	}

	c.JSON(http.StatusOK, out)
}

// GET /api/reef/products/:slug (R-8.1)
func (s *server) getProduct(c *gin.Context) {
	product, err := s.deps.DbClient.ReefProduct().FindBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	resp := productResponse{ReefProduct: *product}
	if product.Kind == models.ReefProductKindFixed {
		variants, err := s.deps.DbClient.ReefProductVariant().FindByProductID(c.Request.Context(), product.ID)
		if err != nil {
			internalError(c, "load variants", err)
			return
		}
		resp.Variants = variants
	}

	c.JSON(http.StatusOK, resp)
}

// GET /api/reef/products/:slug/schema (R-8.1, R-4.4: the single source of
// parameter truth — the TS client renders its configurator form from
// exactly this document, fetched at runtime, with no hand-written form
// fields on the frontend.
func (s *server) getProductSchema(c *gin.Context) {
	product, err := s.deps.DbClient.ReefProduct().FindBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if product.Kind != models.ReefProductKindConfigurable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product is not configurable"})
		return
	}

	schema, err := s.deps.DbClient.ReefParameterSchema().FindActiveByProductID(c.Request.Context(), product.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active parameter schema for this product"})
		return
	}

	c.Data(http.StatusOK, "application/json", schema.Schema)
}

func internalError(c *gin.Context, action string, err error) {
	log.Printf("[reef] %s: %v", action, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": action + " failed"})
}
