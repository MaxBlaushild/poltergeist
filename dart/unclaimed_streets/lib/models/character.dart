import 'character_location.dart';

class CharacterRelationshipState {
  final int trust;
  final int respect;
  final int fear;
  final int debt;

  const CharacterRelationshipState({
    this.trust = 0,
    this.respect = 0,
    this.fear = 0,
    this.debt = 0,
  });

  bool get isZero => trust == 0 && respect == 0 && fear == 0 && debt == 0;

  factory CharacterRelationshipState.fromJson(Map<String, dynamic> json) {
    int parseValue(dynamic raw) {
      if (raw is num) return raw.toInt();
      if (raw is String) return int.tryParse(raw.trim()) ?? 0;
      return 0;
    }

    return CharacterRelationshipState(
      trust: parseValue(json['trust']),
      respect: parseValue(json['respect']),
      fear: parseValue(json['fear']),
      debt: parseValue(json['debt']),
    );
  }
}

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
  final List<CharacterLocation> locations;
  final bool hasAvailableQuest;
  final bool hasAvailableMainStoryQuest;
  final CharacterRelationshipState? relationship;

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
    this.locations = const [],
    this.hasAvailableQuest = false,
    this.hasAvailableMainStoryQuest = false,
    this.relationship,
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
      hasAvailableQuest: json['hasAvailableQuest'] as bool? ?? false,
      hasAvailableMainStoryQuest:
          json['hasAvailableMainStoryQuest'] as bool? ?? false,
      relationship: json['relationship'] is Map<String, dynamic>
          ? CharacterRelationshipState.fromJson(
              json['relationship'] as Map<String, dynamic>,
            )
          : null,
    );
  }
}
