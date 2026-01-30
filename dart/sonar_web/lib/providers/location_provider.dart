import 'package:flutter/foundation.dart';

import '../models/location.dart';
import '../services/location_service.dart';

class LocationProvider with ChangeNotifier {
  final LocationService _locationService;
  AppLocation? _location;
  bool _loading = true;
  String? _error;

  LocationProvider(this._locationService) {
    _init();
  }

  AppLocation? get location => _location;
  bool get loading => _loading;
  String? get error => _error;

  Future<void> _init() async {
    _location = await _locationService.getCurrentLocation();
    _loading = false;
    notifyListeners();
  }

  Future<void> refresh() async {
    _loading = true;
    _error = null;
    notifyListeners();
    _location = await _locationService.getCurrentLocation();
    _loading = false;
    notifyListeners();
  }
}
