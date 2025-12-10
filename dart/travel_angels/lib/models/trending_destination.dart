import 'package:json_annotation/json_annotation.dart';
import 'package:travel_angels/models/document_location.dart';

part 'trending_destination.g.dart';

@JsonSerializable()
class TrendingDestination {
  @JsonKey(name: 'id', fromJson: _idFromJson)
  final String id;

  @JsonKey(name: 'name')
  final String name;

  @JsonKey(name: 'formattedAddress')
  final String formattedAddress;

  @JsonKey(name: 'documentCount')
  final int documentCount;

  @JsonKey(name: 'placeId')
  final String placeId;

  @JsonKey(name: 'latitude')
  final double latitude;

  @JsonKey(name: 'longitude')
  final double longitude;

  @JsonKey(name: 'locationType', fromJson: _locationTypeFromJson)
  final LocationType locationType;

  @JsonKey(name: 'rank')
  final int rank;

  TrendingDestination({
    required this.id,
    required this.name,
    required this.formattedAddress,
    required this.documentCount,
    required this.placeId,
    required this.latitude,
    required this.longitude,
    required this.locationType,
    required this.rank,
  });

  factory TrendingDestination.fromJson(Map<String, dynamic> json) =>
      _$TrendingDestinationFromJson(json);

  Map<String, dynamic> toJson() => _$TrendingDestinationToJson(this);

  static String _idFromJson(dynamic json) {
    if (json == null) return '';
    return json.toString();
  }

  static LocationType _locationTypeFromJson(dynamic json) {
    if (json == null) return LocationType.city;
    final str = json.toString().toLowerCase();
    switch (str) {
      case 'country':
        return LocationType.country;
      case 'continent':
        return LocationType.continent;
      case 'city':
      default:
        return LocationType.city;
    }
  }
}
