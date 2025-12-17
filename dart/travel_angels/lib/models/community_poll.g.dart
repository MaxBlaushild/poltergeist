// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'community_poll.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

CommunityPoll _$CommunityPollFromJson(Map<String, dynamic> json) =>
    CommunityPoll(
      id: CommunityPoll._idFromJson(json['id']),
      createdAt: CommunityPoll._dateTimeFromJson(json['createdAt']),
      updatedAt: CommunityPoll._dateTimeFromJson(json['updatedAt']),
      userId: CommunityPoll._idFromJson(json['userId']),
      question: json['question'] as String,
      options: (json['options'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
    );

Map<String, dynamic> _$CommunityPollToJson(CommunityPoll instance) =>
    <String, dynamic>{
      'id': instance.id,
      'createdAt': instance.createdAt?.toIso8601String(),
      'updatedAt': instance.updatedAt?.toIso8601String(),
      'userId': instance.userId,
      'question': instance.question,
      'options': instance.options,
    };
