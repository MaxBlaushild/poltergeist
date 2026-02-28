class CharacterProficiency {
  final String proficiency;
  final int level;

  const CharacterProficiency({required this.proficiency, required this.level});

  factory CharacterProficiency.fromJson(Map<String, dynamic> json) {
    final rawLevel = json['level'];
    final levelValue = rawLevel is num
        ? rawLevel.toInt()
        : int.tryParse(rawLevel?.toString() ?? '') ?? 0;
    return CharacterProficiency(
      proficiency: json['proficiency']?.toString() ?? '',
      level: levelValue,
    );
  }
}

class CharacterStatus {
  final String id;
  final String name;
  final String description;
  final String effect;
  final bool positive;
  final String effectType;
  final DateTime? startedAt;
  final DateTime? expiresAt;

  const CharacterStatus({
    required this.id,
    required this.name,
    required this.description,
    required this.effect,
    required this.positive,
    required this.effectType,
    this.startedAt,
    this.expiresAt,
  });

  factory CharacterStatus.fromJson(Map<String, dynamic> json) {
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

    return CharacterStatus(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      effect: json['effect']?.toString() ?? '',
      positive: parseBool(json['positive'], fallback: true),
      effectType: json['effectType']?.toString() ?? '',
      startedAt: parseDate(json['startedAt']),
      expiresAt: parseDate(json['expiresAt']),
    );
  }
}

class CharacterStats {
  static const int healthPerConstitutionPoint = 10;
  static const int manaPerMentalStatPoint = 5;

  final int strength;
  final int dexterity;
  final int constitution;
  final int intelligence;
  final int wisdom;
  final int charisma;
  final int health;
  final int mana;
  final Map<String, int> equipmentBonuses;
  final Map<String, int> statusBonuses;
  final int unspentPoints;
  final int level;
  final List<CharacterProficiency> proficiencies;
  final List<CharacterStatus> statuses;

  const CharacterStats({
    required this.strength,
    required this.dexterity,
    required this.constitution,
    required this.intelligence,
    required this.wisdom,
    required this.charisma,
    required this.health,
    required this.mana,
    this.equipmentBonuses = const {},
    this.statusBonuses = const {},
    required this.unspentPoints,
    required this.level,
    this.proficiencies = const [],
    this.statuses = const [],
  });

  factory CharacterStats.fromJson(Map<String, dynamic> json) {
    int intValue(String key, [String? fallback]) {
      final value = json[key] ?? (fallback != null ? json[fallback] : null);
      if (value is num) return value.toInt();
      return int.tryParse(value?.toString() ?? '') ?? 0;
    }

    final strength = intValue('strength', 'Strength');
    final dexterity = intValue('dexterity', 'Dexterity');
    final constitution = intValue('constitution', 'Constitution');
    final intelligence = intValue('intelligence', 'Intelligence');
    final wisdom = intValue('wisdom', 'Wisdom');
    final charisma = intValue('charisma', 'Charisma');
    final equipmentBonuses = _parseBonusMap(json['equipmentBonuses']);
    final statusBonuses = _parseBonusMap(json['statusBonuses']);
    final combinedBonuses = _mergeBonusMaps(equipmentBonuses, statusBonuses);
    final effectiveConstitution =
        constitution + (combinedBonuses['constitution'] ?? 0);
    final effectiveIntelligence =
        intelligence + (combinedBonuses['intelligence'] ?? 0);
    final effectiveWisdom = wisdom + (combinedBonuses['wisdom'] ?? 0);
    final health = json['health'] is num
        ? (json['health'] as num).toInt()
        : deriveHealthFromConstitution(effectiveConstitution);
    final mana = json['mana'] is num
        ? (json['mana'] as num).toInt()
        : deriveManaFromMentalStats(effectiveIntelligence, effectiveWisdom);

    return CharacterStats(
      strength: strength,
      dexterity: dexterity,
      constitution: constitution,
      intelligence: intelligence,
      wisdom: wisdom,
      charisma: charisma,
      health: health,
      mana: mana,
      equipmentBonuses: equipmentBonuses,
      statusBonuses: statusBonuses,
      unspentPoints: intValue('unspentPoints', 'unspent_points'),
      level: intValue('level', 'Level'),
      proficiencies:
          (json['proficiencies'] as List<dynamic>?)
              ?.map(
                (entry) => CharacterProficiency.fromJson(
                  entry as Map<String, dynamic>,
                ),
              )
              .where((entry) => entry.proficiency.isNotEmpty)
              .toList() ??
          const [],
      statuses:
          (json['statuses'] as List<dynamic>?)
              ?.map(
                (entry) =>
                    CharacterStatus.fromJson(entry as Map<String, dynamic>),
              )
              .where((entry) => entry.name.isNotEmpty)
              .toList() ??
          const [],
    );
  }

