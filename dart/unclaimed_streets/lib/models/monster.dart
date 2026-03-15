import 'spell.dart';

class MonsterStatus {
  final String id;
  final String name;
  final String description;
  final String effect;
  final bool positive;
  final String effectType;
  final int damagePerTick;
  final int healthPerTick;
  final DateTime? startedAt;
  final DateTime? expiresAt;
  final DateTime? lastTickAt;

  const MonsterStatus({
    required this.id,
    required this.name,
    required this.description,
    required this.effect,
    required this.positive,
    required this.effectType,
    this.damagePerTick = 0,
    this.healthPerTick = 0,
    this.startedAt,
    this.expiresAt,
    this.lastTickAt,
  });

  factory MonsterStatus.fromJson(Map<String, dynamic> json) {
    DateTime? parseDate(dynamic raw) {
      if (raw is String && raw.trim().isNotEmpty) {
        return DateTime.tryParse(raw.trim());
      }
      return null;
    }

    bool parseBool(dynamic raw, {bool fallback = true}) {
      if (raw is bool) return raw;
      if (raw is String) {
        final normalized = raw.trim().toLowerCase();
        if (normalized == 'true') return true;
        if (normalized == 'false') return false;
      }
      return fallback;
    }

    int parseInt(dynamic raw) {
      if (raw is num) return raw.toInt();
      return int.tryParse(raw?.toString() ?? '') ?? 0;
    }

    return MonsterStatus(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      effect: json['effect']?.toString() ?? '',
      positive: parseBool(json['positive'], fallback: true),
      effectType: json['effectType']?.toString() ?? '',
      damagePerTick: parseInt(json['damagePerTick']),
      healthPerTick: parseInt(json['healthPerTick']),
      startedAt: parseDate(json['startedAt']),
      expiresAt: parseDate(json['expiresAt']),
      lastTickAt: parseDate(json['lastTickAt']),
    );
  }
}

class MonsterTemplate {
  final String id;
  final String name;
  final String description;
  final String imageUrl;
  final String thumbnailUrl;
  final int baseStrength;
  final int baseDexterity;
  final int baseConstitution;
  final int baseIntelligence;
  final int baseWisdom;
  final int baseCharisma;
  final List<Spell> spells;
  final List<Spell> techniques;

  const MonsterTemplate({
    required this.id,
    required this.name,
    this.description = '',
    this.imageUrl = '',
    this.thumbnailUrl = '',
    this.baseStrength = 10,
    this.baseDexterity = 10,
    this.baseConstitution = 10,
    this.baseIntelligence = 10,
    this.baseWisdom = 10,
    this.baseCharisma = 10,
    this.spells = const [],
    this.techniques = const [],
  });

  factory MonsterTemplate.fromJson(Map<String, dynamic> json) {
    final spells = <Spell>[];
    final rawSpells = json['spells'];
    if (rawSpells is List) {
      for (final spell in rawSpells) {
        if (spell is Map<String, dynamic>) {
          spells.add(Spell.fromJson(spell));
        } else if (spell is Map) {
          spells.add(Spell.fromJson(Map<String, dynamic>.from(spell)));
        }
      }
    }

    final spellAbilities = spells
        .where((spell) => spell.abilityType.toLowerCase() != 'technique')
        .toList(growable: false);
    final techniques = spells
        .where((spell) => spell.abilityType.toLowerCase() == 'technique')
        .toList(growable: false);

    return MonsterTemplate(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      imageUrl: json['imageUrl']?.toString() ?? '',
      thumbnailUrl: json['thumbnailUrl']?.toString() ?? '',
      baseStrength: (json['baseStrength'] as num?)?.toInt() ?? 10,
      baseDexterity: (json['baseDexterity'] as num?)?.toInt() ?? 10,
      baseConstitution: (json['baseConstitution'] as num?)?.toInt() ?? 10,
      baseIntelligence: (json['baseIntelligence'] as num?)?.toInt() ?? 10,
      baseWisdom: (json['baseWisdom'] as num?)?.toInt() ?? 10,
      baseCharisma: (json['baseCharisma'] as num?)?.toInt() ?? 10,
      spells: spellAbilities,
      techniques: techniques,
    );
  }
}

