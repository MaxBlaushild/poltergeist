import 'package:json_annotation/json_annotation.dart';

part 'user.g.dart';

@JsonSerializable()
class User {
  @JsonKey(name: 'id', fromJson: _idFromJson)
  final String? id;
  @JsonKey(name: 'createdAt', fromJson: _dateTimeFromJson)
  final DateTime? createdAt;
  @JsonKey(name: 'updatedAt', fromJson: _dateTimeFromJson)
  final DateTime? updatedAt;
  @JsonKey(fromJson: _nameFromJson)
  final String? name;
  @JsonKey(name: 'phoneNumber', fromJson: _phoneNumberFromJson)
  final String? phoneNumber;
  @JsonKey(fromJson: _activeFromJson)
  final bool? active;
  @JsonKey(name: 'profilePictureUrl', fromJson: _profilePictureUrlFromJson)
  final String? profilePictureUrl;
  @JsonKey(name: 'hasSeenTutorial', fromJson: _hasSeenTutorialFromJson)
  final bool? hasSeenTutorial;
  @JsonKey(name: 'partyId', fromJson: _partyIdFromJson)
  final String? partyId;
  @JsonKey(fromJson: _usernameFromJson)
  final String? username;
  @JsonKey(name: 'isActive', fromJson: _isActiveFromJson)
  final bool? isActive;
  @JsonKey(fromJson: _creditsFromJson)
  final int? credits;

  User({
    this.id,
    this.createdAt,
    this.updatedAt,
    this.name,
    this.phoneNumber,
    this.active,
    this.profilePictureUrl,
    this.hasSeenTutorial,
    this.partyId,
    this.username,
    this.isActive,
    this.credits,
  });

  static String? _idFromJson(dynamic json) {
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

  static String? _phoneNumberFromJson(dynamic json) {
    if (json == null) return null;
    return json as String?;
  }

  static bool? _activeFromJson(dynamic json) {
    if (json == null) return null;
    return json as bool?;
  }

  static bool? _hasSeenTutorialFromJson(dynamic json) {
    if (json == null) return null;
    return json as bool?;
  }

  static String? _nameFromJson(dynamic json) {
    if (json == null || json == '') return null;
    return json as String?;
  }

  static String? _profilePictureUrlFromJson(dynamic json) {
    if (json == null || json == '') return null;
    return json as String?;
  }

  static String? _partyIdFromJson(dynamic json) {
    if (json == null) return null;
    return json.toString();
  }

  static String? _usernameFromJson(dynamic json) {
    if (json == null) return null;
    return json as String?;
  }

  static bool? _isActiveFromJson(dynamic json) {
    if (json == null) return null;
    return json as bool?;
  }

  static int? _creditsFromJson(dynamic json) {
    if (json == null) return null;
    if (json is int) return json;
    if (json is String) {
      return int.tryParse(json);
    }
    return null;
  }

  factory User.fromJson(Map<String, dynamic> json) => _$UserFromJson(json);

  Map<String, dynamic> toJson() => _$UserToJson(this);
}

