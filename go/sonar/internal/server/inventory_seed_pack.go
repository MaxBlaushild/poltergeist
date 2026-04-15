package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type seededInventoryPackItem struct {
	Name     string `json:"name"`
	Action   string `json:"action"`
	Category string `json:"category"`
}

type seededInventoryPackResponse struct {
	ProcessedCount int                       `json:"processedCount"`
	CreatedCount   int                       `json:"createdCount"`
	UpdatedCount   int                       `json:"updatedCount"`
	Items          []seededInventoryPackItem `json:"items"`
}

type inventorySeedPackRequest struct {
	Category string
	Request  inventoryItemUpsertRequest
}

type inventorySeedSetConfig struct {
	Theme                        string
	TargetLevel                  int
	RarityTier                   string
	MajorStat                    string
	MinorStat                    string
	InternalTags                 []string
	DamageBonusAffinity          string
	SecondaryDamageBonusAffinity string
	ResistanceAffinity           string
	SecondaryResistanceAffinity  string
}

type inventorySeedMaterialSpec struct {
	Name       string
	ItemLevel  int
	RarityTier string
	FlavorText string
	EffectText string
	BuyPrice   int
	Tags       []string
}

type inventorySeedUtilitySpec struct {
	Name                string
	ItemLevel           int
	RarityTier          string
	FlavorText          string
	EffectText          string
	BuyPrice            int
	UnlockLocksStrength *int
	ConsumeCreateBase   bool
	Tags                []string
}

func (s *server) seedInventoryCorePack(ctx *gin.Context) {
	if _, err := s.getAuthenticatedUser(ctx); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	requests := inventoryCoreSeedPackRequests()
	existingItems, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	existingByKey := make(map[string]models.InventoryItem, len(existingItems))
	for _, existing := range existingItems {
		key := seededInventoryItemKey(existing.Name)
		if key == "" {
			continue
		}
		existingByKey[key] = existing
	}

	response := seededInventoryPackResponse{
		Items: make([]seededInventoryPackItem, 0, len(requests)),
	}

	for _, seed := range requests {
		key := seededInventoryItemKey(seed.Request.Name)
		existing, exists := existingByKey[key]

		request := seed.Request
		if exists && strings.TrimSpace(request.ImageURL) == "" {
			request.ImageURL = existing.ImageURL
		}

		var normalized *models.InventoryItem
		if exists {
			normalized, err = s.normalizeInventoryItemUpsertRequest(ctx, request, &existing)
		} else {
			normalized, err = s.normalizeInventoryItemUpsertRequest(ctx, request, nil)
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to normalize seed item %q: %v", request.Name, err),
			})
			return
		}

		if exists {
			if err := s.dbClient.InventoryItem().UpdateInventoryItem(
				ctx,
				existing.ID,
				inventoryItemUpdatesFromModel(normalized, strings.TrimSpace(request.ImageURL) != ""),
			); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("failed to update seed item %q: %v", normalized.Name, err),
				})
				return
			}
			response.UpdatedCount++
			response.Items = append(response.Items, seededInventoryPackItem{
				Name:     normalized.Name,
				Action:   "updated",
				Category: seed.Category,
			})
			response.ProcessedCount++
			continue
		}

		if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, normalized); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to create seed item %q: %v", normalized.Name, err),
			})
			return
		}
		existingByKey[key] = *normalized
		response.CreatedCount++
		response.Items = append(response.Items, seededInventoryPackItem{
			Name:     normalized.Name,
			Action:   "created",
			Category: seed.Category,
		})
		response.ProcessedCount++
	}

	ctx.JSON(http.StatusOK, response)
}

func seededInventoryItemKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func inventoryItemUpdatesFromModel(
	item *models.InventoryItem,
	hasExplicitImage bool,
) map[string]interface{} {
	if item == nil {
		return map[string]interface{}{}
	}
	updates := map[string]interface{}{
		"archived":                                       item.Archived,
		"name":                                           item.Name,
		"image_url":                                      item.ImageURL,
		"flavor_text":                                    item.FlavorText,
		"effect_text":                                    item.EffectText,
		"rarity_tier":                                    item.RarityTier,
		"resource_type_id":                               item.ResourceTypeID,
		"is_capture_type":                                item.IsCaptureType,
		"buy_price":                                      item.BuyPrice,
		"unlock_tier":                                    item.UnlockTier,
		"unlock_locks_strength":                          item.UnlockLocksStrength,
		"item_level":                                     item.ItemLevel,
		"equip_slot":                                     item.EquipSlot,
		"strength_mod":                                   item.StrengthMod,
		"dexterity_mod":                                  item.DexterityMod,
		"constitution_mod":                               item.ConstitutionMod,
		"intelligence_mod":                               item.IntelligenceMod,
		"wisdom_mod":                                     item.WisdomMod,
		"charisma_mod":                                   item.CharismaMod,
		"physical_damage_bonus_percent":                  item.PhysicalDamageBonusPercent,
		"piercing_damage_bonus_percent":                  item.PiercingDamageBonusPercent,
		"slashing_damage_bonus_percent":                  item.SlashingDamageBonusPercent,
		"bludgeoning_damage_bonus_percent":               item.BludgeoningDamageBonusPercent,
		"fire_damage_bonus_percent":                      item.FireDamageBonusPercent,
		"ice_damage_bonus_percent":                       item.IceDamageBonusPercent,
		"lightning_damage_bonus_percent":                 item.LightningDamageBonusPercent,
		"poison_damage_bonus_percent":                    item.PoisonDamageBonusPercent,
		"arcane_damage_bonus_percent":                    item.ArcaneDamageBonusPercent,
		"holy_damage_bonus_percent":                      item.HolyDamageBonusPercent,
		"shadow_damage_bonus_percent":                    item.ShadowDamageBonusPercent,
		"physical_resistance_percent":                    item.PhysicalResistancePercent,
		"piercing_resistance_percent":                    item.PiercingResistancePercent,
		"slashing_resistance_percent":                    item.SlashingResistancePercent,
		"bludgeoning_resistance_percent":                 item.BludgeoningResistancePercent,
		"fire_resistance_percent":                        item.FireResistancePercent,
		"ice_resistance_percent":                         item.IceResistancePercent,
		"lightning_resistance_percent":                   item.LightningResistancePercent,
		"poison_resistance_percent":                      item.PoisonResistancePercent,
		"arcane_resistance_percent":                      item.ArcaneResistancePercent,
		"holy_resistance_percent":                        item.HolyResistancePercent,
		"shadow_resistance_percent":                      item.ShadowResistancePercent,
		"hand_item_category":                             item.HandItemCategory,
		"handedness":                                     item.Handedness,
		"damage_min":                                     item.DamageMin,
		"damage_max":                                     item.DamageMax,
		"damage_affinity":                                item.DamageAffinity,
		"swipes_per_attack":                              item.SwipesPerAttack,
		"block_percentage":                               item.BlockPercentage,
		"damage_blocked":                                 item.DamageBlocked,
		"spell_damage_bonus_percent":                     item.SpellDamageBonusPercent,
		"consume_health_delta":                           item.ConsumeHealthDelta,
		"consume_mana_delta":                             item.ConsumeManaDelta,
		"consume_revive_party_member_health":             item.ConsumeRevivePartyMemberHealth,
		"consume_revive_all_downed_party_members_health": item.ConsumeReviveAllDownedPartyMembersHealth,
		"consume_deal_damage":                            item.ConsumeDealDamage,
		"consume_deal_damage_hits":                       item.ConsumeDealDamageHits,
		"consume_deal_damage_all_enemies":                item.ConsumeDealDamageAllEnemies,
		"consume_deal_damage_all_enemies_hits":           item.ConsumeDealDamageAllEnemiesHits,
		"consume_create_base":                            item.ConsumeCreateBase,
		"consume_statuses_to_add":                        item.ConsumeStatusesToAdd,
		"consume_statuses_to_remove":                     item.ConsumeStatusesToRemove,
		"consume_spell_ids":                              item.ConsumeSpellIDs,
		"consume_teach_recipe_ids":                       item.ConsumeTeachRecipeIDs,
		"alchemy_recipes":                                item.AlchemyRecipes,
		"workshop_recipes":                               item.WorkshopRecipes,
		"internal_tags":                                  item.InternalTags,
	}
	if hasExplicitImage {
		updates["image_generation_status"] = models.InventoryImageGenerationStatusComplete
		updates["image_generation_error"] = ""
	}
	return updates
}

