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
  final String? description;
  final String? clue;
  final String? originalName;
  final String? googleMapsPlaceId;
  final List<PoiTag> tags;

  const PointOfInterest({
    required this.id,
    required this.name,
    required this.lat,
    required this.lng,
    this.imageURL,
    this.description,
    this.clue,
    this.originalName,
    this.googleMapsPlaceId,
    this.tags = const [],
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
    return PointOfInterest(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      lat: json['lat']?.toString() ?? '0',
      lng: json['lng']?.toString() ?? '0',
      imageURL: json['imageURL'] as String?,
      description: json['description'] as String?,
      clue: json['clue'] as String?,
      originalName: json['originalName'] as String?,
      googleMapsPlaceId: json['googleMapsPlaceId'] as String?,
      tags: tags,
    );
  }
}
