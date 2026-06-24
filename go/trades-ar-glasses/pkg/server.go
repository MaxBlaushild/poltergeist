package pkg

import (
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/trades-ar-glasses/internal/server"
	"github.com/gin-gonic/gin"
)

type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

func NewServer(dbClient db.DbClient) Server {
	return server.NewServer(dbClient)
}

func NewServerFromDependencies(dbClient db.DbClient) Server {
	return NewServer(dbClient)
}
