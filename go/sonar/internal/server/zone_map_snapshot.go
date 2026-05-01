package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type treasureChestWithUserStatus struct {
	models.TreasureChest
	OpenedByUser bool `json:"openedByUser"`
}

type zonePinSnapshot struct {
	PointsOfInterest          []models.PointOfInterest
	Characters                []*models.Character
	IncludesQuestAvailability bool
}

type zoneBaseContentSnapshot struct {
	TreasureChests   []treasureChestWithUserStatus
	HealingFountains []healingFountainWithUserStatus
	Shrines          []shrineWithUserStatus
	Resources        []resourceWithUserStatus
	Scenarios        []scenarioWithUserStatus
	Expositions      []models.Exposition
	Monsters         []models.MonsterEncounter
	Challenges       []models.Challenge
}

type zoneRequestTrace struct {
	ID      string
	Route   string
	ZoneID  string
	Started time.Time
}

func newZoneRequestTrace(
	ctx *gin.Context,
	route string,
	zoneID uuid.UUID,
) *zoneRequestTrace {
	traceID := strings.TrimSpace(ctx.GetHeader("X-Map-Trace-Id"))
	if traceID == "" {
		traceID = fmt.Sprintf("%s-%d", route, time.Now().UnixNano())
	}
	return &zoneRequestTrace{
		ID:      traceID,
		Route:   route,
		ZoneID:  zoneID.String(),
		Started: time.Now(),
	}
}

func (t *zoneRequestTrace) Logf(format string, args ...interface{}) {
	if t == nil {
		return
	}
	log.Printf(
		"zone-trace route=%s trace=%s zone=%s %s",
		t.Route,
		t.ID,
		t.ZoneID,
		fmt.Sprintf(format, args...),
	)
}

func (t *zoneRequestTrace) Step(label string, startedAt time.Time, format string, args ...interface{}) {
	if t == nil {
		return
	}
	message := ""
	if strings.TrimSpace(format) != "" {
		message = " " + fmt.Sprintf(format, args...)
	}
	t.Logf(
		"%s elapsedMs=%d%s",
		label,
		time.Since(startedAt).Milliseconds(),
		message,
	)
}

func serializeCharacterLocationMapSummary(
	location models.CharacterLocation,
) gin.H {
	return gin.H{
		"id":          location.ID,
		"characterId": location.CharacterID,
		"latitude":    location.Latitude,
		"longitude":   location.Longitude,
	}
}

func serializeCharacterRelationshipMapSummary(
	relationship *models.CharacterRelationshipState,
) interface{} {
	if relationship == nil {
		return nil
	}
	return gin.H{
		"trust":   relationship.Trust,
		"respect": relationship.Respect,
		"fear":    relationship.Fear,
		"debt":    relationship.Debt,
	}
}

func serializeCharacterMapSummary(character *models.Character) gin.H {
	if character == nil {
		return gin.H{}
	}
	locations := make([]gin.H, 0, len(character.Locations))
	for _, location := range character.Locations {
		locations = append(
			locations,
			serializeCharacterLocationMapSummary(location),
		)
	}
	var pointOfInterest interface{}
	if character.PointOfInterest != nil {
		pointOfInterest = gin.H{
			"lat": character.PointOfInterest.Lat,
			"lng": character.PointOfInterest.Lng,
		}
	}
	return gin.H{
		"id":                         character.ID,
		"name":                       character.Name,
		"description":                character.Description,
		"mapIconUrl":                 character.MapIconURL,
		"dialogueImageUrl":           character.DialogueImageURL,
		"thumbnailUrl":               character.ThumbnailURL,
		"pointOfInterestId":          character.PointOfInterestID,
		"pointOfInterest":            pointOfInterest,
		"locations":                  locations,
		"hasAvailableQuest":          character.HasAvailableQuest,
		"hasAvailableMainStoryQuest": character.HasAvailableMainStoryQuest,
		"relationship":               serializeCharacterRelationshipMapSummary(character.Relationship),
	}
}

func serializePoiTagMapSummary(tag models.Tag) gin.H {
	return gin.H{
		"id":   tag.ID,
		"name": tag.Value,
	}
}

func serializePointOfInterestMapSummary(pointOfInterest models.PointOfInterest) gin.H {
	characters := make([]gin.H, 0, len(pointOfInterest.Characters))
	for i := range pointOfInterest.Characters {
		characters = append(
			characters,
			serializeCharacterMapSummary(&pointOfInterest.Characters[i]),
		)
	}
	tags := make([]gin.H, 0, len(pointOfInterest.Tags))
	for _, tag := range pointOfInterest.Tags {
		tags = append(tags, serializePoiTagMapSummary(tag))
	}
	return gin.H{
		"id":                         pointOfInterest.ID,
		"name":                       pointOfInterest.Name,
		"lat":                        pointOfInterest.Lat,
		"lng":                        pointOfInterest.Lng,
		"imageURL":                   pointOfInterest.ImageUrl,
		"thumbnailUrl":               pointOfInterest.ThumbnailURL,
		"description":                pointOfInterest.Description,
		"originalName":               pointOfInterest.OriginalName,
		"googleMapsPlaceId":          pointOfInterest.GoogleMapsPlaceID,
		"mapMarkerUrl":               pointOfInterest.MapMarkerURL,
		"markerCategory":             pointOfInterest.MarkerCategory,
		"tags":                       tags,
		"characters":                 characters,
		"hasAvailableQuest":          pointOfInterest.HasAvailableQuest,
		"hasAvailableMainStoryQuest": pointOfInterest.HasAvailableMainStoryQuest,
	}
}

