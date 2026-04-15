package server

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/paulmach/orb/geo"
)

const (
	zoneSeedAutoSquareFeetPerSquareMeter = 10.763910416709722
	zoneSeedAutoSquareFeetPerAcre        = 43560.0
)

type zoneSeedDraftResolutionError struct {
	statusCode int
	err        error
}

func (e *zoneSeedDraftResolutionError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *zoneSeedDraftResolutionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func newZoneSeedDraftResolutionError(statusCode int, err error) error {
	if err == nil {
		return nil
	}
	return &zoneSeedDraftResolutionError{
		statusCode: statusCode,
		err:        err,
	}
}

type zoneSeedDraftCountOverrides struct {
	PlaceCount           *int
	MonsterCount         *int
	BossEncounterCount   *int
	RaidEncounterCount   *int
	InputEncounterCount  *int
	OptionEncounterCount *int
	TreasureChestCount   *int
	HealingFountainCount *int
}

func normalizeZoneSeedDraftMode(mode string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		return models.ZoneSeedModeManual, nil
	}
	if !models.IsValidZoneSeedMode(normalized) {
		return "", fmt.Errorf("invalid zone seed mode: %s", mode)
	}
	return normalized, nil
}

func normalizeZoneSeedDraftTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(tags))
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.ToLower(strings.TrimSpace(tag))
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	return normalized
}

func zoneSeedDraftCountOverridesFromRequest(requestBody zoneSeedDraftRequest) (zoneSeedDraftCountOverrides, error) {
	copyCount := func(value *int, label string) (*int, error) {
		if value == nil {
			return nil, nil
		}
		if *value < 0 {
			return nil, fmt.Errorf("%s must be zero or greater", label)
		}
		normalized := *value
		return &normalized, nil
	}

	var err error
	overrides := zoneSeedDraftCountOverrides{}
	if overrides.PlaceCount, err = copyCount(requestBody.PlaceCount, "placeCount"); err != nil {
		return zoneSeedDraftCountOverrides{}, err
	}
	if overrides.MonsterCount, err = copyCount(requestBody.MonsterCount, "monsterCount"); err != nil {
		return zoneSeedDraftCountOverrides{}, err
	}
	if overrides.BossEncounterCount, err = copyCount(requestBody.BossEncounterCount, "bossEncounterCount"); err != nil {
		return zoneSeedDraftCountOverrides{}, err
	}
	if overrides.RaidEncounterCount, err = copyCount(requestBody.RaidEncounterCount, "raidEncounterCount"); err != nil {
		return zoneSeedDraftCountOverrides{}, err
	}
	if overrides.InputEncounterCount, err = copyCount(requestBody.InputEncounterCount, "inputEncounterCount"); err != nil {
		return zoneSeedDraftCountOverrides{}, err
	}
	if overrides.OptionEncounterCount, err = copyCount(requestBody.OptionEncounterCount, "optionEncounterCount"); err != nil {
		return zoneSeedDraftCountOverrides{}, err
	}
	if overrides.TreasureChestCount, err = copyCount(requestBody.TreasureChestCount, "treasureChestCount"); err != nil {
		return zoneSeedDraftCountOverrides{}, err
	}
	if overrides.HealingFountainCount, err = copyCount(requestBody.HealingFountainCount, "healingFountainCount"); err != nil {
		return zoneSeedDraftCountOverrides{}, err
	}

	return overrides, nil
}

func zoneSeedResolvedCountsFromOverrides(overrides zoneSeedDraftCountOverrides) models.ZoneSeedResolvedCounts {
	counts := models.ZoneSeedResolvedCounts{}
	if overrides.PlaceCount != nil {
		counts.PlaceCount = *overrides.PlaceCount
	}
	if overrides.MonsterCount != nil {
		counts.MonsterCount = *overrides.MonsterCount
	}
	if overrides.BossEncounterCount != nil {
		counts.BossEncounterCount = *overrides.BossEncounterCount
	}
	if overrides.RaidEncounterCount != nil {
		counts.RaidEncounterCount = *overrides.RaidEncounterCount
	}
	if overrides.InputEncounterCount != nil {
		counts.InputEncounterCount = *overrides.InputEncounterCount
	}
	if overrides.OptionEncounterCount != nil {
		counts.OptionEncounterCount = *overrides.OptionEncounterCount
	}
	if overrides.TreasureChestCount != nil {
		counts.TreasureChestCount = *overrides.TreasureChestCount
	}
	if overrides.HealingFountainCount != nil {
		counts.HealingFountainCount = *overrides.HealingFountainCount
	}
	return counts
}

