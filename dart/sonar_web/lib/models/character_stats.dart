class CharacterStats {
  final int strength;
  final int dexterity;
  final int constitution;
  final int intelligence;
  final int wisdom;
  final int charisma;
  final int unspentPoints;
  final int level;

  const CharacterStats({
    required this.strength,
    required this.dexterity,
    required this.constitution,
    required this.intelligence,
    required this.wisdom,
    required this.charisma,
    required this.unspentPoints,
    required this.level,
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
