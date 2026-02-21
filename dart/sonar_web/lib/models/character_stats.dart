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
}
