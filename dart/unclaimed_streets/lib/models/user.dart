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
  });

  factory User.fromJson(Map<String, dynamic> json) {
    String _str(String primary, String fallback) =>
        (json[primary] ?? json[fallback]) as String? ?? '';

    return User(
      id: _str('id', 'ID'),
      phoneNumber: _str('phoneNumber', 'PhoneNumber'),
      name: _str('name', 'Name'),
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
  };
}
