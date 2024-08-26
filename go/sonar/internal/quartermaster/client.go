package quartermaster

import (
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type client struct {
}

type Quartermaster interface {
	GetInventoryItems(c *gin.Context, teamID uuid.UUID) ([]models.InventoryItem, error)
}
