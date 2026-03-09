class TreasureChestItem {
  final String id;
  final int inventoryItemId;
  final int quantity;

  const TreasureChestItem({
    required this.id,
    required this.inventoryItemId,
    required this.quantity,
  });

  factory TreasureChestItem.fromJson(Map<String, dynamic> json) {
    return TreasureChestItem(
      id: json['id']?.toString() ?? '',
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt() ?? 0,
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
    );
  }
}

class TreasureChest {
  final String id;
  final double latitude;
  final double longitude;
  final String zoneId;
  final int? gold;
  final bool? openedByUser;
  final int? unlockTier;
  final List<TreasureChestItem> items;

  const TreasureChest({
    required this.id,
    required this.latitude,
    required this.longitude,
    required this.zoneId,
    this.gold,
    this.openedByUser,
    this.unlockTier,
    this.items = const [],
  });

  factory TreasureChest.fromJson(Map<String, dynamic> json) {
    final raw = json['items'];
    final itemList = <TreasureChestItem>[];
    if (raw is List) {
      for (final e in raw) {
        if (e is Map<String, dynamic>) {
          try {
            itemList.add(TreasureChestItem.fromJson(e));
          } catch (_) {}
        }
      }
    }
    return TreasureChest(
      id: json['id'] as String,
      latitude: (json['latitude'] as num).toDouble(),
      longitude: (json['longitude'] as num).toDouble(),
      zoneId: json['zoneId'] as String,
      gold: (json['gold'] as num?)?.toInt(),
      openedByUser: json['openedByUser'] as bool?,
      unlockTier: (json['unlockTier'] as num?)?.toInt(),
      items: itemList,
    );
  }
}
