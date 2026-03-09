import 'package:flutter/foundation.dart';

import '../models/activity_feed.dart';
import '../services/activity_service.dart';

class ActivityFeedProvider with ChangeNotifier {
  final ActivityService _service;

  ActivityFeedProvider(this._service);

  List<ActivityFeed> _activities = [];
  bool _loading = false;

  List<ActivityFeed> get activities => _activities;
  List<ActivityFeed> get unseenActivities =>
      _activities.where((a) => !a.seen).toList();
  bool get loading => _loading;

  Future<void> refresh() async {
    _loading = true;
    notifyListeners();
    try {
      _activities = await _service.getActivities();
    } catch (_) {
      _activities = [];
    }
    _loading = false;
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
