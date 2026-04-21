package server

import (
	"context"
	stdErrors "errors"
	"fmt"
	"math"
	"strings"

	"github.com/MaxBlaushild/poltergeist/pkg/googlemaps"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/paulmach/orb/geo"
	"gorm.io/gorm"
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
	ResourceCount        *int
}

type zoneSeedCurrentContentSnapshot struct {
	ExistingCounts             models.ZoneSeedResolvedCounts
	RemainingRequiredPlaceTags []string
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

func normalizeZoneSeedDraftCountMode(mode string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		return models.ZoneSeedCountModeAbsolute, nil
	}
	if !models.IsValidZoneSeedCountMode(normalized) {
		return "", fmt.Errorf("invalid zone seed count mode: %s", mode)
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
	if overrides.ResourceCount, err = copyCount(requestBody.ResourceCount, "resourceCount"); err != nil {
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
	if overrides.ResourceCount != nil {
		counts.ResourceCount = *overrides.ResourceCount
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
	if overrides.ResourceCount != nil {
		counts.ResourceCount = *overrides.ResourceCount
	}
	return counts
}

func zoneSeedCountsToNormalizedRequest(
	zoneKind string,
	mode string,
	countMode string,
	counts models.ZoneSeedResolvedCounts,
	requiredTags []string,
	shopkeeperTags []string,
	autoSeedAudit models.ZoneSeedAutoAudit,
	countAudit models.ZoneSeedCountAudit,
) *normalizedZoneSeedDraftRequest {
	return &normalizedZoneSeedDraftRequest{
		SeedMode:             mode,
		CountMode:            countMode,
		ZoneKind:             zoneKind,
		PlaceCount:           counts.PlaceCount,
		MonsterCount:         counts.MonsterCount,
		BossEncounterCount:   counts.BossEncounterCount,
		RaidEncounterCount:   counts.RaidEncounterCount,
		InputEncounterCount:  counts.InputEncounterCount,
		OptionEncounterCount: counts.OptionEncounterCount,
		TreasureChestCount:   counts.TreasureChestCount,
		HealingFountainCount: counts.HealingFountainCount,
		ResourceCount:        counts.ResourceCount,
		RequiredPlaceTags:    requiredTags,
		ShopkeeperItemTags:   shopkeeperTags,
		AutoSeedAudit:        autoSeedAudit,
		CountAudit:           countAudit,
	}
}

func zoneSeedCountsFromNormalizedRequest(settings *normalizedZoneSeedDraftRequest) models.ZoneSeedResolvedCounts {
	if settings == nil {
		return models.ZoneSeedResolvedCounts{}
	}
	return models.ZoneSeedResolvedCounts{
		PlaceCount:           settings.PlaceCount,
		MonsterCount:         settings.MonsterCount,
		BossEncounterCount:   settings.BossEncounterCount,
		RaidEncounterCount:   settings.RaidEncounterCount,
		InputEncounterCount:  settings.InputEncounterCount,
		OptionEncounterCount: settings.OptionEncounterCount,
		TreasureChestCount:   settings.TreasureChestCount,
		HealingFountainCount: settings.HealingFountainCount,
		ResourceCount:        settings.ResourceCount,
	}
}

func zoneSeedSubtractExistingCounts(
	target models.ZoneSeedResolvedCounts,
	existing models.ZoneSeedResolvedCounts,
) models.ZoneSeedResolvedCounts {
	clamp := func(value int) int {
		if value < 0 {
			return 0
		}
		return value
	}

	return models.ZoneSeedResolvedCounts{
		PlaceCount:           clamp(target.PlaceCount - existing.PlaceCount),
		MonsterCount:         clamp(target.MonsterCount - existing.MonsterCount),
		BossEncounterCount:   clamp(target.BossEncounterCount - existing.BossEncounterCount),
		RaidEncounterCount:   clamp(target.RaidEncounterCount - existing.RaidEncounterCount),
		InputEncounterCount:  clamp(target.InputEncounterCount - existing.InputEncounterCount),
		OptionEncounterCount: clamp(target.OptionEncounterCount - existing.OptionEncounterCount),
		TreasureChestCount:   clamp(target.TreasureChestCount - existing.TreasureChestCount),
		HealingFountainCount: clamp(target.HealingFountainCount - existing.HealingFountainCount),
		ResourceCount:        clamp(target.ResourceCount - existing.ResourceCount),
	}
}

func zoneSeedCurrentAwareWarnings(
	target models.ZoneSeedResolvedCounts,
	existing models.ZoneSeedResolvedCounts,
	queued models.ZoneSeedResolvedCounts,
) models.StringArray {
	type countWarning struct {
		label    string
		target   int
		existing int
		queued   int
	}

	items := []countWarning{
		{label: "POIs", target: target.PlaceCount, existing: existing.PlaceCount, queued: queued.PlaceCount},
		{label: "monster encounters", target: target.MonsterCount, existing: existing.MonsterCount, queued: queued.MonsterCount},
		{label: "boss encounters", target: target.BossEncounterCount, existing: existing.BossEncounterCount, queued: queued.BossEncounterCount},
		{label: "raid encounters", target: target.RaidEncounterCount, existing: existing.RaidEncounterCount, queued: queued.RaidEncounterCount},
		{label: "input scenarios", target: target.InputEncounterCount, existing: existing.InputEncounterCount, queued: queued.InputEncounterCount},
		{label: "option scenarios", target: target.OptionEncounterCount, existing: existing.OptionEncounterCount, queued: queued.OptionEncounterCount},
		{label: "treasure chests", target: target.TreasureChestCount, existing: existing.TreasureChestCount, queued: queued.TreasureChestCount},
		{label: "healing fountains", target: target.HealingFountainCount, existing: existing.HealingFountainCount, queued: queued.HealingFountainCount},
		{label: "resources", target: target.ResourceCount, existing: existing.ResourceCount, queued: queued.ResourceCount},
	}

	warnings := models.StringArray{}
	for _, item := range items {
		if item.target <= 0 || item.existing <= 0 || item.queued == item.target {
			continue
		}
		if item.queued == 0 {
			warnings = append(
				warnings,
				fmt.Sprintf(
					"Skipped new %s because the zone already has %d and the target is %d.",
					item.label,
					item.existing,
					item.target,
				),
			)
			continue
		}
		warnings = append(
			warnings,
			fmt.Sprintf(
				"Reduced queued %s from %d to %d because %d already exist in the zone.",
				item.label,
				item.target,
				item.queued,
				item.existing,
			),
		)
	}

	return warnings
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
		ResourceCount:        zoneSeedAutoCurveCount(areaAcres, 1.6),
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

func zoneSeedEffectiveKind(zone *models.Zone, requestedKind string) string {
	if normalized := models.NormalizeZoneKind(requestedKind); normalized != "" {
		return normalized
	}
	if zone == nil {
		return ""
	}
	return models.NormalizeZoneKind(zone.Kind)
}

func (s *server) applyZoneKindRatios(
	ctx context.Context,
	zoneKindSlug string,
	counts models.ZoneSeedResolvedCounts,
	warnings models.StringArray,
) (models.ZoneSeedResolvedCounts, models.StringArray, error) {
	normalizedZoneKind := models.NormalizeZoneKind(zoneKindSlug)
	if normalizedZoneKind == "" {
		return counts, warnings, nil
	}

	zoneKind, err := s.dbClient.ZoneKind().FindBySlug(ctx, normalizedZoneKind)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			warnings = append(
				warnings,
				fmt.Sprintf(
					`Zone kind "%s" is not defined yet, so baseline auto counts were used.`,
					normalizedZoneKind,
				),
			)
			return counts, warnings, nil
		}
		return counts, warnings, err
	}

	return zoneKind.ApplyToCounts(counts), append(
		warnings,
		fmt.Sprintf(
			`Applied zone kind ratios from "%s" to the auto seed recommendation.`,
			zoneKind.Name,
		),
	), nil
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

func isZoneSeedAutoFallbackPlace(place googlemaps.Place) bool {
	lowerTypes := normalizeZoneSeedAutoPlaceTypes(place)
	if len(lowerTypes) == 0 {
		return strings.TrimSpace(place.DisplayName.Text) != ""
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
	return !zoneSeedAutoPlaceHasAnyType(lowerTypes, blocked)
}

func isZoneSeedAutoEnjoyablePlace(place googlemaps.Place) bool {
	lowerTypes := normalizeZoneSeedAutoPlaceTypes(place)
	if len(lowerTypes) == 0 {
		return false
	}

	if !isZoneSeedAutoFallbackPlace(place) {
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

func zoneSeedAutoPlaceSearchAttemptLimit(desired int, hasRequiredTags bool) int {
	attempts := 6
	if desired > attempts {
		attempts = desired
	}
	if attempts > 18 {
		attempts = 18
	}
	if hasRequiredTags && attempts < 12 {
		attempts = 12
	}
	return attempts
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
	maxAttempts := zoneSeedAutoPlaceSearchAttemptLimit(desired, len(requiredTags) > 0)

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
			if !isZoneSeedAutoFallbackPlace(place) && !zoneSeedAutoPlaceMatchesAnyTag(place, requiredTags) {
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

func zoneSeedPointOfInterestMarkerTagHints(category models.PointOfInterestMarkerCategory) []string {
	switch models.NormalizePointOfInterestMarkerCategory(string(category)) {
	case models.PointOfInterestMarkerCategoryCoffeehouse:
		return []string{"coffeehouse", "cafe", "coffee_shop", "coffee", "bakery", "espresso", "tea"}
	case models.PointOfInterestMarkerCategoryTavern:
		return []string{"tavern", "bar", "pub", "cocktail", "wine", "beer", "restaurant"}
	case models.PointOfInterestMarkerCategoryEatery:
		return []string{"eatery", "restaurant", "bakery", "dessert", "ice_cream", "food"}
	case models.PointOfInterestMarkerCategoryMarket:
		return []string{"market", "shopping_mall", "store", "shopping", "supermarket", "clothing_store", "florist"}
	case models.PointOfInterestMarkerCategoryArchive:
		return []string{"archive", "library", "book_store", "bookstore", "book"}
	case models.PointOfInterestMarkerCategoryPark:
		return []string{"park", "garden", "playground", "trail", "hiking", "natural_feature"}
	case models.PointOfInterestMarkerCategoryWaterfront:
		return []string{"waterfront", "marina", "beach", "harbor", "scenic"}
	case models.PointOfInterestMarkerCategoryMuseum:
		return []string{"museum", "gallery", "art_gallery", "exhibit"}
	case models.PointOfInterestMarkerCategoryTheater:
		return []string{"theater", "movie_theater", "cinema", "music", "entertainment"}
	case models.PointOfInterestMarkerCategoryLandmark:
		return []string{"landmark", "plaza", "square", "bridge", "view", "scenic"}
	case models.PointOfInterestMarkerCategoryArena:
		return []string{"arena", "stadium", "sports_complex", "entertainment"}
	default:
		return nil
	}
}

func zoneSeedPointOfInterestMatchesTag(pointOfInterest models.PointOfInterest, tag string) bool {
	needle := strings.ToLower(strings.TrimSpace(tag))
	if needle == "" {
		return false
	}

	aliases := zoneSeedAutoExpandRequiredTagAliases(needle)
	candidates := make([]string, 0, len(pointOfInterest.Tags)+6)
	candidates = append(candidates, zoneSeedPointOfInterestMarkerTagHints(pointOfInterest.MarkerCategory)...)
	if pointOfInterest.GoogleMapsPlaceName != nil {
		candidates = append(candidates, *pointOfInterest.GoogleMapsPlaceName)
	}
	candidates = append(candidates, pointOfInterest.Name, pointOfInterest.OriginalName)
	for _, poiTag := range pointOfInterest.Tags {
		candidates = append(candidates, poiTag.Value)
	}

	for _, candidate := range candidates {
		normalizedCandidate := strings.ToLower(strings.TrimSpace(candidate))
		if normalizedCandidate == "" {
			continue
		}
		for _, alias := range aliases {
			if normalizedCandidate == alias ||
				strings.Contains(normalizedCandidate, alias) ||
				strings.Contains(alias, normalizedCandidate) {
				return true
			}
		}
	}

	return false
}

func zoneSeedRemainingRequiredPlaceTags(
	requiredTags []string,
	pointsOfInterest []models.PointOfInterest,
) []string {
	if len(requiredTags) == 0 {
		return []string{}
	}

	remaining := make([]string, 0, len(requiredTags))
	for _, tag := range requiredTags {
		matched := false
		for _, pointOfInterest := range pointsOfInterest {
			if zoneSeedPointOfInterestMatchesTag(pointOfInterest, tag) {
				matched = true
				break
			}
		}
		if !matched {
			remaining = append(remaining, tag)
		}
	}

	return remaining
}

func (s *server) zoneSeedCurrentContentSnapshot(
	ctx context.Context,
	zone *models.Zone,
	requiredTags []string,
) (zoneSeedCurrentContentSnapshot, error) {
	if zone == nil {
		return zoneSeedCurrentContentSnapshot{}, fmt.Errorf("zone is required")
	}

	pointsOfInterest, err := s.dbClient.PointOfInterest().FindAllForZone(ctx, zone.ID)
	if err != nil {
		return zoneSeedCurrentContentSnapshot{}, err
	}

	encounters, err := s.dbClient.MonsterEncounter().FindByZoneIDExcludingQuestNodes(ctx, zone.ID)
	if err != nil {
		return zoneSeedCurrentContentSnapshot{}, err
	}

	scenarios, err := s.dbClient.Scenario().FindByZoneIDExcludingQuestNodes(ctx, zone.ID)
	if err != nil {
		return zoneSeedCurrentContentSnapshot{}, err
	}

	treasureChests, err := s.dbClient.TreasureChest().FindByZoneID(ctx, zone.ID)
	if err != nil {
		return zoneSeedCurrentContentSnapshot{}, err
	}

	healingFountains, err := s.dbClient.HealingFountain().FindByZoneID(ctx, zone.ID)
	if err != nil {
		return zoneSeedCurrentContentSnapshot{}, err
	}

	resources, err := s.dbClient.Resource().FindByZoneID(ctx, zone.ID)
	if err != nil {
		return zoneSeedCurrentContentSnapshot{}, err
	}

	counts := models.ZoneSeedResolvedCounts{
		PlaceCount:           len(pointsOfInterest),
		TreasureChestCount:   len(treasureChests),
		HealingFountainCount: len(healingFountains),
		ResourceCount:        len(resources),
	}

	for _, encounter := range encounters {
		switch encounter.EncounterType {
		case models.MonsterEncounterTypeBoss:
			counts.BossEncounterCount++
		case models.MonsterEncounterTypeRaid:
			counts.RaidEncounterCount++
		default:
			counts.MonsterCount++
		}
	}

	for _, scenario := range scenarios {
		if scenario.OpenEnded {
			counts.InputEncounterCount++
		} else {
			counts.OptionEncounterCount++
		}
	}

	return zoneSeedCurrentContentSnapshot{
		ExistingCounts:             counts,
		RemainingRequiredPlaceTags: zoneSeedRemainingRequiredPlaceTags(requiredTags, pointsOfInterest),
	}, nil
}

func zoneSeedResolveCurrentAwareCounts(
	targetCounts models.ZoneSeedResolvedCounts,
	snapshot zoneSeedCurrentContentSnapshot,
) (models.ZoneSeedResolvedCounts, models.ZoneSeedCountAudit) {
	queuedCounts := zoneSeedSubtractExistingCounts(targetCounts, snapshot.ExistingCounts)
	warnings := zoneSeedCurrentAwareWarnings(targetCounts, snapshot.ExistingCounts, queuedCounts)
	if len(snapshot.RemainingRequiredPlaceTags) > queuedCounts.PlaceCount {
		warnings = append(
			warnings,
			fmt.Sprintf(
				"Increased queued POIs from %d to %d because %d required place tags are still unmet by existing POIs.",
				queuedCounts.PlaceCount,
				len(snapshot.RemainingRequiredPlaceTags),
				len(snapshot.RemainingRequiredPlaceTags),
			),
		)
		queuedCounts.PlaceCount = len(snapshot.RemainingRequiredPlaceTags)
	}

	remainingRequiredTags := append([]string{}, snapshot.RemainingRequiredPlaceTags...)
	if queuedCounts.PlaceCount == 0 {
		remainingRequiredTags = []string{}
	}

	return queuedCounts, models.ZoneSeedCountAudit{
		TargetCounts:               targetCounts,
		ExistingCounts:             snapshot.ExistingCounts,
		QueuedCounts:               queuedCounts,
		RemainingRequiredPlaceTags: models.StringArray(remainingRequiredTags),
		Warnings:                   warnings,
	}
}

func (s *server) applyZoneSeedCurrentAwareMode(
	ctx context.Context,
	zone *models.Zone,
	settings *normalizedZoneSeedDraftRequest,
) (*normalizedZoneSeedDraftRequest, error) {
	if settings == nil {
		return nil, fmt.Errorf("zone seed settings are required")
	}
	if settings.CountMode != models.ZoneSeedCountModeCurrentAware {
		return settings, nil
	}
	if zone == nil {
		return nil, fmt.Errorf("zone is required")
	}

	targetCounts := zoneSeedCountsFromNormalizedRequest(settings)
	snapshot, err := s.zoneSeedCurrentContentSnapshot(ctx, zone, settings.RequiredPlaceTags)
	if err != nil {
		return nil, err
	}

	queuedCounts, countAudit := zoneSeedResolveCurrentAwareCounts(targetCounts, snapshot)
	return zoneSeedCountsToNormalizedRequest(
		settings.ZoneKind,
		settings.SeedMode,
		settings.CountMode,
		queuedCounts,
		settings.RequiredPlaceTags,
		settings.ShopkeeperItemTags,
		settings.AutoSeedAudit,
		countAudit,
	), nil
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
	effectiveZoneKind := zoneSeedEffectiveKind(zone, requestBody.ZoneKind)

	if mode == models.ZoneSeedModeManual {
		settings, err := normalizeZoneSeedDraftRequest(requestBody)
		if err != nil {
			return nil, newZoneSeedDraftResolutionError(400, err)
		}
		if settings.ZoneKind == "" {
			settings.ZoneKind = effectiveZoneKind
		}
		settings, err = s.applyZoneSeedCurrentAwareMode(ctx, zone, settings)
		if err != nil {
			return nil, newZoneSeedDraftResolutionError(500, err)
		}
		return settings, nil
	}

	if zone == nil {
		return nil, newZoneSeedDraftResolutionError(400, fmt.Errorf("zone is required"))
	}

	requiredTags := normalizeZoneSeedDraftTags(requestBody.RequiredPlaceTags)
	shopkeeperTags := normalizeZoneSeedDraftTags(requestBody.ShopkeeperItemTags)
	countMode, err := normalizeZoneSeedDraftCountMode(requestBody.CountMode)
	if err != nil {
		return nil, newZoneSeedDraftResolutionError(400, err)
	}
	overrides, err := zoneSeedDraftCountOverridesFromRequest(requestBody)
	if err != nil {
		return nil, newZoneSeedDraftResolutionError(400, err)
	}

	areaSquareFeet, areaAcres, err := zoneSeedAreaForAudit(zone)
	if err != nil {
		return nil, newZoneSeedDraftResolutionError(400, err)
	}

	recommendedCounts, warnings := zoneSeedInferAutoCounts(areaAcres, requiredTags)
	recommendedCounts, warnings, err = s.applyZoneKindRatios(ctx, effectiveZoneKind, recommendedCounts, warnings)
	if err != nil {
		return nil, newZoneSeedDraftResolutionError(500, fmt.Errorf("failed to apply zone kind ratios: %w", err))
	}
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
	if overrides.PlaceCount == nil && recommendedCounts.PlaceCount > eligiblePlaceCount {
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

	targetCounts := zoneSeedApplyCountOverrides(recommendedCounts, overrides)
	if targetCounts.PlaceCount > 0 && len(requiredTags) > targetCounts.PlaceCount {
		return nil, newZoneSeedDraftResolutionError(400, fmt.Errorf("requiredPlaceTags cannot exceed placeCount"))
	}
	if targetCounts.PlaceCount == 0 {
		requiredTags = []string{}
	}

	queuedCounts := targetCounts
	countAudit := models.ZoneSeedCountAudit{}
	if countMode == models.ZoneSeedCountModeCurrentAware {
		snapshot, err := s.zoneSeedCurrentContentSnapshot(ctx, zone, requiredTags)
		if err != nil {
			return nil, newZoneSeedDraftResolutionError(500, fmt.Errorf("failed to inspect current zone content: %w", err))
		}
		queuedCounts, countAudit = zoneSeedResolveCurrentAwareCounts(targetCounts, snapshot)
	}

	return zoneSeedCountsToNormalizedRequest(
		effectiveZoneKind,
		mode,
		countMode,
		queuedCounts,
		requiredTags,
		shopkeeperTags,
		models.ZoneSeedAutoAudit{
			ZoneAreaSquareFeet: areaSquareFeet,
			ZoneAreaAcres:      areaAcres,
			RecommendedCounts:  recommendedCounts,
			FinalCounts:        targetCounts,
			Warnings:           warnings,
		},
		countAudit,
	), nil
}
