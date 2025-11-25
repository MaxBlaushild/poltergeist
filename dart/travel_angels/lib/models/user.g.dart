// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

User _$UserFromJson(Map<String, dynamic> json) => User(
  id: User._idFromJson(json['id']),
  createdAt: User._dateTimeFromJson(json['createdAt']),
  updatedAt: User._dateTimeFromJson(json['updatedAt']),
  name: User._nameFromJson(json['name']),
  phoneNumber: User._phoneNumberFromJson(json['phoneNumber']),
  active: User._activeFromJson(json['active']),
  profilePictureUrl: User._profilePictureUrlFromJson(json['profilePictureUrl']),
  hasSeenTutorial: User._hasSeenTutorialFromJson(json['hasSeenTutorial']),
  partyId: User._partyIdFromJson(json['partyId']),
  username: User._usernameFromJson(json['username']),
  isActive: User._isActiveFromJson(json['isActive']),
  credits: User._creditsFromJson(json['credits']),
  dateOfBirth: User._dateTimeFromJson(json['dateOfBirth']),
  gender: User._genderFromJson(json['gender']),
  latitude: User._latitudeFromJson(json['latitude']),
  longitude: User._longitudeFromJson(json['longitude']),
  locationAddress: User._locationAddressFromJson(json['locationAddress']),
  bio: User._bioFromJson(json['bio']),
);

Map<String, dynamic> _$UserToJson(User instance) => <String, dynamic>{
  'id': instance.id,
  'createdAt': instance.createdAt?.toIso8601String(),
  'updatedAt': instance.updatedAt?.toIso8601String(),
  'name': instance.name,
  'phoneNumber': instance.phoneNumber,
  'active': instance.active,
  'profilePictureUrl': instance.profilePictureUrl,
  'hasSeenTutorial': instance.hasSeenTutorial,
  'partyId': instance.partyId,
  'username': instance.username,
  'isActive': instance.isActive,
  'credits': instance.credits,
  'dateOfBirth': instance.dateOfBirth?.toIso8601String(),
  'gender': instance.gender,
  'latitude': instance.latitude,
  'longitude': instance.longitude,
  'locationAddress': instance.locationAddress,
  'bio': instance.bio,
};
