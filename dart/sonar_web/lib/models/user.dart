class User {
  final String id;
  final String phoneNumber;
  final String name;
  final String username;
  final String profilePictureUrl;
  final String? partyId;
  final bool? isActive;
  final int gold;

  const User({
    required this.id,
    required this.phoneNumber,
    required this.name,
    required this.username,
    required this.profilePictureUrl,
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
        'partyId': partyId,
        'isActive': isActive,
        'gold': gold,
      };
}
