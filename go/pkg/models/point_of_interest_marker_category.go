package models

import "strings"

type PointOfInterestMarkerCategory string

const (
	PointOfInterestMarkerCategoryGeneric     PointOfInterestMarkerCategory = "generic"
	PointOfInterestMarkerCategoryCoffeehouse PointOfInterestMarkerCategory = "coffeehouse"
	PointOfInterestMarkerCategoryTavern      PointOfInterestMarkerCategory = "tavern"
	PointOfInterestMarkerCategoryEatery      PointOfInterestMarkerCategory = "eatery"
	PointOfInterestMarkerCategoryMarket      PointOfInterestMarkerCategory = "market"
	PointOfInterestMarkerCategoryArchive     PointOfInterestMarkerCategory = "archive"
	PointOfInterestMarkerCategoryPark        PointOfInterestMarkerCategory = "park"
	PointOfInterestMarkerCategoryWaterfront  PointOfInterestMarkerCategory = "waterfront"
	PointOfInterestMarkerCategoryMuseum      PointOfInterestMarkerCategory = "museum"
	PointOfInterestMarkerCategoryTheater     PointOfInterestMarkerCategory = "theater"
	PointOfInterestMarkerCategoryLandmark    PointOfInterestMarkerCategory = "landmark"
	PointOfInterestMarkerCategoryCivic       PointOfInterestMarkerCategory = "civic"
	PointOfInterestMarkerCategoryArena       PointOfInterestMarkerCategory = "arena"
)

func NormalizePointOfInterestMarkerCategory(raw string) PointOfInterestMarkerCategory {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(PointOfInterestMarkerCategoryCoffeehouse):
		return PointOfInterestMarkerCategoryCoffeehouse
	case string(PointOfInterestMarkerCategoryTavern):
		return PointOfInterestMarkerCategoryTavern
	case string(PointOfInterestMarkerCategoryEatery):
		return PointOfInterestMarkerCategoryEatery
	case string(PointOfInterestMarkerCategoryMarket):
		return PointOfInterestMarkerCategoryMarket
	case string(PointOfInterestMarkerCategoryArchive):
		return PointOfInterestMarkerCategoryArchive
	case string(PointOfInterestMarkerCategoryPark):
		return PointOfInterestMarkerCategoryPark
	case string(PointOfInterestMarkerCategoryWaterfront):
		return PointOfInterestMarkerCategoryWaterfront
	case string(PointOfInterestMarkerCategoryMuseum):
		return PointOfInterestMarkerCategoryMuseum
	case string(PointOfInterestMarkerCategoryTheater):
		return PointOfInterestMarkerCategoryTheater
	case string(PointOfInterestMarkerCategoryLandmark):
		return PointOfInterestMarkerCategoryLandmark
	case string(PointOfInterestMarkerCategoryCivic):
		return PointOfInterestMarkerCategoryCivic
	case string(PointOfInterestMarkerCategoryArena):
		return PointOfInterestMarkerCategoryArena
	default:
		return PointOfInterestMarkerCategoryGeneric
	}
}

func InferPointOfInterestMarkerCategoryFromPointOfInterest(
	pointOfInterest *PointOfInterest,
) PointOfInterestMarkerCategory {
	if pointOfInterest == nil {
		return PointOfInterestMarkerCategoryGeneric
	}

	typeHints := make([]string, 0, len(pointOfInterest.Tags))
	for _, tag := range pointOfInterest.Tags {
		if value := strings.TrimSpace(tag.Value); value != "" {
			typeHints = append(typeHints, value)
		}
	}

	displayText := ""
	if pointOfInterest.GoogleMapsPlaceName != nil {
		displayText = strings.TrimSpace(*pointOfInterest.GoogleMapsPlaceName)
	}
	if displayText == "" {
		displayText = strings.TrimSpace(pointOfInterest.OriginalName)
	}

	return InferPointOfInterestMarkerCategory(
		"",
		typeHints,
		displayText,
		false,
		false,
		false,
		false,
		false,
	)
}

func InferPointOfInterestMarkerCategory(
	primaryType string,
	types []string,
	displayText string,
	servesCoffee bool,
	servesBeer bool,
	servesWine bool,
	servesCocktails bool,
	liveMusic bool,
) PointOfInterestMarkerCategory {
	normalizedTypes := normalizedPointOfInterestMarkerTypes(primaryType, types)
	exactCategory := pointOfInterestMarkerCategoryFromExactTypes(normalizedTypes)
	if exactCategory != "" &&
		exactCategory != PointOfInterestMarkerCategoryEatery {
		return exactCategory
	}
	if category := pointOfInterestMarkerCategoryFromFoodDrinkHints(
		normalizedTypes,
		servesCoffee,
		servesBeer,
		servesWine,
		servesCocktails,
		liveMusic,
	); category != "" {
		return category
	}
	if exactCategory != "" {
		return exactCategory
	}
	if category := pointOfInterestMarkerCategoryFromLooseTypes(normalizedTypes); category != "" {
		return category
	}
	if category := pointOfInterestMarkerCategoryFromText(displayText); category != "" {
		return category
	}
	return PointOfInterestMarkerCategoryGeneric
}

