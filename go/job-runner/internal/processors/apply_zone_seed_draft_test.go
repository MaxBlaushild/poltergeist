package processors

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestZoneSeedEncounterMemberCountRange(t *testing.T) {
	for i := 0; i < 500; i++ {
		count := zoneSeedEncounterMemberCount(models.MonsterEncounterTypeMonster)
		if count < 1 || count > 3 {
			t.Fatalf("expected encounter member count between 1 and 3, got %d", count)
		}
	}
}

func TestZoneSeedEncounterMemberCountForBossAndRaid(t *testing.T) {
	if got := zoneSeedEncounterMemberCount(models.MonsterEncounterTypeBoss); got != 1 {
		t.Fatalf("expected boss encounter member count to be 1, got %d", got)
	}
	if got := zoneSeedEncounterMemberCount(models.MonsterEncounterTypeRaid); got != 5 {
		t.Fatalf("expected raid encounter member count to be 5, got %d", got)
	}
}

func TestPreferredZoneSeedTemplatesForEncounterType(t *testing.T) {
	templates := []models.MonsterTemplate{
		{Name: "Field Beast", MonsterType: models.MonsterTemplateTypeMonster},
		{Name: "Boss Beast", MonsterType: models.MonsterTemplateTypeBoss},
		{Name: "Raid Beast", MonsterType: models.MonsterTemplateTypeRaid},
	}

	bossTemplates := preferredZoneSeedTemplatesForEncounterType(templates, models.MonsterEncounterTypeBoss)
	if len(bossTemplates) != 1 || bossTemplates[0].MonsterType != models.MonsterTemplateTypeBoss {
		t.Fatalf("expected boss encounter to prefer boss templates, got %+v", bossTemplates)
	}

	raidTemplates := preferredZoneSeedTemplatesForEncounterType(templates, models.MonsterEncounterTypeRaid)
	if len(raidTemplates) != 1 || raidTemplates[0].MonsterType != models.MonsterTemplateTypeRaid {
		t.Fatalf("expected raid encounter to prefer raid templates, got %+v", raidTemplates)
	}

	standardTemplates := preferredZoneSeedTemplatesForEncounterType(templates, models.MonsterEncounterTypeMonster)
	if len(standardTemplates) != 1 || standardTemplates[0].MonsterType != models.MonsterTemplateTypeMonster {
		t.Fatalf("expected standard encounter to prefer monster templates, got %+v", standardTemplates)
	}
}

func TestFilterZoneSeedMonsterTemplatesByZoneKind(t *testing.T) {
	templates := []models.MonsterTemplate{
		{Name: "Forest Wolf", ZoneKind: "forest"},
		{Name: "Swamp Hag", ZoneKind: "swamp"},
		{Name: "Neutral Ooze", ZoneKind: ""},
	}

	filtered := filterZoneSeedMonsterTemplatesByZoneKind(templates, "forest")
	if len(filtered) != 1 || filtered[0].Name != "Forest Wolf" {
		t.Fatalf("expected only forest templates, got %+v", filtered)
	}
}

func TestFilterZoneSeedMonsterTemplatesByZoneKindLeavesLegacyZonesBroad(t *testing.T) {
	templates := []models.MonsterTemplate{
		{Name: "Forest Wolf", ZoneKind: "forest"},
		{Name: "Swamp Hag", ZoneKind: "swamp"},
	}

	filtered := filterZoneSeedMonsterTemplatesByZoneKind(templates, "")
	if len(filtered) != len(templates) {
		t.Fatalf("expected blank zone kind to keep all templates, got %+v", filtered)
	}
}

func TestFilterZoneSeedQuestArchetypesByZoneKind(t *testing.T) {
	forestQuest := &models.QuestArchetype{Name: "Forest Watch", ZoneKind: "forest", Category: models.QuestCategorySide}
	swampQuest := &models.QuestArchetype{Name: "Swamp Patrol", ZoneKind: "swamp", Category: models.QuestCategorySide}
	mainStoryQuest := &models.QuestArchetype{Name: "Forest Chapter", ZoneKind: "forest", Category: models.QuestCategoryMainStory}

	filtered := filterZoneSeedQuestArchetypes(
		[]*models.QuestArchetype{forestQuest, swampQuest, mainStoryQuest},
		"forest",
	)
	if len(filtered) != 1 || filtered[0].Name != "Forest Watch" {
		t.Fatalf("expected only forest side quests, got %+v", filtered)
	}
}

