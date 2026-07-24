package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/config"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	DbClient      db.DbClient
	Config        *config.Config
	AwsClient     aws.AWSClient
	JobsClient    jobs.Client
	EmailClient   email.EmailClient
	BillingClient billing.Client
}

type server struct {
	deps    Deps
	limiter *previewRateLimiter
}

func NewServer(deps Deps) *server {
	return &server{deps: deps, limiter: newPreviewRateLimiter()}
}

func (s *server) SetupRoutes(r *gin.Engine) {
	group := r.Group("/api/reef")
	group.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	group.GET("/products", s.listProducts)
	group.GET("/products/:slug", s.getProduct)
	group.GET("/products/:slug/schema", s.getProductSchema)
	group.GET("/tanks", s.listTanks)

	group.POST("/configure/preview", s.configurePreview)
	group.POST("/configure/validate", s.configureValidate)
	group.GET("/configurations/:id", s.getConfiguration)

	group.POST("/cart", s.postCart)
	group.POST("/checkout", s.postCheckout)
	group.POST("/webhooks/stripe", s.postStripeWebhook)
	group.GET("/orders/:token", s.getOrder)

	group.POST("/events", s.postEvent)
	group.GET("/operator/metrics", s.getOperatorMetrics)
}

// permissiveCORS mirrors go/core's own CORS config (gin-contrib/cors would
// pull in a gin-gonic/gin major bump requiring Go 1.25, ahead of this
// repo's go.work toolchain — not worth it for a dev-only convenience path,
// so this is the same handful of lines written directly).
func permissiveCORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}

// ListenAndServe is only used by go/reef-site's own standalone cmd/server
// (local dev / a future dedicated ECS service — see INVENTORY.md). When
// mounted into go/core, core's own top-level CORS config already covers
// these routes, so it isn't duplicated in SetupRoutes.
func (s *server) ListenAndServe(port string) {
	router := gin.Default()
	router.Use(permissiveCORS)
	s.SetupRoutes(router)
	router.Run(":" + port)
}
