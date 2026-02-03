class CharacterLocation {
  final String id;
  final String characterId;
  final double latitude;
  final double longitude;

  const CharacterLocation({
    required this.id,
    required this.characterId,
    required this.latitude,
    required this.longitude,
  });

  factory CharacterLocation.fromJson(Map<String, dynamic> json) {
    return CharacterLocation(
      id: json['id'] as String? ?? '',
      characterId: json['characterId'] as String? ?? '',
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0,
    );
  }
}
