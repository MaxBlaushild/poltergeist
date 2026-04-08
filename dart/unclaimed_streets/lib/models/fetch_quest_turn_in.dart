import 'inventory_item.dart';

class FetchQuestTurnInRequirement {
  final int inventoryItemId;
  final int quantity;
  final int ownedQuantity;
  final InventoryItem? inventoryItem;

  const FetchQuestTurnInRequirement({
    required this.inventoryItemId,
    required this.quantity,
    required this.ownedQuantity,
    this.inventoryItem,
  });

  factory FetchQuestTurnInRequirement.fromJson(Map<String, dynamic> json) {
    final rawItem = json['inventoryItem'];
    InventoryItem? inventoryItem;
    if (rawItem is Map<String, dynamic>) {
      inventoryItem = InventoryItem.fromJson(rawItem);
    } else if (rawItem is Map) {
      inventoryItem = InventoryItem.fromJson(
        Map<String, dynamic>.from(rawItem),
      );
    }
    return FetchQuestTurnInRequirement(
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt() ?? 0,
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
      ownedQuantity: (json['ownedQuantity'] as num?)?.toInt() ?? 0,
      inventoryItem: inventoryItem,
    );
  }

  bool get hasEnough => ownedQuantity >= quantity;
}

class FetchQuestTurnInDetails {
  final String questId;
  final String questName;
  final String questDescription;
  final String questNodeId;
  final String characterId;
  final String characterName;
  final List<FetchQuestTurnInRequirement> requirements;
  final bool canDeliver;

  const FetchQuestTurnInDetails({
    required this.questId,
    required this.questName,
    required this.questDescription,
    required this.questNodeId,
    required this.characterId,
    required this.characterName,
    this.requirements = const [],
    this.canDeliver = false,
  });

  factory FetchQuestTurnInDetails.fromJson(Map<String, dynamic> json) {
    return FetchQuestTurnInDetails(
      questId: json['questId']?.toString() ?? '',
      questName: json['questName']?.toString() ?? '',
      questDescription: json['questDescription']?.toString() ?? '',
      questNodeId: json['questNodeId']?.toString() ?? '',
      characterId: json['characterId']?.toString() ?? '',
      characterName: json['characterName']?.toString() ?? '',
      requirements:
          (json['requirements'] as List<dynamic>?)
              ?.map((entry) {
                if (entry is Map<String, dynamic>) {
                  return FetchQuestTurnInRequirement.fromJson(entry);
                }
                if (entry is Map) {
                  return FetchQuestTurnInRequirement.fromJson(
                    Map<String, dynamic>.from(entry),
                  );
                }
                return null;
              })
              .whereType<FetchQuestTurnInRequirement>()
              .toList() ??
          const [],
      canDeliver: json['canDeliver'] as bool? ?? false,
    );
  }
}
