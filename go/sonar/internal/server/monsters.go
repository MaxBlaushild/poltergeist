package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

type monsterTemplateUpsertRequest struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	ImageURL         string   `json:"imageUrl"`
	ThumbnailURL     string   `json:"thumbnailUrl"`
	BaseStrength     int      `json:"baseStrength"`
	BaseDexterity    int      `json:"baseDexterity"`
	BaseConstitution int      `json:"baseConstitution"`
	BaseIntelligence int      `json:"baseIntelligence"`
	BaseWisdom       int      `json:"baseWisdom"`
	BaseCharisma     int      `json:"baseCharisma"`
	SpellIDs         []string `json:"spellIds"`
}

type monsterRewardItemPayload struct {
	InventoryItemID int `json:"inventoryItemId"`
	Quantity        int `json:"quantity"`
}

type monsterUpsertRequest struct {
	Name                        string                     `json:"name"`
	Description                 string                     `json:"description"`
	ImageURL                    string                     `json:"imageUrl"`
	ThumbnailURL                string                     `json:"thumbnailUrl"`
	ZoneID                      string                     `json:"zoneId"`
	Latitude                    float64                    `json:"latitude"`
	Longitude                   float64                    `json:"longitude"`
	TemplateID                  string                     `json:"templateId"`
	DominantHandInventoryItemID *int                       `json:"dominantHandInventoryItemId"`
	OffHandInventoryItemID      *int                       `json:"offHandInventoryItemId"`
	WeaponInventoryItemID       *int                       `json:"weaponInventoryItemId"`
	Level                       int                        `json:"level"`
	RewardExperience            int                        `json:"rewardExperience"`
	RewardGold                  int                        `json:"rewardGold"`
	ItemRewards                 []monsterRewardItemPayload `json:"itemRewards"`
}

type monsterTemplateResponse struct {
	ID                    uuid.UUID      `json:"id"`
	CreatedAt             time.Time      `json:"createdAt"`
	UpdatedAt             time.Time      `json:"updatedAt"`
	Name                  string         `json:"name"`
	Description           string         `json:"description"`
	ImageURL              string         `json:"imageUrl"`
	ThumbnailURL          string         `json:"thumbnailUrl"`
	BaseStrength          int            `json:"baseStrength"`
	BaseDexterity         int            `json:"baseDexterity"`
	BaseConstitution      int            `json:"baseConstitution"`
	BaseIntelligence      int            `json:"baseIntelligence"`
	BaseWisdom            int            `json:"baseWisdom"`
	BaseCharisma          int            `json:"baseCharisma"`
	Spells                []models.Spell `json:"spells"`
	ImageGenerationStatus string         `json:"imageGenerationStatus"`
	ImageGenerationError  *string        `json:"imageGenerationError,omitempty"`
}

type monsterResponse struct {
	ID                          uuid.UUID                  `json:"id"`
	CreatedAt                   time.Time                  `json:"createdAt"`
	UpdatedAt                   time.Time                  `json:"updatedAt"`
	Name                        string                     `json:"name"`
	Description                 string                     `json:"description"`
	ImageURL                    string                     `json:"imageUrl"`
	ThumbnailURL                string                     `json:"thumbnailUrl"`
	ZoneID                      uuid.UUID                  `json:"zoneId"`
	Zone                        models.Zone                `json:"zone"`
	Latitude                    float64                    `json:"latitude"`
	Longitude                   float64                    `json:"longitude"`
	TemplateID                  *uuid.UUID                 `json:"templateId,omitempty"`
	Template                    *monsterTemplateResponse   `json:"template,omitempty"`
	DominantHandInventoryItemID *int                       `json:"dominantHandInventoryItemId,omitempty"`
	DominantHandInventoryItem   *models.InventoryItem      `json:"dominantHandInventoryItem,omitempty"`
	OffHandInventoryItemID      *int                       `json:"offHandInventoryItemId,omitempty"`
	OffHandInventoryItem        *models.InventoryItem      `json:"offHandInventoryItem,omitempty"`
	WeaponInventoryItemID       *int                       `json:"weaponInventoryItemId,omitempty"`
	WeaponInventoryItem         *models.InventoryItem      `json:"weaponInventoryItem,omitempty"`
	Level                       int                        `json:"level"`
	Strength                    int                        `json:"strength"`
	Dexterity                   int                        `json:"dexterity"`
	Constitution                int                        `json:"constitution"`
	Intelligence                int                        `json:"intelligence"`
	Wisdom                      int                        `json:"wisdom"`
	Charisma                    int                        `json:"charisma"`
	Health                      int                        `json:"health"`
	MaxHealth                   int                        `json:"maxHealth"`
	Mana                        int                        `json:"mana"`
	MaxMana                     int                        `json:"maxMana"`
	AttackDamageMin             int                        `json:"attackDamageMin"`
	AttackDamageMax             int                        `json:"attackDamageMax"`
	AttackSwipesPerAttack       int                        `json:"attackSwipesPerAttack"`
	Spells                      []models.Spell             `json:"spells"`
	RewardExperience            int                        `json:"rewardExperience"`
	RewardGold                  int                        `json:"rewardGold"`
	ItemRewards                 []models.MonsterItemReward `json:"itemRewards"`
	ImageGenerationStatus       string                     `json:"imageGenerationStatus"`
	ImageGenerationError        *string                    `json:"imageGenerationError,omitempty"`
}

