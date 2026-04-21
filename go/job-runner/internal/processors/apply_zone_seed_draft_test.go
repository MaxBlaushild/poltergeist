package processors

import (
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
