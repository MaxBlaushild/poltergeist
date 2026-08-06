package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/bgi-site/internal/config"
	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/aws"
	"github.com/MaxBlaushild/poltergeist/pkg/billing"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/email"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	DbClient      db.DbClient
	Config        *config.Config
	AwsClient     aws.AWSClient
	JobsClient    jobs.Client
	EmailClient   email.EmailClient
	BillingClient billing.Client
	AuthClient    auth.Client
}

type server struct {
	deps    Deps
	limiter *previewRateLimiter
}

func NewServer(deps Deps) *server {
	return &server{deps: deps, limiter: newPreviewRateLimiter()}
}

// SetupRoutes mirrors go/reef-site's route shape at /api/bgi instead of
// /api/reef (R-8.1). No /me/orders or account routes — v1 is guest-only
// (R-1.3 leaves accounts out of scope), so there's no auth.go/user session
// surface to mirror here.
func (s *server) SetupRoutes(r *gin.Engine) {
	group := r.Group("/api/bgi")
	group.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	group.GET("/games", s.listGames)
	group.GET("/games/:slug", s.getGame)
	group.GET("/games/:slug/schema", s.getGameSchema)
	group.GET("/games/:slug/compatibility", s.getGameCompatibility)
	group.GET("/sleeve-profiles", s.listSleeveProfiles)

	group.POST("/configure/preview", s.configurePreview)
	group.POST("/configure/validate", s.configureValidate)
	group.GET("/configurations/:id", s.getConfiguration)

	group.POST("/cart", s.postCart)
	group.POST("/checkout", s.postCheckout)
	group.POST("/webhooks/stripe", s.postStripeWebhook)
	group.GET("/orders/:token", s.getOrder)

	group.POST("/events", s.postEvent)
	group.POST("/waitlist", s.postWaitlist)

	// Same reasoning as reef-site's operator group: real customer data, so
	// gated behind a shared password rather than an unlisted-URL page.
	operatorGroup := group.Group("/operator", gin.BasicAuth(gin.Accounts{
		"operator": s.deps.Config.Secret.AdminToken,
	}))
	operatorGroup.PATCH("/orders/:id/fulfillment", s.updateOrderFulfillment)
}

// permissiveCORS mirrors go/reef-site's own — see that file's comment for
// why this is written directly rather than pulling in gin-contrib/cors.
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

// ListenAndServe is only used by bgi-site's own standalone cmd/server (local
// dev). When mounted into go/core, core's own top-level CORS config already
// covers these routes.
func (s *server) ListenAndServe(port string) {
	router := gin.Default()
	router.Use(permissiveCORS)
	s.SetupRoutes(router)
	router.Run(":" + port)
}
