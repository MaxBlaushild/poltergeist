class Zone {
  final String id;
  final String name;
  final String? description;
  final double latitude;
  final double longitude;
  final double radius;
  final String? boundary; // WKT string format
  final List<LatLngCoords>? boundaryCoords;
  final List<LatLngCoords>? points;

  const Zone({
    required this.id,
    required this.name,
    this.description,
    required this.latitude,
    required this.longitude,
    this.radius = 0,
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
      radius: (json['radius'] as num?)?.toDouble() ?? 0,
      boundary: json['boundary'] as String?,
      boundaryCoords: (json['boundaryCoords'] as List<dynamic>?)
          ?.map((e) => LatLngCoords.fromJson(e as Map<String, dynamic>))
          .toList(),
      points: (json['points'] as List<dynamic>?)
          ?.map((e) => LatLngCoords.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  /// Ordered ring (lat/lng) for polygon outline. Uses points, boundaryCoords, or parsed boundary WKT. Null if none.
  List<LatLngCoords>? get ring {
    if (points != null && points!.isNotEmpty) return points;
    if (boundaryCoords != null && boundaryCoords!.isNotEmpty) return boundaryCoords;
    final coords = _parseBoundaryWkt(boundary);
    return coords != null && coords.isNotEmpty ? coords : null;
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
}

class LatLngCoords {
  final double latitude;
  final double longitude;

  const LatLngCoords({required this.latitude, required this.longitude});

  factory LatLngCoords.fromJson(Map<String, dynamic> json) {
    return LatLngCoords(
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
    );
  }
}
