package googlemaps

import (
	"database/sql/driver"
	"strings"
)

type PlaceType string

// PlaceTypeSlice is a custom type for []PlaceType that implements Scanner and Valuer
type PlaceTypeSlice []PlaceType

var AutomotivePlaceTypes = []PlaceType{
	TypeCarDealer,
	TypeCarRental,
	TypeCarRepair,
	TypeCarWash,
	TypeElectricVehicleChargingStation,
	TypeGasStation,
	TypeParking,
	TypeRestStop,
}

var BusinessPlaceTypes = []PlaceType{
	TypeCorporateOffice,
	TypeFarm,
	TypeRanch,
}

var CulturalPlaceTypes = []PlaceType{
	TypeArtGallery,
	TypeArtStudio,
	TypeAuditorium,
	TypeCulturalLandmark,
	TypeHistoricalPlace,
	TypeMonument,
	TypeMuseum,
	TypePerformingArtsTheater,
	TypeSculpture,
}

var EntertainmentPlaceTypes = []PlaceType{
	TypeAdventureSportsCenter,
	TypeAmphitheatre,
	TypeAmusementCenter,
	TypeAmusementPark,
	TypeAquarium,
	TypeBanquetHall,
	TypeBarbecueArea,
	TypeBotanicalGarden,
	TypeBowlingAlley,
	TypeCasino,
	TypeChildrensCamp,
	TypeComedyClub,
	TypeCommunityCenter,
	TypeConcertHall,
	TypeConventionCenter,
	TypeCulturalCenter,
	TypeCyclingPark,
	TypeDanceHall,
	TypeDogPark,
	TypeEventVenue,
	TypeFerrisWheel,
	TypeGarden,
	TypeHikingArea,
	TypeHistoricalLandmark,
	TypeInternetCafe,
	TypeKaraoke,
	TypeMarina,
	TypeMovieRental,
	TypeMovieTheater,
	TypeNationalPark,
	TypeNightClub,
	TypeObservationDeck,
	TypeOffRoadingArea,
	TypeOperaHouse,
	TypePark,
	TypePhilharmonicHall,
	TypePicnicGround,
	TypePlanetarium,
	TypePlaza,
	TypeRollerCoaster,
	TypeSkateboardPark,
	TypeStatePark,
	TypeTouristAttraction,
	TypeVideoArcade,
	TypeVisitorCenter,
	TypeWaterPark,
	TypeWeddingVenue,
	TypeWildlifePark,
	TypeWildlifeRefuge,
	TypeZoo,
}

var FoodAndDrinkPlaceTypes = []PlaceType{
	TypeAcaiShop,
	TypeAfghaniRestaurant,
	TypeAfricanRestaurant,
	TypeAmericanRestaurant,
	TypeAsianRestaurant,
	TypeBagelShop,
	TypeBakery,
	TypeBar,
	TypeBarAndGrill,
	TypeBarbecueRestaurant,
	TypeBrazilianRestaurant,
	TypeBreakfastRestaurant,
	TypeBrunchRestaurant,
	TypeBuffetRestaurant,
	TypeCafe,
	TypeCafeteria,
	TypeCandyStore,
	TypeCatCafe,
	TypeChineseRestaurant,
	TypeChocolateFactory,
	TypeChocolateShop,
	TypeCoffeeShop,
	TypeConfectionery,
	TypeDeli,
	TypeDessertRestaurant,
	TypeDessertShop,
	TypeDiner,
	TypeDogCafe,
	TypeDonutShop,
	TypeFastFoodRestaurant,
	TypeFineDiningRestaurant,
	TypeFoodCourt,
	TypeFrenchRestaurant,
	TypeGreekRestaurant,
	TypeHamburgerRestaurant,
	TypeIceCreamShop,
	TypeIndianRestaurant,
	TypeIndonesianRestaurant,
	TypeItalianRestaurant,
	TypeJapaneseRestaurant,
	TypeJuiceShop,
	TypeKoreanRestaurant,
	TypeLebaneseRestaurant,
	TypeMediterraneanRestaurant,
	TypeMexicanRestaurant,
	TypeMiddleEasternRestaurant,
	TypePizzaRestaurant,
	TypePub,
	TypeRamenRestaurant,
	TypeRestaurant,
	TypeSandwichShop,
	TypeSeafoodRestaurant,
	TypeSpanishRestaurant,
	TypeSteakHouse,
	TypeSushiRestaurant,
	TypeTeaHouse,
	TypeThaiRestaurant,
	TypeTurkishRestaurant,
	TypeVeganRestaurant,
	TypeVegetarianRestaurant,
	TypeVietnameseRestaurant,
	TypeWineBar,
}

