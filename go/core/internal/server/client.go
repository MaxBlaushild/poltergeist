package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	finalfete "github.com/MaxBlaushild/poltergeist/final-fete/pkg"
	"github.com/MaxBlaushild/poltergeist/pkg/texter"
	sonar "github.com/MaxBlaushild/poltergeist/sonar/pkg"
	travelangels "github.com/MaxBlaushild/poltergeist/travel-angels/pkg"
	verifiablesn "github.com/MaxBlaushild/poltergeist/verifiable-sn/pkg"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type server struct {
	router             *gin.Engine
	finalFeteServer    finalfete.Server
	sonarServer        sonar.Server
	travelAngelsServer travelangels.Server
	verifiableSnServer verifiablesn.Server
	texterClient       texter.Client
}

// NewServer creates a new server instance
func NewServer(finalFeteServer finalfete.Server, sonarServer sonar.Server, travelAngelsServer travelangels.Server, verifiableSnServer verifiablesn.Server, texterClient texter.Client) *server {
	return &server{
		finalFeteServer:    finalFeteServer,
		sonarServer:        sonarServer,
		travelAngelsServer: travelAngelsServer,
		verifiableSnServer: verifiableSnServer,
		texterClient:       texterClient,
	}
}

func (s *server) ListenAndServe(port string) {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept", "Authorization", "X-User-Location"},
	}))

	fountUrl, _ := url.Parse("http://localhost:8081")
	texterUrl, _ := url.Parse("http://localhost:8084")
	scorekeeperUrl, _ := url.Parse("http://localhost:8086")
	authenticatorUrl, _ := url.Parse("http://localhost:8089")
	adminDashboardUrl, _ := url.Parse("http://localhost:9093")
	billingUrl, _ := url.Parse("http://localhost:8022")

	fountProxy := httputil.NewSingleHostReverseProxy(fountUrl)
	texterProxy := httputil.NewSingleHostReverseProxy(texterUrl)
	scorekeeperProxy := httputil.NewSingleHostReverseProxy(scorekeeperUrl)
	authenticatorProxy := httputil.NewSingleHostReverseProxy(authenticatorUrl)
	adminDashboardProxy := httputil.NewSingleHostReverseProxy(adminDashboardUrl)
	billingProxy := httputil.NewSingleHostReverseProxy(billingUrl)

	router.POST("/consult", func(c *gin.Context) {
		fountProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.Any("/texter/*any", func(c *gin.Context) {
		texterProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.Any("/scorekeeper/*any", func(c *gin.Context) {
		scorekeeperProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.Any("/authenticator/*any", func(c *gin.Context) {
		authenticatorProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.Any("/admin/*any", func(c *gin.Context) {
		adminDashboardProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.Any("/billing/*any", func(c *gin.Context) {
		billingProxy.ServeHTTP(c.Writer, c.Request)
	})

	s.finalFeteServer.SetupRoutes(router)
	s.sonarServer.SetupRoutes(router)
	s.travelAngelsServer.SetupRoutes(router)
	s.verifiableSnServer.SetupRoutes(router)

	// Champagne endpoint - sends celebratory text
	router.POST("/champagne", func(c *gin.Context) {
		// Get phone number from environment or use default
		fromPhoneNumber := os.Getenv("PHONE_NUMBER")
		if fromPhoneNumber == "" {
			fromPhoneNumber = "+18445206851" // Default fallback
		}

		// Send champagne text
		err := s.texterClient.Text(c.Request.Context(), &texter.Text{
			Body:     "Time for champagne! 🍾",
			To:       "+12154354713",
			From:     fromPhoneNumber,
			TextType: "champagne",
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to send text",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Champagne text sent successfully",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Goodbye, World!")
	})

	router.Run(":8080")
}
