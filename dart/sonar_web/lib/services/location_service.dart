import 'package:geolocator/geolocator.dart';

import '../models/location.dart';

class LocationService {
  static const LocationSettings _defaultStreamSettings = LocationSettings(
    accuracy: LocationAccuracy.high,
    distanceFilter: 5,
  );

  Future<bool> checkPermission() async {
    bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) return false;
    LocationPermission p = await Geolocator.checkPermission();
    if (p == LocationPermission.denied) {
      p = await Geolocator.requestPermission();
    }
    return p == LocationPermission.whileInUse ||
        p == LocationPermission.always;
  }

  Future<AppLocation?> getCurrentLocation({bool requestPermission = true}) async {
    if (requestPermission) {
      final ok = await checkPermission();
      if (!ok) return null;
    }
    try {
      final pos = await Geolocator.getCurrentPosition(
        locationSettings: const LocationSettings(
          accuracy: LocationAccuracy.high,
        ),
      );
      return AppLocation(
        latitude: pos.latitude,
        longitude: pos.longitude,
        accuracy: pos.accuracy,
      );
    } catch (_) {
      return null;
    }
  }

  Stream<AppLocation> getLocationStream({bool requestPermission = true}) async* {
    if (requestPermission) {
      final ok = await checkPermission();
      if (!ok) return;
    }
    yield* Geolocator.getPositionStream(
      locationSettings: _defaultStreamSettings,
    ).map(
      (pos) => AppLocation(
        latitude: pos.latitude,
        longitude: pos.longitude,
        accuracy: pos.accuracy,
      ),
    );
  }
}