class MonsterItemReward {
  final String id;
  final int inventoryItemId;
  final int quantity;
  final String inventoryItemName;
  final String inventoryItemImageUrl;

  const MonsterItemReward({
    required this.id,
    required this.inventoryItemId,
    required this.quantity,
    this.inventoryItemName = '',
    this.inventoryItemImageUrl = '',
  });

  factory MonsterItemReward.fromJson(Map<String, dynamic> json) {
    final rawInventoryItem = json['inventoryItem'];
    String inventoryItemName = '';
    String inventoryItemImageUrl = '';
    if (rawInventoryItem is Map<String, dynamic>) {
      inventoryItemName = rawInventoryItem['name']?.toString() ?? '';
      inventoryItemImageUrl = rawInventoryItem['imageUrl']?.toString() ?? '';
    } else if (rawInventoryItem is Map) {
      final cast = Map<String, dynamic>.from(rawInventoryItem);
      inventoryItemName = cast['name']?.toString() ?? '';
      inventoryItemImageUrl = cast['imageUrl']?.toString() ?? '';
    }

    return MonsterItemReward(
      id: json['id']?.toString() ?? '',
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt() ?? 0,
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
      inventoryItemName: inventoryItemName,
      inventoryItemImageUrl: inventoryItemImageUrl,
    );
  }
}

class Monster {
  final String id;
  final String name;
  final String description;
  final String imageUrl;
  final String thumbnailUrl;
  final bool scaleWithUserLevel;
  final String zoneId;
  final double latitude;
  final double longitude;
  final String templateId;
  final MonsterTemplate? template;
  final int? weaponInventoryItemId;
  final String weaponInventoryItemName;
  final int level;
  final int strength;
  final int dexterity;
  final int constitution;
  final int intelligence;
  final int wisdom;
  final int charisma;
  final int health;
  final int maxHealth;
  final int mana;
  final int maxMana;
  final int attackDamageMin;
  final int attackDamageMax;
  final int attackSwipesPerAttack;
  final int rewardExperience;
  final int rewardGold;
  final String imageGenerationStatus;
  final String? imageGenerationError;
  final List<Spell> spells;
  final List<Spell> techniques;
  final List<MonsterItemReward> itemRewards;
  final List<MonsterStatus> statuses;

  const Monster({
    required this.id,
    required this.name,
    this.description = '',
    this.imageUrl = '',
    this.thumbnailUrl = '',
    this.scaleWithUserLevel = false,
    required this.zoneId,
    required this.latitude,
    required this.longitude,
    this.templateId = '',
    this.template,
    this.weaponInventoryItemId,
    this.weaponInventoryItemName = '',
    this.level = 1,
    this.strength = 10,
    this.dexterity = 10,
    this.constitution = 10,
    this.intelligence = 10,
    this.wisdom = 10,
    this.charisma = 10,
    this.health = 1,
    this.maxHealth = 1,
    this.mana = 1,
    this.maxMana = 1,
    this.attackDamageMin = 1,
    this.attackDamageMax = 1,
    this.attackSwipesPerAttack = 1,
    this.rewardExperience = 0,
    this.rewardGold = 0,
    this.imageGenerationStatus = 'none',
    this.imageGenerationError,
    this.spells = const [],
    this.techniques = const [],
    this.itemRewards = const [],
    this.statuses = const [],
  });

