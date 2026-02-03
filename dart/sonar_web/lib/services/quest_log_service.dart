import '../models/quest.dart';
import 'api_client.dart';

class QuestLogService {
  final ApiClient _api;

  QuestLogService(this._api);

  /// GET /sonar/questlog?zoneId=...&tags=name1,name2
  Future<QuestLog> getQuestLog(String zoneId, {List<String> tags = const []}) async {
    final params = <String, dynamic>{'zoneId': zoneId};
    if (tags.isNotEmpty) {
      params['tags'] = tags.join(',');
    }
    final raw = await _api.get<dynamic>('/sonar/questlog', params: params);
    final map = raw is Map ? Map<String, dynamic>.from(raw as Map<dynamic, dynamic>) : <String, dynamic>{};
    return QuestLog.fromJson(map);
  }

  /// POST /sonar/trackedPointOfInterestGroups { pointOfInterestGroupID: questId }
  Future<void> trackQuest(String questId) async {
    await _api.post<dynamic>(
      '/sonar/trackedPointOfInterestGroups',
      data: {'pointOfInterestGroupID': questId},
    );
  }

  /// DELETE /sonar/trackedPointOfInterestGroups/:id
  Future<void> untrackQuest(String questId) async {
    await _api.delete<dynamic>('/sonar/trackedPointOfInterestGroups/$questId');
  }

  /// DELETE /sonar/trackedPointOfInterestGroups
  Future<void> untrackAllQuests() async {
    await _api.delete<dynamic>('/sonar/trackedPointOfInterestGroups');
  }

  /// POST /sonar/quests/:questId/turnIn
  /// Returns { goldAwarded: int, itemAwarded?: { id, name, imageUrl } }
  Future<Map<String, dynamic>> turnInQuest(String questId) async {
    final raw = await _api.post<dynamic>('/sonar/quests/turnIn/$questId');
    final map = raw is Map
        ? Map<String, dynamic>.from(raw as Map<dynamic, dynamic>)
        : <String, dynamic>{};
    return map;
  }
}
