class CharacterProficiency {
  final String proficiency;
  final int level;

  const CharacterProficiency({
    required this.proficiency,
    required this.level,
  });

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
  final int strength;
  final int dexterity;
  final int constitution;
  final int intelligence;
  final int wisdom;
  final int charisma;
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
    this.equipmentBonuses = const {},
    required this.unspentPoints,
    required this.level,
    this.proficiencies = const [],
  });

  factory CharacterStats.fromJson(Map<String, dynamic> json) {
    int _int(String key, [String? fallback]) {
      final value = json[key] ?? (fallback != null ? json[fallback] : null);
      if (value is num) return value.toInt();
      return int.tryParse(value?.toString() ?? '') ?? 0;
    }

    return CharacterStats(
      strength: _int('strength', 'Strength'),
      dexterity: _int('dexterity', 'Dexterity'),
      constitution: _int('constitution', 'Constitution'),
      intelligence: _int('intelligence', 'Intelligence'),
      wisdom: _int('wisdom', 'Wisdom'),
      charisma: _int('charisma', 'Charisma'),
      equipmentBonuses: _parseBonusMap(json['equipmentBonuses']),
      unspentPoints: _int('unspentPoints', 'unspent_points'),
      level: _int('level', 'Level'),
      proficiencies: (json['proficiencies'] as List<dynamic>?)
              ?.map((entry) =>
                  CharacterProficiency.fromJson(entry as Map<String, dynamic>))
              .where((entry) => entry.proficiency.isNotEmpty)
              .toList() ??
          const [],
    );
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