func zoneSeedApplyCountOverrides(
	base models.ZoneSeedResolvedCounts,
	overrides zoneSeedDraftCountOverrides,
) models.ZoneSeedResolvedCounts {
	counts := base
	if overrides.PlaceCount != nil {
		counts.PlaceCount = *overrides.PlaceCount
	}
	if overrides.MonsterCount != nil {
		counts.MonsterCount = *overrides.MonsterCount
	}
	if overrides.BossEncounterCount != nil {
		counts.BossEncounterCount = *overrides.BossEncounterCount
	}
	if overrides.RaidEncounterCount != nil {
		counts.RaidEncounterCount = *overrides.RaidEncounterCount
	}
	if overrides.InputEncounterCount != nil {
		counts.InputEncounterCount = *overrides.InputEncounterCount
	}
	if overrides.OptionEncounterCount != nil {
		counts.OptionEncounterCount = *overrides.OptionEncounterCount
	}
	if overrides.TreasureChestCount != nil {
		counts.TreasureChestCount = *overrides.TreasureChestCount
	}
	if overrides.HealingFountainCount != nil {
		counts.HealingFountainCount = *overrides.HealingFountainCount
	}
	return counts
}

func zoneSeedCountsToNormalizedRequest(
	mode string,
	counts models.ZoneSeedResolvedCounts,
	requiredTags []string,
	shopkeeperTags []string,
	autoSeedAudit models.ZoneSeedAutoAudit,
) *normalizedZoneSeedDraftRequest {
	return &normalizedZoneSeedDraftRequest{
		SeedMode:             mode,
		PlaceCount:           counts.PlaceCount,
		MonsterCount:         counts.MonsterCount,
		BossEncounterCount:   counts.BossEncounterCount,
		RaidEncounterCount:   counts.RaidEncounterCount,
		InputEncounterCount:  counts.InputEncounterCount,
		OptionEncounterCount: counts.OptionEncounterCount,
		TreasureChestCount:   counts.TreasureChestCount,
		HealingFountainCount: counts.HealingFountainCount,
		RequiredPlaceTags:    requiredTags,
		ShopkeeperItemTags:   shopkeeperTags,
		AutoSeedAudit:        autoSeedAudit,
	}
}

func zoneSeedAreaForAudit(zone *models.Zone) (float64, float64, error) {
	if zone == nil {
		return 0, 0, fmt.Errorf("zone is required")
	}

	polygon := zone.GetPolygon()
	if polygon == nil || len(polygon) == 0 {
		return 0, 0, fmt.Errorf("zone boundary is missing or invalid")
	}

	areaSquareMeters := geo.Area(polygon)
	if areaSquareMeters <= 0 {
		return 0, 0, fmt.Errorf("zone boundary area must be greater than zero")
	}

	areaSquareFeet := areaSquareMeters * zoneSeedAutoSquareFeetPerSquareMeter
	areaAcres := areaSquareFeet / zoneSeedAutoSquareFeetPerAcre
	return areaSquareFeet, areaAcres, nil
}

func zoneSeedAutoCurveCount(areaAcres float64, multiplier float64) int {
	normalizedArea := math.Max(areaAcres, 0)
	value := int(math.Round(math.Log1p(normalizedArea) * multiplier))
	if value < 1 {
		return 1
	}
	return value
}