func TestFilterZoneSeedQuestArchetypesLeavesBlankZoneKindBroadForSideQuests(t *testing.T) {
	forestQuest := &models.QuestArchetype{Name: "Forest Watch", ZoneKind: "forest", Category: models.QuestCategorySide}
	swampQuest := &models.QuestArchetype{Name: "Swamp Patrol", ZoneKind: "swamp", Category: models.QuestCategorySide}
	mainStoryQuest := &models.QuestArchetype{Name: "Forest Chapter", ZoneKind: "forest", Category: models.QuestCategoryMainStory}

	filtered := filterZoneSeedQuestArchetypes(
		[]*models.QuestArchetype{forestQuest, swampQuest, mainStoryQuest},
		"",
	)
	if len(filtered) != 2 {
		t.Fatalf("expected all side quests for blank zone kind, got %+v", filtered)
	}
}

func TestFilterZoneSeedExpositionTemplatesByZoneKindPrefersMatchesAndKeepsGenericFallback(t *testing.T) {
	templates := []models.ExpositionTemplate{
		{Title: "Forest Warning", ZoneKind: "forest"},
		{Title: "Swamp Murmur", ZoneKind: "swamp"},
		{Title: "Generic Omen", ZoneKind: ""},
	}

	filtered := filterZoneSeedExpositionTemplates(templates, "forest")
	if len(filtered) != 2 {
		t.Fatalf("expected matched and generic exposition templates, got %+v", filtered)
	}
	if filtered[0].Title != "Forest Warning" || filtered[1].Title != "Generic Omen" {
		t.Fatalf("expected forest template first with generic fallback, got %+v", filtered)
	}
}

func TestFilterZoneSeedExpositionTemplatesByZoneKindFallsBackToGenericWhenNoMatchExists(t *testing.T) {
	templates := []models.ExpositionTemplate{
		{Title: "Swamp Murmur", ZoneKind: "swamp"},
		{Title: "Generic Omen", ZoneKind: ""},
	}

	filtered := filterZoneSeedExpositionTemplates(templates, "forest")
	if len(filtered) != 1 || filtered[0].Title != "Generic Omen" {
		t.Fatalf("expected generic exposition fallback, got %+v", filtered)
	}
}

func TestMatchZoneSeedPOILocationArchetypePrefersMostSpecificIncludedTypes(t *testing.T) {
	archetypes := []*models.LocationArchetype{
		{
			Name:          "Coffeehouse",
			IncludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeCafe},
		},
		{
			Name:          "Bakery Cafe",
			IncludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeCafe, googlemaps.TypeBakery},
		},
		{
			Name:          "Park",
			IncludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypePark},
		},
	}

	match := matchZoneSeedPOILocationArchetype(archetypes, &models.ZoneSeedPointOfInterestDraft{
		Types: []string{"cafe", "bakery", "food_store"},
	})
	if match == nil || match.Name != "Bakery Cafe" {
		t.Fatalf("expected Bakery Cafe archetype, got %+v", match)
	}
}

func TestMatchZoneSeedPOILocationArchetypeSkipsExcludedTypes(t *testing.T) {
	archetypes := []*models.LocationArchetype{
		{
			Name:          "Coffeehouse",
			IncludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeCafe},
			ExcludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeMuseum},
		},
		{
			Name:          "Museum",
			IncludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeMuseum},
		},
	}

	match := matchZoneSeedPOILocationArchetype(archetypes, &models.ZoneSeedPointOfInterestDraft{
		Types: []string{"cafe", "museum"},
	})
	if match == nil || match.Name != "Museum" {
		t.Fatalf("expected Museum archetype, got %+v", match)
	}
}

func TestFallbackZoneSeedReusableChallengeTemplateAvoidsSpecificPOIName(t *testing.T) {
	draft := &models.ZoneSeedPointOfInterestDraft{
		Name:    "Moonwake Coffee",
		Address: "123 Lantern St",
		Types:   []string{"cafe", "coffee_shop"},
	}

	spec := fallbackZoneSeedReusableChallengeTemplate(&models.LocationArchetype{Name: "Coffeehouse"}, draft)
	lowerQuestion := strings.ToLower(spec.Question)
	lowerDescription := strings.ToLower(spec.Description)
	if strings.Contains(lowerQuestion, strings.ToLower(draft.Name)) {
		t.Fatalf("expected reusable challenge question to omit POI name, got %q", spec.Question)
	}
	if strings.Contains(lowerDescription, strings.ToLower(draft.Address)) {
		t.Fatalf("expected reusable challenge description to omit POI address, got %q", spec.Description)
	}
	if !strings.Contains(lowerQuestion, "this coffeehouse") {
		t.Fatalf("expected reusable challenge question to stay generic, got %q", spec.Question)
	}
}