  static int deriveHealthFromConstitution(int constitution) {
    final normalized = constitution < 1 ? 1 : constitution;
    return normalized * healthPerConstitutionPoint;
  }

  static int deriveManaFromMentalStats(int intelligence, int wisdom) {
    final mental = intelligence + wisdom;
    final normalized = mental < 1 ? 1 : mental;
    return normalized * manaPerMentalStatPoint;
  }

  Map<String, int> toMap() => {
    'strength': strength,
    'dexterity': dexterity,
    'constitution': constitution,
    'intelligence': intelligence,
    'wisdom': wisdom,
    'charisma': charisma,
  };

  Map<String, int> equipmentBonusMap() => {
    'strength': equipmentBonuses['strength'] ?? 0,
    'dexterity': equipmentBonuses['dexterity'] ?? 0,
    'constitution': equipmentBonuses['constitution'] ?? 0,
    'intelligence': equipmentBonuses['intelligence'] ?? 0,
    'wisdom': equipmentBonuses['wisdom'] ?? 0,
    'charisma': equipmentBonuses['charisma'] ?? 0,
  };

  Map<String, int> statusBonusMap() => {
    'strength': statusBonuses['strength'] ?? 0,
    'dexterity': statusBonuses['dexterity'] ?? 0,
    'constitution': statusBonuses['constitution'] ?? 0,
    'intelligence': statusBonuses['intelligence'] ?? 0,
    'wisdom': statusBonuses['wisdom'] ?? 0,
    'charisma': statusBonuses['charisma'] ?? 0,
  };

  Map<String, int> bonusMap() =>
      _mergeBonusMaps(equipmentBonusMap(), statusBonusMap());

  Map<String, int> effectiveMap() {
    final base = toMap();
    final bonus = bonusMap();
    return {
      for (final entry in base.entries)
        entry.key: entry.value + (bonus[entry.key] ?? 0),
    };
  }

  static Map<String, int> _parseBonusMap(dynamic value) {
    if (value is Map<String, dynamic>) {
      return value.map((key, v) {
        if (v is num) return MapEntry(key, v.toInt());
        return MapEntry(key, int.tryParse(v?.toString() ?? '') ?? 0);
      });
    }
    return const {};
  }

  static Map<String, int> _mergeBonusMaps(
    Map<String, int> first,
    Map<String, int> second,
  ) {
    return {
      'strength': (first['strength'] ?? 0) + (second['strength'] ?? 0),
      'dexterity': (first['dexterity'] ?? 0) + (second['dexterity'] ?? 0),
      'constitution':
          (first['constitution'] ?? 0) + (second['constitution'] ?? 0),
      'intelligence':
          (first['intelligence'] ?? 0) + (second['intelligence'] ?? 0),
      'wisdom': (first['wisdom'] ?? 0) + (second['wisdom'] ?? 0),
      'charisma': (first['charisma'] ?? 0) + (second['charisma'] ?? 0),
    };
  }
}
