package quartermaster

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

type client struct {
	db db.DbClient
}

type Quartermaster interface {
	UseItem(ctx context.Context, ownedInventoryItemID uuid.UUID, metadata *UseItemMetadata) error
	GetItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID) (InventoryItem, error)
	FindItemForItemID(itemID int) (InventoryItem, error)
	GetInventoryItems() []InventoryItem
	ApplyInventoryItemEffects(ctx context.Context, userID uuid.UUID, match *models.Match) error
	GetItemSpecificItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, itemID int) (InventoryItem, error)
}

type UseItemMetadata struct {
	PointOfInterestID uuid.UUID `json:"pointOfInterestId"`
	TargetTeamID      uuid.UUID `json:"targetTeamId"`
	ChallengeID       uuid.UUID `json:"challengeId"`
}

func NewClient(db db.DbClient) Quartermaster {
	return &client{db: db}
}

func (c *client) GetInventoryItems() []InventoryItem {
	ctx := context.Background()
	dbItems, err := c.db.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		return []InventoryItem{}
	}

	items := make([]InventoryItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = InventoryItem{
			ID:            dbItem.ID,
			Name:          dbItem.Name,
			ImageURL:      dbItem.ImageURL,
			FlavorText:    dbItem.FlavorText,
			EffectText:    dbItem.EffectText,
			RarityTier:    Rarity(dbItem.RarityTier),
			IsCaptureType: dbItem.IsCaptureType,
		}
	}
	return items
}

func (c *client) FindItemForItemID(itemID int) (InventoryItem, error) {
	ctx := context.Background()
	dbItem, err := c.db.InventoryItem().FindInventoryItemByID(ctx, itemID)
	if err != nil {
		return InventoryItem{}, fmt.Errorf("item not found: %w", err)
	}

	return InventoryItem{
		ID:            dbItem.ID,
		Name:          dbItem.Name,
		ImageURL:      dbItem.ImageURL,
		FlavorText:    dbItem.FlavorText,
		EffectText:    dbItem.EffectText,
		RarityTier:    Rarity(dbItem.RarityTier),
		IsCaptureType: dbItem.IsCaptureType,
	}, nil
}

func (c *client) GetItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID) (InventoryItem, error) {
	item, err := c.getRandomItem()
	if err != nil {
		return InventoryItem{}, err
	}

	if err := c.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, teamID, userID, item.ID, 1); err != nil {
		return InventoryItem{}, err
	}

	return item, nil
}

func (c *client) GetItemSpecificItem(ctx context.Context, teamID *uuid.UUID, userID *uuid.UUID, itemID int) (InventoryItem, error) {
	item, err := c.FindItemForItemID(itemID)
	if err != nil {
		return InventoryItem{}, err
	}

	if err := c.db.InventoryItem().CreateOrIncrementInventoryItem(ctx, teamID, userID, item.ID, 1); err != nil {
		return InventoryItem{}, err
	}

	return item, nil
}

func (c *client) UseItem(ctx context.Context, ownedInventoryItemID uuid.UUID, metadata *UseItemMetadata) error {
	ownedInventoryItem, err := c.db.InventoryItem().FindByID(ctx, ownedInventoryItemID)
	if err != nil {
		return err
	}

	item, err := c.db.InventoryItem().FindInventoryItemByID(ctx, ownedInventoryItem.InventoryItemID)
	if err != nil {
		return err
	}
	hasConfiguredConsumeEffect := inventoryItemHasConfiguredConsumeEffects(item) && ownedInventoryItem.UserID != nil

	legacyEffectErr := c.ApplyItemEffectByID(ctx, *ownedInventoryItem, metadata)
	if legacyEffectErr != nil && !errors.Is(legacyEffectErr, ErrNoLegacyItemEffect) {
		return legacyEffectErr
	}

	if hasConfiguredConsumeEffect {
		if err := c.applyConfiguredConsumeEffects(ctx, *ownedInventoryItem, item); err != nil {
			return err
		}
	}

	hasLegacyEffect := legacyEffectErr == nil
	if !hasLegacyEffect && !hasConfiguredConsumeEffect && !item.IsCaptureType {
		return errors.New("item has no consumable effect")
	}

	if err := c.db.InventoryItem().UseInventoryItem(ctx, ownedInventoryItem.ID); err != nil {
		return err
	}

	return nil
}

func inventoryItemHasConfiguredConsumeEffects(item *models.InventoryItem) bool {
	if item == nil {
		return false
	}
	if item.ConsumeHealthDelta != 0 || item.ConsumeManaDelta != 0 {
		return true
	}
	return len(item.ConsumeStatusesToAdd) > 0 || len(item.ConsumeStatusesToRemove) > 0
}

