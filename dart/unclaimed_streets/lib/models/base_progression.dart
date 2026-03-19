import 'base.dart';

class BaseResourceBalanceData {
  const BaseResourceBalanceData({
    required this.resourceKey,
    required this.amount,
  });

  final String resourceKey;
  final int amount;

  factory BaseResourceBalanceData.fromJson(Map<String, dynamic> json) {
    return BaseResourceBalanceData(
      resourceKey: json['resourceKey']?.toString() ?? '',
      amount: (json['amount'] as num?)?.toInt() ?? 0,
    );
  }
}

class BaseStructureCostData {
  const BaseStructureCostData({
    required this.level,
    required this.resourceKey,
    required this.amount,
  });

  final int level;
  final String resourceKey;
  final int amount;

  factory BaseStructureCostData.fromJson(Map<String, dynamic> json) {
    return BaseStructureCostData(
      level: (json['level'] as num?)?.toInt() ?? 0,
      resourceKey: json['resourceKey']?.toString() ?? '',
      amount: (json['amount'] as num?)?.toInt() ?? 0,
    );
  }
}

class BaseStructureLevelVisualData {
  const BaseStructureLevelVisualData({
    required this.level,
    required this.imageUrl,
    required this.thumbnailUrl,
    required this.imageGenerationStatus,
    required this.imageGenerationError,
    required this.topDownImageUrl,
    required this.topDownThumbnailUrl,
    required this.topDownImageGenerationStatus,
    required this.topDownImageGenerationError,
  });

  final int level;
  final String imageUrl;
  final String thumbnailUrl;
  final String imageGenerationStatus;
  final String? imageGenerationError;
  final String topDownImageUrl;
  final String topDownThumbnailUrl;
  final String topDownImageGenerationStatus;
  final String? topDownImageGenerationError;

  factory BaseStructureLevelVisualData.fromJson(Map<String, dynamic> json) {
    return BaseStructureLevelVisualData(
      level: (json['level'] as num?)?.toInt() ?? 0,
      imageUrl: json['imageUrl']?.toString() ?? '',
      thumbnailUrl: json['thumbnailUrl']?.toString() ?? '',
      imageGenerationStatus: json['imageGenerationStatus']?.toString() ?? '',
      imageGenerationError: json['imageGenerationError']?.toString(),
      topDownImageUrl: json['topDownImageUrl']?.toString() ?? '',
      topDownThumbnailUrl: json['topDownThumbnailUrl']?.toString() ?? '',
      topDownImageGenerationStatus:
          json['topDownImageGenerationStatus']?.toString() ?? '',
      topDownImageGenerationError: json['topDownImageGenerationError']
          ?.toString(),
    );
  }
}

class BaseStructureDefinitionData {
  const BaseStructureDefinitionData({
    required this.key,
    required this.name,
    required this.description,
    required this.category,
    required this.maxLevel,
    required this.sortOrder,
    required this.effectType,
    required this.prereqConfig,
    required this.levelCosts,
    required this.levelVisuals,
  });

  final String key;
  final String name;
  final String description;
  final String category;
  final int maxLevel;
  final int sortOrder;
  final String effectType;
  final Map<String, dynamic> prereqConfig;
  final List<BaseStructureCostData> levelCosts;
  final List<BaseStructureLevelVisualData> levelVisuals;

  factory BaseStructureDefinitionData.fromJson(Map<String, dynamic> json) {
    final rawCosts = json['levelCosts'];
    final costs = rawCosts is List
        ? rawCosts
              .whereType<Map>()
              .map(
                (e) => BaseStructureCostData.fromJson(
                  Map<String, dynamic>.from(e),
                ),
              )
              .toList()
        : const <BaseStructureCostData>[];
    final rawVisuals = json['levelVisuals'];
    final rawPrereq = json['prereqConfig'];
    return BaseStructureDefinitionData(
      key: json['key']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      category: json['category']?.toString() ?? '',
      maxLevel: (json['maxLevel'] as num?)?.toInt() ?? 1,
      sortOrder: (json['sortOrder'] as num?)?.toInt() ?? 0,
      effectType: json['effectType']?.toString() ?? '',
      prereqConfig: rawPrereq is Map<String, dynamic>
          ? rawPrereq
          : rawPrereq is Map
          ? Map<String, dynamic>.from(rawPrereq)
          : const <String, dynamic>{},
      levelCosts: costs,
      levelVisuals: rawVisuals is List
          ? rawVisuals
                .whereType<Map>()
                .map(
                  (e) => BaseStructureLevelVisualData.fromJson(
                    Map<String, dynamic>.from(e),
                  ),
                )
                .toList()
          : const <BaseStructureLevelVisualData>[],
    );
  }
}

