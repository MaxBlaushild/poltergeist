class UserLevel {
  final int level;
  final int experiencePointsOnLevel;
  final int experienceToNextLevel;

  const UserLevel({
    required this.level,
    required this.experiencePointsOnLevel,
    required this.experienceToNextLevel,
  });

  factory UserLevel.fromJson(Map<String, dynamic> json) {
    int _int(String key, [String? fallback]) {
      final value = json[key] ?? (fallback != null ? json[fallback] : null);
      if (value is num) return value.toInt();
      return int.tryParse(value?.toString() ?? '') ?? 0;
    }

    return UserLevel(
      level: _int('level', 'Level'),
      experiencePointsOnLevel: _int(
        'experiencePointsOnLevel',
        'experience_points_on_level',
      ),
      experienceToNextLevel: _int(
        'experienceToNextLevel',
        'experience_to_next_level',
      ),
    );
  }
}
