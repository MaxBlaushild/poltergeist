import 'inventory_item.dart';
import 'quest_node.dart';

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
          ? InventoryItem.fromJson(json['inventoryItem'] as Map<String, dynamic>)
          : null,
    );
  }
}

class Quest {
  final String id;
  final String name;
  final String description;
  final List<String> acceptanceDialogue;
  final String? imageUrl;
  final String? zoneId;
  final String? questArchetypeId;
  final String? questGiverCharacterId;
  final int gold;
  final List<QuestItemReward> itemRewards;
  final List<QuestNode> nodes;
  final bool isAccepted;
  final DateTime? turnedInAt;
  final bool readyToTurnIn;
  final QuestNode? currentNode;

  const Quest({
    required this.id,
    required this.name,
    required this.description,
    this.acceptanceDialogue = const [],
    this.imageUrl,
    this.zoneId,
    this.questArchetypeId,
    this.questGiverCharacterId,
    this.gold = 0,
    this.itemRewards = const [],
    this.nodes = const [],
    this.isAccepted = false,
    this.turnedInAt,
    this.readyToTurnIn = false,
    this.currentNode,
  });

  factory Quest.fromJson(Map<String, dynamic> json) {
    return Quest(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      description: json['description'] as String? ?? '',
      acceptanceDialogue: (json['acceptanceDialogue'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          const [],
      imageUrl: json['imageUrl'] as String?,
      zoneId: json['zoneId'] as String?,
      questArchetypeId: json['questArchetypeId'] as String?,
      questGiverCharacterId: json['questGiverCharacterId'] as String?,
      gold: (json['gold'] as num?)?.toInt() ?? 0,
      itemRewards: (json['itemRewards'] as List<dynamic>?)
              ?.map((e) => QuestItemReward.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      nodes: (json['nodes'] as List<dynamic>?)
              ?.map((e) => QuestNode.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      isAccepted: json['isAccepted'] as bool? ?? false,
      turnedInAt: json['turnedInAt'] != null
          ? DateTime.tryParse(json['turnedInAt'] as String)
          : null,
      readyToTurnIn: json['readyToTurnIn'] as bool? ?? false,
      currentNode: json['currentNode'] is Map<String, dynamic>
          ? QuestNode.fromJson(json['currentNode'] as Map<String, dynamic>)
          : null,
    );
  }
}
