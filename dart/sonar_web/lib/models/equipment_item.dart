import 'inventory_item.dart';

class EquippedItem {
  final String slot;
  final String ownedInventoryItemId;
  final int inventoryItemId;
  final InventoryItem? inventoryItem;

  const EquippedItem({
    required this.slot,
    required this.ownedInventoryItemId,
    required this.inventoryItemId,
    this.inventoryItem,
  });

  factory EquippedItem.fromJson(Map<String, dynamic> json) {
    return EquippedItem(
      slot: json['slot'] as String? ?? '',
      ownedInventoryItemId: json['ownedInventoryItemId'] as String? ?? '',
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt() ?? 0,
      inventoryItem: json['inventoryItem'] is Map<String, dynamic>
          ? InventoryItem.fromJson(json['inventoryItem'] as Map<String, dynamic>)
          : null,
    );
  }
}