func monsterTemplateResponseFrom(template *models.MonsterTemplate) *monsterTemplateResponse {
	if template == nil {
		return nil
	}
	spells := make([]models.Spell, 0, len(template.Spells))
	for _, templateSpell := range template.Spells {
		if templateSpell.Spell.ID == uuid.Nil {
			continue
		}
		spells = append(spells, templateSpell.Spell)
	}
	return &monsterTemplateResponse{
		ID:                    template.ID,
		CreatedAt:             template.CreatedAt,
		UpdatedAt:             template.UpdatedAt,
		Name:                  template.Name,
		Description:           template.Description,
		ImageURL:              template.ImageURL,
		ThumbnailURL:          template.ThumbnailURL,
		BaseStrength:          template.BaseStrength,
		BaseDexterity:         template.BaseDexterity,
		BaseConstitution:      template.BaseConstitution,
		BaseIntelligence:      template.BaseIntelligence,
		BaseWisdom:            template.BaseWisdom,
		BaseCharisma:          template.BaseCharisma,
		Spells:                spells,
		ImageGenerationStatus: template.ImageGenerationStatus,
		ImageGenerationError:  template.ImageGenerationError,
	}
}

func monsterResponseFrom(monster *models.Monster) monsterResponse {
	stats := monster.EffectiveStats()
	maxHealth := monster.DerivedMaxHealth()
	maxMana := monster.DerivedMaxMana()
	damageMin, damageMax, swipes := monster.DerivedAttackProfile()
	spells := []models.Spell{}
	if monster.Template != nil {
		for _, templateSpell := range monster.Template.Spells {
			if templateSpell.Spell.ID == uuid.Nil {
				continue
			}
			spells = append(spells, templateSpell.Spell)
		}
	}
	imageURL := monster.ImageURL
	if imageURL == "" && monster.Template != nil {
		imageURL = monster.Template.ImageURL
	}
	thumbnailURL := monster.ThumbnailURL
	if thumbnailURL == "" && monster.Template != nil {
		thumbnailURL = monster.Template.ThumbnailURL
	}
	if thumbnailURL == "" {
		thumbnailURL = imageURL
	}

	dominantItemID := monster.DominantHandInventoryItemID
	dominantItem := monster.DominantHandInventoryItem
	if dominantItemID == nil {
		dominantItemID = monster.WeaponInventoryItemID
	}
	if dominantItem == nil {
		dominantItem = monster.WeaponInventoryItem
	}

	return monsterResponse{
		ID:                          monster.ID,
		CreatedAt:                   monster.CreatedAt,
		UpdatedAt:                   monster.UpdatedAt,
		Name:                        monster.Name,
		Description:                 monster.Description,
		ImageURL:                    imageURL,
		ThumbnailURL:                thumbnailURL,
		ZoneID:                      monster.ZoneID,
		Zone:                        monster.Zone,
		Latitude:                    monster.Latitude,
		Longitude:                   monster.Longitude,
		TemplateID:                  monster.TemplateID,
		Template:                    monsterTemplateResponseFrom(monster.Template),
		DominantHandInventoryItemID: dominantItemID,
		DominantHandInventoryItem:   dominantItem,
		OffHandInventoryItemID:      monster.OffHandInventoryItemID,
		OffHandInventoryItem:        monster.OffHandInventoryItem,
		WeaponInventoryItemID:       dominantItemID,
		WeaponInventoryItem:         dominantItem,
		Level:                       monster.EffectiveLevel(),
		Strength:                    stats.Strength,
		Dexterity:                   stats.Dexterity,
		Constitution:                stats.Constitution,
		Intelligence:                stats.Intelligence,
		Wisdom:                      stats.Wisdom,
		Charisma:                    stats.Charisma,
		Health:                      maxHealth,
		MaxHealth:                   maxHealth,
		Mana:                        maxMana,
		MaxMana:                     maxMana,
		AttackDamageMin:             damageMin,
		AttackDamageMax:             damageMax,
		AttackSwipesPerAttack:       swipes,
		Spells:                      spells,
		RewardExperience:            monster.RewardExperience,
		RewardGold:                  monster.RewardGold,
		ItemRewards:                 monster.ItemRewards,
		ImageGenerationStatus:       monster.ImageGenerationStatus,
		ImageGenerationError:        monster.ImageGenerationError,
	}
}

