class BaseOwner {
  final String id;
  final String name;
  final String username;
  final String profilePictureUrl;

  const BaseOwner({
    required this.id,
    required this.name,
    required this.username,
    required this.profilePictureUrl,
  });

  factory BaseOwner.fromJson(Map<String, dynamic> json) {
    return BaseOwner(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      username: json['username']?.toString() ?? '',
      profilePictureUrl: json['profilePictureUrl']?.toString() ?? '',
    );
  }

  String get displayName {
    if (username.trim().isNotEmpty) {
      return '@${username.trim()}';
    }
    if (name.trim().isNotEmpty) {
      return name.trim();
    }
    return 'Unknown adventurer';
  }

  String get secondaryName {
    final trimmedName = name.trim();
    if (trimmedName.isEmpty || trimmedName == displayName) {
      return '';
    }
    return trimmedName;
  }
}

class BasePin {
  final String id;
  final String userId;
  final BaseOwner owner;
  final double latitude;
  final double longitude;
  final String thumbnailUrl;

  const BasePin({
    required this.id,
    required this.userId,
    required this.owner,
    required this.latitude,
    required this.longitude,
    required this.thumbnailUrl,
  });

  factory BasePin.fromJson(Map<String, dynamic> json) {
    final rawOwner = json['owner'];
    final ownerMap = rawOwner is Map<String, dynamic>
        ? rawOwner
        : rawOwner is Map
        ? Map<String, dynamic>.from(rawOwner)
        : const <String, dynamic>{};
    return BasePin(
      id: json['id']?.toString() ?? '',
      userId: json['userId']?.toString() ?? '',
      owner: BaseOwner.fromJson(ownerMap),
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      thumbnailUrl: json['thumbnailUrl']?.toString() ?? '',
    );
  }
}