var GovernmentPlaceTypes = []PlaceType{
	TypeCityHall,
	TypeCourthouse,
	TypeEmbassy,
	TypeFireStation,
	TypeGovernmentOffice,
	TypeLocalGovernmentOffice,
	TypeNeighborhoodPolice,
	TypePolice,
	TypePostOffice,
}

var LodgingPlaceTypes = []PlaceType{
	TypeBedAndBreakfast,
	TypeBudgetJapaneseInn,
	TypeCampground,
	TypeCampingCabin,
	TypeCottage,
	TypeExtendedStayHotel,
	TypeFarmstay,
	TypeGuestHouse,
	TypeHostel,
	TypeHotel,
	TypeInn,
	TypeJapaneseInn,
	TypeLodging,
	TypeMobileHomePark,
	TypeMotel,
	TypePrivateGuestRoom,
	TypeResortHotel,
	TypeRvPark,
}

var PlaceOfWorshipPlaceTypes = []PlaceType{
	TypeChurch,
	TypeHinduTemple,
	TypeMosque,
	TypeSynagogue,
}

var PlaceNaturalPlaceTypes = []PlaceType{
	TypeBeach,
	TypeNationalPark,
	TypePark,
}

var ShoppingPlaceTypes = []PlaceType{
	TypeAsianGroceryStore,
	TypeAutoPartsStore,
	TypeBicycleStore,
	TypeBookStore,
	TypeButcherShop,
	TypeCellPhoneStore,
	TypeClothingStore,
	TypeConvenienceStore,
	TypeDepartmentStore,
	TypeDiscountStore,
	TypeElectronicsStore,
	TypeFoodStore,
	TypeFurnitureStore,
	TypeGiftShop,
	TypeGroceryStore,
	TypeHardwareStore,
	TypeHomeGoodsStore,
	TypeHomeImprovementStore,
	TypeJewelryStore,
	TypeLiquorStore,
	TypeMarket,
	TypePetStore,
	TypeShoeStore,
	TypeShoppingMall,
	TypeSportingGoodsStore,
	TypeStore,
	TypeSupermarket,
	TypeWarehouseStore,
	TypeWholesaler,
}

var SportsPlaceTypes = []PlaceType{
	TypeArena,
	TypeAthleticField,
	TypeFishingCharter,
	TypeFishingPond,
	TypeFitnessCenter,
	TypeGolfCourse,
	TypeGym,
	TypeIceSkatingRink,
	TypePlayground,
	TypeSkiResort,
	TypeSportsActivityLocation,
	TypeSportsClub,
	TypeSportsCoaching,
	TypeSportsComplex,
	TypeStadium,
	TypeSwimmingPool,
}

var TransportationPlaceTypes = []PlaceType{
	TypeAirport,
	TypeAirstrip,
	TypeBusStation,
	TypeBusStop,
	TypeFerryTerminal,
	TypeHeliport,
	TypeInternationalAirport,
	TypeLightRailStation,
	TypeParkAndRide,
	TypeSubwayStation,
	TypeTaxiStand,
	TypeTrainStation,
	TypeTransitDepot,
	TypeTransitStation,
	TypeTruckStop,
}