func (s *server) parseMonsterTemplateUpsertRequest(
	ctx context.Context,
	body monsterTemplateUpsertRequest,
) (*models.MonsterTemplate, []models.MonsterTemplateSpell, error) {
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	if body.BaseStrength < 1 ||
		body.BaseDexterity < 1 ||
		body.BaseConstitution < 1 ||
		body.BaseIntelligence < 1 ||
		body.BaseWisdom < 1 ||
		body.BaseCharisma < 1 {
		return nil, nil, fmt.Errorf("all base stats must be positive")
	}

	spells := []models.MonsterTemplateSpell{}
	seenSpellIDs := map[uuid.UUID]bool{}
	for index, spellIDString := range body.SpellIDs {
		spellID, err := uuid.Parse(strings.TrimSpace(spellIDString))
		if err != nil {
			return nil, nil, fmt.Errorf("spellIds[%d] must be a valid UUID", index)
		}
		if seenSpellIDs[spellID] {
			continue
		}
		seenSpellIDs[spellID] = true
		if _, err := s.dbClient.Spell().FindByID(ctx, spellID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fmt.Errorf("spellIds[%d] not found", index)
			}
			return nil, nil, err
		}
		spells = append(spells, models.MonsterTemplateSpell{SpellID: spellID})
	}

	template := &models.MonsterTemplate{
		Name:             name,
		Description:      strings.TrimSpace(body.Description),
		ImageURL:         strings.TrimSpace(body.ImageURL),
		ThumbnailURL:     strings.TrimSpace(body.ThumbnailURL),
		BaseStrength:     body.BaseStrength,
		BaseDexterity:    body.BaseDexterity,
		BaseConstitution: body.BaseConstitution,
		BaseIntelligence: body.BaseIntelligence,
		BaseWisdom:       body.BaseWisdom,
		BaseCharisma:     body.BaseCharisma,
	}
	if template.ThumbnailURL == "" && template.ImageURL != "" {
		template.ThumbnailURL = template.ImageURL
	}
	return template, spells, nil
}