func inventoryCoreSeedPackRequests() []inventorySeedPackRequest {
	requests := make([]inventorySeedPackRequest, 0, 320)
	for _, config := range inventorySeedSetConfigs() {
		requests = append(requests, buildInventorySeedSetRequests(config)...)
	}
	for _, spec := range inventorySeedMaterialSpecs() {
		requests = append(requests, inventorySeedPackRequest{
			Category: "material",
			Request:  seededMaterialRequest(spec),
		})
	}
	for _, spec := range inventorySeedUtilitySpecs() {
		requests = append(requests, inventorySeedPackRequest{
			Category: "utility",
			Request:  seededUtilityRequest(spec),
		})
	}
	return requests
}

func inventorySeedSetConfigs() []inventorySeedSetConfig {
	return []inventorySeedSetConfig{
		{
			Theme:                        "Ironbound Vanguard",
			TargetLevel:                  12,
			RarityTier:                   "Common",
			MajorStat:                    "strength",
			MinorStat:                    "constitution",
			InternalTags:                 []string{"martial", "guard", "frontline"},
			DamageBonusAffinity:          "physical",
			SecondaryDamageBonusAffinity: "slashing",
			ResistanceAffinity:           "physical",
			SecondaryResistanceAffinity:  "piercing",
		},
		{
			Theme:                        "Nightstep Courier",
			TargetLevel:                  18,
			RarityTier:                   "Common",
			MajorStat:                    "dexterity",
			MinorStat:                    "charisma",
			InternalTags:                 []string{"rogue", "courier", "street"},
			DamageBonusAffinity:          "piercing",
			SecondaryDamageBonusAffinity: "slashing",
			ResistanceAffinity:           "shadow",
			SecondaryResistanceAffinity:  "physical",
		},
		{
			Theme:                        "Hearthwild Ranger",
			TargetLevel:                  24,
			RarityTier:                   "Common",
			MajorStat:                    "dexterity",
			MinorStat:                    "wisdom",
			InternalTags:                 []string{"ranger", "scout", "wild"},
			DamageBonusAffinity:          "poison",
			SecondaryDamageBonusAffinity: "piercing",
			ResistanceAffinity:           "poison",
			SecondaryResistanceAffinity:  "ice",
		},
		{
			Theme:                        "Emberline Duelist",
			TargetLevel:                  30,
			RarityTier:                   "Uncommon",
			MajorStat:                    "strength",
			MinorStat:                    "dexterity",
			InternalTags:                 []string{"duelist", "fire", "martial"},
			DamageBonusAffinity:          "fire",
			SecondaryDamageBonusAffinity: "slashing",
			ResistanceAffinity:           "fire",
			SecondaryResistanceAffinity:  "physical",
		},
		{
			Theme:                        "Mosswake Warden",
			TargetLevel:                  36,
			RarityTier:                   "Uncommon",
			MajorStat:                    "constitution",
			MinorStat:                    "wisdom",
			InternalTags:                 []string{"warden", "nature", "bulwark"},
			DamageBonusAffinity:          "bludgeoning",
			SecondaryDamageBonusAffinity: "poison",
			ResistanceAffinity:           "poison",
			SecondaryResistanceAffinity:  "bludgeoning",
		},
		{
			Theme:                        "Runebinder Vestments",
			TargetLevel:                  42,
			RarityTier:                   "Uncommon",
			MajorStat:                    "intelligence",
			MinorStat:                    "wisdom",
			InternalTags:                 []string{"caster", "arcane", "scholar"},
			DamageBonusAffinity:          "arcane",
			SecondaryDamageBonusAffinity: "holy",
			ResistanceAffinity:           "arcane",
			SecondaryResistanceAffinity:  "shadow",
		},
		{
			Theme:                        "Stormglass Regalia",
			TargetLevel:                  52,
			RarityTier:                   "Epic",
			MajorStat:                    "intelligence",
			MinorStat:                    "dexterity",
			InternalTags:                 []string{"storm", "caster", "tempo"},
			DamageBonusAffinity:          "lightning",
			SecondaryDamageBonusAffinity: "arcane",
			ResistanceAffinity:           "lightning",
			SecondaryResistanceAffinity:  "ice",
		},
		{
			Theme:                        "Gravebloom Reliquary",
			TargetLevel:                  60,
			RarityTier:                   "Epic",
			MajorStat:                    "wisdom",
			MinorStat:                    "constitution",
			InternalTags:                 []string{"occult", "shadow", "survivor"},
			DamageBonusAffinity:          "shadow",
			SecondaryDamageBonusAffinity: "poison",
			ResistanceAffinity:           "shadow",
			SecondaryResistanceAffinity:  "poison",
		},
		{
			Theme:                        "Sunforged Command",
			TargetLevel:                  68,
			RarityTier:                   "Epic",
			MajorStat:                    "charisma",
			MinorStat:                    "strength",
			InternalTags:                 []string{"leader", "holy", "martial"},
			DamageBonusAffinity:          "holy",
			SecondaryDamageBonusAffinity: "physical",
			ResistanceAffinity:           "holy",
			SecondaryResistanceAffinity:  "fire",
		},
		{
			Theme:                        "Frostveil Oracle",
			TargetLevel:                  76,
			RarityTier:                   "Epic",
			MajorStat:                    "wisdom",
			MinorStat:                    "intelligence",
			InternalTags:                 []string{"oracle", "ice", "seer"},
			DamageBonusAffinity:          "ice",
			SecondaryDamageBonusAffinity: "arcane",
			ResistanceAffinity:           "ice",
			SecondaryResistanceAffinity:  "shadow",
		},
		{
			Theme:                        "Ashen Crown Ascendant",
			TargetLevel:                  90,
			RarityTier:                   "Mythic",
			MajorStat:                    "intelligence",
			MinorStat:                    "charisma",
			InternalTags:                 []string{"mythic", "fire", "shadow", "sovereign"},
			DamageBonusAffinity:          "fire",
			SecondaryDamageBonusAffinity: "shadow",
			ResistanceAffinity:           "arcane",
			SecondaryResistanceAffinity:  "fire",
		},
		{
			Theme:                        "Bronzewind Skirmisher",
			TargetLevel:                  20,
			RarityTier:                   "Common",
			MajorStat:                    "dexterity",
			MinorStat:                    "strength",
			InternalTags:                 []string{"skirmisher", "storm", "martial"},
			DamageBonusAffinity:          "lightning",
			SecondaryDamageBonusAffinity: "piercing",
			ResistanceAffinity:           "lightning",
			SecondaryResistanceAffinity:  "physical",
		},
		{
			Theme:                        "Pale Lantern Keeper",
			TargetLevel:                  26,
			RarityTier:                   "Common",
			MajorStat:                    "wisdom",
			MinorStat:                    "constitution",
			InternalTags:                 []string{"warden", "holy", "guide"},
			DamageBonusAffinity:          "holy",
			SecondaryDamageBonusAffinity: "ice",
			ResistanceAffinity:           "holy",
			SecondaryResistanceAffinity:  "shadow",
		},
		{
			Theme:                        "Verdigris Saboteur",
			TargetLevel:                  32,
			RarityTier:                   "Uncommon",
			MajorStat:                    "dexterity",
			MinorStat:                    "intelligence",
			InternalTags:                 []string{"saboteur", "poison", "street"},
			DamageBonusAffinity:          "poison",
			SecondaryDamageBonusAffinity: "lightning",
			ResistanceAffinity:           "shadow",
			SecondaryResistanceAffinity:  "piercing",
		},
		{
			Theme:                        "Deepforge Bastion",
			TargetLevel:                  38,
			RarityTier:                   "Uncommon",
			MajorStat:                    "constitution",
			MinorStat:                    "strength",
			InternalTags:                 []string{"tank", "forge", "frontline"},
			DamageBonusAffinity:          "bludgeoning",
			SecondaryDamageBonusAffinity: "fire",
			ResistanceAffinity:           "fire",
			SecondaryResistanceAffinity:  "bludgeoning",
		},
		{
			Theme:                        "Mirrorwake Envoy",
			TargetLevel:                  44,
			RarityTier:                   "Uncommon",
			MajorStat:                    "charisma",
			MinorStat:                    "intelligence",
			InternalTags:                 []string{"envoy", "arcane", "social"},
			DamageBonusAffinity:          "arcane",
			SecondaryDamageBonusAffinity: "holy",
			ResistanceAffinity:           "arcane",
			SecondaryResistanceAffinity:  "shadow",
		},
		{
			Theme:                        "Thorncourt Harrier",
			TargetLevel:                  46,
			RarityTier:                   "Uncommon",
			MajorStat:                    "dexterity",
			MinorStat:                    "constitution",
			InternalTags:                 []string{"hunter", "poison", "wild"},
			DamageBonusAffinity:          "piercing",
			SecondaryDamageBonusAffinity: "poison",
			ResistanceAffinity:           "poison",
			SecondaryResistanceAffinity:  "physical",
		},
		{
			Theme:                        "Runecoil Artillerist",
			TargetLevel:                  54,
			RarityTier:                   "Epic",
			MajorStat:                    "intelligence",
			MinorStat:                    "strength",
			InternalTags:                 []string{"artillerist", "arcane", "storm"},
			DamageBonusAffinity:          "lightning",
			SecondaryDamageBonusAffinity: "arcane",
			ResistanceAffinity:           "lightning",
			SecondaryResistanceAffinity:  "arcane",
		},
		{
			Theme:                        "Kingshade Broker",
			TargetLevel:                  58,
			RarityTier:                   "Epic",
			MajorStat:                    "charisma",
			MinorStat:                    "dexterity",
			InternalTags:                 []string{"broker", "shadow", "social"},
			DamageBonusAffinity:          "shadow",
			SecondaryDamageBonusAffinity: "piercing",
			ResistanceAffinity:           "shadow",
			SecondaryResistanceAffinity:  "arcane",
		},
		{
			Theme:                        "Ashroot Hierophant",
			TargetLevel:                  62,
			RarityTier:                   "Epic",
			MajorStat:                    "wisdom",
			MinorStat:                    "intelligence",
			InternalTags:                 []string{"hierophant", "nature", "ritual"},
			DamageBonusAffinity:          "poison",
			SecondaryDamageBonusAffinity: "ice",
			ResistanceAffinity:           "poison",
			SecondaryResistanceAffinity:  "holy",
		},
		{
			Theme:                        "Stonehymn Anchor",
			TargetLevel:                  66,
			RarityTier:                   "Epic",
			MajorStat:                    "constitution",
			MinorStat:                    "charisma",
			InternalTags:                 []string{"anchor", "holy", "defender"},
			DamageBonusAffinity:          "bludgeoning",
			SecondaryDamageBonusAffinity: "holy",
			ResistanceAffinity:           "physical",
			SecondaryResistanceAffinity:  "holy",
		},
		{
			Theme:                        "Gloamwatch Marshal",
			TargetLevel:                  70,
			RarityTier:                   "Epic",
			MajorStat:                    "charisma",
			MinorStat:                    "wisdom",
			InternalTags:                 []string{"marshal", "shadow", "leader"},
			DamageBonusAffinity:          "shadow",
			SecondaryDamageBonusAffinity: "holy",
			ResistanceAffinity:           "shadow",
			SecondaryResistanceAffinity:  "holy",
		},
		{
			Theme:                        "Ivory Tempest Savant",
			TargetLevel:                  74,
			RarityTier:                   "Epic",
			MajorStat:                    "intelligence",
			MinorStat:                    "charisma",
			InternalTags:                 []string{"savant", "lightning", "court"},
			DamageBonusAffinity:          "lightning",
			SecondaryDamageBonusAffinity: "holy",
			ResistanceAffinity:           "lightning",
			SecondaryResistanceAffinity:  "ice",
		},
		{
			Theme:                        "Leviathan Corsair",
			TargetLevel:                  82,
			RarityTier:                   "Mythic",
			MajorStat:                    "strength",
			MinorStat:                    "dexterity",
			InternalTags:                 []string{"corsair", "mythic", "sea"},
			DamageBonusAffinity:          "slashing",
			SecondaryDamageBonusAffinity: "ice",
			ResistanceAffinity:           "ice",
			SecondaryResistanceAffinity:  "physical",
		},
		{
			Theme:                        "Astral Magistrate",
			TargetLevel:                  86,
			RarityTier:                   "Mythic",
			MajorStat:                    "charisma",
			MinorStat:                    "wisdom",
			InternalTags:                 []string{"magistrate", "arcane", "holy"},
			DamageBonusAffinity:          "holy",
			SecondaryDamageBonusAffinity: "arcane",
			ResistanceAffinity:           "holy",
			SecondaryResistanceAffinity:  "arcane",
		},
		{
			Theme:                        "Mireglass Hexer",
			TargetLevel:                  88,
			RarityTier:                   "Mythic",
			MajorStat:                    "intelligence",
			MinorStat:                    "wisdom",
			InternalTags:                 []string{"hexer", "poison", "shadow"},
			DamageBonusAffinity:          "poison",
			SecondaryDamageBonusAffinity: "shadow",
			ResistanceAffinity:           "poison",
			SecondaryResistanceAffinity:  "shadow",
		},
		{
			Theme:                        "Thunderking Standard",
			TargetLevel:                  92,
			RarityTier:                   "Mythic",
			MajorStat:                    "strength",
			MinorStat:                    "charisma",
			InternalTags:                 []string{"storm", "leader", "mythic"},
			DamageBonusAffinity:          "lightning",
			SecondaryDamageBonusAffinity: "physical",
			ResistanceAffinity:           "lightning",
			SecondaryResistanceAffinity:  "fire",
		},
		{
			Theme:                        "Winterrose Covenant",
			TargetLevel:                  94,
			RarityTier:                   "Mythic",
			MajorStat:                    "wisdom",
			MinorStat:                    "charisma",
			InternalTags:                 []string{"winter", "holy", "mythic"},
			DamageBonusAffinity:          "ice",
			SecondaryDamageBonusAffinity: "holy",
			ResistanceAffinity:           "ice",
			SecondaryResistanceAffinity:  "holy",
		},
		{
			Theme:                        "Cinder Saint Reliquary",
			TargetLevel:                  96,
			RarityTier:                   "Mythic",
			MajorStat:                    "constitution",
			MinorStat:                    "wisdom",
			InternalTags:                 []string{"saint", "fire", "relic"},
			DamageBonusAffinity:          "fire",
			SecondaryDamageBonusAffinity: "holy",
			ResistanceAffinity:           "fire",
			SecondaryResistanceAffinity:  "holy",
		},
	}
}

