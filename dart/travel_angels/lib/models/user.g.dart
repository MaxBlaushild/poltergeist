// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

User _$UserFromJson(Map<String, dynamic> json) => User(
  id: User._idFromJson(json['ID']),
  createdAt: User._dateTimeFromJson(json['CreatedAt']),
  updatedAt: User._dateTimeFromJson(json['UpdatedAt']),
  name: User._nameFromJson(json['name']),
  phoneNumber: User._phoneNumberFromJson(json['phoneNumber']),
  active: User._activeFromJson(json['active']),
  profilePictureUrl: User._profilePictureUrlFromJson(json['profilePictureUrl']),
  hasSeenTutorial: User._hasSeenTutorialFromJson(json['hasSeenTutorial']),
  partyId: User._partyIdFromJson(json['partyId']),
  username: User._usernameFromJson(json['username']),
  isActive: User._isActiveFromJson(json['isActive']),
);

Map<String, dynamic> _$UserToJson(User instance) => <String, dynamic>{
  'ID': instance.id,
  'CreatedAt': instance.createdAt?.toIso8601String(),
  'UpdatedAt': instance.updatedAt?.toIso8601String(),
  'name': instance.name,
  'phoneNumber': instance.phoneNumber,
  'active': instance.active,
  'profilePictureUrl': instance.profilePictureUrl,
  'hasSeenTutorial': instance.hasSeenTutorial,
  'partyId': instance.partyId,
  'username': instance.username,
  'isActive': instance.isActive,
};