func (s *server) parseMonsterUpsertRequest(
	ctx context.Context,
	body monsterUpsertRequest,
) (*models.Monster, []models.MonsterItemReward, error) {
	zoneID, err := uuid.Parse(strings.TrimSpace(body.ZoneID))
	if err != nil {
		return nil, nil, fmt.Errorf("zoneId must be a valid UUID")
	}
	if _, err := s.dbClient.Zone().FindByID(ctx, zoneID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("zone not found")
		}
		return nil, nil, err
	}

	templateID, err := uuid.Parse(strings.TrimSpace(body.TemplateID))
	if err != nil {
		return nil, nil, fmt.Errorf("templateId must be a valid UUID")
	}
	template, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("template not found")
		}
		return nil, nil, err
	}

	if body.Level < 1 {
		return nil, nil, fmt.Errorf("level must be positive")
	}
	if body.RewardExperience < 0 {
		return nil, nil, fmt.Errorf("rewardExperience must be zero or greater")
	}
	if body.RewardGold < 0 {
		return nil, nil, fmt.Errorf("rewardGold must be zero or greater")
	}

	dominantItemID := body.DominantHandInventoryItemID
	if dominantItemID == nil || (dominantItemID != nil && *dominantItemID <= 0) {
		if body.WeaponInventoryItemID != nil && *body.WeaponInventoryItemID > 0 {
			dominantItemID = body.WeaponInventoryItemID
		}
	}
	if dominantItemID == nil || *dominantItemID <= 0 {
		return nil, nil, fmt.Errorf("dominantHandInventoryItemId is required")
	}
	dominantItem, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, *dominantItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("dominantHandInventoryItemId not found")
		}
		return nil, nil, err
	}
	if !isEligibleMonsterDominantHandItem(dominantItem) {
		return nil, nil, fmt.Errorf("dominantHandInventoryItemId must reference an eligible dominant-hand weapon or staff")
	}

	var offHandItemID *int
	var offHandItem *models.InventoryItem
	if body.OffHandInventoryItemID != nil && *body.OffHandInventoryItemID <= 0 {
		return nil, nil, fmt.Errorf("offHandInventoryItemId must be positive when set")
	}
	if body.OffHandInventoryItemID != nil && *body.OffHandInventoryItemID > 0 {
		offHandItemID = body.OffHandInventoryItemID
		offHandItem, err = s.dbClient.InventoryItem().FindInventoryItemByID(ctx, *offHandItemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fmt.Errorf("offHandInventoryItemId not found")
			}
			return nil, nil, err
		}
		if !isEligibleMonsterOffHandItem(offHandItem) {
			return nil, nil, fmt.Errorf("offHandInventoryItemId must reference an eligible off-hand item or one-handed weapon")
		}
		if isTwoHandedDominantItem(dominantItem) {
			return nil, nil, fmt.Errorf("offHandInventoryItemId cannot be set when dominant hand item is two_handed")
		}
	}

	imageURL := strings.TrimSpace(body.ImageURL)
	thumbnailURL := strings.TrimSpace(body.ThumbnailURL)
	if imageURL == "" {
		imageURL = strings.TrimSpace(template.ImageURL)
	}
	if thumbnailURL == "" {
		thumbnailURL = strings.TrimSpace(template.ThumbnailURL)
	}
	if thumbnailURL == "" && imageURL != "" {
		thumbnailURL = imageURL
	}

	itemQtyByID := map[int]int{}
	for index, reward := range body.ItemRewards {
		if reward.InventoryItemID <= 0 {
			return nil, nil, fmt.Errorf("itemRewards[%d].inventoryItemId must be positive", index)
		}
		if reward.Quantity <= 0 {
			return nil, nil, fmt.Errorf("itemRewards[%d].quantity must be positive", index)
		}
		if _, err := s.dbClient.InventoryItem().FindInventoryItemByID(ctx, reward.InventoryItemID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fmt.Errorf("itemRewards[%d].inventoryItemId not found", index)
			}
			return nil, nil, err
		}
		itemQtyByID[reward.InventoryItemID] += reward.Quantity
	}
	itemRewards := make([]models.MonsterItemReward, 0, len(itemQtyByID))
	for inventoryItemID, quantity := range itemQtyByID {
		itemRewards = append(itemRewards, models.MonsterItemReward{
			InventoryItemID: inventoryItemID,
			Quantity:        quantity,
		})
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = template.Name
	}
	description := strings.TrimSpace(body.Description)
	if description == "" {
		description = template.Description
	}

	monster := &models.Monster{
		Name:                        name,
		Description:                 description,
		ImageURL:                    imageURL,
		ThumbnailURL:                thumbnailURL,
		ZoneID:                      zoneID,
		Latitude:                    body.Latitude,
		Longitude:                   body.Longitude,
		TemplateID:                  &templateID,
		DominantHandInventoryItemID: dominantItemID,
		OffHandInventoryItemID:      offHandItemID,
		WeaponInventoryItemID:       dominantItemID,
		Level:                       body.Level,
		RewardExperience:            body.RewardExperience,
		RewardGold:                  body.RewardGold,
		ImageGenerationStatus:       models.MonsterImageGenerationStatusNone,
	}
	return monster, itemRewards, nil
}

func isEligibleMonsterDominantHandItem(item *models.InventoryItem) bool {
	if item == nil || item.EquipSlot == nil {
		return false
	}
	if strings.TrimSpace(*item.EquipSlot) != string(models.EquipmentSlotDominantHand) {
		return false
	}
	if item.HandItemCategory == nil {
		return false
	}
	category := strings.TrimSpace(*item.HandItemCategory)
	if category != string(models.HandItemCategoryWeapon) && category != string(models.HandItemCategoryStaff) {
		return false
	}
	return item.DamageMin != nil && item.DamageMax != nil && item.SwipesPerAttack != nil
}

