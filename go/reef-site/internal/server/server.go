package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/config"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	DbClient    db.DbClient
	Config      *config.Config
	AwsClient   aws.AWSClient
	JobsClient  jobs.Client
	EmailClient email.EmailClient
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

func (s *server) ListenAndServe(port string) {
	router := gin.Default()
	s.SetupRoutes(router)
	router.Run(":" + port)
}
