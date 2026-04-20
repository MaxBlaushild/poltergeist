package server

import (
	"math"
	"testing"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/liveness"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
)

func TestMonsterBattleInviteAnchorPrefersEncounterCoordinates(t *testing.T) {
	monster := &models.Monster{Latitude: 40.7128, Longitude: -74.0060}
	encounter := &models.MonsterEncounter{Latitude: 40.7580, Longitude: -73.9855}

	lat, lng, ok := monsterBattleInviteAnchor(monster, encounter)
	if !ok {
		t.Fatal("expected encounter-backed battle to provide an invite anchor")
	}
	if lat != encounter.Latitude || lng != encounter.Longitude {
		t.Fatalf(
			"expected invite anchor to use encounter coordinates (%f, %f), got (%f, %f)",
			encounter.Latitude,
			encounter.Longitude,
			lat,
			lng,
		)
	}
}

func TestMonsterBattleInviteAnchorFallsBackToMonsterCoordinates(t *testing.T) {
	monster := &models.Monster{Latitude: 34.0522, Longitude: -118.2437}
	encounter := &models.MonsterEncounter{
		Latitude:  math.Inf(1),
		Longitude: -118.2437,
	}

	lat, lng, ok := monsterBattleInviteAnchor(monster, encounter)
	if !ok {
		t.Fatal("expected monster coordinates to remain a usable fallback")
	}
	if lat != monster.Latitude || lng != monster.Longitude {
		t.Fatalf(
			"expected invite anchor to fall back to monster coordinates (%f, %f), got (%f, %f)",
			monster.Latitude,
			monster.Longitude,
			lat,
			lng,
		)
	}
}

func TestMonsterBattleInviteAnchorReturnsUnavailableWithoutValidCoordinates(t *testing.T) {
	monster := &models.Monster{Latitude: math.NaN(), Longitude: math.NaN()}
	encounter := &models.MonsterEncounter{Latitude: 95, Longitude: 200}

	_, _, ok := monsterBattleInviteAnchor(monster, encounter)
	if ok {
		t.Fatal("expected invite anchor to be unavailable when neither source has valid coordinates")
	}
}

func TestClassifyMonsterBattleInviteProximityInvitesFreshNearbyMember(t *testing.T) {
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	snapshot := &liveness.LocationSnapshot{
		Location: "40.7128,-74.0060",
		SeenAt:   now.Add(-2 * time.Minute),
	}

	decision, distanceMeters, age := classifyMonsterBattleInviteProximity(
		snapshot,
		40.7128,
		-74.0060,
		now,
	)
	if decision != monsterBattleInviteProximityDecisionInvite {
		t.Fatalf("expected fresh nearby member to be invited, got %s", decision)
	}
	if distanceMeters != 0 {
		t.Fatalf("expected zero distance for matching coordinates, got %.2f", distanceMeters)
	}
	if age != 2*time.Minute {
		t.Fatalf("expected age to remain 2m, got %s", age)
	}
}

func TestClassifyMonsterBattleInviteProximitySkipsFreshKnownFarMember(t *testing.T) {
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	snapshot := &liveness.LocationSnapshot{
		Location: "40.7128,-74.0060",
		SeenAt:   now.Add(-2 * time.Minute),
	}

	decision, distanceMeters, _ := classifyMonsterBattleInviteProximity(
		snapshot,
		40.7308,
		-73.9975,
		now,
	)
	if decision != monsterBattleInviteProximityDecisionKnownFar {
		t.Fatalf("expected fresh far member to be filtered out, got %s", decision)
	}
	if distanceMeters < monsterBattlePartyKnownFarFreshMeters {
		t.Fatalf(
			"expected test fixture to exceed fresh known-far threshold %.2f, got %.2f",
			monsterBattlePartyKnownFarFreshMeters,
			distanceMeters,
		)
	}
}

func TestClassifyMonsterBattleInviteProximitySkipsStaleClearlyFarMember(t *testing.T) {
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	snapshot := &liveness.LocationSnapshot{
		Location: "40.7128,-74.0060",
		SeenAt:   now.Add(-20 * time.Minute),
	}

	decision, distanceMeters, age := classifyMonsterBattleInviteProximity(
		snapshot,
		40.7580,
		-73.9855,
		now,
	)
	if decision != monsterBattleInviteProximityDecisionKnownFar {
		t.Fatalf("expected stale but clearly far member to be filtered out, got %s", decision)
	}
	if distanceMeters < monsterBattlePartyKnownFarStaleMeters {
		t.Fatalf(
			"expected test fixture to exceed stale known-far threshold %.2f, got %.2f",
			monsterBattlePartyKnownFarStaleMeters,
			distanceMeters,
		)
	}
	if age != 20*time.Minute {
		t.Fatalf("expected age to remain 20m, got %s", age)
	}
}

func TestClassifyMonsterBattleInviteProximityFailsOpenWhenLocationIsAmbiguous(t *testing.T) {
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)
	snapshot := &liveness.LocationSnapshot{
		Location: "40.7128,-74.0060",
		SeenAt:   now.Add(-12 * time.Minute),
	}

	decision, distanceMeters, age := classifyMonsterBattleInviteProximity(
		snapshot,
		40.7134,
		-74.0055,
		now,
	)
	if decision != monsterBattleInviteProximityDecisionUnknown {
		t.Fatalf("expected ambiguous middle-distance member to fail open, got %s", decision)
	}
	if distanceMeters == 0 {
		t.Fatal("expected ambiguous test fixture to produce a non-zero distance")
	}
	if age != 12*time.Minute {
		t.Fatalf("expected age to remain 12m, got %s", age)
	}
}
