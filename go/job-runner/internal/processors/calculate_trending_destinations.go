package processors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/hibiken/asynq"
)

type CalculateTrendingDestinationsProcessor struct {
	dbClient db.DbClient
}

func NewCalculateTrendingDestinationsProcessor(dbClient db.DbClient) CalculateTrendingDestinationsProcessor {
	log.Println("Initializing CalculateTrendingDestinationsProcessor")
	return CalculateTrendingDestinationsProcessor{
		dbClient: dbClient,
	}
}

func (p *CalculateTrendingDestinationsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing calculate trending destinations task: %v", task.Type())

	// Calculate the date 7 days ago
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	log.Printf("Calculating trending destinations for documents created after: %v", sevenDaysAgo)

	// Query to get top locations by document count in last 7 days
	// We'll do separate queries for cities and countries
	trendingHandler := p.dbClient.TrendingDestination()

	// Clear existing trending destinations
	log.Println("Clearing existing trending destinations")
	if err := trendingHandler.DeleteAll(ctx); err != nil {
		log.Printf("Failed to clear existing trending destinations: %v", err)
		return fmt.Errorf("failed to clear existing trending destinations: %w", err)
	}

	// Process cities
	if err := p.processLocationType(ctx, models.LocationTypeCity, sevenDaysAgo); err != nil {
		log.Printf("Failed to process cities: %v", err)
		return fmt.Errorf("failed to process cities: %w", err)
	}

	// Process countries
	if err := p.processLocationType(ctx, models.LocationTypeCountry, sevenDaysAgo); err != nil {
		log.Printf("Failed to process countries: %v", err)
		return fmt.Errorf("failed to process countries: %w", err)
	}

	log.Println("Completed processing calculate trending destinations task")
	return nil
}

func (p *CalculateTrendingDestinationsProcessor) processLocationType(ctx context.Context, locationType models.LocationType, since time.Time) error {
	log.Printf("Processing %s locations", locationType)

	trendingHandler := p.dbClient.TrendingDestination()

	// Get top 5 locations for this type
	results, err := trendingHandler.GetTopLocationsByType(ctx, locationType, since, 5)
	if err != nil {
		log.Printf("Failed to query %s locations: %v", locationType, err)
		return fmt.Errorf("failed to query %s locations: %w", locationType, err)
	}

	log.Printf("Found %d trending %s locations", len(results), locationType)

	// Insert results into trending_destinations table
	for rank, result := range results {
		destination := &models.TrendingDestination{
			LocationType:     models.LocationType(result.LocationType),
			PlaceID:          result.PlaceID,
			Name:             result.Name,
			FormattedAddress: result.FormattedAddress,
			DocumentCount:    result.DocumentCount,
			Rank:             rank + 1, // Rank is 1-indexed
			Latitude:         result.Latitude,
			Longitude:        result.Longitude,
		}

		if err := trendingHandler.Upsert(ctx, destination); err != nil {
			log.Printf("Failed to upsert trending destination %s (rank %d): %v", result.Name, rank+1, err)
			return fmt.Errorf("failed to upsert trending destination: %w", err)
		}
		log.Printf("Upserted trending %s: %s (rank %d, %d documents)", locationType, result.Name, rank+1, result.DocumentCount)
	}

	return nil
}
