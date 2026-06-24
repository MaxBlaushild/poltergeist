package server

import (
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/gin-gonic/gin"
)

type server struct {
	leadStore db.TradesARGlassesLeadHandle
}

type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

func NewServer(dbClient db.DbClient) Server {
	return newServerFromLeadStore(dbClient.TradesARGlassesLead())
}

func newServerFromLeadStore(leadStore db.TradesARGlassesLeadHandle) Server {
	return &server{
		leadStore: leadStore,
	}
}

func (s *server) SetupRoutes(r *gin.Engine) {
	r.GET("/trades-ar-glasses/health", s.getHealth)
	r.POST("/trades-ar-glasses/interest", s.createInterestLead)

	// Compatibility route for the static landing page while it is served from
	// its own package root.
	r.POST("/api/interest", s.createInterestLead)
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()
	r.Use(devCORS)
	s.SetupRoutes(r)
	r.Run(fmt.Sprintf(":%s", port))
}

func devCORS(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(204)
		return
	}
	ctx.Next()
}