func isEligibleMonsterOffHandItem(item *models.InventoryItem) bool {
	if item == nil || item.EquipSlot == nil {
		return false
	}
	equipSlot := strings.TrimSpace(*item.EquipSlot)
	category := ""
	if item.HandItemCategory != nil {
		category = strings.TrimSpace(*item.HandItemCategory)
	}
	handedness := ""
	if item.Handedness != nil {
		handedness = strings.TrimSpace(*item.Handedness)
	}
	if equipSlot == string(models.EquipmentSlotOffHand) &&
		handedness == string(models.HandednessOneHanded) &&
		(category == string(models.HandItemCategoryShield) || category == string(models.HandItemCategoryOrb)) {
		return true
	}
	if equipSlot == string(models.EquipmentSlotDominantHand) &&
		handedness == string(models.HandednessOneHanded) &&
		category == string(models.HandItemCategoryWeapon) &&
		item.DamageMin != nil &&
		item.DamageMax != nil &&
		item.SwipesPerAttack != nil {
		return true
	}
	return false
}

func (s *server) getMonsterTemplates(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	templates, err := s.dbClient.MonsterTemplate().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]monsterTemplateResponse, 0, len(templates))
	for i := range templates {
		response = append(response, *monsterTemplateResponseFrom(&templates[i]))
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getMonsterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	template, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, monsterTemplateResponseFrom(template))
}

func (s *server) createMonsterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody monsterTemplateUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, spells, err := s.parseMonsterTemplateUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if template.ImageURL != "" {
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusComplete
		emptyError := ""
		template.ImageGenerationError = &emptyError
	} else {
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusNone
		emptyError := ""
		template.ImageGenerationError = &emptyError
	}

	if err := s.dbClient.MonsterTemplate().Create(ctx, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MonsterTemplate().ReplaceSpells(ctx, template.ID, spells); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.dbClient.MonsterTemplate().FindByID(ctx, template.ID)
	if err != nil {
		ctx.JSON(http.StatusCreated, monsterTemplateResponseFrom(template))
		return
	}
	ctx.JSON(http.StatusCreated, monsterTemplateResponseFrom(created))
}

func (s *server) updateMonsterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}
	existingTemplate, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody monsterTemplateUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, spells, err := s.parseMonsterTemplateUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if template.ImageURL != "" {
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusComplete
		emptyError := ""
		template.ImageGenerationError = &emptyError
	} else if existingTemplate.ImageGenerationStatus == models.MonsterTemplateImageGenerationStatusQueued ||
		existingTemplate.ImageGenerationStatus == models.MonsterTemplateImageGenerationStatusInProgress {
		template.ImageGenerationStatus = existingTemplate.ImageGenerationStatus
		template.ImageGenerationError = existingTemplate.ImageGenerationError
	} else {
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusNone
		emptyError := ""
		template.ImageGenerationError = &emptyError
	}

	if err := s.dbClient.MonsterTemplate().Update(ctx, templateID, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.MonsterTemplate().ReplaceSpells(ctx, templateID, spells); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"id": templateID})
		return
	}
	ctx.JSON(http.StatusOK, monsterTemplateResponseFrom(updated))
}

func (s *server) deleteMonsterTemplate(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	if _, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	usageCount, err := s.dbClient.Monster().CountByTemplateID(ctx, templateID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if usageCount > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "template is in use by existing monsters"})
		return
	}

	if err := s.dbClient.MonsterTemplate().Delete(ctx, templateID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "template deleted successfully"})
}

func (s *server) generateMonsterTemplateImage(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	templateID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	template, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusQueued
	emptyError := ""
	template.ImageGenerationError = &emptyError
	if err := s.dbClient.MonsterTemplate().Update(ctx, templateID, template); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue monster template image generation: " + err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.GenerateMonsterTemplateImageTaskPayload{MonsterTemplateID: templateID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build monster template image generation payload"})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateMonsterTemplateImageTaskType, payloadBytes)); err != nil {
		template.ImageGenerationStatus = models.MonsterTemplateImageGenerationStatusFailed
		errMessage := err.Error()
		template.ImageGenerationError = &errMessage
		_ = s.dbClient.MonsterTemplate().Update(ctx, templateID, template)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue monster template image generation: " + err.Error()})
		return
	}

	updatedTemplate, err := s.dbClient.MonsterTemplate().FindByID(ctx, templateID)
	if err != nil {
		ctx.JSON(http.StatusOK, monsterTemplateResponseFrom(template))
		return
	}
	ctx.JSON(http.StatusOK, monsterTemplateResponseFrom(updatedTemplate))
}

