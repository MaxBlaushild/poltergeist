import 'package:flutter/foundation.dart';

import '../models/point_of_interest_discovery.dart';
import '../providers/auth_provider.dart';
import '../services/poi_service.dart';

class DiscoveriesProvider with ChangeNotifier {
  final PoiService _poi;
  final AuthProvider _auth;

  List<PointOfInterestDiscovery> _discoveries = [];
  bool _loading = false;

  DiscoveriesProvider(this._poi, this._auth) {
    _auth.addListener(_onAuthChanged);
  }

  void _onAuthChanged() {
    final u = _auth.user;
    if (u == null) {
      _discoveries = [];
      notifyListeners();
      return;
    }
    // User just became available (e.g. after app load). Fetch discoveries so POIs
    // show correct unlocked state; otherwise they stay "locked" on first open.
    if (_discoveries.isEmpty && !_loading) {
      Future.microtask(() => refresh());
    }
  }

  List<PointOfInterestDiscovery> get discoveries => _discoveries;
  bool get loading => _loading;

  /// True if the current user has discovered the given POI.
  bool hasDiscovered(String pointOfInterestId) {
    final uid = _auth.user?.id;
    if (uid == null || uid.isEmpty) return false;
    return hasDiscoveredPointOfInterest(pointOfInterestId, uid, _discoveries);
  }

  Future<void> refresh() async {
    final uid = _auth.user?.id;
    if (uid == null || uid.isEmpty) {
      _discoveries = [];
      notifyListeners();
      return;
    }
    _loading = true;
    notifyListeners();
    try {
      _discoveries = await _poi.getDiscoveries();
    } catch (_) {
      _discoveries = [];
    }
    _loading = false;
    notifyListeners();
  }
}
