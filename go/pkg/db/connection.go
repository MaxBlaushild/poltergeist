package db

import (
	"fmt"

	"gorm.io/gorm"
)

// GormConnection lets a service keep its domain-specific repositories beside
// its business rules while sharing the application's configured connection pool.
// It deliberately leaves DbClient unchanged, including existing test doubles.
func GormConnection(database DbClient) (*gorm.DB, error) {
	c, ok := database.(*client)
	if !ok || c == nil || c.db == nil {
		return nil, fmt.Errorf("database client does not expose a SQL connection")
	}
	return c.db, nil
}
