package server

import (
	"fmt"
	"math/rand"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
)

type inventorySetFamilyGenerationRequest struct {
	Count                     int      `json:"count"`
	LevelMin                  int      `json:"levelMin"`
	LevelMax                  int      `json:"levelMax"`
	RarityTiers               []string `json:"rarityTiers"`
	Themes                    []string `json:"themes"`
	MajorStats                []string `json:"majorStats"`
	MinorStats                []string `json:"minorStats"`
	Profiles                  []string `json:"profiles"`
	DamageAffinities          []string `json:"damageAffinities"`
	ResistanceAffinities      []string `json:"resistanceAffinities"`
	RequiredInternalTags      []string `json:"requiredInternalTags"`
	ForbiddenTags             []string `json:"forbiddenTags"`
	SlotScope                 string   `json:"slotScope"`
	PowerBias                 string   `json:"powerBias"`
	NamingStyle               string   `json:"namingStyle"`
	AllowHybridAffinities     bool     `json:"allowHybridAffinities"`
	AvoidExistingThemeOverlap bool     `json:"avoidExistingThemeOverlap"`
	QueueImages               bool     `json:"queueImages"`
}

type inventorySetFamilyGenerationResponse struct {
	RequestedCount     int                              `json:"requestedCount"`
	CreatedFamilyCount int                              `json:"createdFamilyCount"`
	CreatedItemCount   int                              `json:"createdItemCount"`
	SkippedFamilyCount int                              `json:"skippedFamilyCount"`
	Families           []inventorySetGenerationResponse `json:"families"`
	SkippedReasons     []string                         `json:"skippedReasons"`
}

type inventorySetGenerationResponse struct {
	SourceItemID    *int                   `json:"sourceItemId,omitempty"`
	SetTheme        string                 `json:"setTheme"`
	TargetLevel     int                    `json:"targetLevel"`
	MajorStat       string                 `json:"majorStat"`
	MinorStat       string                 `json:"minorStat"`
	RarityTier      string                 `json:"rarityTier"`
	CreatedItems    []models.InventoryItem `json:"createdItems"`
	SkippedSlots    []string               `json:"skippedSlots"`
	EnqueueWarnings []string               `json:"enqueueWarnings,omitempty"`
	Message         string                 `json:"message"`
}

type inventorySetFamilyCandidate struct {
	Config inventorySeedSetConfig
	Slots  []string
}

var defaultInventoryFamilyProfiles = []string{
	"martial",
	"tank",
	"caster",
	"skirmisher",
	"support",
	"hybrid",
}

var defaultInventoryFamilyAffinities = []string{
	"physical",
	"piercing",
	"slashing",
	"bludgeoning",
	"fire",
	"ice",
	"lightning",
	"poison",
	"arcane",
	"holy",
	"shadow",
}

