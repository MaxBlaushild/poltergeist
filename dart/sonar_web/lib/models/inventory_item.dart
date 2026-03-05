class InventoryItem {
  final int id;
  final String name;
  final String imageUrl;
  final String flavorText;
  final String effectText;
  final String rarityTier;
  final int? sellValue;
  final int? unlockTier;
  final int itemLevel;
  final String? equipSlot;
  final int strengthMod;
  final int dexterityMod;
  final int constitutionMod;
  final int intelligenceMod;
  final int wisdomMod;
  final int charismaMod;
  final String? handItemCategory;
  final String? handedness;
  final int? damageMin;
  final int? damageMax;
  final int? swipesPerAttack;
  final int? blockPercentage;
  final int? damageBlocked;
  final int? spellDamageBonusPercent;
  final int consumeHealthDelta;
  final int consumeManaDelta;
  final List<InventoryConsumeStatus> consumeStatusesToAdd;
  final List<String> consumeStatusesToRemove;
  final List<String> consumeSpellIds;

  const InventoryItem({
    required this.id,
    required this.name,
    required this.imageUrl,
    required this.flavorText,
    required this.effectText,
    this.rarityTier = '',
    this.sellValue,
    this.unlockTier,
    this.itemLevel = 1,
    this.equipSlot,
    this.strengthMod = 0,
    this.dexterityMod = 0,
    this.constitutionMod = 0,
    this.intelligenceMod = 0,
    this.wisdomMod = 0,
    this.charismaMod = 0,
    this.handItemCategory,
    this.handedness,
    this.damageMin,
    this.damageMax,
    this.swipesPerAttack,
    this.blockPercentage,
    this.damageBlocked,
    this.spellDamageBonusPercent,
    this.consumeHealthDelta = 0,
    this.consumeManaDelta = 0,
    this.consumeStatusesToAdd = const [],
    this.consumeStatusesToRemove = const [],
    this.consumeSpellIds = const [],
  });

  factory InventoryItem.fromJson(Map<String, dynamic> json) {
    return InventoryItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: json['name'] as String? ?? '',
      imageUrl: json['imageUrl'] as String? ?? '',
      flavorText: json['flavorText'] as String? ?? '',
      effectText: json['effectText'] as String? ?? '',
      rarityTier: json['rarityTier'] as String? ?? '',
      sellValue: (json['sellValue'] as num?)?.toInt(),
      unlockTier: (json['unlockTier'] as num?)?.toInt(),
      itemLevel: (json['itemLevel'] as num?)?.toInt() ?? 1,
      equipSlot: json['equipSlot'] as String?,
      strengthMod: (json['strengthMod'] as num?)?.toInt() ?? 0,
      dexterityMod: (json['dexterityMod'] as num?)?.toInt() ?? 0,
      constitutionMod: (json['constitutionMod'] as num?)?.toInt() ?? 0,
      intelligenceMod: (json['intelligenceMod'] as num?)?.toInt() ?? 0,
      wisdomMod: (json['wisdomMod'] as num?)?.toInt() ?? 0,
      charismaMod: (json['charismaMod'] as num?)?.toInt() ?? 0,
      handItemCategory: json['handItemCategory'] as String?,
      handedness: json['handedness'] as String?,
      damageMin: (json['damageMin'] as num?)?.toInt(),
      damageMax: (json['damageMax'] as num?)?.toInt(),
      swipesPerAttack: (json['swipesPerAttack'] as num?)?.toInt(),
      blockPercentage: (json['blockPercentage'] as num?)?.toInt(),
      damageBlocked: (json['damageBlocked'] as num?)?.toInt(),
      spellDamageBonusPercent: (json['spellDamageBonusPercent'] as num?)
          ?.toInt(),
      consumeHealthDelta: (json['consumeHealthDelta'] as num?)?.toInt() ?? 0,
      consumeManaDelta: (json['consumeManaDelta'] as num?)?.toInt() ?? 0,
      consumeStatusesToAdd:
          (json['consumeStatusesToAdd'] as List<dynamic>?)
              ?.map(
                (entry) => InventoryConsumeStatus.fromJson(
                  entry as Map<String, dynamic>,
                ),
              )
              .where((entry) => entry.name.isNotEmpty)
              .toList() ??
          const [],
      consumeStatusesToRemove:
          (json['consumeStatusesToRemove'] as List<dynamic>?)
              ?.map((entry) => entry.toString().trim())
              .where((entry) => entry.isNotEmpty)
              .toList() ??
          const [],
      consumeSpellIds:
          (json['consumeSpellIds'] as List<dynamic>?)
              ?.map((entry) => entry.toString().trim())
              .where((entry) => entry.isNotEmpty)
              .toList() ??
          const [],
    );
  }
}

class InventoryConsumeStatus {
  final String name;
  final String description;
  final String effect;
  final bool positive;
  final int durationSeconds;
  final int strengthMod;
  final int dexterityMod;
  final int constitutionMod;
  final int intelligenceMod;
  final int wisdomMod;
  final int charismaMod;

  const InventoryConsumeStatus({
    required this.name,
    required this.description,
    required this.effect,
    required this.positive,
    required this.durationSeconds,
    required this.strengthMod,
    required this.dexterityMod,
    required this.constitutionMod,
    required this.intelligenceMod,
    required this.wisdomMod,
    required this.charismaMod,
  });

  factory InventoryConsumeStatus.fromJson(Map<String, dynamic> json) {
    bool parseBool(dynamic raw, {bool fallback = true}) {
      if (raw is bool) return raw;
      if (raw is String) {
        final normalized = raw.trim().toLowerCase();
        if (normalized == 'true') return true;
        if (normalized == 'false') return false;
      }
      return fallback;
    }

    int intValue(String key) {
      final raw = json[key];
      if (raw is num) return raw.toInt();
      return int.tryParse(raw?.toString() ?? '') ?? 0;
    }

    return InventoryConsumeStatus(
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      effect: json['effect']?.toString() ?? '',
      positive: parseBool(json['positive'], fallback: true),
      durationSeconds: intValue('durationSeconds'),
      strengthMod: intValue('strengthMod'),
      dexterityMod: intValue('dexterityMod'),
      constitutionMod: intValue('constitutionMod'),
      intelligenceMod: intValue('intelligenceMod'),
      wisdomMod: intValue('wisdomMod'),
      charismaMod: intValue('charismaMod'),
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
