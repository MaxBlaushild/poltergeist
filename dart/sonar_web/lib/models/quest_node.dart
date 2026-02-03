import 'quest_node_challenge.dart';

class QuestNode {
  final String id;
  final String questId;
  final int orderIndex;
  final String? pointOfInterestId;
  final String? polygon;
  final List<QuestNodeChallenge> challenges;

  const QuestNode({
    required this.id,
    required this.questId,
    required this.orderIndex,
    this.pointOfInterestId,
    this.polygon,
    this.challenges = const [],
  });

  factory QuestNode.fromJson(Map<String, dynamic> json) {
    return QuestNode(
      id: json['id'] as String? ?? '',
      questId: json['questId'] as String? ?? '',
      orderIndex: (json['orderIndex'] as num?)?.toInt() ?? 0,
      pointOfInterestId: json['pointOfInterestId'] as String?,
      polygon: json['polygon'] as String?,
      challenges: (json['challenges'] as List<dynamic>?)
              ?.map((e) => QuestNodeChallenge.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );
  }
}
