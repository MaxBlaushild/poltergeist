package processors

import (
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
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