func TestSelectZoneSeedQuestGiverCandidatePrefersRootLocationMatchedDraftCharacter(t *testing.T) {
	bookshopPOIID := uuid.New()
	cafePOIID := uuid.New()
	bookshopPlaceID := "bookshop-place"
	cafePlaceID := "cafe-place"

	archetype := &models.QuestArchetype{
		Root: models.QuestArchetypeNode{
			LocationArchetype: &models.LocationArchetype{
				Name:          "Bookshop",
				IncludedTypes: googlemaps.PlaceTypeSlice{googlemaps.TypeBookStore},
			},
		},
	}
	bookshopCharacter := &models.Character{
		ID:                uuid.New(),
		Name:              "Archivist Vale",
		PointOfInterestID: &bookshopPOIID,
	}
	cafeCharacter := &models.Character{
		ID:                uuid.New(),
		Name:              "Bramble",
		PointOfInterestID: &cafePOIID,
	}

	context := &zoneSeedQuestGiverContext{
		draftCharacters: []*models.Character{cafeCharacter, bookshopCharacter},
		pointOfInterestByID: map[uuid.UUID]*models.PointOfInterest{
			bookshopPOIID: {
				ID:                  bookshopPOIID,
				GoogleMapsPlaceID:   &bookshopPlaceID,
				GoogleMapsPlaceName: nil,
			},
			cafePOIID: {
				ID:                  cafePOIID,
				GoogleMapsPlaceID:   &cafePlaceID,
				GoogleMapsPlaceName: nil,
			},
		},
		poiDraftByPlaceID: map[string]*models.ZoneSeedPointOfInterestDraft{
			bookshopPlaceID: {
				PlaceID: bookshopPlaceID,
				Types:   []string{"book_store"},
			},
			cafePlaceID: {
				PlaceID: cafePlaceID,
				Types:   []string{"cafe"},
			},
		},
		assignments: map[uuid.UUID]int{},
	}

	selected := selectZoneSeedQuestGiverCandidate(archetype, nil, context, false)
	if selected == nil || selected.ID != bookshopCharacter.ID {
		t.Fatalf("expected bookshop-aligned draft character, got %+v", selected)
	}
}

func TestSelectZoneSeedQuestGiverCandidateBalancesAssignments(t *testing.T) {
	first := &models.Character{ID: uuid.New(), Name: "First Draft"}
	second := &models.Character{ID: uuid.New(), Name: "Second Draft"}

	context := &zoneSeedQuestGiverContext{
		draftCharacters: []*models.Character{first, second},
		assignments: map[uuid.UUID]int{
			first.ID:  2,
			second.ID: 0,
		},
	}

	selected := selectZoneSeedQuestGiverCandidate(&models.QuestArchetype{}, nil, context, false)
	if selected == nil || selected.ID != second.ID {
		t.Fatalf("expected lower-assignment draft character, got %+v", selected)
	}
}

func TestSelectZoneSeedQuestGiverCandidateRequiresTagMatchWhenRequested(t *testing.T) {
	matching := &models.Character{
		ID:           uuid.New(),
		Name:         "Watch Captain",
		InternalTags: models.StringArray{"guard", "watch"},
	}
	nonMatching := &models.Character{
		ID:           uuid.New(),
		Name:         "Tea Seller",
		InternalTags: models.StringArray{"merchant"},
	}

	context := &zoneSeedQuestGiverContext{
		zoneCharacters: []*models.Character{nonMatching, matching},
		assignments:    map[uuid.UUID]int{},
	}

	selected := selectZoneSeedQuestGiverCandidate(
		&models.QuestArchetype{},
		map[string]struct{}{"guard": {}},
		context,
		true,
	)
	if selected == nil || selected.ID != matching.ID {
		t.Fatalf("expected tag-matching quest giver, got %+v", selected)
	}
}

