import 'package:json_annotation/json_annotation.dart';

part 'user_level.g.dart';

@JsonSerializable()
class UserLevel {
  @JsonKey(name: 'id', fromJson: _idFromJson)
  final String? id;
  @JsonKey(name: 'createdAt', fromJson: _dateTimeFromJson)
  final DateTime? createdAt;
  @JsonKey(name: 'updatedAt', fromJson: _dateTimeFromJson)
  final DateTime? updatedAt;
  @JsonKey(name: 'userId', fromJson: _userIdFromJson)
  final String? userId;
  @JsonKey(fromJson: _levelFromJson)
  final int level;
  @JsonKey(name: 'experiencePointsOnLevel', fromJson: _experiencePointsOnLevelFromJson)
  final int experiencePointsOnLevel;
  @JsonKey(name: 'totalExperiencePoints', fromJson: _totalExperiencePointsFromJson)
  final int totalExperiencePoints;
  @JsonKey(name: 'levelsGained', fromJson: _levelsGainedFromJson)
  final int? levelsGained;
  @JsonKey(name: 'experienceToNextLevel', fromJson: _experienceToNextLevelFromJson)
  final int experienceToNextLevel;

  UserLevel({
    this.id,
    this.createdAt,
    this.updatedAt,
    this.userId,
    required this.level,
    required this.experiencePointsOnLevel,
    required this.totalExperiencePoints,
    this.levelsGained,
    required this.experienceToNextLevel,
  });

  static String? _idFromJson(dynamic json) {
    if (json == null) return null;
    return json.toString();
  }

  static String? _userIdFromJson(dynamic json) {
    if (json == null) return null;
    return json.toString();
  }

  static DateTime? _dateTimeFromJson(dynamic json) {
    if (json == null) return null;
    if (json is String) {
      try {
        return DateTime.parse(json);
      } catch (e) {
        return null;
      }
    }
    return null;
  }

  static int _levelFromJson(dynamic json) {
    if (json == null) return 1;
    if (json is int) return json;
    if (json is String) {
      return int.tryParse(json) ?? 1;
    }
    return 1;
  }

  static int _experiencePointsOnLevelFromJson(dynamic json) {
    if (json == null) return 0;
    if (json is int) return json;
    if (json is String) {
      return int.tryParse(json) ?? 0;
    }
    return 0;
  }

  static int _totalExperiencePointsFromJson(dynamic json) {
    if (json == null) return 0;
    if (json is int) return json;
    if (json is String) {
      return int.tryParse(json) ?? 0;
    }
    return 0;
  }

  static int? _levelsGainedFromJson(dynamic json) {
    if (json == null) return null;
    if (json is int) return json;
    if (json is String) {
      return int.tryParse(json);
    }
    return null;
  }

  static int _experienceToNextLevelFromJson(dynamic json) {
    if (json == null) return 100; // Default for level 1
    if (json is int) return json;
    if (json is String) {
      return int.tryParse(json) ?? 100;
    }
    return 100;
  }

  factory UserLevel.fromJson(Map<String, dynamic> json) =>
      _$UserLevelFromJson(json);

  Map<String, dynamic> toJson() => _$UserLevelToJson(this);
}