func zoneSeedInferAutoCounts(
	areaAcres float64,
	requiredTags []string,
) (models.ZoneSeedResolvedCounts, models.StringArray) {
	counts := models.ZoneSeedResolvedCounts{
		PlaceCount:           zoneSeedAutoCurveCount(areaAcres, 2.75),
		MonsterCount:         zoneSeedAutoCurveCount(areaAcres, 1.9),
		BossEncounterCount:   zoneSeedAutoCurveCount(areaAcres, 0.85),
		RaidEncounterCount:   zoneSeedAutoCurveCount(areaAcres, 0.55),
		InputEncounterCount:  zoneSeedAutoCurveCount(areaAcres, 1.1),
		OptionEncounterCount: zoneSeedAutoCurveCount(areaAcres, 1.1),
		TreasureChestCount:   zoneSeedAutoCurveCount(areaAcres, 1.35),
		HealingFountainCount: zoneSeedAutoCurveCount(areaAcres, 0.75),
	}

	warnings := models.StringArray{}
	if len(requiredTags) > counts.PlaceCount {
		warnings = append(
			warnings,
			fmt.Sprintf(
				"Increased POI recommendation from %d to %d to cover %d required place tags.",
				counts.PlaceCount,
				len(requiredTags),
				len(requiredTags),
			),
		)
		counts.PlaceCount = len(requiredTags)
	}

	return counts, warnings
}

func zoneSeedAutoPlaceSearchRadius(zone models.Zone) float64 {
	polygon := zone.GetPolygon()
	if polygon == nil {
		return 0
	}
	bounds := polygon.Bound()
	if bounds.IsEmpty() {
		return 0
	}

	centerLng := (bounds.Min.X() + bounds.Max.X()) / 2
	centerLat := (bounds.Min.Y() + bounds.Max.Y()) / 2
	corners := [][2]float64{
		{bounds.Min.Y(), bounds.Min.X()},
		{bounds.Min.Y(), bounds.Max.X()},
		{bounds.Max.Y(), bounds.Min.X()},
		{bounds.Max.Y(), bounds.Max.X()},
	}

	maxDistance := 0.0
	for _, corner := range corners {
		distance := util.HaversineDistance(centerLat, centerLng, corner[0], corner[1])
		if distance > maxDistance {
			maxDistance = distance
		}
	}

	if maxDistance < 500 {
		return 500
	}
	if maxDistance > 5000 {
		return 5000
	}
	return maxDistance
}

func normalizeZoneSeedAutoPlaceTypes(place googlemaps.Place) []string {
	types := make([]string, 0, len(place.Types)+1)
	for _, placeType := range place.Types {
		normalized := strings.ToLower(strings.TrimSpace(placeType))
		if normalized != "" {
			types = append(types, normalized)
		}
	}
	if primaryType := strings.ToLower(strings.TrimSpace(place.PrimaryType)); primaryType != "" {
		types = append(types, primaryType)
	}
	return types
}

func zoneSeedAutoPlaceHasAnyType(types []string, candidates []string) bool {
	for _, placeType := range types {
		for _, candidate := range candidates {
			if placeType == candidate || strings.Contains(placeType, candidate) {
				return true
			}
		}
	}
	return false
}

