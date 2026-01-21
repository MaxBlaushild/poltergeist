package server

import (
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/ethereum"
	"github.com/gin-gonic/gin"
)

type server struct {
	dbClient       db.DbClient
	ethereumClient ethereum.EthereumClient
	chainID        int64
}

type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

func NewServer(
	dbClient db.DbClient,
	ethereumClient ethereum.EthereumClient,
	chainID int64,
) Server {
	return &server{
		dbClient:       dbClient,
		ethereumClient: ethereumClient,
		chainID:        chainID,
	}
}

func (s *server) SetupRoutes(r *gin.Engine) {
	r.GET("/ethereum-transactor/health", s.GetHealth)

	// Transaction routes
	r.POST("/ethereum-transactor/transactions", s.CreateTransaction)
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()
	s.SetupRoutes(r)
	r.Run(fmt.Sprintf(":%s", port))
}
