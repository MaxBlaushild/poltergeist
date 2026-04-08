import 'inventory_item.dart';

class QuestNodeFetchRequirement {
  final int inventoryItemId;
  final int quantity;
  final InventoryItem? inventoryItem;

  const QuestNodeFetchRequirement({
    required this.inventoryItemId,
    required this.quantity,
    this.inventoryItem,
  });

  factory QuestNodeFetchRequirement.fromJson(Map<String, dynamic> json) {
    final rawItem = json['inventoryItem'];
    InventoryItem? inventoryItem;
    if (rawItem is Map<String, dynamic>) {
      inventoryItem = InventoryItem.fromJson(rawItem);
    } else if (rawItem is Map) {
      inventoryItem = InventoryItem.fromJson(
        Map<String, dynamic>.from(rawItem),
      );
    }
    return QuestNodeFetchRequirement(
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt() ?? 0,
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
      inventoryItem: inventoryItem,
    );
  }
}

class QuestNodeObjective {
  static const typeChallenge = 'challenge';
  static const typeFetchQuest = 'fetch_quest';
  static const typeScenario = 'scenario';
  static const typeExposition = 'exposition';
  static const typeMonsterEncounter = 'monster_encounter';
  static const typeMonster = 'monster';

  final String id;
  final String type;
  final String prompt;
  final String description;
  final String imageUrl;
  final String thumbnailUrl;
  final int reward;
  final int? inventoryItemId;
  final String submissionType;
  final int difficulty;
  final List<String> statTags;
  final String? proficiency;
  final String? characterId;
  final String characterName;
  final List<QuestNodeFetchRequirement> fetchRequirements;

  const QuestNodeObjective({
    required this.id,
    required this.type,
    required this.prompt,
    this.description = '',
    this.imageUrl = '',
    this.thumbnailUrl = '',
    this.reward = 0,
    this.inventoryItemId,
    this.submissionType = 'photo',
    this.difficulty = 0,
    this.statTags = const [],
    this.proficiency,
    this.characterId,
    this.characterName = '',
    this.fetchRequirements = const [],
  });

  factory QuestNodeObjective.fromJson(Map<String, dynamic> json) {
    return QuestNodeObjective(
      id: json['id'] as String? ?? '',
      type: json['type'] as String? ?? '',
      prompt: json['prompt'] as String? ?? '',
      description: json['description'] as String? ?? '',
      imageUrl: json['imageUrl'] as String? ?? '',
      thumbnailUrl: json['thumbnailUrl'] as String? ?? '',
      reward: (json['reward'] as num?)?.toInt() ?? 0,
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt(),
      submissionType: json['submissionType'] as String? ?? 'photo',
      difficulty: (json['difficulty'] as num?)?.toInt() ?? 0,
      statTags:
          (json['statTags'] as List<dynamic>?)
              ?.map((tag) => tag.toString())
              .toList() ??
          const [],
      proficiency: json['proficiency'] as String?,
      characterId: json['characterId']?.toString(),
      characterName: json['characterName'] as String? ?? '',
      fetchRequirements:
          (json['fetchRequirements'] as List<dynamic>?)
              ?.map((entry) {
                if (entry is Map<String, dynamic>) {
                  return QuestNodeFetchRequirement.fromJson(entry);
                }
                if (entry is Map) {
                  return QuestNodeFetchRequirement.fromJson(
                    Map<String, dynamic>.from(entry),
                  );
                }
                return null;
              })
              .whereType<QuestNodeFetchRequirement>()
              .toList() ??
          const [],
    );
  }
}
