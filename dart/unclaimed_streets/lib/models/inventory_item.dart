import 'resource_type.dart';

class InventoryItem {
  final int id;
  final String name;
  final String imageUrl;
  final String flavorText;
  final String effectText;
  final String rarityTier;
  final String? resourceTypeId;
  final ResourceType? resourceType;
  final int? buyPrice;
  final int? unlockTier;
  final int? unlockLocksStrength;
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
  final int consumeRevivePartyMemberHealth;
  final int consumeReviveAllDownedPartyMembersHealth;
  final int consumeDealDamage;
  final int consumeDealDamageHits;
  final int consumeDealDamageAllEnemies;
  final int consumeDealDamageAllEnemiesHits;
  final bool consumeCreateBase;
  final List<InventoryConsumeStatus> consumeStatusesToAdd;
  final List<String> consumeStatusesToRemove;
  final List<String> consumeSpellIds;
  final List<String> consumeTeachRecipeIds;
  final List<InventoryRecipe> alchemyRecipes;
  final List<InventoryRecipe> workshopRecipes;

  const InventoryItem({
    required this.id,
    required this.name,
    required this.imageUrl,
    required this.flavorText,
    required this.effectText,
    this.rarityTier = '',
    this.resourceTypeId,
    this.resourceType,
    this.buyPrice,
    this.unlockTier,
    this.unlockLocksStrength,
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
    this.consumeRevivePartyMemberHealth = 0,
    this.consumeReviveAllDownedPartyMembersHealth = 0,
    this.consumeDealDamage = 0,
    this.consumeDealDamageHits = 0,
    this.consumeDealDamageAllEnemies = 0,
    this.consumeDealDamageAllEnemiesHits = 0,
    this.consumeCreateBase = false,
    this.consumeStatusesToAdd = const [],
    this.consumeStatusesToRemove = const [],
    this.consumeSpellIds = const [],
    this.consumeTeachRecipeIds = const [],
    this.alchemyRecipes = const [],
    this.workshopRecipes = const [],
  });

  factory InventoryItem.fromJson(Map<String, dynamic> json) {
    final rawResourceType = json['resourceType'];
    return InventoryItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      name: json['name'] as String? ?? '',
      imageUrl: json['imageUrl'] as String? ?? '',
      flavorText: json['flavorText'] as String? ?? '',
      effectText: json['effectText'] as String? ?? '',
      rarityTier: json['rarityTier'] as String? ?? '',
      resourceTypeId: json['resourceTypeId']?.toString(),
      resourceType: rawResourceType is Map<String, dynamic>
          ? ResourceType.fromJson(rawResourceType)
          : null,
      buyPrice: (json['buyPrice'] as num?)?.toInt(),
      unlockTier: (json['unlockTier'] as num?)?.toInt(),
      unlockLocksStrength: (json['unlockLocksStrength'] as num?)?.toInt(),
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
      consumeRevivePartyMemberHealth:
          (json['consumeRevivePartyMemberHealth'] as num?)?.toInt() ?? 0,
      consumeReviveAllDownedPartyMembersHealth:
          (json['consumeReviveAllDownedPartyMembersHealth'] as num?)?.toInt() ??
          0,
      consumeDealDamage: (json['consumeDealDamage'] as num?)?.toInt() ?? 0,
      consumeDealDamageHits:
          (json['consumeDealDamageHits'] as num?)?.toInt() ?? 0,
      consumeDealDamageAllEnemies:
          (json['consumeDealDamageAllEnemies'] as num?)?.toInt() ?? 0,
      consumeDealDamageAllEnemiesHits:
          (json['consumeDealDamageAllEnemiesHits'] as num?)?.toInt() ?? 0,
      consumeCreateBase: json['consumeCreateBase'] as bool? ?? false,
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
      consumeTeachRecipeIds:
          (json['consumeTeachRecipeIds'] as List<dynamic>?)
              ?.map((entry) => entry.toString().trim())
              .where((entry) => entry.isNotEmpty)
              .toList() ??
          const [],
      alchemyRecipes:
          (json['alchemyRecipes'] as List<dynamic>?)
              ?.map(
                (entry) =>
                    InventoryRecipe.fromJson(entry as Map<String, dynamic>),
              )
              .where((entry) => entry.ingredients.isNotEmpty)
              .toList() ??
          const [],
      workshopRecipes:
          (json['workshopRecipes'] as List<dynamic>?)
              ?.map(
                (entry) =>
                    InventoryRecipe.fromJson(entry as Map<String, dynamic>),
              )
              .where((entry) => entry.ingredients.isNotEmpty)
              .toList() ??
          const [],
    );
  }
}

class InventoryRecipe {
  final String id;
  final int tier;
  final bool isPublic;
  final List<InventoryRecipeIngredient> ingredients;

  const InventoryRecipe({
    required this.id,
    required this.tier,
    required this.isPublic,
    required this.ingredients,
  });

  factory InventoryRecipe.fromJson(Map<String, dynamic> json) {
    return InventoryRecipe(
      id: json['id']?.toString() ?? '',
      tier: (json['tier'] as num?)?.toInt() ?? 1,
      isPublic: json['isPublic'] == true,
      ingredients:
          (json['ingredients'] as List<dynamic>?)
              ?.map(
                (entry) => InventoryRecipeIngredient.fromJson(
                  entry as Map<String, dynamic>,
                ),
              )
              .where((entry) => entry.itemId > 0 && entry.quantity > 0)
              .toList() ??
          const [],
    );
  }
}

class InventoryRecipeIngredient {
  final int itemId;
  final int quantity;

  const InventoryRecipeIngredient({
    required this.itemId,
    required this.quantity,
  });

  factory InventoryRecipeIngredient.fromJson(Map<String, dynamic> json) {
    return InventoryRecipeIngredient(
      itemId: (json['itemId'] as num?)?.toInt() ?? 0,
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
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
