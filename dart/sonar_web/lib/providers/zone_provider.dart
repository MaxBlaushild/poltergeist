import 'package:flutter/foundation.dart';
import '../models/zone.dart';

class ZoneProvider extends ChangeNotifier {
  List<Zone> _zones = [];
  Zone? _selectedZone;

  List<Zone> get zones => _zones;
  Zone? get selectedZone => _selectedZone;

  void setZones(List<Zone> zones) {
    _zones = zones;
    notifyListeners();
  }

  void setSelectedZone(Zone? zone) {
    if (_selectedZone?.id != zone?.id) {
      _selectedZone = zone;
      notifyListeners();
    }
  }

  /// Simple point-in-polygon check using ray casting algorithm
  Zone? findZoneAtCoordinate(double latitude, double longitude) {
    if (_zones.isEmpty) return null;

    for (final zone in _zones) {
      final ring = zone.ring;
      if (ring == null || ring.isEmpty) continue;

      if (_isPointInPolygon(latitude, longitude, ring)) {
        return zone;
      }
    }

    return null;
  }

  bool _isPointInPolygon(double lat, double lng, List<LatLngCoords> polygon) {
    if (polygon.length < 3) return false;

    bool inside = false;
    int j = polygon.length - 1;

    for (int i = 0; i < polygon.length; i++) {
      final xi = polygon[i].longitude;
      final yi = polygon[i].latitude;
      final xj = polygon[j].longitude;
      final yj = polygon[j].latitude;

      final intersect = ((yi > lat) != (yj > lat)) &&
          (lng < (xj - xi) * (lat - yi) / (yj - yi) + xi);
      if (intersect) inside = !inside;
      j = i;
    }

    return inside;
  }
}