  factory Monster.fromJson(Map<String, dynamic> json) {
    MonsterTemplate? template;
    final rawTemplate = json['template'];
    if (rawTemplate is Map<String, dynamic>) {
      template = MonsterTemplate.fromJson(rawTemplate);
    } else if (rawTemplate is Map) {
      template = MonsterTemplate.fromJson(
        Map<String, dynamic>.from(rawTemplate),
      );
    }

    int? weaponInventoryItemId;
    final rawWeaponInventoryItemId = json['weaponInventoryItemId'];
    if (rawWeaponInventoryItemId is num) {
      weaponInventoryItemId = rawWeaponInventoryItemId.toInt();
    }

    String weaponInventoryItemName = '';
    final rawWeaponInventoryItem = json['weaponInventoryItem'];
    if (rawWeaponInventoryItem is Map<String, dynamic>) {
      weaponInventoryItemName =
          rawWeaponInventoryItem['name']?.toString() ?? '';
    } else if (rawWeaponInventoryItem is Map) {
      final cast = Map<String, dynamic>.from(rawWeaponInventoryItem);
      weaponInventoryItemName = cast['name']?.toString() ?? '';
    }

    final allAbilities = <Spell>[];
    final rawSpells = json['spells'];
    if (rawSpells is List) {
      for (final spell in rawSpells) {
        if (spell is Map<String, dynamic>) {
          allAbilities.add(Spell.fromJson(spell));
        } else if (spell is Map) {
          allAbilities.add(Spell.fromJson(Map<String, dynamic>.from(spell)));
        }
      }
    } else if (template != null) {
      allAbilities.addAll(template.spells);
      allAbilities.addAll(template.techniques);
    }
    final spells = allAbilities
        .where((spell) => spell.abilityType.toLowerCase() != 'technique')
        .toList(growable: false);
    final techniques = allAbilities
        .where((spell) => spell.abilityType.toLowerCase() == 'technique')
        .toList(growable: false);

    final itemRewards = <MonsterItemReward>[];
    final rawItemRewards = json['itemRewards'];
    if (rawItemRewards is List) {
      for (final reward in rawItemRewards) {
        if (reward is Map<String, dynamic>) {
          itemRewards.add(MonsterItemReward.fromJson(reward));
        } else if (reward is Map) {
          itemRewards.add(
            MonsterItemReward.fromJson(Map<String, dynamic>.from(reward)),
          );
        }
      }
    }

    final statuses = <MonsterStatus>[];
    final rawStatuses = json['statuses'];
    if (rawStatuses is List) {
      for (final status in rawStatuses) {
        if (status is Map<String, dynamic>) {
          statuses.add(MonsterStatus.fromJson(status));
        } else if (status is Map) {
          statuses.add(
            MonsterStatus.fromJson(Map<String, dynamic>.from(status)),
          );
        }
      }
    }

    return Monster(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      imageUrl: json['imageUrl']?.toString() ?? '',
      thumbnailUrl: json['thumbnailUrl']?.toString() ?? '',
      scaleWithUserLevel: json['scaleWithUserLevel'] as bool? ?? false,
      zoneId: json['zoneId']?.toString() ?? '',
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0,
      templateId: json['templateId']?.toString() ?? template?.id ?? '',
      template: template,
      weaponInventoryItemId: weaponInventoryItemId,
      weaponInventoryItemName: weaponInventoryItemName,
      level: (json['level'] as num?)?.toInt() ?? 1,
      strength: (json['strength'] as num?)?.toInt() ?? 10,
      dexterity: (json['dexterity'] as num?)?.toInt() ?? 10,
      constitution: (json['constitution'] as num?)?.toInt() ?? 10,
      intelligence: (json['intelligence'] as num?)?.toInt() ?? 10,
      wisdom: (json['wisdom'] as num?)?.toInt() ?? 10,
      charisma: (json['charisma'] as num?)?.toInt() ?? 10,
      health: (json['health'] as num?)?.toInt() ?? 1,
      maxHealth: (json['maxHealth'] as num?)?.toInt() ?? 1,
      mana: (json['mana'] as num?)?.toInt() ?? 1,
      maxMana: (json['maxMana'] as num?)?.toInt() ?? 1,
      attackDamageMin: (json['attackDamageMin'] as num?)?.toInt() ?? 1,
      attackDamageMax: (json['attackDamageMax'] as num?)?.toInt() ?? 1,
      attackSwipesPerAttack:
          (json['attackSwipesPerAttack'] as num?)?.toInt() ?? 1,
      rewardExperience: (json['rewardExperience'] as num?)?.toInt() ?? 0,
      rewardGold: (json['rewardGold'] as num?)?.toInt() ?? 0,
      imageGenerationStatus:
          json['imageGenerationStatus']?.toString() ?? 'none',
      imageGenerationError: json['imageGenerationError']?.toString(),
      spells: spells,
      techniques: techniques,
      itemRewards: itemRewards,
      statuses: statuses,
    );
  }
}

