package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ReefParameterSchema is the single source of parameter truth (R-4.4): Go
// validates configurator input against Schema server-side, and the TS client
// renders the configurator form from the same document fetched at runtime.
type ReefParameterSchema struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	ProductID        uuid.UUID      `json:"productId" gorm:"type:uuid;index"`
	Version          int            `json:"version"`
	Schema           datatypes.JSON `json:"schema"`
	GeneratorModule  string         `json:"generatorModule" gorm:"column:generator_module"`
	GeneratorVersion string         `json:"generatorVersion" gorm:"column:generator_version"`
	Active           bool           `json:"active"`
}

func (ReefParameterSchema) TableName() string {
	return "reef_parameter_schemas"
}
