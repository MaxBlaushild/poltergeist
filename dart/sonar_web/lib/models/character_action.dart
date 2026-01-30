import 'character.dart';

class DialogueMessage {
  final String speaker;
  final String text;
  final int order;

  const DialogueMessage({
    required this.speaker,
    required this.text,
    required this.order,
  });

  factory DialogueMessage.fromJson(Map<String, dynamic> json) {
    return DialogueMessage(
      speaker: json['speaker'] as String? ?? 'character',
      text: json['text'] as String? ?? '',
      order: (json['order'] as num?)?.toInt() ?? 0,
    );
  }
}

class ShopInventoryItem {
  final int itemId;
  final int price;

  const ShopInventoryItem({
    required this.itemId,
    required this.price,
  });

  factory ShopInventoryItem.fromJson(Map<String, dynamic> json) {
    return ShopInventoryItem(
      itemId: (json['itemId'] as num?)?.toInt() ?? 0,
      price: (json['price'] as num?)?.toInt() ?? 0,
    );
  }
}

class CharacterAction {
  final String id;
  final String createdAt;
  final String updatedAt;
  final String characterId;
  final String actionType;
  final List<DialogueMessage> dialogue;
  final Map<String, dynamic>? metadata;

  const CharacterAction({
    required this.id,
    required this.createdAt,
    required this.updatedAt,
    required this.characterId,
    required this.actionType,
    required this.dialogue,
    this.metadata,
  });

  factory CharacterAction.fromJson(Map<String, dynamic> json) {
    return CharacterAction(
      id: json['id'] as String,
      createdAt: json['createdAt']?.toString() ?? '',
      updatedAt: json['updatedAt']?.toString() ?? '',
      characterId: json['characterId'] as String? ?? '',
      actionType: json['actionType'] as String? ?? '',
      dialogue: (json['dialogue'] as List<dynamic>?)
              ?.map((e) => DialogueMessage.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      metadata: json['metadata'] as Map<String, dynamic>?,
    );
  }

  List<ShopInventoryItem>? get shopInventory {
    final inv = metadata?['inventory'] as List<dynamic>?;
    if (inv == null) return null;
    return inv
        .map((e) => ShopInventoryItem.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  String? get pointOfInterestGroupId => metadata?['pointOfInterestGroupId'] as String?;
}
