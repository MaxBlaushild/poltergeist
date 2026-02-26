import 'dart:math' as math;

class Zone {
  final String id;
  final String name;
  final String? description;
  final double latitude;
  final double longitude;
  final String? boundary; // WKT string format
  final List<LatLngCoords>? boundaryCoords;
  final List<LatLngCoords>? points;

  const Zone({
    required this.id,
    required this.name,
    this.description,
    required this.latitude,
    required this.longitude,
    this.boundary,
    this.boundaryCoords,
    this.points,
  });

  factory Zone.fromJson(Map<String, dynamic> json) {
    return Zone(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      description: json['description'] as String?,
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0,
      boundary: json['boundary'] as String?,
      boundaryCoords: (json['boundaryCoords'] as List<dynamic>?)
          ?.map((e) => LatLngCoords.fromJsonSafe(e as Map<String, dynamic>))
          .whereType<LatLngCoords>()
          .toList(),
      points: (json['points'] as List<dynamic>?)
          ?.map((e) => LatLngCoords.fromJsonSafe(e as Map<String, dynamic>))
          .whereType<LatLngCoords>()
          .toList(),
    );
  }

  /// Ordered ring (lat/lng) for polygon outline. Uses points, boundaryCoords, or parsed boundary WKT. Null if none.
  List<LatLngCoords>? get ring {
    if (points != null && points!.isNotEmpty) return _orderPointsByAngle(points!);
    if (boundaryCoords != null && boundaryCoords!.isNotEmpty) return boundaryCoords;
    final coords = _parseBoundaryWkt(boundary);
    if (coords != null && coords.isNotEmpty) return coords;
    return null;
  }

  /// Parse POLYGON((lng lat, lng lat, ...)) WKT into [LatLngCoords] (lat, lng).
  static List<LatLngCoords>? _parseBoundaryWkt(String? wkt) {
    if (wkt == null || wkt.isEmpty) return null;
    final match = RegExp(r'POLYGON\s*\(\s*\(\s*(.+?)\s*\)\s*\)', caseSensitive: false).firstMatch(wkt);
    if (match == null) return null;
    final parts = match.group(1)!.split(',');
    final coords = <LatLngCoords>[];
    for (final p in parts) {
      final nums = p.trim().split(RegExp(r'\s+'));
      if (nums.length >= 2) {
        final lng = double.tryParse(nums[0]);
        final lat = double.tryParse(nums[1]);
        if (lng != null && lat != null) coords.add(LatLngCoords(latitude: lat, longitude: lng));
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
      final angleA = math.atan2(a.latitude - centerLat, a.longitude - centerLng);
      final angleB = math.atan2(b.latitude - centerLat, b.longitude - centerLng);
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
