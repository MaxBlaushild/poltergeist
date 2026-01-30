class User {
  final String? id;
  final DateTime? createdAt;
  final DateTime? updatedAt;
  final String? name;
  final String? phoneNumber;
  final bool? active;
  final String? profilePictureUrl;
  final bool? hasSeenTutorial;
  final String? partyId;
  final String? username;
  final bool? isActive;
  final int? credits;
  final DateTime? dateOfBirth;
  final String? gender;
  final double? latitude;
  final double? longitude;
  final String? locationAddress;
  final String? bio;
  final String? category;
  final String? ageRange;

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
    this.dateOfBirth,
    this.gender,
    this.latitude,
    this.longitude,
    this.locationAddress,
    this.bio,
    this.category,
    this.ageRange,
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

  static String? _genderFromJson(dynamic json) {
    if (json == null || json == '') return null;
    return json as String?;
  }

  static double? _latitudeFromJson(dynamic json) {
    if (json == null) return null;
    if (json is double) return json;
    if (json is int) return json.toDouble();
    if (json is String) {
      return double.tryParse(json);
    }
    return null;
  }

  static double? _longitudeFromJson(dynamic json) {
    if (json == null) return null;
    if (json is double) return json;
    if (json is int) return json.toDouble();
    if (json is String) {
      return double.tryParse(json);
    }
    return null;
  }

  static String? _locationAddressFromJson(dynamic json) {
    if (json == null || json == '') return null;
    return json as String?;
  }

  static String? _bioFromJson(dynamic json) {
    if (json == null || json == '') return null;
    return json as String?;
  }

  static String? _categoryFromJson(dynamic json) {
    if (json == null || json == '') return null;
    return json as String?;
  }

  static String? _ageRangeFromJson(dynamic json) {
    if (json == null || json == '') return null;
    return json as String?;
  }

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: _idFromJson(json['id']),
      createdAt: _dateTimeFromJson(json['createdAt']),
      updatedAt: _dateTimeFromJson(json['updatedAt']),
      name: _nameFromJson(json['name']),
      phoneNumber: _phoneNumberFromJson(json['phoneNumber']),
      active: _activeFromJson(json['active']),
      profilePictureUrl: _profilePictureUrlFromJson(json['profilePictureUrl']),
      hasSeenTutorial: _hasSeenTutorialFromJson(json['hasSeenTutorial']),
      partyId: _partyIdFromJson(json['partyId']),
      username: _usernameFromJson(json['username']),
      isActive: _isActiveFromJson(json['isActive']),
      credits: _creditsFromJson(json['credits']),
      dateOfBirth: _dateTimeFromJson(json['dateOfBirth']),
      gender: _genderFromJson(json['gender']),
      latitude: _latitudeFromJson(json['latitude']),
      longitude: _longitudeFromJson(json['longitude']),
      locationAddress: _locationAddressFromJson(json['locationAddress']),
      bio: _bioFromJson(json['bio']),
      category: _categoryFromJson(json['category']),
      ageRange: _ageRangeFromJson(json['ageRange']),
    );
  }
}