func isZoneSeedAutoEnjoyablePlace(place googlemaps.Place) bool {
	lowerTypes := normalizeZoneSeedAutoPlaceTypes(place)
	if len(lowerTypes) == 0 {
		return false
	}

	blocked := []string{
		"accounting", "insurance_agency", "real_estate_agency", "lawyer",
		"bank", "atm", "police", "fire_station", "post_office", "courthouse",
		"city_hall", "local_government_office",
		"doctor", "dentist", "hospital", "health", "pharmacy", "drugstore", "veterinary_care",
		"school", "primary_school", "secondary_school", "university",
		"locksmith", "hardware_store", "home_improvement_store",
		"gas_station", "car_repair", "car_wash", "parking", "storage",
		"train_station", "subway_station", "bus_station", "transit_station", "airport",
		"lodging",
	}
	if zoneSeedAutoPlaceHasAnyType(lowerTypes, blocked) {
		return false
	}

	allowed := []string{
		"cafe", "coffee_shop", "bakery", "restaurant", "bar", "meal_takeaway", "meal_delivery",
		"ice_cream_shop", "dessert",
		"park", "garden", "playground", "trail", "hiking_area", "campground", "natural_feature",
		"beach", "marina", "harbor", "water",
		"museum", "art_gallery", "gallery", "tourist_attraction", "point_of_interest", "landmark",
		"library", "book_store", "movie_theater", "theater", "night_club", "music_venue",
		"stadium", "sports_complex", "gym",
		"zoo", "aquarium", "amusement_park", "bowling_alley",
		"plaza", "square", "bridge",
		"store", "shopping_mall", "market", "supermarket", "convenience_store",
		"clothing_store", "department_store", "electronics_store", "furniture_store", "home_goods_store",
		"pet_store", "florist", "spa", "beauty_salon", "barber", "hair_care",
	}
	if zoneSeedAutoPlaceHasAnyType(lowerTypes, allowed) {
		return true
	}

	display := strings.ToLower(strings.TrimSpace(place.PrimaryTypeDisplayName.Text))
	if display == "" {
		return false
	}

	keywords := []string{
		"cafe", "coffee", "bakery", "restaurant", "bar", "park", "garden", "museum",
		"gallery", "library", "book", "theater", "cinema", "market", "shop", "store",
		"plaza", "square", "trail", "beach", "zoo", "aquarium", "amusement",
	}
	for _, keyword := range keywords {
		if strings.Contains(display, keyword) {
			return true
		}
	}

	return false
}

func zoneSeedAutoExpandRequiredTagAliases(tag string) []string {
	aliases := map[string][]string{
		"coffee_shop":   {"coffee_shop", "cafe", "coffee", "bakery", "espresso", "tea"},
		"cafe":          {"cafe", "coffee_shop", "coffee", "bakery", "espresso", "tea"},
		"park":          {"park", "garden", "playground", "trail", "hiking", "natural_feature"},
		"museum":        {"museum", "gallery", "art_gallery", "exhibit"},
		"gallery":       {"gallery", "art_gallery", "museum"},
		"library":       {"library", "book_store", "bookstore", "book"},
		"book_store":    {"book_store", "bookstore", "library", "book"},
		"market":        {"market", "shopping_mall", "store", "shopping", "supermarket"},
		"shopping":      {"shopping_mall", "store", "market", "shopping", "clothing_store"},
		"scenic":        {"plaza", "square", "bridge", "beach", "marina", "harbor", "view"},
		"theater":       {"theater", "movie_theater", "cinema"},
		"entertainment": {"theater", "movie_theater", "cinema", "music", "stadium", "amusement", "zoo", "aquarium"},
	}
	if expanded, ok := aliases[tag]; ok {
		return expanded
	}
	return []string{tag}
}

func zoneSeedAutoPlaceMatchesTag(place googlemaps.Place, tag string) bool {
	needle := strings.ToLower(strings.TrimSpace(tag))
	if needle == "" {
		return false
	}

	aliases := zoneSeedAutoExpandRequiredTagAliases(needle)
	for _, placeType := range normalizeZoneSeedAutoPlaceTypes(place) {
		for _, alias := range aliases {
			if placeType == alias || strings.Contains(placeType, alias) {
				return true
			}
		}
	}

	display := strings.ToLower(strings.TrimSpace(place.PrimaryTypeDisplayName.Text))
	for _, alias := range aliases {
		if display != "" && strings.Contains(display, alias) {
			return true
		}
	}

	name := strings.ToLower(strings.TrimSpace(place.DisplayName.Text))
	for _, alias := range aliases {
		if name != "" && strings.Contains(name, alias) {
			return true
		}
	}

	return false
}