func buildInventorySeedSetRequests(config inventorySeedSetConfig) []inventorySeedPackRequest {
	targetLevel := maxInt(config.TargetLevel, 1)
	sourceItem := &models.InventoryItem{
		Name:          fmt.Sprintf("%s Core", strings.TrimSpace(config.Theme)),
		RarityTier:    config.RarityTier,
		ItemLevel:     targetLevel,
		UnlockTier:    intPtr(targetLevel),
		EquipSlot:     stringPtr(string(models.EquipmentSlotChest)),
		IsCaptureType: false,
	}
	setInventoryStatValue(
		sourceItem,
		config.MajorStat,
		inventorySetPrimaryStatPointsForTargetLevel(targetLevel, config.RarityTier),
	)
	setInventoryStatValue(
		sourceItem,
		config.MinorStat,
		maxInt(1, int(inventorySetPrimaryStatPointsForTargetLevel(targetLevel, config.RarityTier)*3/5)),
	)
	profile := deriveInventorySetProfile(sourceItem)
	slots := inventorySetAllEquippableSlots()
	requests := make([]inventorySeedPackRequest, 0, len(slots))
	for _, slot := range slots {
		name, handCategory := inventorySetItemName(sourceItem, config.Theme, slot, profile)
		strengthMod, dexterityMod, constitutionMod, intelligenceMod, wisdomMod, charismaMod :=
			inventorySetScaledStats(sourceItem, slot, profile)

		buyPrice := inventorySeedBuyPrice(targetLevel, config.RarityTier, slot, "equipment")
		internalTags := parseInventoryInternalTags(
			append(
				append([]string{"core_seed_pack", "equipment", slugifyInventorySeedTheme(config.Theme)}, config.InternalTags...),
				slot,
			),
		)

		request := inventoryItemUpsertRequest{
			Name:            name,
			FlavorText:      inventorySetFlavorText(config.Theme, slot, handCategory),
			EffectText:      "",
			RarityTier:      config.RarityTier,
			BuyPrice:        intPtr(buyPrice),
			UnlockTier:      intPtr(targetLevel),
			ItemLevel:       intPtr(targetLevel),
			EquipSlot:       stringPtr(slot),
			StrengthMod:     strengthMod,
			DexterityMod:    dexterityMod,
			ConstitutionMod: constitutionMod,
			IntelligenceMod: intelligenceMod,
			WisdomMod:       wisdomMod,
			CharismaMod:     charismaMod,
			InternalTags:    []string(internalTags),
		}

		if models.IsHandEquipSlot(slot) {
			attrs, _ := inventorySetHandAttributesForSlot(sourceItem, slot, profile)
			request.HandItemCategory = attrs.HandItemCategory
			request.Handedness = attrs.Handedness
			request.DamageMin = attrs.DamageMin
			request.DamageMax = attrs.DamageMax
			request.DamageAffinity = attrs.DamageAffinity
			request.SwipesPerAttack = attrs.SwipesPerAttack
			request.BlockPercentage = attrs.BlockPercentage
			request.DamageBlocked = attrs.DamageBlocked
			request.SpellDamageBonusPercent = attrs.SpellDamageBonusPercent
			if slot == string(models.EquipmentSlotDominantHand) && request.DamageAffinity != nil && strings.TrimSpace(config.DamageBonusAffinity) != "" {
				request.DamageAffinity = stringPtr(config.DamageBonusAffinity)
			}
		}

		applyInventorySeedAffinityIdentity(&request, config, slot)
		request.EffectText = buildInventorySeedSetEffectText(&request, handCategory)
		requests = append(requests, inventorySeedPackRequest{
			Category: "equipment",
			Request:  request,
		})
	}
	return requests
}

