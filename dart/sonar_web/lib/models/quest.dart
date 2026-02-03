import 'quest_node.dart';

class Quest {
  final String id;
  final String name;
  final String description;
  final String? imageUrl;
  final String? zoneId;
  final String? questArchetypeId;
  final String? questGiverCharacterId;
  final int gold;
  final List<QuestNode> nodes;

  const Quest({
    required this.id,
    required this.name,
    required this.description,
    this.imageUrl,
    this.zoneId,
    this.questArchetypeId,
    this.questGiverCharacterId,
    this.gold = 0,
    this.nodes = const [],
  });

  factory Quest.fromJson(Map<String, dynamic> json) {
    return Quest(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      description: json['description'] as String? ?? '',
      imageUrl: json['imageUrl'] as String?,
      zoneId: json['zoneId'] as String?,
      questArchetypeId: json['questArchetypeId'] as String?,
      questGiverCharacterId: json['questGiverCharacterId'] as String?,
      gold: (json['gold'] as num?)?.toInt() ?? 0,
      nodes: (json['nodes'] as List<dynamic>?)
              ?.map((e) => QuestNode.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );
  }
}
