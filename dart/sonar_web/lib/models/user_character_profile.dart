import 'character_stats.dart';
import 'user.dart';
import 'user_level.dart';

class UserCharacterProfile {
  final User user;
  final CharacterStats stats;
  final UserLevel userLevel;

  const UserCharacterProfile({
    required this.user,
    required this.stats,
    required this.userLevel,
  });

  factory UserCharacterProfile.fromJson(Map<String, dynamic> json) {
    return UserCharacterProfile(
      user: User.fromJson(json['user'] as Map<String, dynamic>),
      stats: CharacterStats.fromJson(json['stats'] as Map<String, dynamic>),
      userLevel: UserLevel.fromJson(json['userLevel'] as Map<String, dynamic>),
    );
  }
}