func normalizedPointOfInterestMarkerTypes(primaryType string, types []string) []string {
	normalized := make([]string, 0, len(types)+1)
	seen := map[string]struct{}{}

	add := func(raw string) {
		candidate := strings.ToLower(strings.TrimSpace(raw))
		if candidate == "" {
			return
		}
		if _, exists := seen[candidate]; exists {
			return
		}
		seen[candidate] = struct{}{}
		normalized = append(normalized, candidate)
	}

	add(primaryType)
	for _, placeType := range types {
		add(placeType)
	}

	return normalized
}

func pointOfInterestMarkerCategoryFromExactTypes(types []string) PointOfInterestMarkerCategory {
	exact := map[string]PointOfInterestMarkerCategory{
		"coffee_shop":             PointOfInterestMarkerCategoryCoffeehouse,
		"cafe":                    PointOfInterestMarkerCategoryCoffeehouse,
		"tea_house":               PointOfInterestMarkerCategoryCoffeehouse,
		"bar":                     PointOfInterestMarkerCategoryTavern,
		"pub":                     PointOfInterestMarkerCategoryTavern,
		"wine_bar":                PointOfInterestMarkerCategoryTavern,
		"cocktail_bar":            PointOfInterestMarkerCategoryTavern,
		"brewpub":                 PointOfInterestMarkerCategoryTavern,
		"irish_pub":               PointOfInterestMarkerCategoryTavern,
		"night_club":              PointOfInterestMarkerCategoryTavern,
		"restaurant":              PointOfInterestMarkerCategoryEatery,
		"bakery":                  PointOfInterestMarkerCategoryEatery,
		"dessert_shop":            PointOfInterestMarkerCategoryEatery,
		"ice_cream_shop":          PointOfInterestMarkerCategoryEatery,
		"sandwich_shop":           PointOfInterestMarkerCategoryEatery,
		"pizza_restaurant":        PointOfInterestMarkerCategoryEatery,
		"food_court":              PointOfInterestMarkerCategoryEatery,
		"market":                  PointOfInterestMarkerCategoryMarket,
		"store":                   PointOfInterestMarkerCategoryMarket,
		"shopping_mall":           PointOfInterestMarkerCategoryMarket,
		"department_store":        PointOfInterestMarkerCategoryMarket,
		"convenience_store":       PointOfInterestMarkerCategoryMarket,
		"grocery_store":           PointOfInterestMarkerCategoryMarket,
		"supermarket":             PointOfInterestMarkerCategoryMarket,
		"gift_shop":               PointOfInterestMarkerCategoryMarket,
		"clothing_store":          PointOfInterestMarkerCategoryMarket,
		"electronics_store":       PointOfInterestMarkerCategoryMarket,
		"book_store":              PointOfInterestMarkerCategoryArchive,
		"library":                 PointOfInterestMarkerCategoryArchive,
		"park":                    PointOfInterestMarkerCategoryPark,
		"garden":                  PointOfInterestMarkerCategoryPark,
		"botanical_garden":        PointOfInterestMarkerCategoryPark,
		"hiking_area":             PointOfInterestMarkerCategoryPark,
		"picnic_ground":           PointOfInterestMarkerCategoryPark,
		"state_park":              PointOfInterestMarkerCategoryPark,
		"national_park":           PointOfInterestMarkerCategoryPark,
		"playground":              PointOfInterestMarkerCategoryPark,
		"dog_park":                PointOfInterestMarkerCategoryPark,
		"marina":                  PointOfInterestMarkerCategoryWaterfront,
		"beach":                   PointOfInterestMarkerCategoryWaterfront,
		"museum":                  PointOfInterestMarkerCategoryMuseum,
		"art_gallery":             PointOfInterestMarkerCategoryMuseum,
		"art_studio":              PointOfInterestMarkerCategoryMuseum,
		"history_museum":          PointOfInterestMarkerCategoryMuseum,
		"performing_arts_theater": PointOfInterestMarkerCategoryTheater,
		"movie_theater":           PointOfInterestMarkerCategoryTheater,
		"concert_hall":            PointOfInterestMarkerCategoryTheater,
		"opera_house":             PointOfInterestMarkerCategoryTheater,
		"live_music_venue":        PointOfInterestMarkerCategoryTheater,
		"tourist_attraction":      PointOfInterestMarkerCategoryLandmark,
		"historical_landmark":     PointOfInterestMarkerCategoryLandmark,
		"monument":                PointOfInterestMarkerCategoryLandmark,
		"plaza":                   PointOfInterestMarkerCategoryLandmark,
		"observation_deck":        PointOfInterestMarkerCategoryLandmark,
		"castle":                  PointOfInterestMarkerCategoryLandmark,
		"fountain":                PointOfInterestMarkerCategoryLandmark,
		"bridge":                  PointOfInterestMarkerCategoryLandmark,
		"post_office":             PointOfInterestMarkerCategoryCivic,
		"city_hall":               PointOfInterestMarkerCategoryCivic,
		"courthouse":              PointOfInterestMarkerCategoryCivic,
		"government_office":       PointOfInterestMarkerCategoryCivic,
		"local_government_office": PointOfInterestMarkerCategoryCivic,
		"stadium":                 PointOfInterestMarkerCategoryArena,
		"arena":                   PointOfInterestMarkerCategoryArena,
		"sports_complex":          PointOfInterestMarkerCategoryArena,
		"gym":                     PointOfInterestMarkerCategoryArena,
		"golf_course":             PointOfInterestMarkerCategoryArena,
		"skating_rink":            PointOfInterestMarkerCategoryArena,
	}

	for _, placeType := range types {
		if category, ok := exact[placeType]; ok {
			return category
		}
	}

	return ""
}