func serializeTreasureChestItemMapSummary(
	item models.TreasureChestItem,
) gin.H {
	return gin.H{
		"id":              item.ID,
		"inventoryItemId": item.InventoryItemID,
		"quantity":        item.Quantity,
	}
}

func serializeTreasureChestMapSummary(
	chest treasureChestWithUserStatus,
) gin.H {
	items := make([]gin.H, 0, len(chest.Items))
	for _, item := range chest.Items {
		items = append(items, serializeTreasureChestItemMapSummary(item))
	}
	return gin.H{
		"id":           chest.ID,
		"latitude":     chest.Latitude,
		"longitude":    chest.Longitude,
		"zoneId":       chest.ZoneID,
		"gold":         chest.Gold,
		"openedByUser": chest.OpenedByUser,
		"unlockTier":   chest.UnlockTier,
		"mapMarkerUrl": chest.MapMarkerURL,
		"items":        items,
	}
}

func serializeHealingFountainMapSummary(
	fountain healingFountainWithUserStatus,
) gin.H {
	return gin.H{
		"id":                       fountain.ID,
		"name":                     fountain.Name,
		"description":              fountain.Description,
		"thumbnailUrl":             fountain.ThumbnailURL,
		"zoneId":                   fountain.ZoneID,
		"latitude":                 fountain.Latitude,
		"longitude":                fountain.Longitude,
		"availableNow":             fountain.AvailableNow,
		"discovered":               fountain.Discovered,
		"lastUsedAt":               fountain.LastUsedAt,
		"nextAvailableAt":          fountain.NextAvailableAt,
		"cooldownSecondsRemaining": fountain.CooldownSecondsRemaining,
	}
}

func serializeShrineMapSummary(shrine shrineWithUserStatus) gin.H {
	return gin.H{
		"id":                       shrine.ID,
		"shrineTemplateId":         shrine.ShrineTemplateID,
		"name":                     shrine.Name,
		"description":              shrine.Description,
		"blessingName":             shrine.BlessingName,
		"effectDescription":        shrine.EffectDescription,
		"effectKind":               shrine.EffectKind,
		"baseMagnitude":            shrine.BaseMagnitude,
		"zoneId":                   shrine.ZoneID,
		"latitude":                 shrine.Latitude,
		"longitude":                shrine.Longitude,
		"cooldownSeconds":          shrine.CooldownSeconds,
		"availableNow":             shrine.AvailableNow,
		"lastUsedAt":               shrine.LastUsedAt,
		"nextAvailableAt":          shrine.NextAvailableAt,
		"cooldownSecondsRemaining": shrine.CooldownSecondsRemaining,
		"mapMarkerUrl":             shrine.MapMarkerURL,
	}
}

func serializeResourceTypeMapSummary(
	resourceType models.ResourceType,
) gin.H {
	return gin.H{
		"id":         resourceType.ID,
		"name":       resourceType.Name,
		"slug":       resourceType.Slug,
		"mapIconUrl": resourceType.MapIconURL,
	}
}

func serializeResourceMapSummary(resource resourceWithUserStatus) gin.H {
	return gin.H{
		"id":             resource.ID,
		"zoneId":         resource.ZoneID,
		"resourceTypeId": resource.ResourceTypeID,
		"resourceType":   serializeResourceTypeMapSummary(resource.ResourceType),
		"latitude":       resource.Latitude,
		"longitude":      resource.Longitude,
		"gatheredByUser": resource.GatheredByUser,
		"hasFullDetails": false,
	}
}

func serializeScenarioMapSummary(scenario scenarioWithUserStatus) gin.H {
	return gin.H{
		"id":                 scenario.ID,
		"zoneId":             scenario.ZoneID,
		"pointOfInterestId":  scenario.PointOfInterestID,
		"latitude":           scenario.Latitude,
		"longitude":          scenario.Longitude,
		"prompt":             scenario.Prompt,
		"imageUrl":           scenario.ImageURL,
		"thumbnailUrl":       scenario.ThumbnailURL,
		"difficulty":         scenario.Difficulty,
		"scaleWithUserLevel": scenario.ScaleWithUserLevel,
		"rewardExperience":   scenario.RewardExperience,
		"rewardGold":         scenario.RewardGold,
		"openEnded":          scenario.OpenEnded,
		"attemptedByUser":    scenario.AttemptedByUser,
		"hasFullDetails":     false,
	}
}

