package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type characterStatsResponse struct {
	Strength         int                            `json:"strength"`
	Dexterity        int                            `json:"dexterity"`
	Constitution     int                            `json:"constitution"`
	Intelligence     int                            `json:"intelligence"`
	Wisdom           int                            `json:"wisdom"`
	Charisma         int                            `json:"charisma"`
	Health           int                            `json:"health"`
	MaxHealth        int                            `json:"maxHealth"`
	Mana             int                            `json:"mana"`
	MaxMana          int                            `json:"maxMana"`
	EquipmentBonuses map[string]int                 `json:"equipmentBonuses"`
	StatusBonuses    map[string]int                 `json:"statusBonuses"`
	UnspentPoints    int                            `json:"unspentPoints"`
	Level            int                            `json:"level"`
	Proficiencies    []characterProficiencyResponse `json:"proficiencies"`
	Statuses         []characterStatusResponse      `json:"statuses"`
	Spells           []characterSpellResponse       `json:"spells"`
}

type characterProficiencyResponse struct {
	Proficiency string `json:"proficiency"`
	Level       int    `json:"level"`
}

type characterStatusResponse struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Effect        string     `json:"effect"`
	Positive      bool       `json:"positive"`
	EffectType    string     `json:"effectType"`
	DamagePerTick int        `json:"damagePerTick"`
	HealthPerTick int        `json:"healthPerTick"`
	ManaPerTick   int        `json:"manaPerTick"`
	StartedAt     time.Time  `json:"startedAt"`
	LastTickAt    *time.Time `json:"lastTickAt,omitempty"`
	ExpiresAt     time.Time  `json:"expiresAt"`
}

type characterSpellResponse struct {
	ID                       uuid.UUID           `json:"id"`
	Name                     string              `json:"name"`
	Description              string              `json:"description"`
	IconURL                  string              `json:"iconUrl"`
	AbilityType              string              `json:"abilityType"`
	AbilityLevel             int                 `json:"abilityLevel"`
	CooldownTurns            int                 `json:"cooldownTurns"`
	CooldownTurnsRemaining   int                 `json:"cooldownTurnsRemaining"`
	CooldownSecondsRemaining int                 `json:"cooldownSecondsRemaining"`
	CooldownExpiresAt        *time.Time          `json:"cooldownExpiresAt,omitempty"`
	EffectText               string              `json:"effectText"`
	SchoolOfMagic            string              `json:"schoolOfMagic"`
	ManaCost                 int                 `json:"manaCost"`
	Effects                  models.SpellEffects `json:"effects"`
}

type characterStatsAllocationRequest struct {
	Allocations map[string]int `json:"allocations"`
}

type userCharacterProfileResponse struct {
	User      models.User                 `json:"user"`
	Stats     characterStatsResponse      `json:"stats"`
	UserLevel *models.UserLevel           `json:"userLevel"`
	Equipment []equipmentSlotResponse     `json:"equipment"`
	Inventory []userInventoryItemResponse `json:"inventory"`
}

type userInventoryItemResponse struct {
	OwnedInventoryItem models.OwnedInventoryItem `json:"ownedInventoryItem"`
	InventoryItem      *models.InventoryItem     `json:"inventoryItem,omitempty"`
	EquippedSlots      []string                  `json:"equippedSlots,omitempty"`
}

func (s *server) getCharacterStats(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := s.dbClient.UserCharacterStats().EnsureLevelPoints(ctx, user.ID, userLevel.Level)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	proficiencies, err := s.dbClient.UserProficiency().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusBonuses, statuses, err := s.getActiveStatusBonusesAndStatuses(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	spells, err := s.dbClient.UserSpell().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, characterStatsResponseFrom(stats, userLevel.Level, proficiencies, equipmentBonuses, statusBonuses, statuses, spells))
}

