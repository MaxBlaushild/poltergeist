import 'package:flutter/foundation.dart';

import '../models/activity_feed.dart';
import '../services/activity_service.dart';

class ActivityFeedProvider with ChangeNotifier {
  static const int _pageSize = 20;

  final ActivityService _service;

  ActivityFeedProvider(this._service);

  List<ActivityFeed> _activities = [];
  bool _loading = false;
  bool _loadingMore = false;
  bool _hasMore = true;
  int _nextOffset = 0;

  List<ActivityFeed> get activities => _activities;
  List<ActivityFeed> get unseenActivities =>
      _activities.where((a) => !a.seen).toList();
  bool get loading => _loading;
  bool get loadingMore => _loadingMore;
  bool get hasMore => _hasMore;

  Future<void> refresh() async {
    _loading = true;
    _loadingMore = false;
    _hasMore = true;
    _nextOffset = 0;
    notifyListeners();
    try {
      final firstPage = await _service.getActivities(limit: _pageSize);
      _activities = firstPage;
      _hasMore = firstPage.length == _pageSize;
      _nextOffset = firstPage.length;
    } catch (_) {
      _activities = [];
      _hasMore = false;
    }
    _loading = false;
    notifyListeners();
  }

  Future<void> loadMore() async {
    if (_loading || _loadingMore || !_hasMore) return;
    _loadingMore = true;
    notifyListeners();
    try {
      final nextPage = await _service.getActivities(
        limit: _pageSize,
        offset: _nextOffset,
      );
      _nextOffset += nextPage.length;
      if (nextPage.isEmpty) {
        _hasMore = false;
      } else {
        final seenIds = _activities.map((a) => a.id).toSet();
        final deduped = nextPage.where((a) => !seenIds.contains(a.id)).toList();
        _activities = [..._activities, ...deduped];
        _hasMore = nextPage.length == _pageSize;
      }
    } catch (_) {
      // Leave current items alone and allow another retry.
    }
    _loadingMore = false;
    notifyListeners();
  }

  Future<void> markAsSeen(List<String> activityIds) async {
    if (activityIds.isEmpty) return;
    await _service.markAsSeen(activityIds);
    _activities = _activities.map((a) {
      if (activityIds.contains(a.id)) {
        return ActivityFeed(
          id: a.id,
          userId: a.userId,
          activityType: a.activityType,
          data: a.data,
          seen: true,
          createdAt: a.createdAt,
          updatedAt: a.updatedAt,
        );
      }
      return a;
    }).toList();
    notifyListeners();
  }
}