func serializeExpositionMapSummary(exposition models.Exposition) gin.H {
	return gin.H{
		"id":                exposition.ID,
		"zoneId":            exposition.ZoneID,
		"pointOfInterestId": exposition.PointOfInterestID,
		"latitude":          exposition.Latitude,
		"longitude":         exposition.Longitude,
		"title":             exposition.Title,
		"description":       exposition.Description,
		"imageUrl":          exposition.ImageURL,
		"thumbnailUrl":      exposition.ThumbnailURL,
		"hasFullDetails":    false,
	}
}

func serializeMonsterMapSummary(monster monsterResponse) gin.H {
	return gin.H{
		"id":           monster.ID,
		"name":         monster.Name,
		"description":  monster.Description,
		"imageUrl":     monster.ImageURL,
		"thumbnailUrl": monster.ThumbnailURL,
		"zoneId":       monster.ZoneID,
		"latitude":     monster.Latitude,
		"longitude":    monster.Longitude,
	}
}

func serializeMonsterEncounterMapSummary(encounter models.MonsterEncounter) gin.H {
	thumbnailURL := strings.TrimSpace(encounter.ThumbnailURL)
	imageURL := strings.TrimSpace(encounter.ImageURL)
	if thumbnailURL == "" {
		thumbnailURL = imageURL
	}
	return gin.H{
		"id":                encounter.ID,
		"name":              encounter.Name,
		"description":       encounter.Description,
		"imageUrl":          imageURL,
		"thumbnailUrl":      thumbnailURL,
		"encounterType":     encounter.EncounterType,
		"zoneId":            encounter.ZoneID,
		"pointOfInterestId": encounter.PointOfInterestID,
		"latitude":          encounter.Latitude,
		"longitude":         encounter.Longitude,
		"monsterCount":      len(encounter.Members),
		"hasFullDetails":    false,
	}
}

func serializeChallengeMapSummary(challenge models.Challenge) gin.H {
	return gin.H{
		"id":                 challenge.ID,
		"zoneId":             challenge.ZoneID,
		"pointOfInterestId":  challenge.PointOfInterestID,
		"latitude":           challenge.Latitude,
		"longitude":          challenge.Longitude,
		"polygonPoints":      challenge.PolygonPoints,
		"question":           challenge.Question,
		"description":        challenge.Description,
		"imageUrl":           challenge.ImageURL,
		"thumbnailUrl":       challenge.ThumbnailURL,
		"reward":             challenge.Reward,
		"inventoryItemId":    challenge.InventoryItemID,
		"submissionType":     challenge.SubmissionType,
		"difficulty":         challenge.Difficulty,
		"scaleWithUserLevel": challenge.ScaleWithUserLevel,
		"statTags":           challenge.StatTags,
		"proficiency":        challenge.Proficiency,
	}
}

func serializeZonePinSnapshot(snapshot zonePinSnapshot) gin.H {
	pointsOfInterest := make([]gin.H, 0, len(snapshot.PointsOfInterest))
	for i := range snapshot.PointsOfInterest {
		pointsOfInterest = append(
			pointsOfInterest,
			serializePointOfInterestMapSummary(snapshot.PointsOfInterest[i]),
		)
	}
	characters := make([]gin.H, 0, len(snapshot.Characters))
	for _, character := range snapshot.Characters {
		characters = append(characters, serializeCharacterMapSummary(character))
	}
	return gin.H{
		"pointsOfInterest":          pointsOfInterest,
		"characters":                characters,
		"includesQuestAvailability": snapshot.IncludesQuestAvailability,
	}
}

func serializeZoneBaseContentSnapshot(snapshot zoneBaseContentSnapshot) gin.H {
	treasureChests := make([]gin.H, 0, len(snapshot.TreasureChests))
	healingFountains := make([]gin.H, 0, len(snapshot.HealingFountains))
	shrines := make([]gin.H, 0, len(snapshot.Shrines))
	resources := make([]gin.H, 0, len(snapshot.Resources))
	scenarios := make([]gin.H, 0, len(snapshot.Scenarios))
	expositions := make([]gin.H, 0, len(snapshot.Expositions))
	monsters := make([]gin.H, 0, len(snapshot.Monsters))
	challenges := make([]gin.H, 0, len(snapshot.Challenges))

	for _, chest := range snapshot.TreasureChests {
		treasureChests = append(treasureChests, serializeTreasureChestMapSummary(chest))
	}
	for _, fountain := range snapshot.HealingFountains {
		healingFountains = append(
			healingFountains,
			serializeHealingFountainMapSummary(fountain),
		)
	}
	for _, shrine := range snapshot.Shrines {
		shrines = append(shrines, serializeShrineMapSummary(shrine))
	}
	for _, resource := range snapshot.Resources {
		resources = append(resources, serializeResourceMapSummary(resource))
	}
	for _, scenario := range snapshot.Scenarios {
		scenarios = append(scenarios, serializeScenarioMapSummary(scenario))
	}
	for _, exposition := range snapshot.Expositions {
		expositions = append(expositions, serializeExpositionMapSummary(exposition))
	}
	for _, monster := range snapshot.Monsters {
		monsters = append(monsters, serializeMonsterEncounterMapSummary(monster))
	}
	for _, challenge := range snapshot.Challenges {
		challenges = append(challenges, serializeChallengeMapSummary(challenge))
	}

	return gin.H{
		"treasureChests":   treasureChests,
		"healingFountains": healingFountains,
		"shrines":          shrines,
		"resources":        resources,
		"scenarios":        scenarios,
		"expositions":      expositions,
		"monsters":         monsters,
		"challenges":       challenges,
	}
}