const (
	TypeAirstrip             PlaceType = "airstrip"
	TypeBusStop              PlaceType = "bus_stop"
	TypeFerryTerminal        PlaceType = "ferry_terminal"
	TypeHeliport             PlaceType = "heliport"
	TypeInternationalAirport PlaceType = "international_airport"
	TypeParkAndRide          PlaceType = "park_and_ride"
	TypeTransitDepot         PlaceType = "transit_depot"
	TypeTruckStop            PlaceType = "truck_stop"

	TypeArena                  PlaceType = "arena"
	TypeAthleticField          PlaceType = "athletic_field"
	TypeFishingCharter         PlaceType = "fishing_charter"
	TypeFishingPond            PlaceType = "fishing_pond"
	TypeFitnessCenter          PlaceType = "fitness_center"
	TypeGolfCourse             PlaceType = "golf_course"
	TypeIceSkatingRink         PlaceType = "ice_skating_rink"
	TypePlayground             PlaceType = "playground"
	TypeSkiResort              PlaceType = "ski_resort"
	TypeSportsActivityLocation PlaceType = "sports_activity_location"
	TypeSportsClub             PlaceType = "sports_club"
	TypeSportsCoaching         PlaceType = "sports_coaching"
	TypeSportsComplex          PlaceType = "sports_complex"
	TypeSwimmingPool           PlaceType = "swimming_pool"

	TypeAsianGroceryStore    PlaceType = "asian_grocery_store"
	TypeAutoPartsStore       PlaceType = "auto_parts_store"
	TypeButcherShop          PlaceType = "butcher_shop"
	TypeCellPhoneStore       PlaceType = "cell_phone_store"
	TypeDiscountStore        PlaceType = "discount_store"
	TypeFoodStore            PlaceType = "food_store"
	TypeGiftShop             PlaceType = "gift_shop"
	TypeGroceryStore         PlaceType = "grocery_store"
	TypeHomeImprovementStore PlaceType = "home_improvement_store"
	TypeMarket               PlaceType = "market"
	TypeSportingGoodsStore   PlaceType = "sporting_goods_store"
	TypeWarehouseStore       PlaceType = "warehouse_store"
	TypeWholesaler           PlaceType = "wholesaler"

	TypeBeach             PlaceType = "beach"
	TypeBedAndBreakfast   PlaceType = "bed_and_breakfast"
	TypeBudgetJapaneseInn PlaceType = "budget_japanese_inn"
	TypeCampingCabin      PlaceType = "camping_cabin"
	TypeCottage           PlaceType = "cottage"
	TypeExtendedStayHotel PlaceType = "extended_stay_hotel"
	TypeFarmstay          PlaceType = "farmstay"
	TypeGuestHouse        PlaceType = "guest_house"
	TypeHostel            PlaceType = "hostel"
	TypeHotel             PlaceType = "hotel"
	TypeInn               PlaceType = "inn"
	TypeJapaneseInn       PlaceType = "japanese_inn"
	TypeMobileHomePark    PlaceType = "mobile_home_park"
	TypeMotel             PlaceType = "motel"
	TypePrivateGuestRoom  PlaceType = "private_guest_room"
	TypeResortHotel       PlaceType = "resort_hotel"

	TypeGovernmentOffice   PlaceType = "government_office"
	TypeNeighborhoodPolice PlaceType = "neighborhood_police_station"

	TypeAcaiShop                PlaceType = "acai_shop"
	TypeAfghaniRestaurant       PlaceType = "afghani_restaurant"
	TypeAfricanRestaurant       PlaceType = "african_restaurant"
	TypeAmericanRestaurant      PlaceType = "american_restaurant"
	TypeAsianRestaurant         PlaceType = "asian_restaurant"
	TypeBagelShop               PlaceType = "bagel_shop"
	TypeBarAndGrill             PlaceType = "bar_and_grill"
	TypeBarbecueRestaurant      PlaceType = "barbecue_restaurant"
	TypeBrazilianRestaurant     PlaceType = "brazilian_restaurant"
	TypeBreakfastRestaurant     PlaceType = "breakfast_restaurant"
	TypeBrunchRestaurant        PlaceType = "brunch_restaurant"
	TypeBuffetRestaurant        PlaceType = "buffet_restaurant"
	TypeCafeteria               PlaceType = "cafeteria"
	TypeCandyStore              PlaceType = "candy_store"
	TypeCatCafe                 PlaceType = "cat_cafe"
	TypeChineseRestaurant       PlaceType = "chinese_restaurant"
	TypeChocolateFactory        PlaceType = "chocolate_factory"
	TypeChocolateShop           PlaceType = "chocolate_shop"
	TypeCoffeeShop              PlaceType = "coffee_shop"
	TypeConfectionery           PlaceType = "confectionery"
	TypeDeli                    PlaceType = "deli"
	TypeDessertRestaurant       PlaceType = "dessert_restaurant"
	TypeDessertShop             PlaceType = "dessert_shop"
	TypeDiner                   PlaceType = "diner"
	TypeDogCafe                 PlaceType = "dog_cafe"
	TypeDonutShop               PlaceType = "donut_shop"
	TypeFastFoodRestaurant      PlaceType = "fast_food_restaurant"
	TypeFineDiningRestaurant    PlaceType = "fine_dining_restaurant"
	TypeFoodCourt               PlaceType = "food_court"
	TypeFrenchRestaurant        PlaceType = "french_restaurant"
	TypeGreekRestaurant         PlaceType = "greek_restaurant"
	TypeHamburgerRestaurant     PlaceType = "hamburger_restaurant"
	TypeIceCreamShop            PlaceType = "ice_cream_shop"
	TypeIndianRestaurant        PlaceType = "indian_restaurant"
	TypeIndonesianRestaurant    PlaceType = "indonesian_restaurant"
	TypeItalianRestaurant       PlaceType = "italian_restaurant"
	TypeJapaneseRestaurant      PlaceType = "japanese_restaurant"
	TypeJuiceShop               PlaceType = "juice_shop"
	TypeKoreanRestaurant        PlaceType = "korean_restaurant"
	TypeLebaneseRestaurant      PlaceType = "lebanese_restaurant"
	TypeMediterraneanRestaurant PlaceType = "mediterranean_restaurant"
	TypeMexicanRestaurant       PlaceType = "mexican_restaurant"
	TypeMiddleEasternRestaurant PlaceType = "middle_eastern_restaurant"
	TypePizzaRestaurant         PlaceType = "pizza_restaurant"
	TypePub                     PlaceType = "pub"
	TypeRamenRestaurant         PlaceType = "ramen_restaurant"
	TypeSandwichShop            PlaceType = "sandwich_shop"
	TypeSeafoodRestaurant       PlaceType = "seafood_restaurant"
	TypeSpanishRestaurant       PlaceType = "spanish_restaurant"
	TypeSteakHouse              PlaceType = "steak_house"
	TypeSushiRestaurant         PlaceType = "sushi_restaurant"
	TypeTeaHouse                PlaceType = "tea_house"
	TypeThaiRestaurant          PlaceType = "thai_restaurant"
	TypeTurkishRestaurant       PlaceType = "turkish_restaurant"
	TypeVeganRestaurant         PlaceType = "vegan_restaurant"
	TypeVegetarianRestaurant    PlaceType = "vegetarian_restaurant"
	TypeVietnameseRestaurant    PlaceType = "vietnamese_restaurant"

	TypeAdventureSportsCenter PlaceType = "adventure_sports_center"
	TypeAmphitheatre          PlaceType = "amphitheatre"
	TypeAmusementCenter       PlaceType = "amusement_center"
	TypeBanquetHall           PlaceType = "banquet_hall"
	TypeBarbecueArea          PlaceType = "barbecue_area"
	TypeBotanicalGarden       PlaceType = "botanical_garden"
	TypeChildrensCamp         PlaceType = "childrens_camp"
	TypeComedyClub            PlaceType = "comedy_club"
	TypeCommunityCenter       PlaceType = "community_center"
	TypeConcertHall           PlaceType = "concert_hall"
	TypeConventionCenter      PlaceType = "convention_center"
	TypeCulturalCenter        PlaceType = "cultural_center"
	TypeCyclingPark           PlaceType = "cycling_park"
	TypeDanceHall             PlaceType = "dance_hall"
	TypeDogPark               PlaceType = "dog_park"
	TypeEventVenue            PlaceType = "event_venue"
	TypeFerrisWheel           PlaceType = "ferris_wheel"
	TypeGarden                PlaceType = "garden"
	TypeHikingArea            PlaceType = "hiking_area"
	TypeHistoricalLandmark    PlaceType = "historical_landmark"
	TypeInternetCafe          PlaceType = "internet_cafe"
	TypeKaraoke               PlaceType = "karaoke"
	TypeMarina                PlaceType = "marina"
	TypeNationalPark          PlaceType = "national_park"
	TypeObservationDeck       PlaceType = "observation_deck"
	TypeOffRoadingArea        PlaceType = "off_roading_area"
	TypeOperaHouse            PlaceType = "opera_house"
	TypePhilharmonicHall      PlaceType = "philharmonic_hall"
	TypePicnicGround          PlaceType = "picnic_ground"
	TypePlanetarium           PlaceType = "planetarium"
	TypePlaza                 PlaceType = "plaza"
	TypeRollerCoaster         PlaceType = "roller_coaster"
	TypeSkateboardPark        PlaceType = "skateboard_park"
	TypeStatePark             PlaceType = "state_park"
	TypeVideoArcade           PlaceType = "video_arcade"
	TypeVisitorCenter         PlaceType = "visitor_center"
	TypeWaterPark             PlaceType = "water_park"
	TypeWeddingVenue          PlaceType = "wedding_venue"
	TypeWildlifePark          PlaceType = "wildlife_park"
	TypeWildlifeRefuge        PlaceType = "wildlife_refuge"

	TypeCorporateOffice                PlaceType = "corporate_office"
	TypeFarm                           PlaceType = "farm"
	TypeRanch                          PlaceType = "ranch"
	TypeElectricVehicleChargingStation PlaceType = "electric_vehicle_charging_station"
	TypeRestStop                       PlaceType = "rest_stop"
	TypeAccounting                     PlaceType = "accounting"
	TypeAirport                        PlaceType = "airport"
	TypeAmusementPark                  PlaceType = "amusement_park"
	TypeAquarium                       PlaceType = "aquarium"
	TypeArtGallery                     PlaceType = "art_gallery"
	TypeAtm                            PlaceType = "atm"
	TypeBakery                         PlaceType = "bakery"
	TypeBank                           PlaceType = "bank"
	TypeBar                            PlaceType = "bar"
	TypeBeautySalon                    PlaceType = "beauty_salon"

	TypeArtStudio             PlaceType = "art_studio"
	TypeAuditorium            PlaceType = "auditorium"
	TypeCulturalLandmark      PlaceType = "cultural_landmark"
	TypeHistoricalPlace       PlaceType = "historical_place"
	TypeMonument              PlaceType = "monument"
	TypePerformingArtsTheater PlaceType = "performing_arts_theater"
	TypeSculpture             PlaceType = "sculpture"

	TypeBicycleStore PlaceType = "bicycle_store"
	TypeBookStore    PlaceType = "book_store"
	TypeBowlingAlley PlaceType = "bowling_alley"
	TypeBusStation   PlaceType = "bus_station"
	TypeCafe         PlaceType = "cafe"

	TypeCampground PlaceType = "campground"
	TypeCarDealer  PlaceType = "car_dealer"
	TypeCarRental  PlaceType = "car_rental"
	TypeCarRepair  PlaceType = "car_repair"
	TypeCarWash    PlaceType = "car_wash"

	TypeCasino           PlaceType = "casino"
	TypeCemetery         PlaceType = "cemetery"
	TypeChurch           PlaceType = "church"
	TypeCityHall         PlaceType = "city_hall"
	TypeClothingStore    PlaceType = "clothing_store"
	TypeConvenienceStore PlaceType = "convenience_store"
	TypeCourthouse       PlaceType = "courthouse"
	TypeDentist          PlaceType = "dentist"
	TypeDepartmentStore  PlaceType = "department_store"
	TypeDoctor           PlaceType = "doctor"

	TypeDrugstore             PlaceType = "drugstore"
	TypeElectrician           PlaceType = "electrician"
	TypeElectronicsStore      PlaceType = "electronics_store"
	TypeEmbassy               PlaceType = "embassy"
	TypeFireStation           PlaceType = "fire_station"
	TypeFlorist               PlaceType = "florist"
	TypeFuneralHome           PlaceType = "funeral_home"
	TypeFurnitureStore        PlaceType = "furniture_store"
	TypeGasStation            PlaceType = "gas_station"
	TypeGym                   PlaceType = "gym"
	TypeHairCare              PlaceType = "hair_care"
	TypeHardwareStore         PlaceType = "hardware_store"
	TypeHinduTemple           PlaceType = "hindu_temple"
	TypeHomeGoodsStore        PlaceType = "home_goods_store"
	TypeHospital              PlaceType = "hospital"
	TypeInsuranceAgency       PlaceType = "insurance_agency"
	TypeJewelryStore          PlaceType = "jewelry_store"
	TypeLaundry               PlaceType = "laundry"
	TypeLawyer                PlaceType = "lawyer"
	TypeLibrary               PlaceType = "library"
	TypeLightRailStation      PlaceType = "light_rail_station"
	TypeLiquorStore           PlaceType = "liquor_store"
	TypeLocalGovernmentOffice PlaceType = "local_government_office"
	TypeLocksmith             PlaceType = "locksmith"
	TypeLodging               PlaceType = "lodging"
	TypeMealDelivery          PlaceType = "meal_delivery"
	TypeMealTakeaway          PlaceType = "meal_takeaway"
	TypeMosque                PlaceType = "mosque"
	TypeMovieRental           PlaceType = "movie_rental"
	TypeMovieTheater          PlaceType = "movie_theater"
	TypeMovingCompany         PlaceType = "moving_company"
	TypeMuseum                PlaceType = "museum"
	TypeNightClub             PlaceType = "night_club"
	TypePainter               PlaceType = "painter"
	TypePark                  PlaceType = "park"
	TypeParking               PlaceType = "parking"
	TypePetStore              PlaceType = "pet_store"
	TypePharmacy              PlaceType = "pharmacy"
	TypePhysiotherapist       PlaceType = "physiotherapist"
	TypePlumber               PlaceType = "plumber"
	TypePolice                PlaceType = "police"
	TypePostOffice            PlaceType = "post_office"
	TypePrimarySchool         PlaceType = "primary_school"
	TypeRealEstateAgency      PlaceType = "real_estate_agency"
	TypeRestaurant            PlaceType = "restaurant"
	TypeRoofingContractor     PlaceType = "roofing_contractor"
	TypeRvPark                PlaceType = "rv_park"
	TypeSchool                PlaceType = "school"
	TypeSecondarySchool       PlaceType = "secondary_school"
	TypeShoeStore             PlaceType = "shoe_store"
	TypeShoppingMall          PlaceType = "shopping_mall"
	TypeSpa                   PlaceType = "spa"
	TypeStadium               PlaceType = "stadium"
	TypeStorage               PlaceType = "storage"
	TypeStore                 PlaceType = "store"
	TypeSubwayStation         PlaceType = "subway_station"
	TypeSupermarket           PlaceType = "supermarket"
	TypeSynagogue             PlaceType = "synagogue"
	TypeTaxiStand             PlaceType = "taxi_stand"
	TypeTouristAttraction     PlaceType = "tourist_attraction"
	TypeTrainStation          PlaceType = "train_station"
	TypeTransitStation        PlaceType = "transit_station"
	TypeTravelAgency          PlaceType = "travel_agency"
	TypeUniversity            PlaceType = "university"
	TypeVeterinaryCare        PlaceType = "veterinary_care"
	TypeWineBar               PlaceType = "wine_bar"
	TypeZoo                   PlaceType = "zoo"
)

