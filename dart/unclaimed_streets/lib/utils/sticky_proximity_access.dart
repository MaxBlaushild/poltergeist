import '../models/location.dart';

class StickyProximityAccess {
  AppLocation? _grantedLocation;

  bool resolve({
    required AppLocation? currentLocation,
    required bool withinRange,
  }) {
    if (withinRange && currentLocation != null) {
      _grantedLocation ??= currentLocation;
    }
    return _grantedLocation != null || withinRange;
  }

  AppLocation? get grantedLocation => _grantedLocation;

  bool get granted => _grantedLocation != null;

  void reset() {
    _grantedLocation = null;
  }
}