func (s *server) generateInventorySetFamilies(ctx *gin.Context) {
	var requestBody inventorySetFamilyGenerationRequest
	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	normalized, err := normalizeInventorySetFamilyGenerationRequest(requestBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingItems, err := s.dbClient.InventoryItem().FindAllInventoryItems(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	existingKeys := make(map[string]struct{}, len(existingItems))
	existingThemeSlugs := map[string]struct{}{}
	for _, existing := range existingItems {
		if existing.EquipSlot != nil && strings.TrimSpace(*existing.EquipSlot) != "" {
			existingKeys[inventorySetItemKey(normalizeInventorySetSlot(*existing.EquipSlot), existing.Name)] = struct{}{}
			themeSlug := slugifyInventorySeedTheme(inventorySetThemeFromName(existing.Name))
			if themeSlug != "" {
				existingThemeSlugs[themeSlug] = struct{}{}
			}
		}
	}

	usedThemeSlugs := map[string]struct{}{}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	families := make([]inventorySetGenerationResponse, 0, normalized.Count)
	skippedReasons := make([]string, 0)
	createdItemCount := 0
	attemptLimit := maxInt(normalized.Count*10, 12)

	for attempts := 0; len(families) < normalized.Count && attempts < attemptLimit; attempts++ {
		candidate, skipReason, ok := buildInventorySetFamilyCandidate(normalized, rng, existingThemeSlugs, usedThemeSlugs)
		if !ok {
			if skipReason != "" {
				skippedReasons = append(skippedReasons, skipReason)
			}
			continue
		}

		response, createdCount, err := s.createInventorySetFamily(ctx, candidate.Config, candidate.Slots, normalized.QueueImages, existingKeys)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if createdCount == 0 {
			skippedReasons = append(skippedReasons, fmt.Sprintf("%s: all target slots already exist", candidate.Config.Theme))
			continue
		}

		families = append(families, response)
		createdItemCount += createdCount
		usedThemeSlugs[slugifyInventorySeedTheme(candidate.Config.Theme)] = struct{}{}
	}

	ctx.JSON(http.StatusOK, inventorySetFamilyGenerationResponse{
		RequestedCount:     normalized.Count,
		CreatedFamilyCount: len(families),
		CreatedItemCount:   createdItemCount,
		SkippedFamilyCount: len(skippedReasons),
		Families:           families,
		SkippedReasons:     skippedReasons,
	})
}

func normalizeInventorySetFamilyGenerationRequest(
	request inventorySetFamilyGenerationRequest,
) (inventorySetFamilyGenerationRequest, error) {
	if request.Count <= 0 {
		request.Count = 6
	}
	if request.Count > 40 {
		return request, fmt.Errorf("count must be between 1 and 40")
	}

	if request.LevelMin <= 0 {
		request.LevelMin = 1
	}
	if request.LevelMax <= 0 {
		request.LevelMax = 100
	}
	if request.LevelMin > request.LevelMax {
		return request, fmt.Errorf("levelMin must be less than or equal to levelMax")
	}
	if request.LevelMin < 1 || request.LevelMax > 100 {
		return request, fmt.Errorf("level range must stay between 1 and 100")
	}

	request.RarityTiers = normalizeInventoryFamilyRarities(request.RarityTiers)
	request.Themes = normalizeUniqueStringList(request.Themes)
	request.MajorStats = normalizeInventoryFamilyStats(request.MajorStats)
	request.MinorStats = normalizeInventoryFamilyStats(request.MinorStats)
	request.Profiles = normalizeInventoryFamilyProfiles(request.Profiles)
	request.DamageAffinities = normalizeInventoryFamilyAffinities(request.DamageAffinities)
	request.ResistanceAffinities = normalizeInventoryFamilyAffinities(request.ResistanceAffinities)
	request.RequiredInternalTags = normalizeUniqueStringList(request.RequiredInternalTags)
	request.ForbiddenTags = normalizeUniqueStringList(request.ForbiddenTags)

	request.SlotScope = normalizeInventoryFamilySlotScope(request.SlotScope)
	if request.SlotScope == "" {
		return request, fmt.Errorf("slotScope must be one of: full_set, armor_only, jewelry_only, hand_items_only")
	}

	request.PowerBias = normalizeInventoryFamilyPowerBias(request.PowerBias)
	if request.PowerBias == "" {
		request.PowerBias = "balanced"
	}

	request.NamingStyle = normalizeInventoryFamilyNamingStyle(request.NamingStyle)
	if request.NamingStyle == "" {
		request.NamingStyle = "grounded"
	}

	if len(request.RarityTiers) == 0 {
		request.RarityTiers = []string{}
	}
	if len(request.Profiles) == 0 {
		request.Profiles = append([]string{}, defaultInventoryFamilyProfiles...)
	}
	if len(request.DamageAffinities) == 0 {
		request.DamageAffinities = append([]string{}, defaultInventoryFamilyAffinities...)
	}
	if len(request.ResistanceAffinities) == 0 {
		request.ResistanceAffinities = append([]string{}, defaultInventoryFamilyAffinities...)
	}
	if !request.QueueImages {
		request.QueueImages = false
	}

	return request, nil
}

func normalizeInventoryFamilyStats(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		stat := normalizeInventorySetStatKey(value)
		if stat == "" {
			continue
		}
		if _, exists := seen[stat]; exists {
			continue
		}
		seen[stat] = struct{}{}
		normalized = append(normalized, stat)
	}
	return normalized
}

func normalizeInventoryFamilyRarities(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		rarity := normalizeInventorySetRarityTier(value)
		if rarity == "" {
			continue
		}
		if _, exists := seen[rarity]; exists {
			continue
		}
		seen[rarity] = struct{}{}
		normalized = append(normalized, rarity)
	}
	return normalized
}