func GetAllPlaceTypes() []PlaceType {
	return []PlaceType{
		TypeAirstrip,
		TypeBusStop,
		TypeFerryTerminal,
		TypeHeliport,
		TypeInternationalAirport,
		TypeParkAndRide,
		TypeTransitDepot,
		TypeTruckStop,
		TypeArena,
		TypeAthleticField,
		TypeFishingCharter,
		TypeFishingPond,
		TypeFitnessCenter,
		TypeGolfCourse,
		TypeIceSkatingRink,
		TypePlayground,
		TypeSkiResort,
		TypeSportsActivityLocation,
		TypeSportsClub,
		TypeSportsCoaching,
		TypeSportsComplex,
		TypeSwimmingPool,
		TypeAsianGroceryStore,
		TypeAutoPartsStore,
		TypeButcherShop,
		TypeCellPhoneStore,
		TypeDiscountStore,
		TypeFoodStore,
		TypeGiftShop,
		TypeGroceryStore,
		TypeHomeImprovementStore,
		TypeMarket,
		TypeSportingGoodsStore,
		TypeWarehouseStore,
		TypeWholesaler,
		TypeBeach,
		TypeBedAndBreakfast,
		TypeBudgetJapaneseInn,
		TypeCampingCabin,
		TypeCottage,
		TypeExtendedStayHotel,
		TypeFarmstay,
		TypeGuestHouse,
		TypeHostel,
		TypeHotel,
		TypeInn,
		TypeJapaneseInn,
		TypeMobileHomePark,
		TypeMotel,
		TypePrivateGuestRoom,
		TypeResortHotel,
		TypeGovernmentOffice,
		TypeNeighborhoodPolice,
		TypeAcaiShop,
		TypeAfghaniRestaurant,
		TypeAfricanRestaurant,
		TypeAmericanRestaurant,
		TypeAsianRestaurant,
		TypeBagelShop,
		TypeBarAndGrill,
		TypeBarbecueRestaurant,
		TypeBrazilianRestaurant,
		TypeBreakfastRestaurant,
		TypeBrunchRestaurant,
		TypeBuffetRestaurant,
		TypeCafeteria,
		TypeCandyStore,
		TypeCatCafe,
		TypeChineseRestaurant,
		TypeChocolateFactory,
		TypeChocolateShop,
		TypeCoffeeShop,
		TypeConfectionery,
		TypeDeli,
		TypeDessertRestaurant,
		TypeDessertShop,
		TypeDiner,
		TypeDogCafe,
		TypeDonutShop,
		TypeFastFoodRestaurant,
		TypeFineDiningRestaurant,
		TypeFoodCourt,
		TypeFrenchRestaurant,
		TypeGreekRestaurant,
		TypeHamburgerRestaurant,
		TypeIceCreamShop,
		TypeIndianRestaurant,
		TypeIndonesianRestaurant,
		TypeItalianRestaurant,
		TypeJapaneseRestaurant,
		TypeJuiceShop,
		TypeKoreanRestaurant,
		TypeLebaneseRestaurant,
		TypeMediterraneanRestaurant,
		TypeMexicanRestaurant,
		TypeMiddleEasternRestaurant,
		TypePizzaRestaurant,
		TypePub,
		TypeRamenRestaurant,
		TypeSandwichShop,
		TypeSeafoodRestaurant,
		TypeSpanishRestaurant,
		TypeSteakHouse,
		TypeSushiRestaurant,
		TypeTeaHouse,
		TypeThaiRestaurant,
		TypeTurkishRestaurant,
		TypeVeganRestaurant,
		TypeVegetarianRestaurant,
		TypeVietnameseRestaurant,
		TypeAdventureSportsCenter,
		TypeAmphitheatre,
		TypeAmusementCenter,
		TypeBanquetHall,
		TypeBarbecueArea,
		TypeBotanicalGarden,
		TypeChildrensCamp,
		TypeComedyClub,
		TypeCommunityCenter,
		TypeConcertHall,
		TypeConventionCenter,
		TypeCulturalCenter,
		TypeCyclingPark,
		TypeDanceHall,
		TypeDogPark,
		TypeEventVenue,
		TypeFerrisWheel,
		TypeGarden,
		TypeHikingArea,
		TypeHistoricalLandmark,
		TypeInternetCafe,
		TypeKaraoke,
		TypeMarina,
		TypeNationalPark,
		TypeObservationDeck,
		TypeOffRoadingArea,
		TypeOperaHouse,
		TypePhilharmonicHall,
		TypePicnicGround,
		TypePlanetarium,
		TypePlaza,
		TypeRollerCoaster,
		TypeSkateboardPark,
		TypeStatePark,
		TypeVideoArcade,
		TypeVisitorCenter,
		TypeWaterPark,
		TypeWeddingVenue,
		TypeWildlifePark,
		TypeWildlifeRefuge,
		TypeCorporateOffice,
		TypeFarm,
		TypeRanch,
		TypeElectricVehicleChargingStation,
		TypeRestStop,
		TypeAccounting,
		TypeAirport,
		TypeAmusementPark,
		TypeAquarium,
		TypeArtGallery,
		TypeAtm,
		TypeBakery,
		TypeBank,
		TypeBar,
		TypeBeautySalon,
		TypeArtStudio,
		TypeAuditorium,
		TypeCulturalLandmark,
		TypeHistoricalPlace,
		TypeMonument,
		TypePerformingArtsTheater,
		TypeSculpture,
		TypeBicycleStore,
		TypeBookStore,
		TypeBowlingAlley,
		TypeBusStation,
		TypeCafe,
		TypeCampground,
		TypeCarDealer,
		TypeCarRental,
		TypeCarRepair,
		TypeCarWash,
		TypeCasino,
		TypeCemetery,
		TypeChurch,
		TypeCityHall,
		TypeClothingStore,
		TypeConvenienceStore,
		TypeCourthouse,
		TypeDentist,
		TypeDepartmentStore,
		TypeDoctor,
		TypeDrugstore,
		TypeElectrician,
		TypeElectronicsStore,
		TypeEmbassy,
		TypeFireStation,
		TypeFlorist,
		TypeFuneralHome,
		TypeFurnitureStore,
		TypeGasStation,
		TypeGym,
		TypeHairCare,
		TypeHardwareStore,
		TypeHinduTemple,
		TypeHomeGoodsStore,
		TypeHospital,
		TypeInsuranceAgency,
		TypeJewelryStore,
		TypeLaundry,
		TypeLawyer,
		TypeLibrary,
		TypeLightRailStation,
		TypeLiquorStore,
		TypeLocalGovernmentOffice,
		TypeLocksmith,
		TypeLodging,
		TypeMealDelivery,
		TypeMealTakeaway,
		TypeMosque,
		TypeMovieRental,
		TypeMovieTheater,
		TypeMovingCompany,
		TypeMuseum,
		TypeNightClub,
		TypePainter,
		TypePark,
		TypeParking,
		TypePetStore,
		TypePharmacy,
		TypePhysiotherapist,
		TypePlumber,
		TypePolice,
		TypePostOffice,
		TypePrimarySchool,
		TypeRealEstateAgency,
		TypeRestaurant,
		TypeRoofingContractor,
		TypeRvPark,
		TypeSchool,
		TypeSecondarySchool,
		TypeShoeStore,
		TypeShoppingMall,
		TypeSpa,
		TypeStadium,
		TypeStorage,
		TypeStore,
		TypeSubwayStation,
		TypeSupermarket,
		TypeSynagogue,
		TypeTaxiStand,
		TypeTouristAttraction,
		TypeTrainStation,
		TypeTransitStation,
		TypeTravelAgency,
		TypeUniversity,
		TypeVeterinaryCare,
		TypeWineBar,
		TypeZoo,
	}
}