func TestZoneSeedBuildResourcePoolsFiltersToEligibleTypes(t *testing.T) {
	typeIDOne := uuid.New()
	typeIDTwo := uuid.New()
	typeIDUnused := uuid.New()

	resourceTypes := []models.ResourceType{
		{ID: typeIDOne, Name: "Ore"},
		{ID: typeIDTwo, Name: "Herb"},
		{ID: typeIDUnused, Name: "Crystal"},
	}
	items := []models.InventoryItem{
		{ID: 101, Name: "Iron Ore", ResourceTypeID: &typeIDOne},
		{ID: 102, Name: "Copper Ore", ResourceTypeID: &typeIDOne},
		{ID: 201, Name: "Moonleaf", ResourceTypeID: &typeIDTwo},
		{ID: 301, Name: "Loose Trinket"},
	}

	pools := zoneSeedBuildResourcePools(resourceTypes, items)
	if len(pools) != 2 {
		t.Fatalf("expected 2 eligible resource pools, got %d", len(pools))
	}
	if pools[0].resourceType.ID != typeIDOne || len(pools[0].inventoryItems) != 2 {
		t.Fatalf("expected ore pool with 2 items, got %+v", pools[0])
	}
	if pools[1].resourceType.ID != typeIDTwo || len(pools[1].inventoryItems) != 1 {
		t.Fatalf("expected herb pool with 1 item, got %+v", pools[1])
	}
}

func TestZoneSeedBuildResourcePoolsBySlugGroupsEligibleTypes(t *testing.T) {
	herbalismID := uuid.New()
	miningID := uuid.New()

	resourceTypes := []models.ResourceType{
		{ID: herbalismID, Name: "Herbalism", Slug: "herbalism"},
		{ID: miningID, Name: "Mining", Slug: "mining"},
	}
	items := []models.InventoryItem{
		{ID: 101, Name: "Moonleaf", ResourceTypeID: &herbalismID},
		{ID: 201, Name: "Iron Ore", ResourceTypeID: &miningID},
	}

	poolsBySlug := zoneSeedBuildResourcePoolsBySlug(resourceTypes, items)
	if len(poolsBySlug[zoneSeedHerbalismResourceTypeSlug]) != 1 {
		t.Fatalf("expected one herbalism pool, got %+v", poolsBySlug[zoneSeedHerbalismResourceTypeSlug])
	}
	if len(poolsBySlug[zoneSeedMiningResourceTypeSlug]) != 1 {
		t.Fatalf("expected one mining pool, got %+v", poolsBySlug[zoneSeedMiningResourceTypeSlug])
	}
	if poolsBySlug[zoneSeedHerbalismResourceTypeSlug][0].resourceType.ID != herbalismID {
		t.Fatalf("expected herbalism pool to map to herbalism resource type, got %+v", poolsBySlug[zoneSeedHerbalismResourceTypeSlug][0])
	}
	if poolsBySlug[zoneSeedMiningResourceTypeSlug][0].resourceType.ID != miningID {
		t.Fatalf("expected mining pool to map to mining resource type, got %+v", poolsBySlug[zoneSeedMiningResourceTypeSlug][0])
	}
}

func TestEnsureRequiredSelectionsTopsOffFromCandidatePoolWithoutTags(t *testing.T) {
	base := []googlemaps.Place{
		{ID: "one", DisplayName: googlemaps.LocalizedText{Text: "One"}},
	}
	candidatePool := []googlemaps.Place{
		{ID: "one", DisplayName: googlemaps.LocalizedText{Text: "One"}},
		{ID: "two", DisplayName: googlemaps.LocalizedText{Text: "Two"}},
		{ID: "three", DisplayName: googlemaps.LocalizedText{Text: "Three"}},
	}

	results := ensureRequiredSelections(nil, map[string]googlemaps.Place{}, base, candidatePool, 3)
	if len(results) != 3 {
		t.Fatalf("expected 3 places after top-off, got %d", len(results))
	}
	if results[0].ID != "one" || results[1].ID != "two" || results[2].ID != "three" {
		t.Fatalf("unexpected top-off ordering: %+v", results)
	}
}

func TestPlaceSearchAttemptLimitScalesWithDesiredCount(t *testing.T) {
	if got := placeSearchAttemptLimit(4, false); got != 6 {
		t.Fatalf("expected minimum attempt floor of 6, got %d", got)
	}
	if got := placeSearchAttemptLimit(15, false); got != 15 {
		t.Fatalf("expected desired count to increase attempts, got %d", got)
	}
	if got := placeSearchAttemptLimit(30, true); got != 18 {
		t.Fatalf("expected attempts to clamp at 18, got %d", got)
	}
}
