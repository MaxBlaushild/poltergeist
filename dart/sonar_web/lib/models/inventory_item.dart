class InventoryItem {
  final int id;
  final String name;
  final String imageUrl;
  final String flavorText;
  final String effectText;
  final int? sellValue;
  final int? unlockTier;

  const InventoryItem({
    required this.id,
    required this.name,
    required this.imageUrl,
    required this.flavorText,
    required this.effectText,
    this.sellValue,
    this.unlockTier,
  });

  factory InventoryItem.fromJson(Map<String, dynamic> json) {
    return InventoryItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: json['name'] as String? ?? '',
      imageUrl: json['imageUrl'] as String? ?? '',
      flavorText: json['flavorText'] as String? ?? '',
      effectText: json['effectText'] as String? ?? '',
      sellValue: (json['sellValue'] as num?)?.toInt(),
      unlockTier: (json['unlockTier'] as num?)?.toInt(),
    );
  }
}

class OwnedInventoryItem {
  final String id;
  final int inventoryItemId;
  final int quantity;

  const OwnedInventoryItem({
    required this.id,
    required this.inventoryItemId,
    required this.quantity,
  });

  factory OwnedInventoryItem.fromJson(Map<String, dynamic> json) {
    return OwnedInventoryItem(
      id: json['id'] as String,
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt() ?? 0,
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
    );
  }
}
