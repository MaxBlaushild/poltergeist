package server

import (
	"strings"
	"testing"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/google/uuid"
)

func TestAbilityTomePricingTierForLevel(t *testing.T) {
	cases := []struct {
		level    int
		rarity   string
		buyPrice int
	}{
		{level: 10, rarity: "Common", buyPrice: 150},
		{level: 25, rarity: "Uncommon", buyPrice: 350},
		{level: 50, rarity: "Epic", buyPrice: 900},
		{level: 70, rarity: "Mythic", buyPrice: 1800},
	}

	for _, tc := range cases {
		tier := abilityTomePricingTierForLevel(tc.level)
		if tier.rarity != tc.rarity || tier.buyPrice != tc.buyPrice {
			t.Fatalf(
				"expected level %d tome tier to be %s/%d, got %s/%d",
				tc.level,
				tc.rarity,
				tc.buyPrice,
				tier.rarity,
				tier.buyPrice,
			)
		}
	}
}

func TestBuildAbilityTomeItemForSpell(t *testing.T) {
	ability := &models.Spell{
		ID:            uuid.New(),
		Name:          "Ember Lance",
		Description:   "Launch a focused flame spike that pierces a single target.",
		AbilityType:   models.SpellAbilityTypeSpell,
		AbilityLevel:  25,
		SchoolOfMagic: "Pyromancy",
	}

	item := buildAbilityTomeItem(ability)
	if item.Name != "Tome of Ember Lance" {
		t.Fatalf("expected tome name to match ability name, got %q", item.Name)
	}
	if item.ItemLevel != 25 {
		t.Fatalf("expected tome item level to match ability level, got %d", item.ItemLevel)
	}
	if item.BuyPrice == nil || *item.BuyPrice != 350 {
		t.Fatalf("expected tome buy price to use the uncommon tier, got %v", item.BuyPrice)
	}
	if item.RarityTier != "Uncommon" {
		t.Fatalf("expected uncommon rarity for level 25 tome, got %q", item.RarityTier)
	}
	if len(item.ConsumeSpellIDs) != 1 || item.ConsumeSpellIDs[0] != ability.ID.String() {
		t.Fatalf("expected tome to grant the ability on consume, got %v", item.ConsumeSpellIDs)
	}
	description := strings.ToLower(item.FlavorText)
	if !strings.Contains(description, "grimoire") {
		t.Fatalf("expected spell tome description to look like a book, got %q", item.FlavorText)
	}
	if !strings.Contains(description, "\"ember lance\"") {
		t.Fatalf("expected tome description to play on the spell name, got %q", item.FlavorText)
	}
	if !strings.Contains(description, "focused flame") {
		t.Fatalf("expected tome description to echo the spell description, got %q", item.FlavorText)
	}
}

func TestBuildAbilityTomeItemForTechnique(t *testing.T) {
	ability := &models.Spell{
		ID:            uuid.New(),
		Name:          "Iron Counter",
		Description:   "Time a precise counterstrike immediately after blocking an attack.",
		AbilityType:   models.SpellAbilityTypeTechnique,
		AbilityLevel:  70,
		SchoolOfMagic: "Martial",
	}

	item := buildAbilityTomeItem(ability)
	if item.RarityTier != "Mythic" {
		t.Fatalf("expected mythic rarity for level 70 technique tome, got %q", item.RarityTier)
	}
	if item.BuyPrice == nil || *item.BuyPrice != 1800 {
		t.Fatalf("expected tome buy price to use the mythic tier, got %v", item.BuyPrice)
	}
	if len(item.InternalTags) == 0 || item.InternalTags[0] != "tome" {
		t.Fatalf("expected tome tags to be populated, got %v", item.InternalTags)
	}
	description := strings.ToLower(item.FlavorText)
	if !strings.Contains(description, "manual") {
		t.Fatalf("expected technique tome description to read like a manual, got %q", item.FlavorText)
	}
	if !strings.Contains(description, "\"iron counter\"") {
		t.Fatalf("expected technique tome description to play on the ability name, got %q", item.FlavorText)
	}
	if !strings.Contains(description, "precise counterstrike") {
		t.Fatalf("expected technique tome description to echo the ability description, got %q", item.FlavorText)
	}
}

func TestBuildLegacyAbilityTomeName(t *testing.T) {
	ability := &models.Spell{Name: "Ember Lance."}
	if got := buildAbilityTomeName(ability); got != "Tome of Ember Lance" {
		t.Fatalf("expected trimmed tome name without trailing punctuation, got %q", got)
	}
	if got := buildLegacyAbilityTomeName(ability); got != "Tome of Ember Lance." {
		t.Fatalf("expected legacy tome name with trailing period, got %q", got)
	}
}