func (s *server) getZoneMapSnapshot(ctx *gin.Context) {
	requestStartedAt := time.Now()
	user, err := s.getAuthenticatedUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	zoneID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone ID"})
		return
	}
	trace := newZoneRequestTrace(ctx, "map-snapshot", zoneID)
	trace.Logf("start user=%s", user.ID.String())

	zoneLookupStartedAt := time.Now()
	zone, err := s.dbClient.Zone().FindByID(ctx, zoneID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	trace.Step("zone.lookup", zoneLookupStartedAt, "")

	storyFlagsStartedAt := time.Now()
	activeStoryFlags, err := s.loadUserStoryFlagMap(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	trace.Step(
		"story-flags",
		storyFlagsStartedAt,
		"count=%d",
		len(activeStoryFlags),
	)

	requestCtx := ctx.Request.Context()
	var (
		pinSnapshot  zonePinSnapshot
		baseSnapshot zoneBaseContentSnapshot
		firstErr     error
		errMu        sync.Mutex
		wg           sync.WaitGroup
	)
	setErr := func(err error) {
		if err == nil {
			return
		}
		errMu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMu.Unlock()
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		startedAt := time.Now()
		snapshot, err := s.zonePinSnapshotForUser(
			requestCtx,
			user,
			zone,
			activeStoryFlags,
			false,
			trace,
		)
		if err != nil {
			setErr(err)
			return
		}
		pinSnapshot = snapshot
		trace.Step(
			"pin-snapshot",
			startedAt,
			"pois=%d characters=%d includesQuestAvailability=%t",
			len(snapshot.PointsOfInterest),
			len(snapshot.Characters),
			snapshot.IncludesQuestAvailability,
		)
	}()
	go func() {
		defer wg.Done()
		startedAt := time.Now()
		snapshot, err := s.zoneBaseContentSnapshotForUser(
			requestCtx,
			user,
			zone,
			activeStoryFlags,
			trace,
		)
		if err != nil {
			setErr(err)
			return
		}
		baseSnapshot = snapshot
		trace.Step(
			"base-snapshot",
			startedAt,
			"chests=%d fountains=%d shrines=%d resources=%d scenarios=%d expositions=%d monsters=%d challenges=%d",
			len(snapshot.TreasureChests),
			len(snapshot.HealingFountains),
			len(snapshot.Shrines),
			len(snapshot.Resources),
			len(snapshot.Scenarios),
			len(snapshot.Expositions),
			len(snapshot.Monsters),
			len(snapshot.Challenges),
		)
	}()
	wg.Wait()
	if firstErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": firstErr.Error()})
		return
	}

	response := serializeZonePinSnapshot(pinSnapshot)
	for key, value := range serializeZoneBaseContentSnapshot(baseSnapshot) {
		response[key] = value
	}
	trace.Step("request.total", requestStartedAt, "")
	ctx.JSON(http.StatusOK, response)
}

