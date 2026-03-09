enum UserZoneReputationName {
  neutral,
  friendly,
  honored,
  revered,
  exalted,
  legendary;

  static UserZoneReputationName? fromString(String? value) {
    if (value == null) return null;
    return UserZoneReputationName.values.firstWhere(
      (e) => e.name == value,
      orElse: () => UserZoneReputationName.neutral,
    );
  }
}

class UserZoneReputation {
  final String id;
  final String createdAt;
  final String updatedAt;
  final String userId;
  final String zoneId;
  final int level;
  final int totalReputation;
  final int reputationOnLevel;
  final int levelsGained;
  final UserZoneReputationName name;
  final int reputationToNextLevel;

  const UserZoneReputation({
    required this.id,
    required this.createdAt,
    required this.updatedAt,
    required this.userId,
    required this.zoneId,
    required this.level,
    required this.totalReputation,
    required this.reputationOnLevel,
    required this.levelsGained,
    required this.name,
    required this.reputationToNextLevel,
  });

  factory UserZoneReputation.fromJson(Map<String, dynamic> json) {
    return UserZoneReputation(
      id: json['id'] as String,
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
      userId: json['userId'] as String,
      zoneId: json['zoneId'] as String,
      level: (json['level'] as num).toInt(),
      totalReputation: (json['totalReputation'] as num).toInt(),
      reputationOnLevel: (json['reputationOnLevel'] as num).toInt(),
      levelsGained: (json['levelsGained'] as num).toInt(),
      name: UserZoneReputationName.fromString(json['name'] as String?) ?? UserZoneReputationName.neutral,
      reputationToNextLevel: (json['reputationToNextLevel'] as num).toInt(),
    );
  }
}
