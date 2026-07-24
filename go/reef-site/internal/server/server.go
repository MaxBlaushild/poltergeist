package server

import (
	"net/http"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/reef-site/internal/config"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	DbClient db.DbClient
	Config   *config.Config
}

type server struct {
	deps Deps
}

func NewServer(deps Deps) *server {
	return &server{deps: deps}
}

func (s *server) SetupRoutes(r *gin.Engine) {
	group := r.Group("/api/reef")
	group.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (s *server) ListenAndServe(port string) {
	router := gin.Default()
	s.SetupRoutes(router)
	router.Run(":" + port)
}
