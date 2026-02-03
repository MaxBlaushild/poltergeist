import 'quest.dart';

class QuestLog {
  final List<Quest> quests;
  final List<String> trackedQuestIds;

  const QuestLog({
    this.quests = const [],
    this.trackedQuestIds = const [],
  });

  factory QuestLog.fromJson(Map<String, dynamic> json) {
    return QuestLog(
      quests: (json['quests'] as List<dynamic>?)
              ?.map((e) => Quest.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      trackedQuestIds: (json['trackedQuestIds'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          const [],
    );
  }
}
