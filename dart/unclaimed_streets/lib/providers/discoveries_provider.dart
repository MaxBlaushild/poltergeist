import 'package:flutter/foundation.dart';

import '../models/point_of_interest_discovery.dart';
import '../providers/auth_provider.dart';
import '../services/poi_service.dart';

class DiscoveriesProvider with ChangeNotifier {
  final PoiService _poi;
  final AuthProvider _auth;

  List<PointOfInterestDiscovery> _discoveries = [];
  Set<String> _discoveredPoiIds = <String>{};
  bool _loading = false;
  Future<void>? _refreshFuture;

  DiscoveriesProvider(this._poi, this._auth) {
    _auth.addListener(_onAuthChanged);
  }

  void _onAuthChanged() {
    final u = _auth.user;
    if (u == null) {
      _discoveries = [];
      _discoveredPoiIds = <String>{};
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
  Set<String> get discoveredPoiIds => _discoveredPoiIds;
  bool get loading => _loading;

  /// True if the current user has discovered the given POI.
  bool hasDiscovered(String pointOfInterestId) {
    if (pointOfInterestId.isEmpty) return false;
    return _discoveredPoiIds.contains(pointOfInterestId);
  }

  Future<void> refresh() async {
    if (_refreshFuture != null) {
      return _refreshFuture!;
    }
    final uid = _auth.user?.id;
    if (uid == null || uid.isEmpty) {
      _discoveries = [];
      _discoveredPoiIds = <String>{};
      notifyListeners();
      return;
    }
    _loading = true;
    notifyListeners();

    final future = _refreshForUser(uid);
    _refreshFuture = future;
    await future;
  }

  Future<void> _refreshForUser(String uid) async {
    try {
      final discoveries = await _poi.getDiscoveries();
      if (_auth.user?.id != uid) return;
      _discoveries = discoveries;
      _discoveredPoiIds = discoveries
          .map((discovery) => discovery.pointOfInterestId)
          .where((id) => id.isNotEmpty)
          .toSet();
    } catch (_) {
      if (_auth.user?.id != uid) return;
      _discoveries = [];
      _discoveredPoiIds = <String>{};
    } finally {
      if (_auth.user?.id == uid) {
        _loading = false;
        notifyListeners();
      }
      _refreshFuture = null;
    }
  }
}
