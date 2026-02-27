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
  final int unspentPoints;
  final int level;
  final List<CharacterProficiency> proficiencies;

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
    required this.unspentPoints,
    required this.level,
    this.proficiencies = const [],
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
    final effectiveConstitution =
        constitution + (equipmentBonuses['constitution'] ?? 0);
    final effectiveIntelligence =
        intelligence + (equipmentBonuses['intelligence'] ?? 0);
    final effectiveWisdom = wisdom + (equipmentBonuses['wisdom'] ?? 0);
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

  Map<String, int> bonusMap() => {
    'strength': equipmentBonuses['strength'] ?? 0,
    'dexterity': equipmentBonuses['dexterity'] ?? 0,
    'constitution': equipmentBonuses['constitution'] ?? 0,
    'intelligence': equipmentBonuses['intelligence'] ?? 0,
    'wisdom': equipmentBonuses['wisdom'] ?? 0,
    'charisma': equipmentBonuses['charisma'] ?? 0,
  };

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
}
