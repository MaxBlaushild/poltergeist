// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user_level.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

UserLevel _$UserLevelFromJson(Map<String, dynamic> json) => UserLevel(
  id: UserLevel._idFromJson(json['id']),
  createdAt: UserLevel._dateTimeFromJson(json['createdAt']),
  updatedAt: UserLevel._dateTimeFromJson(json['updatedAt']),
  userId: UserLevel._userIdFromJson(json['userId']),
  level: UserLevel._levelFromJson(json['level']),
  experiencePointsOnLevel: UserLevel._experiencePointsOnLevelFromJson(
    json['experiencePointsOnLevel'],
  ),
  totalExperiencePoints: UserLevel._totalExperiencePointsFromJson(
    json['totalExperiencePoints'],
  ),
  levelsGained: UserLevel._levelsGainedFromJson(json['levelsGained']),
  experienceToNextLevel: UserLevel._experienceToNextLevelFromJson(
    json['experienceToNextLevel'],
  ),
);

Map<String, dynamic> _$UserLevelToJson(UserLevel instance) => <String, dynamic>{
  'id': instance.id,
  'createdAt': instance.createdAt?.toIso8601String(),
  'updatedAt': instance.updatedAt?.toIso8601String(),
  'userId': instance.userId,
  'level': instance.level,
  'experiencePointsOnLevel': instance.experiencePointsOnLevel,
  'totalExperiencePoints': instance.totalExperiencePoints,
  'levelsGained': instance.levelsGained,
  'experienceToNextLevel': instance.experienceToNextLevel,
};
