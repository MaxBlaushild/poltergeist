// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'friend.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Friend _$FriendFromJson(Map<String, dynamic> json) => Friend(
  id: Friend._idFromJson(json['id']),
  createdAt: Friend._dateTimeFromJson(json['createdAt']),
  updatedAt: Friend._dateTimeFromJson(json['updatedAt']),
  firstUserId: Friend._idFromJson(json['firstUserId']),
  secondUserId: Friend._idFromJson(json['secondUserId']),
  firstUser: json['firstUser'] == null
      ? null
      : User.fromJson(json['firstUser'] as Map<String, dynamic>),
  secondUser: json['secondUser'] == null
      ? null
      : User.fromJson(json['secondUser'] as Map<String, dynamic>),
);

Map<String, dynamic> _$FriendToJson(Friend instance) => <String, dynamic>{
  'id': instance.id,
  'createdAt': instance.createdAt?.toIso8601String(),
  'updatedAt': instance.updatedAt?.toIso8601String(),
  'firstUserId': instance.firstUserId,
  'secondUserId': instance.secondUserId,
  'firstUser': instance.firstUser,
  'secondUser': instance.secondUser,
};