class MonsterEncounterMember {
  final int slot;
  final Monster monster;

  const MonsterEncounterMember({required this.slot, required this.monster});

  factory MonsterEncounterMember.fromJson(Map<String, dynamic> json) {
    final rawMonster = json['monster'];
    Monster monster;
    if (rawMonster is Map<String, dynamic>) {
      monster = Monster.fromJson(rawMonster);
    } else if (rawMonster is Map) {
      monster = Monster.fromJson(Map<String, dynamic>.from(rawMonster));
    } else {
      monster = const Monster(
        id: '',
        name: '',
        zoneId: '',
        latitude: 0,
        longitude: 0,
      );
    }
    return MonsterEncounterMember(
      slot: (json['slot'] as num?)?.toInt() ?? 0,
      monster: monster,
    );
  }
}

class MonsterEncounter {
  final String id;
  final String name;
  final String description;
  final String imageUrl;
  final String thumbnailUrl;
  final String encounterType;
  final String zoneId;
  final String? pointOfInterestId;
  final double latitude;
  final double longitude;
  final int monsterCount;
  final List<MonsterEncounterMember> members;
  final List<Monster> monsters;

  const MonsterEncounter({
    required this.id,
    required this.name,
    this.description = '',
    this.imageUrl = '',
    this.thumbnailUrl = '',
    this.encounterType = 'monster',
    required this.zoneId,
    this.pointOfInterestId,
    required this.latitude,
    required this.longitude,
    this.monsterCount = 0,
    this.members = const [],
    this.monsters = const [],
  });

  factory MonsterEncounter.fromJson(Map<String, dynamic> json) {
    final members = <MonsterEncounterMember>[];
    final rawMembers = json['members'];
    if (rawMembers is List) {
      for (final member in rawMembers) {
        if (member is Map<String, dynamic>) {
          members.add(MonsterEncounterMember.fromJson(member));
        } else if (member is Map) {
          members.add(
            MonsterEncounterMember.fromJson(Map<String, dynamic>.from(member)),
          );
        }
      }
    }

    final monsters = <Monster>[];
    final rawMonsters = json['monsters'];
    if (rawMonsters is List) {
      for (final monster in rawMonsters) {
        if (monster is Map<String, dynamic>) {
          monsters.add(Monster.fromJson(monster));
        } else if (monster is Map) {
          monsters.add(Monster.fromJson(Map<String, dynamic>.from(monster)));
        }
      }
    } else {
      for (final member in members) {
        monsters.add(member.monster);
      }
    }

    return MonsterEncounter(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      imageUrl: json['imageUrl']?.toString() ?? '',
      thumbnailUrl: json['thumbnailUrl']?.toString() ?? '',
      encounterType:
          (json['encounterType']?.toString().trim().isNotEmpty ?? false)
          ? json['encounterType']!.toString().trim().toLowerCase()
          : 'monster',
      zoneId: json['zoneId']?.toString() ?? '',
      pointOfInterestId: json['pointOfInterestId']?.toString(),
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0,
      monsterCount:
          (json['monsterCount'] as num?)?.toInt() ??
          (monsters.isNotEmpty ? monsters.length : members.length),
      members: members,
      monsters: monsters,
    );
  }

  int get totalRewardExperience =>
      monsters.fold<int>(0, (sum, monster) => sum + monster.rewardExperience);

  int get totalRewardGold =>
      monsters.fold<int>(0, (sum, monster) => sum + monster.rewardGold);

  bool get isBossEncounter => encounterType == 'boss';

  bool get isRaidEncounter => encounterType == 'raid';

  String get encounterTypeLabel {
    switch (encounterType) {
      case 'boss':
        return 'Boss Encounter';
      case 'raid':
        return 'Raid Encounter';
      default:
        return 'Monster Encounter';
    }
  }
}