func applyInventorySeedAffinityIdentity(
	request *inventoryItemUpsertRequest,
	config inventorySeedSetConfig,
	slot string,
) {
	if request == nil {
		return
	}
	bigDamage := inventorySeedAffinityValue(config.RarityTier, 14, 22, 32, 42)
	mediumDamage := maxInt(4, int(bigDamage*2/3))
	smallDamage := maxInt(2, int(bigDamage/2))
	bigResist := inventorySeedAffinityValue(config.RarityTier, 12, 18, 26, 34)
	mediumResist := maxInt(4, int(bigResist*2/3))
	smallResist := maxInt(2, int(bigResist/2))

	switch slot {
	case string(models.EquipmentSlotDominantHand):
		applyInventorySeedDamageAffinityBonus(request, config.DamageBonusAffinity, bigDamage)
		applyInventorySeedDamageAffinityBonus(request, config.SecondaryDamageBonusAffinity, smallDamage)
	case string(models.EquipmentSlotOffHand):
		applyInventorySeedResistanceAffinityBonus(request, config.ResistanceAffinity, bigResist)
		applyInventorySeedResistanceAffinityBonus(request, config.SecondaryResistanceAffinity, smallResist)
	case string(models.EquipmentSlotRing):
		applyInventorySeedDamageAffinityBonus(request, config.DamageBonusAffinity, mediumDamage)
		applyInventorySeedDamageAffinityBonus(request, config.SecondaryDamageBonusAffinity, mediumDamage)
	case string(models.EquipmentSlotNecklace):
		applyInventorySeedResistanceAffinityBonus(request, config.ResistanceAffinity, mediumResist)
		applyInventorySeedResistanceAffinityBonus(request, config.SecondaryResistanceAffinity, mediumResist)
	case string(models.EquipmentSlotChest):
		applyInventorySeedResistanceAffinityBonus(request, config.ResistanceAffinity, mediumResist)
	case string(models.EquipmentSlotLegs):
		applyInventorySeedResistanceAffinityBonus(request, config.SecondaryResistanceAffinity, mediumResist)
	case string(models.EquipmentSlotHat), string(models.EquipmentSlotGloves):
		applyInventorySeedDamageAffinityBonus(request, config.SecondaryDamageBonusAffinity, smallDamage)
	case string(models.EquipmentSlotShoes):
		applyInventorySeedResistanceAffinityBonus(request, config.SecondaryResistanceAffinity, smallResist)
	}
}

func inventorySeedAffinityValue(rarity string, common int, uncommon int, epic int, mythic int) int {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "uncommon":
		return uncommon
	case "epic":
		return epic
	case "mythic":
		return mythic
	default:
		return common
	}
}

