package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	BgiProductKindConfigurable = "configurable"
	BgiProductKindFixed        = "fixed"
)

type BgiProduct struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	GameID         uuid.UUID      `json:"gameId" gorm:"type:uuid;index"`
	Slug           string         `json:"slug" gorm:"uniqueIndex"`
	Name           string         `json:"name"`
	Kind           string         `json:"kind"`
	Description    string         `json:"description"`
	Material       string         `json:"material"`
	BasePriceCents int64          `json:"basePriceCents" gorm:"column:base_price_cents"`
	Images         datatypes.JSON `json:"images"`
	Active         bool           `json:"active"`
}

func (BgiProduct) TableName() string {
	return "bgi_products"
}

// BgiParameterSchema mirrors ReefParameterSchema exactly (R-4.4's single
// source of parameter truth), with one structural difference: GeneratorModule
// on a bgi product's schema is the sentinel "bgi_tray_set" rather than a real
// generate.Module slug — go/bgi-site's configure handlers recognize this and
// route through go/pkg/reef/set.Assemble instead of generate.Get(), since a
// tray set is many generated parts, not one (R-3.3).
type BgiParameterSchema struct {
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

func (BgiParameterSchema) TableName() string {
	return "bgi_parameter_schemas"
}
