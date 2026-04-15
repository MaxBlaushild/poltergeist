import 'character.dart';

enum PoiMarkerCategory {
  generic,
  coffeehouse,
  tavern,
  eatery,
  market,
  archive,
  park,
  waterfront,
  museum,
  theater,
  landmark,
  civic,
  arena,
}

extension PoiMarkerCategoryX on PoiMarkerCategory {
  String get wireValue {
    switch (this) {
      case PoiMarkerCategory.generic:
        return 'generic';
      case PoiMarkerCategory.coffeehouse:
        return 'coffeehouse';
      case PoiMarkerCategory.tavern:
        return 'tavern';
      case PoiMarkerCategory.eatery:
        return 'eatery';
      case PoiMarkerCategory.market:
        return 'market';
      case PoiMarkerCategory.archive:
        return 'archive';
      case PoiMarkerCategory.park:
        return 'park';
      case PoiMarkerCategory.waterfront:
        return 'waterfront';
      case PoiMarkerCategory.museum:
        return 'museum';
      case PoiMarkerCategory.theater:
        return 'theater';
      case PoiMarkerCategory.landmark:
        return 'landmark';
      case PoiMarkerCategory.civic:
        return 'civic';
      case PoiMarkerCategory.arena:
        return 'arena';
    }
  }
}

PoiMarkerCategory parsePoiMarkerCategory(String? raw) {
  switch (raw?.trim().toLowerCase()) {
    case 'coffeehouse':
      return PoiMarkerCategory.coffeehouse;
    case 'tavern':
      return PoiMarkerCategory.tavern;
    case 'eatery':
      return PoiMarkerCategory.eatery;
    case 'market':
      return PoiMarkerCategory.market;
    case 'archive':
      return PoiMarkerCategory.archive;
    case 'park':
      return PoiMarkerCategory.park;
    case 'waterfront':
      return PoiMarkerCategory.waterfront;
    case 'museum':
      return PoiMarkerCategory.museum;
    case 'theater':
      return PoiMarkerCategory.theater;
    case 'landmark':
      return PoiMarkerCategory.landmark;
    case 'civic':
      return PoiMarkerCategory.civic;
    case 'arena':
      return PoiMarkerCategory.arena;
    default:
      return PoiMarkerCategory.generic;
  }
}

/// Lightweight tag as returned on a POI (id, name).
class PoiTag {
  final String id;
  final String name;

  const PoiTag({required this.id, required this.name});

  factory PoiTag.fromJson(Map<String, dynamic> json) {
    return PoiTag(
      id: json['id']?.toString() ?? '',
      name: json['name'] as String? ?? '',
    );
  }
}

class PointOfInterest {
  final String id;
  final String name;
  final String lat;
  final String lng;
  final String? imageURL;
  final String? thumbnailUrl;
  final String? description;
  final String? clue;
  final String? originalName;
  final String? googleMapsPlaceId;
  final String? googleMapsPlaceName;
  final PoiMarkerCategory markerCategory;
  final List<PoiTag> tags;
  final List<Character> characters;
  final bool hasAvailableQuest;
  final bool hasAvailableMainStoryQuest;

  const PointOfInterest({
    required this.id,
    required this.name,
    required this.lat,
    required this.lng,
    this.imageURL,
    this.thumbnailUrl,
    this.description,
    this.clue,
    this.originalName,
    this.googleMapsPlaceId,
    this.googleMapsPlaceName,
    this.markerCategory = PoiMarkerCategory.generic,
    this.tags = const [],
    this.characters = const [],
    this.hasAvailableQuest = false,
    this.hasAvailableMainStoryQuest = false,
  });

  factory PointOfInterest.fromJson(Map<String, dynamic> json) {
    List<PoiTag> tags = [];
    final raw = json['tags'];
    if (raw is List) {
      for (final t in raw) {
        if (t is Map<String, dynamic>) {
          try {
            tags.add(PoiTag.fromJson(t));
          } catch (_) {}
        }
      }
    }
    List<Character> characters = [];
    final rawCharacters = json['characters'];
    if (rawCharacters is List) {
      for (final c in rawCharacters) {
        if (c is Map<String, dynamic>) {
          try {
            characters.add(Character.fromJson(c));
          } catch (_) {}
        }
      }
    }
    return PointOfInterest(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      lat: json['lat']?.toString() ?? '0',
      lng: json['lng']?.toString() ?? '0',
      imageURL: json['imageURL'] as String?,
      thumbnailUrl: json['thumbnailUrl'] as String?,
      description: json['description'] as String?,
      clue: json['clue'] as String?,
      originalName: json['originalName'] as String?,
      googleMapsPlaceId: json['googleMapsPlaceId'] as String?,
      googleMapsPlaceName: json['googleMapsPlaceName'] as String?,
      markerCategory: parsePoiMarkerCategory(
        json['markerCategory']?.toString(),
      ),
      tags: tags,
      characters: characters,
      hasAvailableQuest: json['hasAvailableQuest'] as bool? ?? false,
      hasAvailableMainStoryQuest:
          json['hasAvailableMainStoryQuest'] as bool? ?? false,
    );
  }
}
