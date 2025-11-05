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

	log.Println("Fetching all inventory items")
	inventoryItems, err := p.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		log.Printf("Failed to fetch inventory items: %v", err)
		return fmt.Errorf("failed to fetch inventory items: %w", err)
	}
	if len(inventoryItems) == 0 {
		log.Println("No inventory items found, skipping chest seeding")
		return nil
	}
	log.Printf("Found %d inventory items", len(inventoryItems))

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

			// Generate random gold between 100-500
			gold := 100 + rand.Intn(401) // 100 + [0-400] = 100-500

			// Assign unlock_tier: 30% chance of tier 1, 70% chance of nil
			var unlockTier *int
			if rand.Intn(100) < 30 {
				tier := 1
				unlockTier = &tier
			}

			// Create the treasure chest
			treasureChest := &models.TreasureChest{
				Latitude:    randomPoint.Y(), // Y() is latitude
				Longitude:   randomPoint.X(), // X() is longitude
				ZoneID:      zone.ID,
				Gold:        &gold,
				UnlockTier:  unlockTier,
				Invalidated: false,
			}

			if err := p.dbClient.TreasureChest().Create(ctx, treasureChest); err != nil {
				log.Printf("Failed to create treasure chest for zone %s: %v", zone.Name, err)
				continue
			}
			log.Printf("Created treasure chest %v with %d gold for zone %s", treasureChest.ID, gold, zone.Name)

			// Generate 1-3 random items with random quantities (1-3 each)
			numItems := 1 + rand.Intn(3) // 1 + [0-2] = 1-3 items
			usedItemIndices := make(map[int]bool)

			for j := 0; j < numItems; j++ {
				// Select a random inventory item that hasn't been used yet
				var selectedItem models.InventoryItem
				attempts := 0
				maxAttempts := 100

				for attempts < maxAttempts {
					itemIndex := rand.Intn(len(inventoryItems))
					if !usedItemIndices[itemIndex] {
						selectedItem = inventoryItems[itemIndex]
						usedItemIndices[itemIndex] = true
						break
					}
					attempts++
				}

				if attempts >= maxAttempts {
					log.Printf("Could not find unique item for chest %v after %d attempts, skipping remaining items", treasureChest.ID, maxAttempts)
					break
				}

				// Generate random quantity between 1-3
				quantity := 1 + rand.Intn(3) // 1 + [0-2] = 1-3

				if err := p.dbClient.TreasureChest().AddItem(ctx, treasureChest.ID, selectedItem.ID, quantity); err != nil {
					log.Printf("Failed to add item %d (quantity %d) to chest %v: %v", selectedItem.ID, quantity, treasureChest.ID, err)
					continue
				}
				log.Printf("Added item %s (ID: %d, quantity: %d) to chest %v", selectedItem.Name, selectedItem.ID, quantity, treasureChest.ID)
			}
		}
	}

	log.Println("Completed processing seed treasure chests task")
	return nil
}