func (c *client) applyConfiguredConsumeEffects(
	ctx context.Context,
	ownedInventoryItem models.OwnedInventoryItem,
	item *models.InventoryItem,
) error {
	if item == nil || ownedInventoryItem.UserID == nil {
		return nil
	}

	userID := *ownedInventoryItem.UserID
	if err := c.applyConfiguredConsumeResourceEffects(ctx, userID, item.ConsumeHealthDelta, item.ConsumeManaDelta); err != nil {
		return err
	}

	if len(item.ConsumeStatusesToRemove) > 0 {
		if err := c.db.UserStatus().DeleteActiveByUserIDAndNames(ctx, userID, []string(item.ConsumeStatusesToRemove)); err != nil {
			return err
		}
	}

	now := time.Now()
	for _, template := range item.ConsumeStatusesToAdd {
		name := strings.TrimSpace(template.Name)
		if name == "" || template.DurationSeconds <= 0 {
			continue
		}
		status := &models.UserStatus{
			UserID:          userID,
			Name:            name,
			Description:     strings.TrimSpace(template.Description),
			Effect:          strings.TrimSpace(template.Effect),
			Positive:        template.Positive,
			EffectType:      models.UserStatusEffectTypeStatModifier,
			StrengthMod:     template.StrengthMod,
			DexterityMod:    template.DexterityMod,
			ConstitutionMod: template.ConstitutionMod,
			IntelligenceMod: template.IntelligenceMod,
			WisdomMod:       template.WisdomMod,
			CharismaMod:     template.CharismaMod,
			StartedAt:       now,
			ExpiresAt:       now.Add(time.Duration(template.DurationSeconds) * time.Second),
		}
		if err := c.db.UserStatus().Create(ctx, status); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) applyConfiguredConsumeResourceEffects(
	ctx context.Context,
	userID uuid.UUID,
	healthDelta int,
	manaDelta int,
) error {
	if healthDelta == 0 && manaDelta == 0 {
		return nil
	}

	stats, err := c.db.UserCharacterStats().FindOrCreateForUser(ctx, userID)
	if err != nil {
		return err
	}
	equipmentBonuses, err := c.db.UserEquipment().GetStatBonuses(ctx, userID)
	if err != nil {
		return err
	}
	statusBonuses, err := c.db.UserStatus().GetActiveStatBonuses(ctx, userID)
	if err != nil {
		return err
	}

	totalBonuses := equipmentBonuses.Add(statusBonuses)
	maxHealth, maxMana, currentHealth, currentMana := deriveResourceState(stats, totalBonuses)
	healthDeficitDelta := inventoryResourceDeficitDelta(healthDelta, maxHealth, currentHealth, stats.HealthDeficit)
	manaDeficitDelta := inventoryResourceDeficitDelta(manaDelta, maxMana, currentMana, stats.ManaDeficit)
	if healthDeficitDelta == 0 && manaDeficitDelta == 0 {
		return nil
	}

	_, err = c.db.UserCharacterStats().AdjustResourceDeficits(ctx, userID, healthDeficitDelta, manaDeficitDelta)
	return err
}

func inventoryResourceDeficitDelta(
	delta int,
	maxValue int,
	currentValue int,
	deficit int,
) int {
	if delta == 0 || maxValue <= 0 {
		return 0
	}

	if delta > 0 {
		restore := minInt(delta, deficit)
		return -restore
	}

	drain := minInt(-delta, currentValue)
	return drain
}

func deriveResourceState(
	stats *models.UserCharacterStats,
	bonuses models.CharacterStatBonuses,
) (maxHealth int, maxMana int, currentHealth int, currentMana int) {
	effectiveConstitution := stats.Constitution + bonuses.Constitution
	effectiveIntelligence := stats.Intelligence + bonuses.Intelligence
	effectiveWisdom := stats.Wisdom + bonuses.Wisdom

	if effectiveConstitution < 1 {
		effectiveConstitution = 1
	}
	mental := effectiveIntelligence + effectiveWisdom
	if mental < 1 {
		mental = 1
	}

	maxHealth = effectiveConstitution * 10
	maxMana = mental * 5

	currentHealth = maxHealth - stats.HealthDeficit
	if currentHealth < 0 {
		currentHealth = 0
	}
	if currentHealth > maxHealth {
		currentHealth = maxHealth
	}

	currentMana = maxMana - stats.ManaDeficit
	if currentMana < 0 {
		currentMana = 0
	}
	if currentMana > maxMana {
		currentMana = maxMana
	}

	return maxHealth, maxMana, currentHealth, currentMana
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *client) getRandomItem() (InventoryItem, error) {
	rand.Seed(uint64(time.Now().UnixNano()))

	const (
		weightCommon       = 50
		weightUncommon     = 30
		weightEpic         = 15
		weightMythic       = 5
		weightNotDroppable = 0
	)

	rarityWeights := map[Rarity]int{
		RarityCommon:   weightCommon,
		RarityUncommon: weightUncommon,
		RarityEpic:     weightEpic,
		RarityMythic:   weightMythic,
		NotDroppable:   weightNotDroppable,
	}

	items := c.GetInventoryItems()
	totalWeight := 0
	for _, item := range items {
		totalWeight += rarityWeights[item.RarityTier]
	}

	for {
		randWeight := rand.Intn(totalWeight + 1)

		for _, item := range items {
			randWeight -= rarityWeights[item.RarityTier]
			if randWeight < 0 {
				return item, nil
			}
		}
	}
}
