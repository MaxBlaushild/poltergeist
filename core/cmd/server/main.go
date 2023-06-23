package main

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	fountUrl, _ := url.Parse("http://localhost:8081")

	fountProxy := httputil.NewSingleHostReverseProxy(fountUrl)

	router.POST("/consult", func(c *gin.Context) {
		fountProxy.ServeHTTP(c.Writer, c.Request)
	})

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Goodbye, World!")
	})

	router.Run(":8080")
}
