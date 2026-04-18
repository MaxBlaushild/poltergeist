import '../models/activity_feed.dart';
import 'api_client.dart';

class ActivityService {
  final ApiClient _api;

  ActivityService(this._api);

  Future<List<ActivityFeed>> getActivities({
    int limit = 20,
    int offset = 0,
  }) async {
    final list = await _api.get<List<dynamic>>(
      '/sonar/activities',
      params: {'limit': limit, 'offset': offset},
    );
    return list
        .map((e) => ActivityFeed.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> markAsSeen(List<String> activityIds) async {
    if (activityIds.isEmpty) return;
    await _api.post<dynamic>(
      '/sonar/activities/markAsSeen',
      data: {'activityIDs': activityIds},
    );
  }
}