func zoneSeedAutoRequiredTagPlaceTypes(tag string) []googlemaps.PlaceType {
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "coffee_shop", "cafe", "coffee":
		return []googlemaps.PlaceType{
			googlemaps.TypeCoffeeShop,
			googlemaps.TypeCafe,
			googlemaps.TypeTeaHouse,
			googlemaps.TypeBakery,
		}
	case "park", "garden", "playground", "trail", "hiking", "natural_feature":
		return []googlemaps.PlaceType{
			googlemaps.TypePark,
			googlemaps.TypeGarden,
			googlemaps.TypePlayground,
			googlemaps.TypeDogPark,
			googlemaps.TypeHikingArea,
			googlemaps.TypePicnicGround,
			googlemaps.TypeBotanicalGarden,
		}
	case "museum", "gallery", "art_gallery":
		return []googlemaps.PlaceType{
			googlemaps.TypeMuseum,
			googlemaps.TypeArtGallery,
			googlemaps.TypeTouristAttraction,
		}
	case "library", "book_store", "bookstore":
		return []googlemaps.PlaceType{
			googlemaps.TypeLibrary,
			googlemaps.TypeBookStore,
		}
	case "market", "shopping", "store", "shopping_mall":
		return []googlemaps.PlaceType{
			googlemaps.TypeMarket,
			googlemaps.TypeShoppingMall,
			googlemaps.TypeStore,
			googlemaps.TypeSupermarket,
			googlemaps.TypeClothingStore,
		}
	case "scenic", "plaza", "square", "beach", "marina":
		return []googlemaps.PlaceType{
			googlemaps.TypePlaza,
			googlemaps.TypeBeach,
			googlemaps.TypeMarina,
		}
	case "theater", "cinema", "movie_theater":
		return []googlemaps.PlaceType{
			googlemaps.TypeMovieTheater,
			googlemaps.TypePerformingArtsTheater,
		}
	case "entertainment", "music_venue":
		return []googlemaps.PlaceType{
			googlemaps.TypeConcertHall,
			googlemaps.TypeAmphitheatre,
			googlemaps.TypeNightClub,
			googlemaps.TypeStadium,
			googlemaps.TypeAmusementPark,
			googlemaps.TypeZoo,
			googlemaps.TypeAquarium,
		}
	default:
		return nil
	}
}

func zoneSeedAutoEstimateEligiblePlaceCount(
	ctx context.Context,
	zone models.Zone,
	desired int,
	requiredTags []string,
	googlemapsClient googlemaps.Client,
) (int, models.StringArray, error) {
	_ = ctx
	if desired <= 0 {
		return 0, models.StringArray{}, nil
	}

	radius := zoneSeedAutoPlaceSearchRadius(zone)
	if radius <= 0 {
		return 0, models.StringArray{}, nil
	}

	seen := make(map[string]googlemaps.Place)
	maxAttempts := 6
	if len(requiredTags) > 0 {
		maxAttempts = 10
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		point := zone.GetRandomPoint()
		if point.X() == 0 && point.Y() == 0 {
			continue
		}

		places, err := googlemapsClient.FindPlaces(googlemaps.PlaceQuery{
			Lat:            point.Y(),
			Long:           point.X(),
			Radius:         radius,
			MaxResultCount: 20,
			RankPreference: googlemaps.RankPreferencePopularity,
		})
		if err != nil {
			return 0, nil, err
		}

		for _, place := range places {
			if place.ID == "" {
				continue
			}
			if !zone.IsPointInBoundary(place.Location.Latitude, place.Location.Longitude) {
				continue
			}
			if !isZoneSeedAutoEnjoyablePlace(place) && !zoneSeedAutoPlaceMatchesAnyTag(place, requiredTags) {
				continue
			}
			if _, ok := seen[place.ID]; ok {
				continue
			}
			seen[place.ID] = place
		}
	}

	warnings := models.StringArray{}
	for _, tag := range requiredTags {
		types := zoneSeedAutoRequiredTagPlaceTypes(tag)
		if len(types) == 0 {
			warnings = append(warnings, fmt.Sprintf("No targeted place search mapping exists for required tag %q.", tag))
			continue
		}

		found := false
		for _, place := range seen {
			if zoneSeedAutoPlaceMatchesTag(place, tag) {
				found = true
				break
			}
		}
		if found {
			continue
		}

		for attempt := 0; attempt < 3; attempt++ {
			point := zone.GetRandomPoint()
			if point.X() == 0 && point.Y() == 0 {
				continue
			}

			places, err := googlemapsClient.FindPlaces(googlemaps.PlaceQuery{
				Lat:            point.Y(),
				Long:           point.X(),
				Radius:         radius,
				MaxResultCount: 20,
				IncludedTypes:  types,
				RankPreference: googlemaps.RankPreferencePopularity,
			})
			if err != nil {
				return 0, nil, err
			}

			for _, place := range places {
				if place.ID == "" {
					continue
				}
				if !zone.IsPointInBoundary(place.Location.Latitude, place.Location.Longitude) {
					continue
				}
				seen[place.ID] = place
				if zoneSeedAutoPlaceMatchesTag(place, tag) {
					found = true
				}
			}

			if found {
				break
			}
		}

		if !found {
			warnings = append(warnings, fmt.Sprintf("No eligible places were found for required tag %q during auto inference.", tag))
		}
	}

	return len(seen), warnings, nil
}