func applyInventorySeedDamageAffinityBonus(
	request *inventoryItemUpsertRequest,
	affinity string,
	value int,
) {
	switch strings.ToLower(strings.TrimSpace(affinity)) {
	case "physical":
		request.PhysicalDamageBonusPercent += value
	case "piercing":
		request.PiercingDamageBonusPercent += value
	case "slashing":
		request.SlashingDamageBonusPercent += value
	case "bludgeoning":
		request.BludgeoningDamageBonusPercent += value
	case "fire":
		request.FireDamageBonusPercent += value
	case "ice":
		request.IceDamageBonusPercent += value
	case "lightning":
		request.LightningDamageBonusPercent += value
	case "poison":
		request.PoisonDamageBonusPercent += value
	case "arcane":
		request.ArcaneDamageBonusPercent += value
	case "holy":
		request.HolyDamageBonusPercent += value
	case "shadow":
		request.ShadowDamageBonusPercent += value
	}
}

func applyInventorySeedResistanceAffinityBonus(
	request *inventoryItemUpsertRequest,
	affinity string,
	value int,
) {
	switch strings.ToLower(strings.TrimSpace(affinity)) {
	case "physical":
		request.PhysicalResistancePercent += value
	case "piercing":
		request.PiercingResistancePercent += value
	case "slashing":
		request.SlashingResistancePercent += value
	case "bludgeoning":
		request.BludgeoningResistancePercent += value
	case "fire":
		request.FireResistancePercent += value
	case "ice":
		request.IceResistancePercent += value
	case "lightning":
		request.LightningResistancePercent += value
	case "poison":
		request.PoisonResistancePercent += value
	case "arcane":
		request.ArcaneResistancePercent += value
	case "holy":
		request.HolyResistancePercent += value
	case "shadow":
		request.ShadowResistancePercent += value
	}
}

func buildInventorySeedSetEffectText(
	request *inventoryItemUpsertRequest,
	handCategory string,
) string {
	item := &models.InventoryItem{
		EquipSlot:                     request.EquipSlot,
		StrengthMod:                   request.StrengthMod,
		DexterityMod:                  request.DexterityMod,
		ConstitutionMod:               request.ConstitutionMod,
		IntelligenceMod:               request.IntelligenceMod,
		WisdomMod:                     request.WisdomMod,
		CharismaMod:                   request.CharismaMod,
		PhysicalDamageBonusPercent:    request.PhysicalDamageBonusPercent,
		PiercingDamageBonusPercent:    request.PiercingDamageBonusPercent,
		SlashingDamageBonusPercent:    request.SlashingDamageBonusPercent,
		BludgeoningDamageBonusPercent: request.BludgeoningDamageBonusPercent,
		FireDamageBonusPercent:        request.FireDamageBonusPercent,
		IceDamageBonusPercent:         request.IceDamageBonusPercent,
		LightningDamageBonusPercent:   request.LightningDamageBonusPercent,
		PoisonDamageBonusPercent:      request.PoisonDamageBonusPercent,
		ArcaneDamageBonusPercent:      request.ArcaneDamageBonusPercent,
		HolyDamageBonusPercent:        request.HolyDamageBonusPercent,
		ShadowDamageBonusPercent:      request.ShadowDamageBonusPercent,
		PhysicalResistancePercent:     request.PhysicalResistancePercent,
		PiercingResistancePercent:     request.PiercingResistancePercent,
		SlashingResistancePercent:     request.SlashingResistancePercent,
		BludgeoningResistancePercent:  request.BludgeoningResistancePercent,
		FireResistancePercent:         request.FireResistancePercent,
		IceResistancePercent:          request.IceResistancePercent,
		LightningResistancePercent:    request.LightningResistancePercent,
		PoisonResistancePercent:       request.PoisonResistancePercent,
		ArcaneResistancePercent:       request.ArcaneResistancePercent,
		HolyResistancePercent:         request.HolyResistancePercent,
		ShadowResistancePercent:       request.ShadowResistancePercent,
	}
	base := inventorySetEffectText(item, handCategory)
	affinities := make([]string, 0, 4)
	appendAffinity := func(label string, value int) {
		if value > 0 {
			affinities = append(affinities, label)
		}
	}
	appendAffinity("physical damage", request.PhysicalDamageBonusPercent)
	appendAffinity("piercing damage", request.PiercingDamageBonusPercent)
	appendAffinity("slashing damage", request.SlashingDamageBonusPercent)
	appendAffinity("bludgeoning damage", request.BludgeoningDamageBonusPercent)
	appendAffinity("fire damage", request.FireDamageBonusPercent)
	appendAffinity("ice damage", request.IceDamageBonusPercent)
	appendAffinity("lightning damage", request.LightningDamageBonusPercent)
	appendAffinity("poison damage", request.PoisonDamageBonusPercent)
	appendAffinity("arcane damage", request.ArcaneDamageBonusPercent)
	appendAffinity("holy damage", request.HolyDamageBonusPercent)
	appendAffinity("shadow damage", request.ShadowDamageBonusPercent)
	appendAffinity("physical warding", request.PhysicalResistancePercent)
	appendAffinity("piercing warding", request.PiercingResistancePercent)
	appendAffinity("slashing warding", request.SlashingResistancePercent)
	appendAffinity("bludgeoning warding", request.BludgeoningResistancePercent)
	appendAffinity("fire warding", request.FireResistancePercent)
	appendAffinity("ice warding", request.IceResistancePercent)
	appendAffinity("lightning warding", request.LightningResistancePercent)
	appendAffinity("poison warding", request.PoisonResistancePercent)
	appendAffinity("arcane warding", request.ArcaneResistancePercent)
	appendAffinity("holy warding", request.HolyResistancePercent)
	appendAffinity("shadow warding", request.ShadowResistancePercent)
	if len(affinities) == 0 {
		return base
	}
	if len(affinities) > 2 {
		affinities = affinities[:2]
	}
	return fmt.Sprintf("%s Reinforces %s.", base, strings.Join(affinities, " and "))
}

func inventorySeedBuyPrice(level int, rarity string, slot string, category string) int {
	rank := 1
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "uncommon":
		rank = 2
	case "epic":
		rank = 3
	case "mythic":
		rank = 4
	}

	if category == "material" {
		return maxInt(2, level*(rank+1)+rank*3)
	}
	if category == "utility" {
		return maxInt(8, level*(rank+3)+rank*8)
	}

	slotMultiplier := 1.0
	switch slot {
	case string(models.EquipmentSlotChest):
		slotMultiplier = 1.5
	case string(models.EquipmentSlotLegs), string(models.EquipmentSlotDominantHand):
		slotMultiplier = 1.3
	case string(models.EquipmentSlotOffHand):
		slotMultiplier = 1.2
	case string(models.EquipmentSlotHat), string(models.EquipmentSlotNecklace):
		slotMultiplier = 1.0
	case string(models.EquipmentSlotGloves), string(models.EquipmentSlotShoes):
		slotMultiplier = 0.9
	case string(models.EquipmentSlotRing):
		slotMultiplier = 1.1
	}
	base := float64(level*(rank+4) + rank*12)
	return maxInt(12, int(base*slotMultiplier))
}

func slugifyInventorySeedTheme(input string) string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(input)))
	if len(parts) == 0 {
		return "seed"
	}
	return strings.Join(parts, "_")
}

func seededMaterialRequest(spec inventorySeedMaterialSpec) inventoryItemUpsertRequest {
	level := maxInt(spec.ItemLevel, 1)
	return inventoryItemUpsertRequest{
		Name:       spec.Name,
		FlavorText: spec.FlavorText,
		EffectText: spec.EffectText,
		RarityTier: spec.RarityTier,
		BuyPrice:   intPtr(spec.BuyPrice),
		UnlockTier: intPtr(level),
		ItemLevel:  intPtr(level),
		InternalTags: []string(parseInventoryInternalTags(
			append([]string{"core_seed_pack", "material"}, spec.Tags...),
		)),
	}
}