func (s *server) getUserCharacterProfile(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	idStr := ctx.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	target, err := s.dbClient.User().FindByID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	userLevel.ExperienceToNextLevel = userLevel.XPToNextLevel()

	stats, err := s.dbClient.UserCharacterStats().EnsureLevelPoints(ctx, userID, userLevel.Level)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	proficiencies, err := s.dbClient.UserProficiency().FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusBonuses, statuses, err := s.getActiveStatusBonusesAndStatuses(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	spells, err := s.dbClient.UserSpell().FindByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	equipment, err := s.buildEquipmentResponse(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ownedItems, err := s.dbClient.InventoryItem().GetItems(ctx, models.OwnedInventoryItem{UserID: &userID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	equippedSlotsByOwnedID := make(map[uuid.UUID][]string, len(equipment))
	for _, slot := range equipment {
		if slot.OwnedInventoryItemID == nil {
			continue
		}
		equippedSlotsByOwnedID[*slot.OwnedInventoryItemID] = append(
			equippedSlotsByOwnedID[*slot.OwnedInventoryItemID],
			slot.Slot,
		)
	}
	inventory := make([]userInventoryItemResponse, 0, len(ownedItems))
	for _, ownedItem := range ownedItems {
		if ownedItem.Quantity <= 0 {
			continue
		}
		item, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, ownedItem.InventoryItemID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		inventory = append(inventory, userInventoryItemResponse{
			OwnedInventoryItem: ownedItem,
			InventoryItem:      item,
			EquippedSlots:      equippedSlotsByOwnedID[ownedItem.ID],
		})
	}

	ctx.JSON(http.StatusOK, userCharacterProfileResponse{
		User:      *target,
		Stats:     characterStatsResponseFrom(stats, userLevel.Level, proficiencies, equipmentBonuses, statusBonuses, statuses, spells),
		UserLevel: userLevel,
		Equipment: equipment,
		Inventory: inventory,
	})
}

func (s *server) allocateCharacterStats(ctx *gin.Context) {
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var req characterStatsAllocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || len(req.Allocations) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "allocations required"})
		return
	}

	userLevel, err := s.dbClient.UserLevel().FindOrCreateForUser(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := s.dbClient.UserCharacterStats().ApplyAllocations(ctx, user.ID, userLevel.Level, req.Allocations)
	if err != nil {
		switch {
		case errors.Is(err, db.ErrNoStatAllocations),
			errors.Is(err, db.ErrInvalidStatAllocation),
			errors.Is(err, db.ErrInsufficientStatPoints):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	proficiencies, err := s.dbClient.UserProficiency().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	equipmentBonuses, err := s.dbClient.UserEquipment().GetStatBonuses(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	statusBonuses, statuses, err := s.getActiveStatusBonusesAndStatuses(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	spells, err := s.dbClient.UserSpell().FindByUserID(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, characterStatsResponseFrom(stats, userLevel.Level, proficiencies, equipmentBonuses, statusBonuses, statuses, spells))
}

func characterStatsResponseFrom(
	stats *models.UserCharacterStats,
	level int,
	proficiencies []models.UserProficiency,
	equipmentBonuses models.CharacterStatBonuses,
	statusBonuses models.CharacterStatBonuses,
	statuses []models.UserStatus,
	spells []models.UserSpell,
) characterStatsResponse {
	totalBonuses := equipmentBonuses.Add(statusBonuses)
	maxHealth, maxMana, currentHealth, currentMana := deriveCharacterResources(stats, totalBonuses)
	proficiencyResponse := make([]characterProficiencyResponse, 0, len(proficiencies))
	for _, proficiency := range proficiencies {
		proficiencyResponse = append(proficiencyResponse, characterProficiencyResponse{
			Proficiency: proficiency.Proficiency,
			Level:       proficiency.Level,
		})
	}
	return characterStatsResponse{
		Strength:         stats.Strength,
		Dexterity:        stats.Dexterity,
		Constitution:     stats.Constitution,
		Intelligence:     stats.Intelligence,
		Wisdom:           stats.Wisdom,
		Charisma:         stats.Charisma,
		Health:           currentHealth,
		MaxHealth:        maxHealth,
		Mana:             currentMana,
		MaxMana:          maxMana,
		EquipmentBonuses: equipmentBonuses.ToMap(),
		StatusBonuses:    statusBonuses.ToMap(),
		UnspentPoints:    stats.UnspentPoints,
		Level:            level,
		Proficiencies:    proficiencyResponse,
		Statuses:         characterStatusResponsesFrom(statuses),
		Spells:           characterSpellResponsesFrom(spells),
	}
}

func characterStatusResponsesFrom(statuses []models.UserStatus) []characterStatusResponse {
	response := make([]characterStatusResponse, 0, len(statuses))
	for _, status := range statuses {
		effectType := string(status.EffectType)
		if effectType == "" {
			effectType = string(models.UserStatusEffectTypeStatModifier)
		}
		response = append(response, characterStatusResponse{
			ID:            status.ID,
			Name:          status.Name,
			Description:   status.Description,
			Effect:        status.Effect,
			Positive:      status.Positive,
			EffectType:    effectType,
			DamagePerTick: status.DamagePerTick,
			HealthPerTick: status.HealthPerTick,
			ManaPerTick:   status.ManaPerTick,
			StartedAt:     status.StartedAt,
			LastTickAt:    status.LastTickAt,
			ExpiresAt:     status.ExpiresAt,
		})
	}
	return response
}

func characterSpellResponsesFrom(spells []models.UserSpell) []characterSpellResponse {
	response := make([]characterSpellResponse, 0, len(spells))
	now := time.Now()
	for _, userSpell := range spells {
		spell := userSpell.Spell
		if spell.ID == uuid.Nil {
			continue
		}
		response = append(response, characterSpellResponse{
			ID:                       spell.ID,
			Name:                     spell.Name,
			Description:              spell.Description,
			IconURL:                  spell.IconURL,
			AbilityType:              string(models.NormalizeSpellAbilityType(string(spell.AbilityType))),
			AbilityLevel:             spell.AbilityLevel,
			CooldownTurns:            spell.CooldownTurns,
			CooldownTurnsRemaining:   cooldownTurnsRemaining(userSpell, now),
			CooldownSecondsRemaining: cooldownSecondsRemaining(userSpell, now),
			CooldownExpiresAt:        userSpell.CooldownExpiresAt,
			EffectText:               spell.EffectText,
			SchoolOfMagic:            spell.SchoolOfMagic,
			ManaCost:                 spell.ManaCost,
			Effects:                  spell.Effects,
		})
	}
	return response
}

func (s *server) getActiveStatusBonusesAndStatuses(
	ctx context.Context,
	userID uuid.UUID,
) (models.CharacterStatBonuses, []models.UserStatus, error) {
	if err := s.applyOutOfBattleUserDamageOverTime(ctx, userID); err != nil {
		return models.CharacterStatBonuses{}, nil, err
	}
	statusBonuses, err := s.dbClient.UserStatus().GetActiveStatBonuses(ctx, userID)
	if err != nil {
		return models.CharacterStatBonuses{}, nil, err
	}
	statuses, err := s.dbClient.UserStatus().FindActiveByUserID(ctx, userID)
	if err != nil {
		return models.CharacterStatBonuses{}, nil, err
	}
	return statusBonuses, statuses, nil
}

func deriveCharacterHealth(constitution int) int {
	if constitution < 1 {
		constitution = 1
	}
	return constitution * 10
}

func deriveCharacterMana(intelligence int, wisdom int) int {
	mental := intelligence + wisdom
	if mental < 1 {
		mental = 1
	}
	return mental * 5
}

func deriveCharacterResources(
	stats *models.UserCharacterStats,
	bonuses models.CharacterStatBonuses,
) (maxHealth int, maxMana int, currentHealth int, currentMana int) {
	effectiveConstitution := stats.Constitution + bonuses.Constitution
	effectiveIntelligence := stats.Intelligence + bonuses.Intelligence
	effectiveWisdom := stats.Wisdom + bonuses.Wisdom
	maxHealth = deriveCharacterHealth(effectiveConstitution)
	maxMana = deriveCharacterMana(effectiveIntelligence, effectiveWisdom)

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
