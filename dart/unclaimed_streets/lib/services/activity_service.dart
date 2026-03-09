import '../models/activity_feed.dart';
import 'api_client.dart';

class ActivityService {
  final ApiClient _api;

  ActivityService(this._api);

  Future<List<ActivityFeed>> getActivities() async {
    try {
      final list = await _api.get<List<dynamic>>('/sonar/activities');
      return list
          .map((e) => ActivityFeed.fromJson(e as Map<String, dynamic>))
          .toList();
    } catch (_) {
      return [];
    }
  }

  Future<void> markAsSeen(List<String> activityIds) async {
    if (activityIds.isEmpty) return;
    await _api.post<dynamic>(
      '/sonar/activities/markAsSeen',
      data: {'activityIDs': activityIds},
    );
  }
}
