import 'dart:async';

import 'package:flutter/foundation.dart';

import '../models/location.dart';
import '../services/location_service.dart';

class LocationProvider with ChangeNotifier {
  final LocationService _locationService;
  AppLocation? _location;
  StreamSubscription<AppLocation>? _subscription;
  bool _loading = true;
  String? _error;
  bool _initialized = false;

  LocationProvider(this._locationService) {
    _loading = false;
  }

  AppLocation? get location => _location;
  bool get loading => _loading;
  String? get error => _error;

  Future<void> ensureLoaded() async {
    if (!_initialized) {
      await _init();
      return;
    }
    await _startUpdates();
  }

  Future<void> _init() async {
    _loading = true;
    _error = null;
    notifyListeners();
    final ok = await _locationService.checkPermission();
    if (!ok) {
      _error = 'Location permission denied or disabled.';
      _loading = false;
      notifyListeners();
      return;
    }
    _location = await _locationService.getCurrentLocation(requestPermission: false);
    _initialized = true;
    _loading = false;
    notifyListeners();
    await _startUpdates(requestPermission: false);
  }

  Future<void> refresh() async {
    _loading = true;
    _error = null;
    notifyListeners();
    final ok = await _locationService.checkPermission();
    if (!ok) {
      _error = 'Location permission denied or disabled.';
      _loading = false;
      notifyListeners();
      return;
    }
    _location = await _locationService.getCurrentLocation(requestPermission: false);
    _initialized = true;
    _loading = false;
    notifyListeners();
    await _startUpdates(requestPermission: false);
  }

  Future<void> _startUpdates({bool requestPermission = true}) async {
    if (_subscription != null) return;
    _subscription = _locationService
        .getLocationStream(requestPermission: requestPermission)
        .listen(
      (loc) {
        if (!_isValidLocation(loc)) return;
        _location = loc;
        _initialized = true;
        _loading = false;
        _error = null;
        notifyListeners();
      },
      onError: (e) {
        _error = e.toString();
        _loading = false;
        notifyListeners();
      },
    );
  }

  bool _isValidLocation(AppLocation loc) {
    final lat = loc.latitude;
    final lng = loc.longitude;
    if (!lat.isFinite || !lng.isFinite) return false;
    if (lat.abs() > 90 || lng.abs() > 180) return false;
    return true;
  }

  @override
  void dispose() {
    _subscription?.cancel();
    super.dispose();
  }
}
