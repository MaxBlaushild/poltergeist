import 'character_stats.dart';
import 'user.dart';

class UserCharacterProfile {
  final User user;
  final CharacterStats stats;

  const UserCharacterProfile({
    required this.user,
    required this.stats,
  });

  factory UserCharacterProfile.fromJson(Map<String, dynamic> json) {
    return UserCharacterProfile(
      user: User.fromJson(json['user'] as Map<String, dynamic>),
      stats: CharacterStats.fromJson(json['stats'] as Map<String, dynamic>),
    );
  }
}