func zoneSeedAutoPlaceMatchesAnyTag(place googlemaps.Place, tags []string) bool {
	for _, tag := range tags {
		if zoneSeedAutoPlaceMatchesTag(place, tag) {
			return true
		}
	}
	return false
}

func (s *server) resolveZoneSeedDraftRequest(
	ctx context.Context,
	zone *models.Zone,
	requestBody zoneSeedDraftRequest,
) (*normalizedZoneSeedDraftRequest, error) {
	mode, err := normalizeZoneSeedDraftMode(requestBody.SeedMode)
	if err != nil {
		return nil, newZoneSeedDraftResolutionError(400, err)
	}

	if mode == models.ZoneSeedModeManual {
		settings, err := normalizeZoneSeedDraftRequest(requestBody)
		if err != nil {
			return nil, newZoneSeedDraftResolutionError(400, err)
		}
		return settings, nil
	}

	if zone == nil {
		return nil, newZoneSeedDraftResolutionError(400, fmt.Errorf("zone is required"))
	}

	requiredTags := normalizeZoneSeedDraftTags(requestBody.RequiredPlaceTags)
	shopkeeperTags := normalizeZoneSeedDraftTags(requestBody.ShopkeeperItemTags)
	overrides, err := zoneSeedDraftCountOverridesFromRequest(requestBody)
	if err != nil {
		return nil, newZoneSeedDraftResolutionError(400, err)
	}

	areaSquareFeet, areaAcres, err := zoneSeedAreaForAudit(zone)
	if err != nil {
		return nil, newZoneSeedDraftResolutionError(400, err)
	}

	recommendedCounts, warnings := zoneSeedInferAutoCounts(areaAcres, requiredTags)
	eligiblePlaceCount, estimateWarnings, err := zoneSeedAutoEstimateEligiblePlaceCount(
		ctx,
		*zone,
		recommendedCounts.PlaceCount,
		requiredTags,
		s.googlemapsClient,
	)
	if err != nil {
		return nil, newZoneSeedDraftResolutionError(500, fmt.Errorf("failed to estimate eligible places for auto mode: %w", err))
	}
	warnings = append(warnings, estimateWarnings...)
	if recommendedCounts.PlaceCount > eligiblePlaceCount {
		warnings = append(
			warnings,
			fmt.Sprintf(
				"Capped POI recommendation from %d to %d because only %d eligible places were found during auto inference.",
				recommendedCounts.PlaceCount,
				eligiblePlaceCount,
				eligiblePlaceCount,
			),
		)
		recommendedCounts.PlaceCount = eligiblePlaceCount
	}

	finalCounts := zoneSeedApplyCountOverrides(recommendedCounts, overrides)
	if finalCounts.PlaceCount > 0 && len(requiredTags) > finalCounts.PlaceCount {
		return nil, newZoneSeedDraftResolutionError(400, fmt.Errorf("requiredPlaceTags cannot exceed placeCount"))
	}
	if finalCounts.PlaceCount == 0 {
		requiredTags = []string{}
	}

	return zoneSeedCountsToNormalizedRequest(
		mode,
		finalCounts,
		requiredTags,
		shopkeeperTags,
		models.ZoneSeedAutoAudit{
			ZoneAreaSquareFeet: areaSquareFeet,
			ZoneAreaAcres:      areaAcres,
			RecommendedCounts:  recommendedCounts,
			FinalCounts:        finalCounts,
			Warnings:           warnings,
		},
	), nil
}
