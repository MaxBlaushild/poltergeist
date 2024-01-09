package main

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	fountUrl, _ := url.Parse("http://localhost:8081")
	trivaiUrl, _ := url.Parse("http://localhost:8082")
	texterUrl, _ := url.Parse("http://localhost:8084")
	scorekeeperUrl, _ := url.Parse("http://localhost:8086")
	authenticatorUrl, _ := url.Parse("http://localhost:8089")
	crystalCrisisUrl, _ := url.Parse("http://localhost:8091")
	adminDashboardUrl, _ := url.Parse("http://localhost:9093")
	billingUrl, _ := url.Parse("http://localhost:8022")
	sonarUrl, _ := url.Parse("http://localhost:8042")

	fountProxy := httputil.NewSingleHostReverseProxy(fountUrl)
	trivaiProxy := httputil.NewSingleHostReverseProxy(trivaiUrl)
	texterProxy := httputil.NewSingleHostReverseProxy(texterUrl)
	scorekeeperProxy := httputil.NewSingleHostReverseProxy(scorekeeperUrl)
	authenticatorProxy := httputil.NewSingleHostReverseProxy(authenticatorUrl)
	crystalCrisisProxy := httputil.NewSingleHostReverseProxy(crystalCrisisUrl)
	adminDashboardProxy := httputil.NewSingleHostReverseProxy(adminDashboardUrl)
	billingProxy := httputil.NewSingleHostReverseProxy(billingUrl)
	sonarProxy := httputil.NewSingleHostReverseProxy(sonarUrl)

	router.POST("/consult", func(c *gin.Context) {
		fountProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.Any("/trivai/*any", func(c *gin.Context) {
		trivaiProxy.ServeHTTP(c.Writer, c.Request)
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

	router.Any("/crystal-crisis/*any", func(c *gin.Context) {
		crystalCrisisProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.Any("/admin/*any", func(c *gin.Context) {
		adminDashboardProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.Any("/billing/*any", func(c *gin.Context) {
		billingProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.Any("/sonar/*any", func(c *gin.Context) {
		sonarProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Goodbye, World!")
	})

	router.Run(":8080")
}