func seededUtilityRequest(spec inventorySeedUtilitySpec) inventoryItemUpsertRequest {
	level := maxInt(spec.ItemLevel, 1)
	return inventoryItemUpsertRequest{
		Name:                spec.Name,
		FlavorText:          spec.FlavorText,
		EffectText:          spec.EffectText,
		RarityTier:          spec.RarityTier,
		BuyPrice:            intPtr(spec.BuyPrice),
		UnlockTier:          intPtr(level),
		ItemLevel:           intPtr(level),
		UnlockLocksStrength: spec.UnlockLocksStrength,
		ConsumeCreateBase:   spec.ConsumeCreateBase,
		InternalTags: []string(parseInventoryInternalTags(
			append([]string{"core_seed_pack", "utility"}, spec.Tags...),
		)),
	}
}

func inventorySeedMaterialSpecs() []inventorySeedMaterialSpec {
	return []inventorySeedMaterialSpec{
		{"Scrap Iron", 4, "Common", "Bent scraps and broken fasteners gathered from alleys and workshops.", "Basic smithing stock for rough workshop recipes.", 6, []string{"metal", "common_craft"}},
		{"Iron Ingot", 8, "Common", "A plain iron bar ready for straightforward forge work.", "Reliable workshop ingredient for early weapons and armor.", 10, []string{"metal", "forge"}},
		{"Treated Leather", 6, "Common", "Hide cured against rot and stitched into usable sheets.", "Basic armorworking material for light gear.", 8, []string{"leather", "armor"}},
		{"Tanned Hide", 10, "Common", "Flexible hide softened for straps, wraps, and lining.", "Common workshop component for rugged field gear.", 9, []string{"leather", "utility"}},
		{"Glass Phial", 5, "Common", "A thin clear vial with room for powders, oils, or tinctures.", "Base alchemy vessel for brews and reagents.", 7, []string{"alchemy", "container"}},
		{"Canvas Roll", 7, "Common", "Heavy woven cloth bundled around a hardwood spindle.", "Workshop textile for packs, wraps, and banners.", 8, []string{"cloth", "utility"}},
		{"Oak Heartwood", 9, "Common", "Dense wood cut from the center of an old oak beam.", "Stable crafting wood for handles and hafts.", 11, []string{"wood", "workshop"}},
		{"Bone Shard", 6, "Common", "Whitened fragments taken from large beasts and cleaned for use.", "Common occult and workshop component.", 7, []string{"bone", "occult"}},
		{"Copper Rivets", 4, "Common", "A pouch of soft copper rivets for quick repairs and assembly.", "Basic workshop fastener material.", 5, []string{"metal", "fastener"}},
		{"Linen Wrap", 5, "Common", "Folded lengths of breathable linen for binding wounds or gear.", "Cloth component used in kits and wrappings.", 6, []string{"cloth", "field"}},
		{"Mason's Chalk", 3, "Common", "Dry chalk sticks that leave bright marks even on damp stone.", "Cheap marking material for survey and ritual work.", 4, []string{"utility", "ritual"}},
		{"River Salt", 4, "Common", "Rough mineral salt harvested where brine meets runoff.", "Simple alchemical stabilizer and preservative.", 5, []string{"alchemy", "mineral"}},
		{"Coal Dust", 5, "Common", "Black dust swept from braziers, kilns, and old engine rooms.", "Fuel additive and reactive alchemy powder.", 6, []string{"fuel", "alchemy"}},
		{"Sulfur Lump", 8, "Common", "A yellow mineral chunk that reeks faintly of struck matches.", "Reactive reagent for aggressive brews and charges.", 9, []string{"alchemy", "volatile"}},
		{"Herbal Binder", 7, "Common", "Twine-bound bundles of drying leaves and stems.", "Mild alchemy base for poultices, powders, and salves.", 8, []string{"herb", "alchemy"}},
		{"Twine Spool", 2, "Common", "A palm-sized spool of tough brown fiber cord.", "Cheap workshop component for binding and rigging.", 3, []string{"cloth", "utility"}},
		{"Steel Ingot", 20, "Uncommon", "A clean steel billet with a bright edge and even grain.", "Improved forge stock for mid-tier workshop gear.", 24, []string{"metal", "forge"}},
		{"Silver Wire", 24, "Uncommon", "Thin silver strands wound carefully around a lacquered bobbin.", "Conductive arcane threading for charms, seals, and catalysts.", 28, []string{"metal", "arcane"}},
		{"Hardened Leather", 22, "Uncommon", "Layered leather treated with oil and pressure until it holds shape.", "Durable armorworking material for serious field gear.", 23, []string{"leather", "armor"}},
		{"Spider Silk Spool", 26, "Uncommon", "Fine silk wound from giant-web nests into a gleaming spool.", "High-flex textile for precise gear and occult bindings.", 30, []string{"cloth", "occult"}},
		{"Yew Stave Core", 28, "Uncommon", "A straight yew shaft chosen for balance and magical receptivity.", "Preferred core for staves, rods, and ritual poles.", 32, []string{"wood", "arcane"}},
		{"Quartz Lens", 25, "Uncommon", "A cut quartz disk polished until it catches every stray glimmer.", "Focusing component for workshop optics and alchemical rigs.", 27, []string{"crystal", "precision"}},
		{"Arcane Dust", 30, "Uncommon", "Soft shimmering dust that clings to the air for a heartbeat.", "Universal catalyst for arcane recipes.", 34, []string{"arcane", "alchemy"}},
		{"Ember Resin", 32, "Uncommon", "Amber-red resin warm to the touch and slow to cool.", "Heat-bearing reagent used in fire aligned crafting.", 35, []string{"fire", "alchemy"}},
		{"Frost Bloom", 30, "Uncommon", "Pale petals that stay cool long after being picked.", "Cooling reagent for ice-aligned brews and wards.", 33, []string{"ice", "alchemy"}},
		{"Stormglass Shard", 34, "Uncommon", "Blue-veined glass that hums softly when storms gather.", "Lightning-tuned crystal fragment for charged tools.", 36, []string{"lightning", "crystal"}},
		{"Nightshade", 29, "Uncommon", "Dark berries and leaves wrapped in waxed paper.", "Toxic reagent for poison work and shadow brews.", 31, []string{"poison", "herb"}},
		{"Glowcap Powder", 27, "Uncommon", "Luminescent mushroom dust stored in a stoppered tube.", "Useful for visibility brews and eerie alchemical mixes.", 29, []string{"fungus", "alchemy"}},
		{"Sunpetal", 31, "Uncommon", "Gold-white petals that smell faintly of warm stone and citrus.", "Radiant botanical for holy mixtures and restoratives.", 34, []string{"holy", "herb"}},
		{"Grave Moss", 33, "Uncommon", "Cold moss gathered from shaded memorial walls and crypt mouths.", "Shadow-leaning reagent for quiet, stubborn magic.", 35, []string{"shadow", "herb"}},
		{"Moon Dew", 35, "Uncommon", "Condensed silver dew bottled before dawn burns it away.", "Rare solvent for delicate alchemy and lucid inks.", 38, []string{"arcane", "liquid"}},
		{"Brass Gearwork", 28, "Uncommon", "Nested brass cogs and escapements packed in greaseproof cloth.", "Workshop mechanism stock for precise moving parts.", 30, []string{"mechanical", "workshop"}},
		{"Cold Iron Ingot", 48, "Epic", "A heavy ingot whose surface drinks in stray warmth.", "Specialist forge material valued against uncanny threats.", 70, []string{"metal", "anti_occult"}},
		{"Starsteel Filament", 56, "Epic", "Bright metal thread so light it seems to float in the spool.", "Premium smithing material for elite precision gear.", 82, []string{"metal", "mythic_forge"}},
		{"Voidglass Shard", 54, "Epic", "Black glass with an interior sheen like distant water.", "Rare occult crystal used in shadow and arcane constructs.", 76, []string{"shadow", "crystal"}},
		{"Echo Crystal", 52, "Epic", "A resonant crystal that answers even quiet taps with layered tones.", "High-tier focus crystal for signal, ward, and spell tools.", 74, []string{"arcane", "crystal"}},
		{"Runeslate Tablet", 50, "Epic", "Thin slate sheets that take etched glyphs with unnatural clarity.", "Workshop and ritual substrate for advanced formulas.", 68, []string{"ritual", "workshop"}},
		{"Phoenix Cinder", 62, "Epic", "Hot crimson ash sealed in a ceramic capsule.", "Powerful ignition reagent for fire-aligned masterwork crafting.", 88, []string{"fire", "mythic"}},
		{"Tempest Coil", 58, "Epic", "A silver coil that snaps with static whenever moved too quickly.", "Charged workshop component for stormbound mechanisms.", 84, []string{"lightning", "mechanical"}},
		{"Saint's Wax", 60, "Epic", "Pale wax cast around flecks of gold leaf and incense dust.", "Holy sealant for sanctified gear and reliquaries.", 86, []string{"holy", "ritual"}},
		{"Onyx Powder", 55, "Epic", "Fine black mineral powder that stains fingertips violet in candlelight.", "Dense shadow reagent for inks, wards, and catalysts.", 78, []string{"shadow", "alchemy"}},
		{"Moonstone Fragment", 57, "Epic", "Milk-bright stone splinters with a quiet lunar shimmer.", "Valuable component for perception and dreamwork craft.", 80, []string{"arcane", "stone"}},
		{"Bloodroot Resin", 53, "Epic", "Dark red resin packed in a waxed tin to keep it from drying.", "Aggressive reagent used in volatile poisons and salves.", 75, []string{"poison", "alchemy"}},
		{"Whisper Reed", 49, "Epic", "Thin gray reeds that seem to rustle even when perfectly still.", "Occult textile material for hidden messages and subtle wards.", 72, []string{"occult", "cloth"}},
		{"Gilded Bearings", 51, "Epic", "Tiny polished bearings stored in velvet to keep their finish.", "Fine workshop components for elite clockwork assemblies.", 73, []string{"mechanical", "precision"}},
		{"Aether Sand", 63, "Epic", "Blue-white grains sealed in glass to keep them from evaporating.", "Rare catalytic powder for advanced alchemy and enchantment.", 90, []string{"arcane", "alchemy"}},
		{"Umbral Thread", 59, "Epic", "Black thread that seems thinner whenever you try to focus on it.", "Shadowworking textile for cloaks, seals, and bindings.", 83, []string{"shadow", "cloth"}},
		{"Dawnsilver Leaf", 61, "Epic", "Hammered silver leaf with a faint warm glow at the edges.", "Radiant finishing material for holy and ceremonial craft.", 87, []string{"holy", "metal"}},
		{"Adamant Plate Blank", 64, "Epic", "A dense unfinished armor plate that rings low and clear when struck.", "Premium workshop base for heavy mythic armor.", 92, []string{"metal", "armor"}},
		{"Wyrmhide Panel", 67, "Epic", "Scaled hide cut into lacquered panels tough enough to turn a blade edge.", "Rare armorworking stock for elite field gear.", 95, []string{"leather", "mythic"}},
		{"Prism Alloy", 65, "Epic", "Iridescent metal that throws colored sparks when filed.", "Advanced alloy for hybrid elemental equipment.", 93, []string{"metal", "arcane"}},
		{"Blessed Thread", 47, "Epic", "Fine pale thread braided with incense ash and tiny seal knots.", "Holy textile used in vestments and relic wraps.", 69, []string{"holy", "cloth"}},
		{"Cinderglass Bead", 45, "Epic", "Red-black glass beads with tiny trapped bubbles like frozen sparks.", "Fire-aligned inlay component for elite gear.", 67, []string{"fire", "crystal"}},
		{"Farseer Vellum", 46, "Epic", "Smooth vellum sheets prepared to hold complex diagrams and wards.", "Rare inscription substrate for advanced formulas.", 68, []string{"ritual", "arcane"}},
		{"Gravetwine", 44, "Epic", "Dark braided cord with a faint scent of wet stone and old cedar.", "Shadowworking cord for bindings and occult kits.", 66, []string{"shadow", "cloth"}},
		{"Skyglass Lens", 43, "Epic", "A pale blue lens that seems clearer against moving cloud or smoke.", "Precision optic for stormbound workshop devices.", 65, []string{"lightning", "precision"}},
		{"Kingsfoil Distillate", 41, "Epic", "A sealed vial of fragrant green concentrate reduced from armfuls of herb.", "Potent botanical base for high-grade restorative craft.", 63, []string{"herb", "alchemy"}},
		{"Mothwing Felt", 39, "Uncommon", "Soft layered felt stitched with iridescent fibers from dusk moths.", "Light crafting textile for stealth and occult gear.", 40, []string{"cloth", "stealth"}},
		{"Anchor Chain Links", 37, "Uncommon", "Short lengths of blackened chain with the weight of shipyard work.", "Workshop metal stock for brute-force tools and armor.", 39, []string{"metal", "utility"}},
		{"Seabright Coral", 42, "Uncommon", "Branching coral trimmed and dried until it glows a soft cream color.", "Oceanic crafting reagent for holy and frost designs.", 44, []string{"holy", "ice"}},
		{"Dusk Copper Filigree", 40, "Uncommon", "Rolled strips of etched copper meant to be wrapped around casings.", "Decorative but functional conductor for refined workshop parts.", 42, []string{"metal", "lightning"}},
		{"Bog Iron Nugget", 18, "Common", "Rough iron lifted from dark wet soil and still streaked with peat.", "Common forge stock with a rugged natural feel.", 19, []string{"metal", "wild"}},
		{"Charred Cedar", 16, "Common", "Aromatic wood blackened on the outside and sound at the core.", "Useful handle and haft material for fire-leaning gear.", 17, []string{"wood", "fire"}},
		{"Gutterglass", 14, "Common", "Cloudy city glass sorted from broken lanterns and bottle heaps.", "Cheap crystal substitute for low-tier workshop builds.", 15, []string{"glass", "urban"}},
		{"Marrow Paste", 12, "Common", "Rendered bone marrow thickened into a pale waxy compound.", "Binding paste for bonecraft and field repairs.", 13, []string{"bone", "utility"}},
		{"Needle Fern", 11, "Common", "Dry serrated fronds bundled into a prickly bundle.", "Simple herb stock for toxins, poultices, and hardy brews.", 12, []string{"herb", "poison"}},
		{"Saintglass Fragment", 69, "Mythic", "Brilliant glass splinters recovered from ruined sanctum windows.", "Relic-grade component for luminous mythic gear.", 112, []string{"holy", "mythic"}},
		{"Stormheart Core", 71, "Mythic", "A dense metallic node that thrums like a caged thunderhead.", "Mythic engine component for apex lightning craft.", 116, []string{"lightning", "mythic"}},
		{"Ashblood Lacquer", 73, "Mythic", "Glossy lacquer that dries into a hard dark-red sheen within seconds.", "Top-tier finishing agent for fire and martial relics.", 120, []string{"fire", "mythic"}},
		{"Oracle Salt", 75, "Mythic", "Pale translucent crystals that whisper faintly when stirred.", "Rare reagent prized for foresight rituals and lucid alchemy.", 124, []string{"arcane", "mythic"}},
		{"Threnody Silk", 77, "Mythic", "Dark silk so smooth it seems to slide away from the hand.", "Mythic textile for shadow cloaks and ceremonial bindings.", 126, []string{"shadow", "mythic"}},
		{"Titan Rivet Kit", 79, "Mythic", "A heavy box of oversized alloy rivets, caps, and setting tools.", "Legendary workshop fasteners for monumental builds.", 130, []string{"metal", "mythic"}},
		{"Frostglass Vein", 81, "Mythic", "A branching seam of icy crystal wrapped in layered wool for transport.", "Mythic-grade frost catalyst for elite arms and wards.", 134, []string{"ice", "mythic"}},
		{"Venom Pearl", 83, "Mythic", "A slick dark pearl with a green sheen under direct light.", "Extremely rare toxin focus for apex alchemy.", 136, []string{"poison", "mythic"}},
		{"Sunspoke Alloy", 85, "Mythic", "Radiant metal rods arranged in a wax-sealed wooden tray.", "Holy masterwork stock for relic poles, crowns, and weapon cores.", 140, []string{"holy", "mythic"}},
		{"Crownsmoke Resin", 87, "Mythic", "Near-black resin that gives off a regal incense scent when warmed.", "Shadow-fire catalyst for sovereign-grade mythic craft.", 144, []string{"shadow", "fire", "mythic"}},
		{"Leviathan Bone", 72, "Mythic", "Dense ocean-pale bone carved from something vastly larger than a cart.", "Legendary crafting stock for mythic arms and relic frames.", 118, []string{"bone", "mythic"}},
		{"Celestial Ink", 78, "Mythic", "Ink that shines like midnight constellations when uncapped.", "Mythic inscription medium for peak ritual and arcane work.", 128, []string{"arcane", "ritual"}},
		{"Radiant Amber", 80, "Mythic", "Golden amber enclosing motes of light that never settle.", "High-holy catalyst for relic-grade crafting.", 132, []string{"holy", "mythic"}},
		{"Abyssal Pearl", 84, "Mythic", "A dark pearl with a sheen like moonlight on deep harbor water.", "Extremely rare occult focus for shadowed masterwork items.", 138, []string{"shadow", "mythic"}},
	}
}