func normalizeInventoryFamilyAffinities(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		affinity := strings.ToLower(strings.TrimSpace(value))
		if affinity == "" || !slices.Contains(defaultInventoryFamilyAffinities, affinity) {
			continue
		}
		if _, exists := seen[affinity]; exists {
			continue
		}
		seen[affinity] = struct{}{}
		normalized = append(normalized, affinity)
	}
	return normalized
}

func normalizeUniqueStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func normalizeInventoryFamilyProfiles(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		profile := strings.ToLower(strings.TrimSpace(value))
		switch profile {
		case "martial", "tank", "caster", "skirmisher", "support", "hybrid":
		default:
			continue
		}
		if _, exists := seen[profile]; exists {
			continue
		}
		seen[profile] = struct{}{}
		normalized = append(normalized, profile)
	}
	return normalized
}

func normalizeInventoryFamilySlotScope(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "full_set", "full":
		return "full_set"
	case "armor_only", "armor":
		return "armor_only"
	case "jewelry_only", "jewelry":
		return "jewelry_only"
	case "hand_items_only", "hands":
		return "hand_items_only"
	default:
		return ""
	}
}

func normalizeInventoryFamilyPowerBias(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "balanced":
		return "balanced"
	case "offense", "defense", "utility":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func normalizeInventoryFamilyNamingStyle(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "grounded":
		return "grounded"
	case "heroic", "occult", "royal", "wild":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func buildInventorySetFamilyCandidate(
	request inventorySetFamilyGenerationRequest,
	rng *rand.Rand,
	existingThemeSlugs map[string]struct{},
	usedThemeSlugs map[string]struct{},
) (inventorySetFamilyCandidate, string, bool) {
	profile := pickRandomString(rng, request.Profiles, "martial")
	majorStat, minorStat := resolveInventoryFamilyStatPair(rng, request.MajorStats, request.MinorStats, profile, request.PowerBias)
	if majorStat == "" || minorStat == "" || majorStat == minorStat {
		return inventorySetFamilyCandidate{}, "unable to resolve a major/minor stat pair", false
	}

	targetLevel := request.LevelMin
	if request.LevelMax > request.LevelMin {
		targetLevel = rng.Intn(request.LevelMax-request.LevelMin+1) + request.LevelMin
	}
	rarityTier := ""
	if len(request.RarityTiers) > 0 {
		rarityTier = pickRandomString(rng, request.RarityTiers, inventorySetRarityForTargetLevel(targetLevel))
	} else {
		rarityTier = inventorySetRarityForTargetLevel(targetLevel)
	}

	primaryDamage := pickAffinityWithBias(rng, request.DamageAffinities, profile, request.PowerBias, "")
	secondaryDamage := ""
	if request.AllowHybridAffinities {
		secondaryDamage = pickAffinityWithBias(rng, request.DamageAffinities, profile, request.PowerBias, primaryDamage)
	}
	primaryResistance := pickAffinityWithBias(rng, request.ResistanceAffinities, profile, "defense", "")
	secondaryResistance := ""
	if request.AllowHybridAffinities {
		secondaryResistance = pickAffinityWithBias(rng, request.ResistanceAffinities, profile, "defense", primaryResistance)
	}

	theme := buildInventoryFamilyTheme(rng, request.Themes, request.NamingStyle, profile, majorStat, primaryDamage, targetLevel)
	themeSlug := slugifyInventorySeedTheme(theme)
	if themeSlug == "" {
		return inventorySetFamilyCandidate{}, "failed to build a non-empty set theme", false
	}
	if request.AvoidExistingThemeOverlap {
		if _, exists := existingThemeSlugs[themeSlug]; exists {
			return inventorySetFamilyCandidate{}, fmt.Sprintf("%s overlaps with an existing theme", theme), false
		}
		if _, exists := usedThemeSlugs[themeSlug]; exists {
			return inventorySetFamilyCandidate{}, fmt.Sprintf("%s duplicates a generated theme in this batch", theme), false
		}
	}

	internalTags := append([]string{}, request.RequiredInternalTags...)
	internalTags = append(internalTags, profile, request.PowerBias, majorStat, minorStat)
	if primaryDamage != "" {
		internalTags = append(internalTags, primaryDamage)
	}
	if primaryResistance != "" {
		internalTags = append(internalTags, primaryResistance+"_ward")
	}
	internalTags = normalizeUniqueStringList(internalTags)
	for _, forbidden := range request.ForbiddenTags {
		if slices.Contains(internalTags, forbidden) || strings.Contains(themeSlug, forbidden) {
			return inventorySetFamilyCandidate{}, fmt.Sprintf("%s hit forbidden tag %s", theme, forbidden), false
		}
	}

	return inventorySetFamilyCandidate{
		Config: inventorySeedSetConfig{
			Theme:                        theme,
			TargetLevel:                  targetLevel,
			RarityTier:                   rarityTier,
			MajorStat:                    majorStat,
			MinorStat:                    minorStat,
			InternalTags:                 internalTags,
			DamageBonusAffinity:          primaryDamage,
			SecondaryDamageBonusAffinity: secondaryDamage,
			ResistanceAffinity:           primaryResistance,
			SecondaryResistanceAffinity:  secondaryResistance,
		},
		Slots: inventoryFamilySlotsForScope(request.SlotScope),
	}, "", true
}

func resolveInventoryFamilyStatPair(
	rng *rand.Rand,
	majorStats []string,
	minorStats []string,
	profile string,
	powerBias string,
) (string, string) {
	if len(majorStats) > 0 {
		major := pickRandomString(rng, majorStats, majorStats[0])
		minorPool := make([]string, 0, len(minorStats))
		for _, candidate := range minorStats {
			if candidate != major {
				minorPool = append(minorPool, candidate)
			}
		}
		if len(minorPool) == 0 {
			minorPool = inventoryFamilyFallbackMinorStats(profile, powerBias, major)
		}
		if len(minorPool) == 0 {
			return "", ""
		}
		return major, pickRandomString(rng, minorPool, minorPool[0])
	}

	pairs := inventoryFamilyStatPairs(profile, powerBias)
	if len(pairs) == 0 {
		return "", ""
	}
	pair := pairs[rng.Intn(len(pairs))]
	if len(minorStats) > 0 && !slices.Contains(minorStats, pair[1]) {
		return resolveInventoryFamilyStatPair(rng, []string{pair[0]}, minorStats, profile, powerBias)
	}
	return pair[0], pair[1]
}

func inventoryFamilyFallbackMinorStats(profile string, powerBias string, major string) []string {
	pairs := inventoryFamilyStatPairs(profile, powerBias)
	minors := make([]string, 0, len(pairs))
	seen := map[string]struct{}{}
	for _, pair := range pairs {
		if pair[0] != major || pair[1] == major {
			continue
		}
		if _, exists := seen[pair[1]]; exists {
			continue
		}
		seen[pair[1]] = struct{}{}
		minors = append(minors, pair[1])
	}
	return minors
}

func inventoryFamilyStatPairs(profile string, powerBias string) [][2]string {
	switch profile {
	case "tank":
		if powerBias == "offense" {
			return [][2]string{{"constitution", "strength"}, {"strength", "constitution"}}
		}
		return [][2]string{{"constitution", "strength"}, {"constitution", "wisdom"}, {"strength", "constitution"}}
	case "caster":
		if powerBias == "utility" {
			return [][2]string{{"wisdom", "intelligence"}, {"intelligence", "wisdom"}, {"wisdom", "charisma"}}
		}
		return [][2]string{{"intelligence", "wisdom"}, {"intelligence", "charisma"}, {"wisdom", "intelligence"}}
	case "skirmisher":
		return [][2]string{{"dexterity", "strength"}, {"dexterity", "charisma"}, {"dexterity", "wisdom"}}
	case "support":
		return [][2]string{{"wisdom", "charisma"}, {"charisma", "wisdom"}, {"wisdom", "constitution"}}
	case "hybrid":
		return [][2]string{{"strength", "intelligence"}, {"dexterity", "wisdom"}, {"charisma", "strength"}, {"intelligence", "constitution"}}
	default:
		switch powerBias {
		case "defense":
			return [][2]string{{"constitution", "strength"}, {"constitution", "wisdom"}, {"strength", "constitution"}}
		case "utility":
			return [][2]string{{"dexterity", "charisma"}, {"wisdom", "charisma"}, {"intelligence", "dexterity"}}
		default:
			return [][2]string{{"strength", "dexterity"}, {"dexterity", "strength"}, {"strength", "constitution"}}
		}
	}
}

func pickAffinityWithBias(
	rng *rand.Rand,
	pool []string,
	profile string,
	powerBias string,
	exclude string,
) string {
	candidates := make([]string, 0, len(pool))
	for _, value := range pool {
		if value != exclude {
			candidates = append(candidates, value)
		}
	}
	if len(candidates) == 0 {
		return exclude
	}

	preferred := []string{}
	switch profile {
	case "tank":
		preferred = []string{"physical", "bludgeoning", "holy", "fire"}
	case "caster":
		preferred = []string{"arcane", "fire", "ice", "lightning", "shadow", "holy"}
	case "skirmisher":
		preferred = []string{"piercing", "slashing", "poison", "lightning"}
	case "support":
		preferred = []string{"holy", "arcane", "ice", "shadow"}
	case "hybrid":
		preferred = []string{"arcane", "physical", "fire", "shadow", "holy"}
	default:
		preferred = []string{"physical", "slashing", "piercing", "fire"}
	}
	if powerBias == "defense" {
		preferred = append([]string{"holy", "physical", "ice"}, preferred...)
	}
	if powerBias == "utility" {
		preferred = append([]string{"arcane", "shadow", "poison"}, preferred...)
	}

	for _, desired := range preferred {
		if slices.Contains(candidates, desired) {
			return desired
		}
	}
	return candidates[rng.Intn(len(candidates))]
}

func buildInventoryFamilyTheme(
	rng *rand.Rand,
	themes []string,
	namingStyle string,
	profile string,
	majorStat string,
	affinity string,
	level int,
) string {
	base := ""
	if len(themes) > 0 {
		base = titleCaseInventoryTheme(themes[rng.Intn(len(themes))])
	} else {
		base = fmt.Sprintf("%s %s", inventoryAffinityThemeWord(affinity), inventoryFamilyArchetypeWord(profile, majorStat))
	}
	base = strings.TrimSpace(base)
	if base == "" {
		base = "Wanderer Regalia"
	}

	suffix := inventoryFamilyThemeSuffix(namingStyle, level)
	if suffix == "" || strings.HasSuffix(strings.ToLower(base), strings.ToLower(suffix)) {
		return base
	}
	return fmt.Sprintf("%s %s", base, suffix)
}

func inventoryAffinityThemeWord(affinity string) string {
	switch affinity {
	case "physical":
		return "Iron"
	case "piercing":
		return "Needle"
	case "slashing":
		return "Edge"
	case "bludgeoning":
		return "Stone"
	case "fire":
		return "Ember"
	case "ice":
		return "Frost"
	case "lightning":
		return "Storm"
	case "poison":
		return "Thorn"
	case "arcane":
		return "Rune"
	case "holy":
		return "Dawn"
	case "shadow":
		return "Gloam"
	default:
		return "Wayfarer"
	}
}

func inventoryFamilyArchetypeWord(profile string, majorStat string) string {
	switch profile {
	case "tank":
		return "Bastion"
	case "caster":
		return "Weaver"
	case "skirmisher":
		return "Strider"
	case "support":
		return "Cantor"
	case "hybrid":
		return "Magister"
	default:
		switch majorStat {
		case "dexterity":
			return "Runner"
		case "constitution":
			return "Bulwark"
		case "intelligence":
			return "Savant"
		case "wisdom":
			return "Oracle"
		case "charisma":
			return "Regent"
		default:
			return "Vanguard"
		}
	}
}

func inventoryFamilyThemeSuffix(namingStyle string, level int) string {
	switch namingStyle {
	case "heroic":
		if level >= 80 {
			return "Ascendant"
		}
		if level >= 45 {
			return "Paragon"
		}
		return "Vanguard"
	case "occult":
		if level >= 70 {
			return "Reliquary"
		}
		return "Hex"
	case "royal":
		if level >= 70 {
			return "Regalia"
		}
		return "Court"
	case "wild":
		if level >= 70 {
			return "Covenant"
		}
		return "Harrier"
	default:
		if level >= 75 {
			return "Kit"
		}
		return ""
	}
}

func titleCaseInventoryTheme(value string) string {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) == 0 {
		return ""
	}
	for idx, part := range parts {
		if part == "" {
			continue
		}
		parts[idx] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, " ")
}

func inventoryFamilySlotsForScope(scope string) []string {
	switch scope {
	case "armor_only":
		return []string{
			string(models.EquipmentSlotHat),
			string(models.EquipmentSlotChest),
			string(models.EquipmentSlotLegs),
			string(models.EquipmentSlotShoes),
			string(models.EquipmentSlotGloves),
		}
	case "jewelry_only":
		return []string{
			string(models.EquipmentSlotNecklace),
			string(models.EquipmentSlotRing),
		}
	case "hand_items_only":
		return []string{
			string(models.EquipmentSlotDominantHand),
			string(models.EquipmentSlotOffHand),
		}
	default:
		return inventorySetAllEquippableSlots()
	}
}

func (s *server) createInventorySetFamily(
	ctx *gin.Context,
	config inventorySeedSetConfig,
	targetSlots []string,
	queueImages bool,
	existingKeys map[string]struct{},
) (inventorySetGenerationResponse, int, error) {
	requests := buildInventorySeedSetRequests(config)
	if len(targetSlots) > 0 {
		targetSet := map[string]struct{}{}
		for _, slot := range targetSlots {
			targetSet[slot] = struct{}{}
		}
		filtered := make([]inventorySeedPackRequest, 0, len(requests))
		for _, request := range requests {
			slot := ""
			if request.Request.EquipSlot != nil {
				slot = strings.TrimSpace(*request.Request.EquipSlot)
			}
			if _, ok := targetSet[slot]; ok {
				filtered = append(filtered, request)
			}
		}
		requests = filtered
	}

	response := inventorySetGenerationResponse{
		SetTheme:     config.Theme,
		TargetLevel:  config.TargetLevel,
		MajorStat:    config.MajorStat,
		MinorStat:    config.MinorStat,
		RarityTier:   config.RarityTier,
		CreatedItems: make([]models.InventoryItem, 0, len(requests)),
		SkippedSlots: make([]string, 0),
	}

	createdCount := 0
	for _, seed := range requests {
		slot := ""
		if seed.Request.EquipSlot != nil {
			slot = strings.TrimSpace(*seed.Request.EquipSlot)
		}
		itemKey := inventorySetItemKey(slot, seed.Request.Name)
		if _, exists := existingKeys[itemKey]; exists {
			response.SkippedSlots = append(response.SkippedSlots, slot)
			continue
		}

		normalized, err := s.normalizeInventoryItemUpsertRequest(ctx, seed.Request, nil)
		if err != nil {
			return response, createdCount, fmt.Errorf("failed to normalize generated set item %q: %w", seed.Request.Name, err)
		}
		if !queueImages {
			normalized.ImageGenerationStatus = models.InventoryImageGenerationStatusNone
			normalized.ImageGenerationError = nil
		} else {
			normalized.ImageGenerationStatus = models.InventoryImageGenerationStatusQueued
		}

		if err := s.dbClient.InventoryItem().CreateInventoryItem(ctx, normalized); err != nil {
			return response, createdCount, fmt.Errorf("failed to create generated set item %q: %w", normalized.Name, err)
		}
		if queueImages {
			if err := s.enqueueInventoryItemImageGeneration(ctx, normalized.ID, normalized.Name, normalized.FlavorText, normalized.RarityTier); err != nil {
				response.EnqueueWarnings = append(response.EnqueueWarnings, fmt.Sprintf("%s: %s", normalized.Name, err.Error()))
			}
		}

		response.CreatedItems = append(response.CreatedItems, *normalized)
		existingKeys[itemKey] = struct{}{}
		createdCount++
	}

	response.Message = fmt.Sprintf("created %d item(s) for %s", createdCount, config.Theme)
	return response, createdCount, nil
}

func pickRandomString(rng *rand.Rand, values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return values[rng.Intn(len(values))]
}
