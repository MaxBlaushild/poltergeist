import 'package:flutter_test/flutter_test.dart';
import 'package:unclaimed_streets/models/zone.dart';
import 'package:unclaimed_streets/providers/zone_provider.dart';

void main() {
  test(
    'findZoneAtCoordinate prefers boundaryCoords over raw points when both exist',
    () {
      final provider = ZoneProvider();
      const zone = Zone(
        id: 'zone-1',
        name: 'Forest Edge',
        latitude: 0,
        longitude: 0,
        boundaryCoords: <LatLngCoords>[
          LatLngCoords(latitude: 0, longitude: 0),
          LatLngCoords(latitude: 0, longitude: 10),
          LatLngCoords(latitude: 10, longitude: 10),
          LatLngCoords(latitude: 10, longitude: 0),
          LatLngCoords(latitude: 0, longitude: 0),
        ],
        points: <LatLngCoords>[
          LatLngCoords(latitude: 20, longitude: 20),
          LatLngCoords(latitude: 20, longitude: 30),
          LatLngCoords(latitude: 30, longitude: 25),
        ],
      );

      provider.setZones(const <Zone>[zone]);

      final foundZone = provider.findZoneAtCoordinate(5, 5);

      expect(foundZone?.id, zone.id);
    },
  );

  test(
    'ring falls back to ordered raw points when no boundary is available',
    () {
      const zone = Zone(
        id: 'zone-2',
        name: 'Fallback',
        latitude: 0,
        longitude: 0,
        points: <LatLngCoords>[
          LatLngCoords(latitude: 1, longitude: 1),
          LatLngCoords(latitude: 0, longitude: 0),
          LatLngCoords(latitude: 0, longitude: 1),
        ],
      );

      final ring = zone.ring;

      expect(ring, isNotNull);
      expect(ring, hasLength(3));
    },
  );
}
