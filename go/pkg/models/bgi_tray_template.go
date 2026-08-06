package models

import (
	"time"

	"github.com/google/uuid"
)

// BgiTrayTemplate is one hand-designed tray (R-1.1's curated-parametric
// model). GeneratorModule matches a go/pkg/reef/generate.Module.Slug() the
// set-assembly service resolves at runtime.
type BgiTrayTemplate struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Slug             string    `json:"slug" gorm:"uniqueIndex"`
	GeneratorModule  string    `json:"generatorModule" gorm:"column:generator_module"`
	GeneratorVersion string    `json:"generatorVersion" gorm:"column:generator_version"`
	ComponentType    string    `json:"componentType" gorm:"column:component_type"`
	Description      string    `json:"description"`
	Active           bool      `json:"active"`
}

func (BgiTrayTemplate) TableName() string {
	return "bgi_tray_templates"
}
