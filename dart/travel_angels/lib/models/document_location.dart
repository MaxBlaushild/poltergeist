import 'package:json_annotation/json_annotation.dart';

part 'document_location.g.dart';

enum LocationType {
  @JsonValue('city')
  city,
  @JsonValue('country')
  country,
  @JsonValue('continent')
  continent,
}

@JsonSerializable()
class DocumentLocation {
  @JsonKey(name: 'id', fromJson: _idFromJson)
  final String id;

  @JsonKey(name: 'documentId', fromJson: _idFromJson)
  final String documentId;

  @JsonKey(name: 'placeId')
  final String placeId;

  @JsonKey(name: 'name')
  final String name;

  @JsonKey(name: 'formattedAddress')
  final String formattedAddress;

  @JsonKey(name: 'latitude')
  final double latitude;

  @JsonKey(name: 'longitude')
  final double longitude;

  @JsonKey(name: 'locationType', fromJson: _locationTypeFromJson)
  final LocationType locationType;

  DocumentLocation({
    required this.id,
    required this.documentId,
    required this.placeId,
    required this.name,
    required this.formattedAddress,
    required this.latitude,
    required this.longitude,
    required this.locationType,
  });

  factory DocumentLocation.fromJson(Map<String, dynamic> json) =>
      _$DocumentLocationFromJson(json);

  Map<String, dynamic> toJson() => _$DocumentLocationToJson(this);

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

