package processors

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/hibiken/asynq"
)

type SeedTreasureChestsProcessor struct {
	dbClient db.DbClient
}

func NewSeedTreasureChestsProcessor(dbClient db.DbClient) SeedTreasureChestsProcessor {
	log.Println("Initializing SeedTreasureChestsProcessor")
	return SeedTreasureChestsProcessor{
		dbClient: dbClient,
	}
}

func (p *SeedTreasureChestsProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	log.Printf("Processing seed treasure chests task: %v", task.Type())

	log.Println("Fetching all zones")
	zones, err := p.dbClient.Zone().FindAll(ctx)
	if err != nil {
		log.Printf("Failed to fetch zones: %v", err)
		return fmt.Errorf("failed to fetch zones: %w", err)
	}
	log.Printf("Found %d zones", len(zones))

	for _, zone := range zones {
		log.Printf("Processing zone: %s (ID: %v)", zone.Name, zone.ID)

		// Invalidate existing chests in this zone
		if err := p.dbClient.TreasureChest().InvalidateByZoneID(ctx, zone.ID); err != nil {
			log.Printf("Failed to invalidate existing chests for zone %s: %v", zone.Name, err)
			continue
		}
		log.Printf("Invalidated existing chests for zone %s", zone.Name)

		// Create 3 new chests for this zone
		for i := 0; i < 3; i++ {
			log.Printf("Creating chest %d/3 for zone %s", i+1, zone.Name)

			// Get random location within zone boundary
			randomPoint := zone.GetRandomPoint()
			if randomPoint.X() == 0 && randomPoint.Y() == 0 {
				log.Printf("Failed to get random point for zone %s, skipping chest", zone.Name)
				continue
			}

			// Assign unlock_tier: 30% chance of tier 1, 70% chance of nil
			var unlockTier *int
			if rand.Intn(100) < 30 {
				tier := 1
				unlockTier = &tier
			}
			size := models.RandomRewardSizeSmall
			switch roll := rand.Intn(100); {
			case roll < 65:
				size = models.RandomRewardSizeSmall
			case roll < 92:
				size = models.RandomRewardSizeMedium
			default:
				size = models.RandomRewardSizeLarge
			}

			// Create the treasure chest
			treasureChest := &models.TreasureChest{
				Latitude:         randomPoint.Y(), // Y() is latitude
				Longitude:        randomPoint.X(), // X() is longitude
				ZoneID:           zone.ID,
				RewardMode:       models.RewardModeRandom,
				RandomRewardSize: size,
				RewardExperience: 0,
				Gold:             nil,
				UnlockTier:       unlockTier,
				Invalidated:      false,
			}

			if err := p.dbClient.TreasureChest().Create(ctx, treasureChest); err != nil {
				log.Printf("Failed to create treasure chest for zone %s: %v", zone.Name, err)
				continue
			}
			log.Printf("Created treasure chest %v with random %s rewards for zone %s", treasureChest.ID, size, zone.Name)
		}
	}

	log.Println("Completed processing seed treasure chests task")
	return nil
}
