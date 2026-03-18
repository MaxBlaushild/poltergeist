import 'inventory_item.dart';
import 'quest_node.dart';
import 'spell.dart';

class QuestItemReward {
  final int inventoryItemId;
  final int quantity;
  final InventoryItem? inventoryItem;

  const QuestItemReward({
    required this.inventoryItemId,
    required this.quantity,
    this.inventoryItem,
  });

  factory QuestItemReward.fromJson(Map<String, dynamic> json) {
    return QuestItemReward(
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt() ?? 0,
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
      inventoryItem: json['inventoryItem'] is Map<String, dynamic>
          ? InventoryItem.fromJson(
              json['inventoryItem'] as Map<String, dynamic>,
            )
          : null,
    );
  }
}

class QuestSpellReward {
  final String spellId;
  final Spell? spell;

  const QuestSpellReward({required this.spellId, this.spell});

  factory QuestSpellReward.fromJson(Map<String, dynamic> json) {
    Spell? spell;
    final rawSpell = json['spell'];
    if (rawSpell is Map<String, dynamic>) {
      spell = Spell.fromJson(rawSpell);
    } else if (rawSpell is Map) {
      spell = Spell.fromJson(Map<String, dynamic>.from(rawSpell));
    }
    return QuestSpellReward(
      spellId: json['spellId']?.toString() ?? '',
      spell: spell,
    );
  }
}

class QuestMaterialReward {
  final String resourceKey;
  final int amount;

  const QuestMaterialReward({
    required this.resourceKey,
    required this.amount,
  });

  factory QuestMaterialReward.fromJson(Map<String, dynamic> json) {
    return QuestMaterialReward(
      resourceKey: json['resourceKey']?.toString() ?? '',
      amount: (json['amount'] as num?)?.toInt() ?? 0,
    );
  }
}

class Quest {
  static const rewardModeExplicit = 'explicit';
  static const rewardModeRandom = 'random';
  static const randomRewardSizeSmall = 'small';
  static const randomRewardSizeMedium = 'medium';
  static const randomRewardSizeLarge = 'large';

  final String id;
  final String name;
  final String description;
  final List<String> acceptanceDialogue;
  final String? imageUrl;
  final String rewardMode;
  final String randomRewardSize;
  final String? zoneId;
  final String? questArchetypeId;
  final String? questGiverCharacterId;
  final String? recurringQuestId;
  final String? recurrenceFrequency;
  final DateTime? nextRecurrenceAt;
  final int gold;
  final List<QuestMaterialReward> materialRewards;
  final List<QuestItemReward> itemRewards;
  final List<QuestSpellReward> spellRewards;
  final List<QuestNode> nodes;
  final bool isAccepted;
  final DateTime? turnedInAt;
  final int completionCount;
  final bool readyToTurnIn;
  final QuestNode? currentNode;

  const Quest({
    required this.id,
    required this.name,
    required this.description,
    this.acceptanceDialogue = const [],
    this.imageUrl,
    this.rewardMode = rewardModeRandom,
    this.randomRewardSize = randomRewardSizeSmall,
    this.zoneId,
    this.questArchetypeId,
    this.questGiverCharacterId,
    this.recurringQuestId,
    this.recurrenceFrequency,
    this.nextRecurrenceAt,
    this.gold = 0,
    this.materialRewards = const [],
    this.itemRewards = const [],
    this.spellRewards = const [],
    this.nodes = const [],
    this.isAccepted = false,
    this.turnedInAt,
    this.completionCount = 0,
    this.readyToTurnIn = false,
    this.currentNode,
  });

  factory Quest.fromJson(Map<String, dynamic> json) {
    return Quest(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      description: json['description'] as String? ?? '',
      acceptanceDialogue:
          (json['acceptanceDialogue'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          const [],
      imageUrl: json['imageUrl'] as String?,
      rewardMode: json['rewardMode']?.toString() ?? rewardModeRandom,
      randomRewardSize:
          json['randomRewardSize']?.toString() ?? randomRewardSizeSmall,
      zoneId: json['zoneId'] as String?,
      questArchetypeId: json['questArchetypeId'] as String?,
      questGiverCharacterId: json['questGiverCharacterId'] as String?,
      recurringQuestId: json['recurringQuestId'] as String?,
      recurrenceFrequency: json['recurrenceFrequency'] as String?,
      nextRecurrenceAt: json['nextRecurrenceAt'] != null
          ? DateTime.tryParse(json['nextRecurrenceAt'] as String)
          : null,
      gold: (json['gold'] as num?)?.toInt() ?? 0,
      materialRewards:
          (json['materialRewards'] as List<dynamic>?)
              ?.map((e) {
                if (e is Map<String, dynamic>) {
                  return QuestMaterialReward.fromJson(e);
                }
                if (e is Map) {
                  return QuestMaterialReward.fromJson(
                    Map<String, dynamic>.from(e),
                  );
                }
                return null;
              })
              .whereType<QuestMaterialReward>()
              .toList() ??
          const [],
      itemRewards:
          (json['itemRewards'] as List<dynamic>?)
              ?.map((e) => QuestItemReward.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      spellRewards:
          (json['spellRewards'] as List<dynamic>?)
              ?.map((e) {
                if (e is Map<String, dynamic>) {
                  return QuestSpellReward.fromJson(e);
                }
                if (e is Map) {
                  return QuestSpellReward.fromJson(
                    Map<String, dynamic>.from(e),
                  );
                }
                return null;
              })
              .whereType<QuestSpellReward>()
              .toList() ??
          const [],
      nodes:
          (json['nodes'] as List<dynamic>?)
              ?.map((e) => QuestNode.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      isAccepted: json['isAccepted'] as bool? ?? false,
      turnedInAt: json['turnedInAt'] != null
          ? DateTime.tryParse(json['turnedInAt'] as String)
          : null,
      completionCount: (json['completionCount'] as num?)?.toInt() ?? 0,
      readyToTurnIn: json['readyToTurnIn'] as bool? ?? false,
      currentNode: json['currentNode'] is Map<String, dynamic>
          ? QuestNode.fromJson(json['currentNode'] as Map<String, dynamic>)
          : null,
    );
  }

  bool get hasRandomRewards => rewardMode == rewardModeRandom;
}
