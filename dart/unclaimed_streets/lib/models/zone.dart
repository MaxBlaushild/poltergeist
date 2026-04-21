import 'dart:math' as math;

class ZoneGenre {
  final String id;
  final String name;
  final int sortOrder;
  final bool active;

  const ZoneGenre({
    required this.id,
    required this.name,
    required this.sortOrder,
    required this.active,
  });

  factory ZoneGenre.fromJson(Map<String, dynamic> json) {
    return ZoneGenre(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      sortOrder: (json['sortOrder'] as num?)?.toInt() ?? 0,
      active: json['active'] != false,
    );
  }
}

class ZoneGenreScore {
  final String genreId;
  final ZoneGenre genre;
  final int score;

  const ZoneGenreScore({
    required this.genreId,
    required this.genre,
    required this.score,
  });

  factory ZoneGenreScore.fromJson(Map<String, dynamic> json) {
    final rawGenre = json['genre'];
    final genre = rawGenre is Map<String, dynamic>
        ? ZoneGenre.fromJson(rawGenre)
        : rawGenre is Map
        ? ZoneGenre.fromJson(Map<String, dynamic>.from(rawGenre))
        : ZoneGenre(
            id: json['genreId']?.toString() ?? '',
            name: '',
            sortOrder: 0,
            active: true,
          );
    return ZoneGenreScore(
      genreId: json['genreId']?.toString() ?? genre.id,
      genre: genre,
      score: (json['score'] as num?)?.toInt() ?? 0,
    );
  }
}

class Zone {
  final String id;
  final String name;
  final String? description;
  final String? kind;
  final String? kindOverlayColor;
  final double latitude;
  final double longitude;
  final bool discovered;
  final String? discoveredAt;
  final String? boundary; // WKT string format
  final List<LatLngCoords>? boundaryCoords;
  final List<LatLngCoords>? points;
  final List<ZoneGenreScore> genreScores;

  const Zone({
    required this.id,
    required this.name,
    this.description,
    this.kind,
    this.kindOverlayColor,
    required this.latitude,
    required this.longitude,
    this.discovered = false,
    this.discoveredAt,
    this.boundary,
    this.boundaryCoords,
    this.points,
    this.genreScores = const [],
  });

  factory Zone.fromJson(Map<String, dynamic> json) {
    return Zone(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      description: json['description'] as String?,
      kind: json['kind']?.toString(),
      kindOverlayColor: json['kindOverlayColor']?.toString(),
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0,
      discovered: json['discovered'] == true,
      discoveredAt: json['discoveredAt']?.toString(),
      boundary: json['boundary'] as String?,
      boundaryCoords: (json['boundaryCoords'] as List<dynamic>?)
          ?.map((e) => LatLngCoords.fromJsonSafe(e as Map<String, dynamic>))
          .whereType<LatLngCoords>()
          .toList(),
      points: (json['points'] as List<dynamic>?)
          ?.map((e) => LatLngCoords.fromJsonSafe(e as Map<String, dynamic>))
          .whereType<LatLngCoords>()
          .toList(),
      genreScores:
          (json['genreScores'] as List<dynamic>?)
              ?.whereType<Map>()
              .map(
                (entry) =>
                    ZoneGenreScore.fromJson(Map<String, dynamic>.from(entry)),
              )
              .toList(growable: false) ??
          const <ZoneGenreScore>[],
    );
  }

  Zone copyWith({
    String? id,
    String? name,
    String? description,
    String? kind,
    String? kindOverlayColor,
    double? latitude,
    double? longitude,
    bool? discovered,
    String? discoveredAt,
    String? boundary,
    List<LatLngCoords>? boundaryCoords,
    List<LatLngCoords>? points,
    List<ZoneGenreScore>? genreScores,
  }) {
    return Zone(
      id: id ?? this.id,
      name: name ?? this.name,
      description: description ?? this.description,
      kind: kind ?? this.kind,
      kindOverlayColor: kindOverlayColor ?? this.kindOverlayColor,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
      discovered: discovered ?? this.discovered,
      discoveredAt: discoveredAt ?? this.discoveredAt,
      boundary: boundary ?? this.boundary,
      boundaryCoords: boundaryCoords ?? this.boundaryCoords,
      points: points ?? this.points,
      genreScores: genreScores ?? this.genreScores,
    );
  }

  /// Ordered ring (lat/lng) for polygon outline.
  ///
  /// Prefer raw boundary points first because they have historically been the
  /// most reliable source for player-facing zone rendering in this client.
  List<LatLngCoords>? get ring {
    if (points != null && points!.isNotEmpty) {
      return _orderPointsByAngle(points!);
    }
    if (boundaryCoords != null && boundaryCoords!.isNotEmpty) {
      return boundaryCoords;
    }
    final coords = _parseBoundaryWkt(boundary);
    if (coords != null && coords.isNotEmpty) return coords;
    return null;
  }

  /// Parse POLYGON((lng lat, lng lat, ...)) WKT into [LatLngCoords] (lat, lng).
  static List<LatLngCoords>? _parseBoundaryWkt(String? wkt) {
    if (wkt == null || wkt.isEmpty) return null;
    final match = RegExp(
      r'POLYGON\s*\(\s*\(\s*(.+?)\s*\)\s*\)',
      caseSensitive: false,
    ).firstMatch(wkt);
    if (match == null) return null;
    final parts = match.group(1)!.split(',');
    final coords = <LatLngCoords>[];
    for (final p in parts) {
      final nums = p.trim().split(RegExp(r'\s+'));
      if (nums.length >= 2) {
        final lng = double.tryParse(nums[0]);
        final lat = double.tryParse(nums[1]);
        if (lng != null && lat != null) {
          coords.add(LatLngCoords(latitude: lat, longitude: lng));
        }
      }
    }
    return coords.isEmpty ? null : coords;
  }

  static List<LatLngCoords> _orderPointsByAngle(List<LatLngCoords> points) {
    if (points.length <= 2) return points;
    double sumLat = 0;
    double sumLng = 0;
    for (final p in points) {
      sumLat += p.latitude;
      sumLng += p.longitude;
    }
    final centerLat = sumLat / points.length;
    final centerLng = sumLng / points.length;
    final ordered = List<LatLngCoords>.from(points);
    ordered.sort((a, b) {
      final angleA = math.atan2(
        a.latitude - centerLat,
        a.longitude - centerLng,
      );
      final angleB = math.atan2(
        b.latitude - centerLat,
        b.longitude - centerLng,
      );
      return angleA.compareTo(angleB);
    });
    return ordered;
  }
}

class LatLngCoords {
  final double latitude;
  final double longitude;

  const LatLngCoords({required this.latitude, required this.longitude});

  factory LatLngCoords.fromJson(Map<String, dynamic> json) {
    final coords = fromJsonSafe(json);
    if (coords == null) {
      throw FormatException('Invalid LatLngCoords: $json');
    }
    return coords;
  }

  static LatLngCoords? fromJsonSafe(Map<String, dynamic> json) {
    final latitude = _parseDouble(json['latitude'] ?? json['lat']);
    final longitude = _parseDouble(json['longitude'] ?? json['lng']);
    if (latitude == null || longitude == null) {
      return null;
    }
    return LatLngCoords(latitude: latitude, longitude: longitude);
  }

  static double? _parseDouble(dynamic value) {
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value);
    return null;
  }
}