class UserBaseStructureData {
  const UserBaseStructureData({
    required this.structureKey,
    required this.level,
    required this.gridX,
    required this.gridY,
  });

  final String structureKey;
  final int level;
  final int gridX;
  final int gridY;

  factory UserBaseStructureData.fromJson(Map<String, dynamic> json) {
    return UserBaseStructureData(
      structureKey: json['structureKey']?.toString() ?? '',
      level: (json['level'] as num?)?.toInt() ?? 0,
      gridX: (json['gridX'] as num?)?.toInt() ?? 0,
      gridY: (json['gridY'] as num?)?.toInt() ?? 0,
    );
  }
}

class BaseDailyEffectData {
  const BaseDailyEffectData({required this.stateKey, required this.state});

  final String stateKey;
  final Map<String, dynamic> state;

  factory BaseDailyEffectData.fromJson(Map<String, dynamic> json) {
    final raw = json['state'];
    return BaseDailyEffectData(
      stateKey: json['stateKey']?.toString() ?? '',
      state: raw is Map<String, dynamic>
          ? raw
          : raw is Map
          ? Map<String, dynamic>.from(raw)
          : const <String, dynamic>{},
    );
  }
}

class BaseProgressionSnapshot {
  const BaseProgressionSnapshot({
    required this.base,
    required this.resources,
    required this.structures,
    required this.activeDailyEffects,
    required this.grassTileUrls,
    required this.canManage,
  });

  final BasePin? base;
  final List<BaseResourceBalanceData> resources;
  final List<UserBaseStructureData> structures;
  final List<BaseDailyEffectData> activeDailyEffects;
  final Map<String, String> grassTileUrls;
  final bool canManage;

  factory BaseProgressionSnapshot.fromJson(Map<String, dynamic> json) {
    final rawBase = json['base'];
    final rawResources = json['resources'];
    final rawStructures = json['structures'];
    final rawEffects = json['activeDailyEffects'];
    final rawGrassTileUrls = json['grassTileUrls'];
    return BaseProgressionSnapshot(
      base: rawBase is Map<String, dynamic>
          ? BasePin.fromJson(rawBase)
          : rawBase is Map
          ? BasePin.fromJson(Map<String, dynamic>.from(rawBase))
          : null,
      resources: rawResources is List
          ? rawResources
                .whereType<Map>()
                .map(
                  (e) => BaseResourceBalanceData.fromJson(
                    Map<String, dynamic>.from(e),
                  ),
                )
                .toList()
          : const <BaseResourceBalanceData>[],
      structures: rawStructures is List
          ? rawStructures
                .whereType<Map>()
                .map(
                  (e) => UserBaseStructureData.fromJson(
                    Map<String, dynamic>.from(e),
                  ),
                )
                .toList()
          : const <UserBaseStructureData>[],
      activeDailyEffects: rawEffects is List
          ? rawEffects
                .whereType<Map>()
                .map(
                  (e) => BaseDailyEffectData.fromJson(
                    Map<String, dynamic>.from(e),
                  ),
                )
                .toList()
          : const <BaseDailyEffectData>[],
      grassTileUrls: rawGrassTileUrls is Map<String, dynamic>
          ? rawGrassTileUrls.map(
              (key, value) => MapEntry(key, value?.toString() ?? ''),
            )
          : rawGrassTileUrls is Map
          ? Map<String, dynamic>.from(
              rawGrassTileUrls,
            ).map((key, value) => MapEntry(key, value?.toString() ?? ''))
          : const <String, String>{},
      canManage: json['canManage'] == true,
    );
  }
}
