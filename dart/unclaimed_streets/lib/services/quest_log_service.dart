import '../models/quest_log.dart';
import '../models/fetch_quest_turn_in.dart';
import 'api_client.dart';

class QuestLogService {
  final ApiClient _api;

  QuestLogService(this._api);

  /// GET /sonar/questlog?zoneId=...&tags=name1,name2
  Future<QuestLog> getQuestLog({
    String? zoneId,
    List<String> tags = const [],
  }) async {
    final params = <String, dynamic>{};
    if (zoneId != null && zoneId.isNotEmpty) {
      params['zoneId'] = zoneId;
    }
    if (tags.isNotEmpty) {
      params['tags'] = tags.join(',');
    }
    final raw = await _api.get<dynamic>('/sonar/questlog', params: params);
    final map = raw is Map
        ? Map<String, dynamic>.from(raw)
        : <String, dynamic>{};
    return QuestLog.fromJson(map);
  }

  /// POST /sonar/trackedQuests { questId }
  Future<void> trackQuest(String questId) async {
    await _api.post<dynamic>(
      '/sonar/trackedQuests',
      data: {'questId': questId},
    );
  }

  /// DELETE /sonar/trackedQuests/:id
  Future<void> untrackQuest(String questId) async {
    await _api.delete<dynamic>('/sonar/trackedQuests/$questId');
  }

  /// DELETE /sonar/trackedQuests
  Future<void> untrackAllQuests() async {
    await _api.delete<dynamic>('/sonar/trackedQuests');
  }

  /// POST /sonar/quests/:questId/turnIn
  /// Returns { goldAwarded: int, itemsAwarded?: [{ id, name, imageUrl, quantity }] }
  Future<Map<String, dynamic>> turnInQuest(String questId) async {
    final raw = await _api.post<dynamic>('/sonar/quests/turnIn/$questId');
    final map = raw is Map
        ? Map<String, dynamic>.from(raw)
        : <String, dynamic>{};
    return map;
  }

  Future<void> shareQuest(String questId, String targetUserId) async {
    await _api.post<dynamic>(
      '/sonar/quests/$questId/share',
      data: {'targetUserId': targetUserId},
    );
  }

  /// DELETE /sonar/questAcceptances/:questId
  Future<void> forgetQuest(String questId) async {
    await _api.delete<dynamic>('/sonar/questAcceptances/$questId');
  }

  /// POST /sonar/questNodes/:id/submit
  /// Returns { successful: bool, reason: string, questCompleted: bool, score?: int, difficulty?: int, combinedScore?: int }
  Future<Map<String, dynamic>> submitQuestNode(
    String questNodeId, {
    String? textSubmission,
    String? imageSubmissionUrl,
    String? videoSubmissionUrl,
  }) async {
    final raw = await _api.post<dynamic>(
      '/sonar/questNodes/$questNodeId/submit',
      data: {
        if (textSubmission != null) 'textSubmission': textSubmission,
        if (imageSubmissionUrl != null)
          'imageSubmissionUrl': imageSubmissionUrl,
        if (videoSubmissionUrl != null)
          'videoSubmissionUrl': videoSubmissionUrl,
      },
    );
    final map = raw is Map
        ? Map<String, dynamic>.from(raw)
        : <String, dynamic>{};
    return map;
  }

  Future<FetchQuestTurnInDetails> getFetchQuestTurnIn(String questId) async {
    final raw = await _api.get<dynamic>('/sonar/quests/$questId/fetch-turn-in');
    final map = raw is Map
        ? Map<String, dynamic>.from(raw)
        : <String, dynamic>{};
    return FetchQuestTurnInDetails.fromJson(map);
  }

  Future<Map<String, dynamic>> submitFetchQuestTurnIn(String questId) async {
    final raw = await _api.post<dynamic>(
      '/sonar/quests/$questId/fetch-turn-in',
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }
}
