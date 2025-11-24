// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'friend_invite.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

FriendInvite _$FriendInviteFromJson(Map<String, dynamic> json) => FriendInvite(
  id: FriendInvite._idFromJson(json['id']),
  createdAt: FriendInvite._dateTimeFromJson(json['createdAt']),
  updatedAt: FriendInvite._dateTimeFromJson(json['updatedAt']),
  inviterId: FriendInvite._idFromJson(json['inviterId']),
  inviteeId: FriendInvite._idFromJson(json['inviteeId']),
  inviter: json['inviter'] == null
      ? null
      : User.fromJson(json['inviter'] as Map<String, dynamic>),
  invitee: json['invitee'] == null
      ? null
      : User.fromJson(json['invitee'] as Map<String, dynamic>),
);

Map<String, dynamic> _$FriendInviteToJson(FriendInvite instance) =>
    <String, dynamic>{
      'id': instance.id,
      'createdAt': instance.createdAt?.toIso8601String(),
      'updatedAt': instance.updatedAt?.toIso8601String(),
      'inviterId': instance.inviterId,
      'inviteeId': instance.inviteeId,
      'inviter': instance.inviter,
      'invitee': instance.invitee,
    };
