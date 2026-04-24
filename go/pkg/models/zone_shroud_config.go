package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type ZoneShroudConfig struct {
	ID                          int       `gorm:"primaryKey" json:"id"`
	PatternTileURL              string    `json:"patternTileUrl" gorm:"column:pattern_tile_url"`
	PatternTilePrompt           string    `json:"patternTilePrompt" gorm:"column:pattern_tile_prompt"`
	PatternTileGenerationStatus string    `json:"patternTileGenerationStatus" gorm:"column:pattern_tile_generation_status"`
	PatternTileGenerationError  string    `json:"patternTileGenerationError" gorm:"column:pattern_tile_generation_error"`
	CreatedAt                   time.Time `json:"createdAt"`
	UpdatedAt                   time.Time `json:"updatedAt"`
}

func (ZoneShroudConfig) TableName() string {
	return "zone_shroud_configs"
}

func (c *ZoneShroudConfig) BeforeSave(tx *gorm.DB) error {
	c.PatternTileURL = strings.TrimSpace(c.PatternTileURL)
	c.PatternTilePrompt = strings.TrimSpace(c.PatternTilePrompt)
	c.PatternTileGenerationError = strings.TrimSpace(c.PatternTileGenerationError)
	c.PatternTileGenerationStatus = normalizeZoneKindPatternTileGenerationStatus(c.PatternTileGenerationStatus)
	return nil
}