func (s *server) getMonsters(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsters, err := s.dbClient.Monster().FindAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]monsterResponse, 0, len(monsters))
	for i := range monsters {
		response = append(response, monsterResponseFrom(&monsters[i]))
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) getMonster(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	monster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, monsterResponseFrom(monster))
}

func (s *server) getMonstersForZone(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}

	monsters, err := s.dbClient.Monster().FindByZoneID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]monsterResponse, 0, len(monsters))
	for i := range monsters {
		response = append(response, monsterResponseFrom(&monsters[i]))
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *server) createMonster(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var requestBody monsterUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	monster, itemRewards, err := s.parseMonsterUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if monster.ImageURL != "" {
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusComplete
		emptyError := ""
		monster.ImageGenerationError = &emptyError
	}

	if err := s.dbClient.Monster().Create(ctx, monster); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Monster().ReplaceItemRewards(ctx, monster.ID, itemRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := s.dbClient.Monster().FindByID(ctx, monster.ID)
	if err != nil {
		ctx.JSON(http.StatusCreated, monsterResponseFrom(monster))
		return
	}
	ctx.JSON(http.StatusCreated, monsterResponseFrom(created))
}

func (s *server) updateMonster(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	existingMonster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var requestBody monsterUpsertRequest
	if err := ctx.Bind(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	monster, itemRewards, err := s.parseMonsterUpsertRequest(ctx, requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if monster.ImageURL != "" {
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusComplete
		emptyError := ""
		monster.ImageGenerationError = &emptyError
	} else if existingMonster.ImageGenerationStatus == models.MonsterImageGenerationStatusQueued ||
		existingMonster.ImageGenerationStatus == models.MonsterImageGenerationStatusInProgress {
		monster.ImageGenerationStatus = existingMonster.ImageGenerationStatus
		monster.ImageGenerationError = existingMonster.ImageGenerationError
	} else {
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusNone
		emptyError := ""
		monster.ImageGenerationError = &emptyError
	}

	if err := s.dbClient.Monster().Update(ctx, monsterID, monster); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := s.dbClient.Monster().ReplaceItemRewards(ctx, monsterID, itemRewards); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"id": monsterID})
		return
	}
	ctx.JSON(http.StatusOK, monsterResponseFrom(updated))
}

func (s *server) deleteMonster(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	if _, err := s.dbClient.Monster().FindByID(ctx, monsterID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := s.dbClient.Monster().Delete(ctx, monsterID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "monster deleted successfully"})
}

func (s *server) generateMonsterImage(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	monsterID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid monster ID"})
		return
	}

	monster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "monster not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	monster.ImageGenerationStatus = models.MonsterImageGenerationStatusQueued
	emptyError := ""
	monster.ImageGenerationError = &emptyError
	if err := s.dbClient.Monster().Update(ctx, monsterID, monster); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue monster image generation: " + err.Error()})
		return
	}

	payloadBytes, err := json.Marshal(jobs.GenerateMonsterImageTaskPayload{MonsterID: monsterID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build monster image generation payload"})
		return
	}

	if _, err := s.asyncClient.Enqueue(asynq.NewTask(jobs.GenerateMonsterImageTaskType, payloadBytes)); err != nil {
		monster.ImageGenerationStatus = models.MonsterImageGenerationStatusFailed
		errMessage := err.Error()
		monster.ImageGenerationError = &errMessage
		_ = s.dbClient.Monster().Update(ctx, monsterID, monster)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue monster image generation: " + err.Error()})
		return
	}

	updatedMonster, err := s.dbClient.Monster().FindByID(ctx, monsterID)
	if err != nil {
		ctx.JSON(http.StatusOK, monsterResponseFrom(monster))
		return
	}
	ctx.JSON(http.StatusOK, monsterResponseFrom(updatedMonster))
}