func (s *server) zonePinSnapshotForUser(
	ctx context.Context,
	user *models.User,
	zone *models.Zone,
	activeStoryFlags map[string]bool,
	includeQuestAvailability bool,
	trace *zoneRequestTrace,
) (zonePinSnapshot, error) {
	totalStartedAt := time.Now()
	pointsOfInterestStartedAt := time.Now()
	pointsOfInterest, err := s.dbClient.PointOfInterest().FindAllForZone(ctx, zone.ID)
	if err != nil {
		return zonePinSnapshot{}, err
	}
	trace.Step(
		"pin.poi-query",
		pointsOfInterestStartedAt,
		"count=%d",
		len(pointsOfInterest),
	)

	zonePoiIDsList := make([]uuid.UUID, 0, len(pointsOfInterest))
	for _, pointOfInterest := range pointsOfInterest {
		if pointOfInterest.ID == uuid.Nil {
			continue
		}
		zonePoiIDsList = append(zonePoiIDsList, pointOfInterest.ID)
	}

	charactersStartedAt := time.Now()
	characters, err := s.dbClient.Character().FindPotentiallyInZone(
		ctx,
		zone,
		zonePoiIDsList,
	)
	if err != nil {
		return zonePinSnapshot{}, err
	}
	trace.Step(
		"pin.character-query",
		charactersStartedAt,
		"count=%d",
		len(characters),
	)

	relevantCharacterIDs := collectCharacterIDsFromPointsOfInterest(pointsOfInterest)
	for _, character := range characters {
		if character == nil || character.ID == uuid.Nil {
			continue
		}
		relevantCharacterIDs = append(relevantCharacterIDs, character.ID)
	}
	seenRelevantCharacterIDs := make(map[uuid.UUID]struct{}, len(relevantCharacterIDs))
	uniqueRelevantCharacterIDs := make([]uuid.UUID, 0, len(relevantCharacterIDs))
	for _, characterID := range relevantCharacterIDs {
		if characterID == uuid.Nil {
			continue
		}
		if _, exists := seenRelevantCharacterIDs[characterID]; exists {
			continue
		}
		seenRelevantCharacterIDs[characterID] = struct{}{}
		uniqueRelevantCharacterIDs = append(uniqueRelevantCharacterIDs, characterID)
	}

	relationshipStartedAt := time.Now()
	relationshipMap, err := s.loadUserCharacterRelationshipMap(
		ctx,
		user.ID,
		uniqueRelevantCharacterIDs,
	)
	if err != nil {
		return zonePinSnapshot{}, err
	}
	trace.Step(
		"pin.relationships",
		relationshipStartedAt,
		"count=%d",
		len(relationshipMap),
	)

	activeFetchStartedAt := time.Now()
	activeFetchCharacterIDs, err := s.activeFetchQuestCharacterIDsForUser(
		ctx,
		user.ID,
	)
	if err != nil {
		return zonePinSnapshot{}, err
	}
	trace.Step(
		"pin.active-fetch-character-ids",
		activeFetchStartedAt,
		"count=%d",
		len(activeFetchCharacterIDs),
	)

	storyPoiStartedAt := time.Now()
	if err := s.applyStoryWorldChangesToPointOfInterests(
		ctx,
		pointsOfInterest,
		activeStoryFlags,
	); err != nil {
		return zonePinSnapshot{}, err
	}
	trace.Step("pin.story-pois", storyPoiStartedAt, "")

	zonePoiIDs := make(map[uuid.UUID]struct{}, len(zonePoiIDsList))
	markerCache := contentMapMarkerExistenceCache{}
	zoneKindSlug := models.NormalizeZoneKind(zone.Kind)
	availability := map[uuid.UUID]characterQuestAvailability{}
	if includeQuestAvailability {
		availabilityStartedAt := time.Now()
		availability, err = s.questAvailabilityByCharacterIDs(
			ctx,
			user.ID,
			uniqueRelevantCharacterIDs,
		)
		if err != nil {
			return zonePinSnapshot{}, err
		}
		trace.Step(
			"pin.quest-availability",
			availabilityStartedAt,
			"count=%d",
			len(availability),
		)
	}

	for i := range pointsOfInterest {
		hasAvailable := false
		hasAvailableMainStory := false
		zonePoiIDs[pointsOfInterest[i].ID] = struct{}{}
		applyPointOfInterestStoryVariant(&pointsOfInterest[i], activeStoryFlags)
		pointsOfInterest[i].MapMarkerURL = s.resolvePointOfInterestMapMarkerURL(
			ctx,
			pointsOfInterest[i].MarkerCategory,
			effectiveContentMapMarkerZoneKind(pointsOfInterest[i].ZoneKind, zone),
			markerCache,
		)
		for j := range pointsOfInterest[i].Characters {
			if includeQuestAvailability {
				pointsOfInterest[i].Characters[j].HasAvailableQuest =
					availability[pointsOfInterest[i].Characters[j].ID].HasAvailableQuest
				pointsOfInterest[i].Characters[j].HasAvailableMainStoryQuest =
					availability[pointsOfInterest[i].Characters[j].ID].HasAvailableMainStoryQuest
			}
			applyCharacterStoryVariant(
				&pointsOfInterest[i].Characters[j],
				activeStoryFlags,
			)
			applyCharacterRelationship(
				&pointsOfInterest[i].Characters[j],
				relationshipMap,
			)
			if includeQuestAvailability &&
				pointsOfInterest[i].Characters[j].HasAvailableQuest {
				hasAvailable = true
			}
			if includeQuestAvailability &&
				pointsOfInterest[i].Characters[j].HasAvailableMainStoryQuest {
				hasAvailableMainStory = true
			}
		}
		if includeQuestAvailability {
			pointsOfInterest[i].HasAvailableQuest = hasAvailable
			pointsOfInterest[i].HasAvailableMainStoryQuest = hasAvailableMainStory
		}
	}

	visibleCharacters := make([]*models.Character, 0, len(characters))
	for i := range characters {
		ch := characters[i]
		if ch == nil ||
			!characterVisibleToUser(user.ID, ch) ||
			!fetchQuestCharacterVisibleToUser(ch, activeFetchCharacterIDs) {
			continue
		}
		if err := s.applyStoryWorldChangesToCharacter(
			ctx,
			ch,
			activeStoryFlags,
		); err != nil {
			return zonePinSnapshot{}, err
		}
		if includeQuestAvailability {
			ch.HasAvailableQuest = availability[ch.ID].HasAvailableQuest
			ch.HasAvailableMainStoryQuest =
				availability[ch.ID].HasAvailableMainStoryQuest
		}
		applyCharacterStoryVariant(ch, activeStoryFlags)
		applyCharacterRelationship(ch, relationshipMap)
		if markerURL := s.resolveSharedContentMapMarkerURL(
			ctx,
			sharedContentMapMarkerDefinitions[7],
			zoneKindSlug,
			ch.MapIconURL,
			markerCache,
		); markerURL != "" {
			ch.MapIconURL = markerURL
		}
		if !characterInZone(zone, zonePoiIDs, ch) {
			continue
		}
		visibleCharacters = append(visibleCharacters, ch)
	}

	snapshot := zonePinSnapshot{
		PointsOfInterest:          pointsOfInterest,
		Characters:                visibleCharacters,
		IncludesQuestAvailability: includeQuestAvailability,
	}
	trace.Step(
		"pin.total",
		totalStartedAt,
		"pois=%d visibleCharacters=%d includeQuestAvailability=%t",
		len(snapshot.PointsOfInterest),
		len(snapshot.Characters),
		snapshot.IncludesQuestAvailability,
	)
	return snapshot, nil
}

