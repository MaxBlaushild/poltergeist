import 'quest.dart';

class QuestLog {
  final List<Quest> quests;
  final List<Quest> completedQuests;
  final List<String> trackedQuestIds;

  const QuestLog({
    this.quests = const [],
    this.completedQuests = const [],
    this.trackedQuestIds = const [],
  });

  factory QuestLog.fromJson(Map<String, dynamic> json) {
    return QuestLog(
      quests: (json['quests'] as List<dynamic>?)
              ?.map((e) => Quest.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      completedQuests: (json['completedQuests'] as List<dynamic>?)
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