// Scan implements the sql.Scanner interface for PlaceType
func (p *PlaceType) Scan(value interface{}) error {
	if value == nil {
		*p = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*p = PlaceType(v)
	case []byte:
		*p = PlaceType(v)
	default:
		return nil
	}
	return nil
}

// Value implements the driver.Valuer interface for PlaceType
func (p PlaceType) Value() (driver.Value, error) {
	return string(p), nil
}

// Scan implements the sql.Scanner interface for PlaceTypeSlice
func (p *PlaceTypeSlice) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		// Remove curly braces and split by comma
		cleaned := strings.Trim(v, "{}")
		if cleaned == "" {
			*p = PlaceTypeSlice{}
			return nil
		}
		parts := strings.Split(cleaned, ",")
		types := make(PlaceTypeSlice, len(parts))
		for i, part := range parts {
			types[i] = PlaceType(strings.TrimSpace(part))
		}
		*p = types
	case []byte:
		return p.Scan(string(v))
	default:
		return nil
	}
	return nil
}

// Value implements the driver.Valuer interface for PlaceTypeSlice
func (p PlaceTypeSlice) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	if len(p) == 0 {
		return "{}", nil
	}

	// Convert to PostgreSQL array format
	parts := make([]string, len(p))
	for i, t := range p {
		parts[i] = string(t)
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

// NewPlaceTypeSlice creates a new PlaceTypeSlice from a string slice
func NewPlaceTypeSlice(types []string) PlaceTypeSlice {
	placeTypes := make(PlaceTypeSlice, len(types))
	for i, t := range types {
		placeTypes[i] = PlaceType(t)
	}
	return placeTypes
}
