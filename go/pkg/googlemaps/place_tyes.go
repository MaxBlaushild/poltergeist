package googlemaps

type PlaceType string

const (
	TypeAccounting            PlaceType = "accounting"
	TypeAirport               PlaceType = "airport"
	TypeAmusementPark         PlaceType = "amusement_park"
	TypeAquarium              PlaceType = "aquarium"
	TypeArtGallery            PlaceType = "art_gallery"
	TypeAtm                   PlaceType = "atm"
	TypeBakery                PlaceType = "bakery"
	TypeBank                  PlaceType = "bank"
	TypeBar                   PlaceType = "bar"
	TypeBeautySalon           PlaceType = "beauty_salon"
	TypeBicycleStore          PlaceType = "bicycle_store"
	TypeBookStore             PlaceType = "book_store"
	TypeBowlingAlley          PlaceType = "bowling_alley"
	TypeBusStation            PlaceType = "bus_station"
	TypeCafe                  PlaceType = "cafe"
	TypeCampground            PlaceType = "campground"
	TypeCarDealer             PlaceType = "car_dealer"
	TypeCarRental             PlaceType = "car_rental"
	TypeCarRepair             PlaceType = "car_repair"
	TypeCarWash               PlaceType = "car_wash"
	TypeCasino                PlaceType = "casino"
	TypeCemetery              PlaceType = "cemetery"
	TypeChurch                PlaceType = "church"
	TypeCityHall              PlaceType = "city_hall"
	TypeClothingStore         PlaceType = "clothing_store"
	TypeConvenienceStore      PlaceType = "convenience_store"
	TypeCourthouse            PlaceType = "courthouse"
	TypeDentist               PlaceType = "dentist"
	TypeDepartmentStore       PlaceType = "department_store"
	TypeDoctor                PlaceType = "doctor"
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
	TypeZoo                   PlaceType = "zoo"
)

var AllPlaceTypes = []PlaceType{
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
	TypeZoo,
}