func inventorySeedUtilitySpecs() []inventorySeedUtilitySpec {
	return []inventorySeedUtilitySpec{
		{
			Name:                "Lockbreaker's Wedge",
			ItemLevel:           12,
			RarityTier:          "Common",
			FlavorText:          "A flat steel wedge and grip-wrap built for stubborn street hardware.",
			EffectText:          "Improves lockbreaking by 18.",
			BuyPrice:            inventorySeedBuyPrice(12, "Common", "", "utility"),
			UnlockLocksStrength: intPtr(18),
			Tags:                []string{"utility", "lockpick", "street"},
		},
		{
			Name:                "Smuggler's False Key",
			ItemLevel:           20,
			RarityTier:          "Common",
			FlavorText:          "A ring of shaved blanks filed to bully cheap mechanisms into opening.",
			EffectText:          "Improves lockbreaking by 26.",
			BuyPrice:            inventorySeedBuyPrice(20, "Common", "", "utility"),
			UnlockLocksStrength: intPtr(26),
			Tags:                []string{"utility", "lockpick", "smuggler"},
		},
		{
			Name:                "Gilded Lockpick Set",
			ItemLevel:           34,
			RarityTier:          "Uncommon",
			FlavorText:          "Roll-wrapped picks, tension bars, and a tiny mirror packed for professionals.",
			EffectText:          "Improves lockbreaking by 38.",
			BuyPrice:            inventorySeedBuyPrice(34, "Uncommon", "", "utility"),
			UnlockLocksStrength: intPtr(38),
			Tags:                []string{"utility", "lockpick", "precision"},
		},
		{
			Name:                "Siege Prybar",
			ItemLevel:           48,
			RarityTier:          "Epic",
			FlavorText:          "A reinforced pry tool meant for doors that were never supposed to yield quietly.",
			EffectText:          "Improves lockbreaking by 54.",
			BuyPrice:            inventorySeedBuyPrice(48, "Epic", "", "utility"),
			UnlockLocksStrength: intPtr(54),
			Tags:                []string{"utility", "breach", "metal"},
		},
		{
			Name:                "Masterwork Breaching Kit",
			ItemLevel:           66,
			RarityTier:          "Epic",
			FlavorText:          "A velvet-lined case of specialized picks, wedges, and pressure tools for elite jobs.",
			EffectText:          "Improves lockbreaking by 72.",
			BuyPrice:            inventorySeedBuyPrice(66, "Epic", "", "utility"),
			UnlockLocksStrength: intPtr(72),
			Tags:                []string{"utility", "breach", "elite"},
		},
		{
			Name:              "Founder's Deed",
			ItemLevel:         24,
			RarityTier:        "Uncommon",
			FlavorText:        "A rolled deed stamped with enough wards and signatures to claim a place of your own.",
			EffectText:        "Consume to establish a base.",
			BuyPrice:          inventorySeedBuyPrice(24, "Uncommon", "", "utility"),
			ConsumeCreateBase: true,
			Tags:              []string{"utility", "base", "settlement"},
		},
		{
			Name:                "Whisperpick Sleeve",
			ItemLevel:           28,
			RarityTier:          "Uncommon",
			FlavorText:          "A stitched arm-wrap hiding slim picks, shims, and wire loops in narrow pockets.",
			EffectText:          "Improves lockbreaking by 32.",
			BuyPrice:            inventorySeedBuyPrice(28, "Uncommon", "", "utility"),
			UnlockLocksStrength: intPtr(32),
			Tags:                []string{"utility", "lockpick", "stealth"},
		},
		{
			Name:                "Vault Needle Set",
			ItemLevel:           56,
			RarityTier:          "Epic",
			FlavorText:          "A lacquered case of hardened needles, probes, and silent tension bars built for fine vault work.",
			EffectText:          "Improves lockbreaking by 60.",
			BuyPrice:            inventorySeedBuyPrice(56, "Epic", "", "utility"),
			UnlockLocksStrength: intPtr(60),
			Tags:                []string{"utility", "lockpick", "elite"},
		},
		{
			Name:              "Surveyor's Claim Stakes",
			ItemLevel:         36,
			RarityTier:        "Uncommon",
			FlavorText:        "A bundled set of iron stakes, measuring cord, and stamped markers meant to lay claim to a site.",
			EffectText:        "Consume to establish a base.",
			BuyPrice:          inventorySeedBuyPrice(36, "Uncommon", "", "utility"),
			ConsumeCreateBase: true,
			Tags:              []string{"utility", "base", "survey"},
		},
		{
			Name:              "Frontier Charter",
			ItemLevel:         52,
			RarityTier:        "Epic",
			FlavorText:        "An embossed charter sealed with enough authority to turn an outpost into recognized ground.",
			EffectText:        "Consume to establish a base.",
			BuyPrice:          inventorySeedBuyPrice(52, "Epic", "", "utility"),
			ConsumeCreateBase: true,
			Tags:              []string{"utility", "base", "charter"},
		},
	}
}