func (s *server) zoneBaseContentSnapshotForUser(
	ctx context.Context,
	user *models.User,
	zone *models.Zone,
	activeStoryFlags map[string]bool,
	trace *zoneRequestTrace,
) (zoneBaseContentSnapshot, error) {
	totalStartedAt := time.Now()
	zoneID := zone.ID
	var (
		treasureChestResponse []treasureChestWithUserStatus
		healingResponse       []healingFountainWithUserStatus
		shrineResponse        []shrineWithUserStatus
		resourceResponse      []resourceWithUserStatus
		scenarioResponse      []scenarioWithUserStatus
		expositionResponse    []models.Exposition
		monsterResponse       []models.MonsterEncounter
		challengeResponse     []models.Challenge
		userLevel             int
		partyLevel            int
		firstErr              error
		errMu                 sync.Mutex
		wg                    sync.WaitGroup
	)
	setErr := func(err error) {
		if err == nil {
			return
		}
		errMu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMu.Unlock()
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		startedAt := time.Now()
		level, err := s.currentUserLevel(ctx, user.ID)
		if err != nil {
			setErr(err)
			return
		}
		userLevel = level
		trace.Step("base.user-level", startedAt, "level=%d", level)
	}()
	go func() {
		defer wg.Done()
		startedAt := time.Now()
		level, err := s.currentPartyMaxLevel(ctx, user)
		if err != nil {
			setErr(err)
			return
		}
		partyLevel = level
		trace.Step("base.party-max-level", startedAt, "level=%d", level)
	}()
	wg.Wait()
	if firstErr != nil {
		return zoneBaseContentSnapshot{}, firstErr
	}

	wg = sync.WaitGroup{}
	wg.Add(8)

	go func() {
		defer wg.Done()
		startedAt := time.Now()
		markerCache := contentMapMarkerExistenceCache{}
		treasureChests, openedMap, err := s.dbClient.TreasureChest().
			FindByZoneIDWithUserStatus(ctx, zoneID, &user.ID)
		if err != nil {
			setErr(err)
			return
		}
		response := make([]treasureChestWithUserStatus, len(treasureChests))
		for i, chest := range treasureChests {
			if markerURL := s.resolveSharedContentMapMarkerURL(
				ctx,
				sharedContentMapMarkerDefinitions[3],
				effectiveContentMapMarkerZoneKind(chest.ZoneKind, zone),
				"",
				markerCache,
			); markerURL != "" {
				chest.MapMarkerURL = markerURL
			}
			response[i] = treasureChestWithUserStatus{
				TreasureChest: chest,
				OpenedByUser:  openedMap[chest.ID],
			}
		}
		treasureChestResponse = response
		trace.Step("base.treasure-chests", startedAt, "count=%d", len(response))
	}()

	go func() {
		defer wg.Done()
		startedAt := time.Now()
		markerCache := contentMapMarkerExistenceCache{}
		fountains, err := s.dbClient.HealingFountain().FindByZoneID(ctx, zoneID)
		if err != nil {
			setErr(err)
			return
		}
		now := time.Now()
		latestVisitsByFountain, err := s.dbClient.HealingFountain().
			FindLatestVisitsByUser(ctx, user.ID)
		if err != nil {
			setErr(err)
			return
		}
		discoveredByFountain, err := s.userHealingFountainDiscoveryMap(ctx, user.ID)
		if err != nil {
			setErr(err)
			return
		}
		response := make([]healingFountainWithUserStatus, 0, len(fountains))
		for _, fountain := range fountains {
			if fountain.Invalidated {
				continue
			}
			discovered := discoveredByFountain[fountain.ID]
			status := healingFountainCooldownStatusFromVisit(
				latestVisitsByFountain[fountain.ID],
				now,
			)
			markerDefinition := sharedContentMapMarkerDefinitions[0]
			if discovered {
				markerDefinition = sharedContentMapMarkerDefinitions[8]
			}
			markerURL := s.resolveSharedContentMapMarkerURL(
				ctx,
				markerDefinition,
				effectiveContentMapMarkerZoneKind(fountain.ZoneKind, zone),
				"",
				markerCache,
			)
			response = append(
				response,
				healingFountainResponseWithStatus(
					fountain,
					status,
					discovered,
					markerURL,
				),
			)
		}
		healingResponse = response
		trace.Step("base.healing-fountains", startedAt, "count=%d", len(response))
	}()

	go func() {
		defer wg.Done()
		startedAt := time.Now()
		shrines, err := s.dbClient.Shrine().FindByZoneID(ctx, zoneID)
		if err != nil {
			setErr(err)
			return
		}
		response, err := s.shrineResponsesForUser(ctx, shrines, &user.ID)
		if err != nil {
			setErr(err)
			return
		}
		shrineResponse = response
		trace.Step("base.shrines", startedAt, "count=%d", len(response))
	}()

	go func() {
		defer wg.Done()
		startedAt := time.Now()
		markerCache := contentMapMarkerExistenceCache{}
		resources, gatheredMap, err := s.dbClient.Resource().
			FindByZoneIDWithUserStatus(ctx, zoneID, &user.ID)
		if err != nil {
			setErr(err)
			return
		}
		response := make([]resourceWithUserStatus, 0, len(resources))
		for _, resource := range resources {
			if resource.Invalidated {
				continue
			}
			if markerURL := s.resolveResourceTypeMapMarkerURL(
				ctx,
				&resource.ResourceType,
				effectiveContentMapMarkerZoneKind(resource.ZoneKind, zone),
				markerCache,
			); markerURL != "" {
				resource.ResourceType.MapIconURL = markerURL
			}
			response = append(response, resourceWithUserStatus{
				Resource:       resource,
				GatheredByUser: gatheredMap[resource.ID],
			})
		}
		resourceResponse = response
		trace.Step("base.resources", startedAt, "count=%d", len(response))
	}()

	go func() {
		defer wg.Done()
		startedAt := time.Now()
		markerCache := contentMapMarkerExistenceCache{}
		scenarios, attemptedMap, err := s.dbClient.Scenario().
			FindByZoneIDWithUserStatusExcludingQuestNodes(ctx, zoneID, &user.ID)
		if err != nil {
			setErr(err)
			return
		}
		response := make([]scenarioWithUserStatus, 0, len(scenarios))
		for _, scenario := range scenarios {
			if !scenarioAvailableForStoryFlags(&scenario, activeStoryFlags) {
				continue
			}
			if attemptedMap[scenario.ID] {
				continue
			}
			scenarioWithMarker := scenarioWithScaledDifficulty(scenario, userLevel)
			if markerURL := s.resolveSharedContentMapMarkerURL(
				ctx,
				sharedContentMapMarkerDefinitions[1],
				effectiveContentMapMarkerZoneKind(scenario.ZoneKind, zone),
				scenarioWithMarker.ThumbnailURL,
				markerCache,
			); markerURL != "" {
				scenarioWithMarker.ThumbnailURL = markerURL
			}
			response = append(response, scenarioWithUserStatus{
				Scenario:        scenarioWithMarker,
				AttemptedByUser: false,
			})
		}
		scenarioResponse = response
		trace.Step("base.scenarios", startedAt, "count=%d", len(response))
	}()

	go func() {
		defer wg.Done()
		startedAt := time.Now()
		markerCache := contentMapMarkerExistenceCache{}
		expositions, err := s.dbClient.Exposition().
			FindByZoneIDExcludingQuestNodes(ctx, zoneID)
		if err != nil {
			setErr(err)
			return
		}
		expositionIDs := make([]uuid.UUID, 0, len(expositions))
		for _, exposition := range expositions {
			if !expositionAvailableForStoryFlags(&exposition, activeStoryFlags) {
				continue
			}
			expositionIDs = append(expositionIDs, exposition.ID)
		}
		completedExpositionIDs, err := s.dbClient.Exposition().
			FindCompletedExpositionIDsByUser(ctx, user.ID, expositionIDs)
		if err != nil {
			setErr(err)
			return
		}
		completedExpositionSet := make(
			map[uuid.UUID]struct{},
			len(completedExpositionIDs),
		)
		for _, id := range completedExpositionIDs {
			completedExpositionSet[id] = struct{}{}
		}
		response := make([]models.Exposition, 0, len(expositions))
		for i := range expositions {
			if !expositionAvailableForStoryFlags(&expositions[i], activeStoryFlags) {
				continue
			}
			if _, completed := completedExpositionSet[expositions[i].ID]; completed {
				continue
			}
			if markerURL := s.resolveSharedContentMapMarkerURL(
				ctx,
				sharedContentMapMarkerDefinitions[2],
				effectiveContentMapMarkerZoneKind(expositions[i].ZoneKind, zone),
				expositions[i].ThumbnailURL,
				markerCache,
			); markerURL != "" {
				expositions[i].ThumbnailURL = markerURL
			}
			response = append(response, expositions[i])
		}
		expositionResponse = response
		trace.Step("base.expositions", startedAt, "count=%d", len(response))
	}()

	go func() {
		defer wg.Done()
		startedAt := time.Now()
		markerCache := contentMapMarkerExistenceCache{}
		encounters, err := s.dbClient.MonsterEncounter().
			FindMapByZoneIDExcludingQuestNodes(ctx, zoneID)
		if err != nil {
			setErr(err)
			return
		}
		defeatedEncounterIDs, err := s.dbClient.UserMonsterEncounterVictory().
			FindEncounterIDsByUserAndZone(ctx, user.ID, zoneID)
		if err != nil {
			setErr(err)
			return
		}
		defeatedSet := make(map[uuid.UUID]struct{}, len(defeatedEncounterIDs))
		for _, encounterID := range defeatedEncounterIDs {
			defeatedSet[encounterID] = struct{}{}
		}
		response := make([]models.MonsterEncounter, 0, len(encounters))
		for i := range encounters {
			if !monsterEncounterAvailableForStoryFlags(
				&encounters[i],
				activeStoryFlags,
			) {
				continue
			}
			if _, defeated := defeatedSet[encounters[i].ID]; defeated {
				continue
			}
			if !monsterEncounterVisibleToUser(user.ID, &encounters[i]) {
				continue
			}
			fallbackThumbnailURL := strings.TrimSpace(encounters[i].ThumbnailURL)
			if fallbackThumbnailURL == "" {
				fallbackThumbnailURL = strings.TrimSpace(encounters[i].ImageURL)
			}
			if markerURL := s.resolveSharedContentMapMarkerURL(
				ctx,
				monsterEncounterContentMapMarkerDefinition(encounters[i].EncounterType),
				effectiveContentMapMarkerZoneKind(encounters[i].ZoneKind, zone),
				fallbackThumbnailURL,
				markerCache,
			); markerURL != "" {
				encounters[i].ThumbnailURL = markerURL
			} else {
				encounters[i].ThumbnailURL = fallbackThumbnailURL
			}
			response = append(response, encounters[i])
		}
		monsterResponse = response
		trace.Step("base.monsters", startedAt, "count=%d", len(response))
	}()

	go func() {
		defer wg.Done()
		startedAt := time.Now()
		challenges, err := s.dbClient.Challenge().
			FindByZoneIDExcludingQuestNodes(ctx, zoneID)
		if err != nil {
			setErr(err)
			return
		}
		challengeIDs := make([]uuid.UUID, 0, len(challenges))
		for _, challenge := range challenges {
			if !challengeAvailableForStoryFlags(&challenge, activeStoryFlags) {
				continue
			}
			challengeIDs = append(challengeIDs, challenge.ID)
		}
		completedChallengeIDs, err := s.dbClient.Challenge().
			FindCompletedChallengeIDsByUser(ctx, user.ID, challengeIDs)
		if err != nil {
			setErr(err)
			return
		}
		completedChallengeSet := make(map[uuid.UUID]struct{}, len(completedChallengeIDs))
		for _, id := range completedChallengeIDs {
			completedChallengeSet[id] = struct{}{}
		}
		response := make([]models.Challenge, 0, len(challenges))
		for i := range challenges {
			if !challengeAvailableForStoryFlags(&challenges[i], activeStoryFlags) {
				continue
			}
			if _, completed := completedChallengeSet[challenges[i].ID]; completed {
				continue
			}
			response = append(
				response,
				challengeWithScaledDifficulty(challenges[i], userLevel),
			)
		}
		challengeResponse = response
		trace.Step("base.challenges", startedAt, "count=%d", len(response))
	}()

	wg.Wait()
	if firstErr != nil {
		return zoneBaseContentSnapshot{}, firstErr
	}

	snapshot := zoneBaseContentSnapshot{
		TreasureChests:   treasureChestResponse,
		HealingFountains: healingResponse,
		Shrines:          shrineResponse,
		Resources:        resourceResponse,
		Scenarios:        scenarioResponse,
		Expositions:      expositionResponse,
		Monsters:         monsterResponse,
		Challenges:       challengeResponse,
	}
	trace.Step(
		"base.total",
		totalStartedAt,
		"chests=%d fountains=%d shrines=%d resources=%d scenarios=%d expositions=%d monsters=%d challenges=%d userLevel=%d partyLevel=%d",
		len(snapshot.TreasureChests),
		len(snapshot.HealingFountains),
		len(snapshot.Shrines),
		len(snapshot.Resources),
		len(snapshot.Scenarios),
		len(snapshot.Expositions),
		len(snapshot.Monsters),
		len(snapshot.Challenges),
		userLevel,
		partyLevel,
	)
	return snapshot, nil
}
