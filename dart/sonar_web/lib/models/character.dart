import 'movement_pattern.dart';
import 'character_location.dart';

class Character {
  final String id;
  final String name;
  final String? description;
  final String? mapIconUrl;
  final String? dialogueImageUrl;
  final String? thumbnailUrl;
  final String? pointOfInterestId;
  final double? pointOfInterestLat;
  final double? pointOfInterestLng;
  final MovementPattern? movementPattern;
  final List<CharacterLocation> locations;
  final bool hasAvailableQuest;

  const Character({
    required this.id,
    required this.name,
    this.description,
    this.mapIconUrl,
    this.dialogueImageUrl,
    this.thumbnailUrl,
    this.pointOfInterestId,
    this.pointOfInterestLat,
    this.pointOfInterestLng,
    this.movementPattern,
    this.locations = const [],
    this.hasAvailableQuest = false,
  });

  factory Character.fromJson(Map<String, dynamic> json) {
    double? parseCoordinate(dynamic raw) {
      if (raw == null) return null;
      if (raw is num) return raw.toDouble();
      if (raw is String) return double.tryParse(raw.trim());
      return null;
    }

    final poi = json['pointOfInterest'];
    double? pointOfInterestLat;
    double? pointOfInterestLng;
    if (poi is Map<String, dynamic>) {
      pointOfInterestLat = parseCoordinate(poi['lat'] ?? poi['latitude']);
      pointOfInterestLng = parseCoordinate(poi['lng'] ?? poi['longitude']);
    }

    return Character(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      description: json['description'] as String?,
      mapIconUrl: json['mapIconUrl'] as String?,
      dialogueImageUrl: json['dialogueImageUrl'] as String?,
      thumbnailUrl: json['thumbnailUrl'] as String?,
      pointOfInterestId: json['pointOfInterestId'] as String?,
      pointOfInterestLat: pointOfInterestLat,
      pointOfInterestLng: pointOfInterestLng,
      locations:
          (json['locations'] as List<dynamic>?)
              ?.map(
                (e) => CharacterLocation.fromJson(e as Map<String, dynamic>),
              )
              .toList() ??
          const [],
      movementPattern: json['movementPattern'] != null
          ? MovementPattern.fromJson(
              json['movementPattern'] as Map<String, dynamic>,
            )
          : null,
      hasAvailableQuest: json['hasAvailableQuest'] as bool? ?? false,
    );
  }

  double get lat => movementPattern?.startingLatitude ?? 0;
  double get lng => movementPattern?.startingLongitude ?? 0;
}