func pointOfInterestMarkerCategoryFromFoodDrinkHints(
	types []string,
	servesCoffee bool,
	servesBeer bool,
	servesWine bool,
	servesCocktails bool,
	liveMusic bool,
) PointOfInterestMarkerCategory {
	if !pointOfInterestMarkerTypesContainAny(types, []string{
		"restaurant",
		"food",
		"meal_takeaway",
		"meal_delivery",
		"diner",
		"bakery",
		"dessert",
		"ice_cream",
		"sandwich",
		"pizza",
	}) {
		return ""
	}

	if servesBeer || servesWine || servesCocktails || liveMusic {
		return PointOfInterestMarkerCategoryTavern
	}
	if servesCoffee || pointOfInterestMarkerTypesContainAny(types, []string{"coffee", "cafe", "tea", "espresso"}) {
		return PointOfInterestMarkerCategoryCoffeehouse
	}
	return PointOfInterestMarkerCategoryEatery
}

func pointOfInterestMarkerCategoryFromLooseTypes(types []string) PointOfInterestMarkerCategory {
	switch {
	case pointOfInterestMarkerTypesContainAny(types, []string{"coffee", "cafe", "tea", "espresso"}):
		return PointOfInterestMarkerCategoryCoffeehouse
	case pointOfInterestMarkerTypesContainAny(types, []string{"bar", "pub", "brew", "cocktail", "wine", "night_club"}):
		return PointOfInterestMarkerCategoryTavern
	case pointOfInterestMarkerTypesContainAny(types, []string{"restaurant", "bakery", "dessert", "ice_cream", "sandwich", "pizza", "food", "diner"}):
		return PointOfInterestMarkerCategoryEatery
	case pointOfInterestMarkerTypesContainAny(types, []string{"market", "store", "shop", "mall", "grocery", "supermarket", "clothing", "electronics", "gift"}):
		return PointOfInterestMarkerCategoryMarket
	case pointOfInterestMarkerTypesContainAny(types, []string{"library", "book"}):
		return PointOfInterestMarkerCategoryArchive
	case pointOfInterestMarkerTypesContainAny(types, []string{"park", "garden", "trail", "hiking", "picnic", "playground", "campground", "arboretum"}):
		return PointOfInterestMarkerCategoryPark
	case pointOfInterestMarkerTypesContainAny(types, []string{"marina", "beach", "waterfront", "pier", "harbor", "harbour", "lake", "river", "canal", "boardwalk"}):
		return PointOfInterestMarkerCategoryWaterfront
	case pointOfInterestMarkerTypesContainAny(types, []string{"museum", "gallery", "history"}):
		return PointOfInterestMarkerCategoryMuseum
	case pointOfInterestMarkerTypesContainAny(types, []string{"theater", "theatre", "movie", "cinema", "concert", "opera", "music_venue", "performing_arts"}):
		return PointOfInterestMarkerCategoryTheater
	case pointOfInterestMarkerTypesContainAny(types, []string{"tourist", "landmark", "monument", "plaza", "square", "observation", "scenic", "lookout", "castle", "bridge", "fountain", "church", "temple", "shrine", "zoo", "aquarium", "amusement"}):
		return PointOfInterestMarkerCategoryLandmark
	case pointOfInterestMarkerTypesContainAny(types, []string{"post_office", "courthouse", "city_hall", "government", "municipal", "town_hall"}):
		return PointOfInterestMarkerCategoryCivic
	case pointOfInterestMarkerTypesContainAny(types, []string{"stadium", "arena", "sports", "gym", "golf", "athletic", "fieldhouse", "skating"}):
		return PointOfInterestMarkerCategoryArena
	default:
		return ""
	}
}

func pointOfInterestMarkerCategoryFromText(raw string) PointOfInterestMarkerCategory {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" {
		return ""
	}
	return pointOfInterestMarkerCategoryFromLooseTypes([]string{text})
}

func pointOfInterestMarkerTypesContainAny(types []string, needles []string) bool {
	for _, placeType := range types {
		for _, needle := range needles {
			if placeType == needle || strings.Contains(placeType, needle) {
				return true
			}
		}
	}
	return false
}
