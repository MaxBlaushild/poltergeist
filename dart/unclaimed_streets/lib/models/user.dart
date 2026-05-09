class User {
  final String id;
  final String phoneNumber;
  final String name;
  final String username;
  final String profilePictureUrl;
  final String backProfilePictureUrl;
  final bool hasCustomizedPortrait;
  final String? partyId;
  final bool? isActive;
  final int gold;
  final double? latitude;
  final double? longitude;

  const User({
    required this.id,
    required this.phoneNumber,
    required this.name,
    required this.username,
    required this.profilePictureUrl,
    this.backProfilePictureUrl = '',
    this.hasCustomizedPortrait = false,
    this.partyId,
    this.isActive,
    this.gold = 0,
    this.latitude,
    this.longitude,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    String strFromJson(String primary, String fallback) =>
        (json[primary] ?? json[fallback]) as String? ?? '';

    double? doubleFromJson(dynamic raw) {
      if (raw is num) return raw.toDouble();
      if (raw is String) return double.tryParse(raw.trim());
      return null;
    }

    return User(
      id: strFromJson('id', 'ID'),
      phoneNumber: strFromJson('phoneNumber', 'PhoneNumber'),
      name: strFromJson('name', 'Name'),
      username: json['username'] as String? ?? '',
      profilePictureUrl: json['profilePictureUrl'] as String? ?? '',
      backProfilePictureUrl:
          (json['backProfilePictureUrl'] ?? json['back_profile_picture_url'])
              as String? ??
          '',
      hasCustomizedPortrait: json['hasCustomizedPortrait'] as bool? ?? false,
      partyId: json['partyId'] as String?,
      isActive: json['isActive'] as bool?,
      gold: (json['gold'] as num?)?.toInt() ?? 0,
      latitude: doubleFromJson(json['latitude'] ?? json['lat']),
      longitude: doubleFromJson(json['longitude'] ?? json['lng']),
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'phoneNumber': phoneNumber,
    'name': name,
    'username': username,
    'profilePictureUrl': profilePictureUrl,
    'backProfilePictureUrl': backProfilePictureUrl,
    'hasCustomizedPortrait': hasCustomizedPortrait,
    'partyId': partyId,
    'isActive': isActive,
    'gold': gold,
    'latitude': latitude,
    'longitude': longitude,
  };
}
