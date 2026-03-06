class HealingFountain {
  final String id;
  final String name;
  final String description;
  final String thumbnailUrl;
  final String zoneId;
  final double latitude;
  final double longitude;
  final bool availableNow;
  final DateTime? lastUsedAt;
  final DateTime? nextAvailableAt;
  final int cooldownSecondsRemaining;

  const HealingFountain({
    required this.id,
    required this.name,
    required this.description,
    required this.thumbnailUrl,
    required this.zoneId,
    required this.latitude,
    required this.longitude,
    this.availableNow = true,
    this.lastUsedAt,
    this.nextAvailableAt,
    this.cooldownSecondsRemaining = 0,
  });

  factory HealingFountain.fromJson(Map<String, dynamic> json) {
    DateTime? parseDateTime(dynamic raw) {
      if (raw == null) return null;
      final text = raw.toString().trim();
      if (text.isEmpty) return null;
      return DateTime.tryParse(text)?.toLocal();
    }

    return HealingFountain(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? 'Healing Fountain',
      description: json['description']?.toString() ?? '',
      thumbnailUrl: json['thumbnailUrl']?.toString() ?? '',
      zoneId: json['zoneId']?.toString() ?? '',
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      availableNow: json['availableNow'] as bool? ?? true,
      lastUsedAt: parseDateTime(json['lastUsedAt']),
      nextAvailableAt: parseDateTime(json['nextAvailableAt']),
      cooldownSecondsRemaining:
          (json['cooldownSecondsRemaining'] as num?)?.toInt() ?? 0,
    );
  }

  HealingFountain copyWith({
    bool? availableNow,
    DateTime? lastUsedAt,
    DateTime? nextAvailableAt,
    int? cooldownSecondsRemaining,
  }) {
    return HealingFountain(
      id: id,
      name: name,
      description: description,
      thumbnailUrl: thumbnailUrl,
      zoneId: zoneId,
      latitude: latitude,
      longitude: longitude,
      availableNow: availableNow ?? this.availableNow,
      lastUsedAt: lastUsedAt ?? this.lastUsedAt,
      nextAvailableAt: nextAvailableAt ?? this.nextAvailableAt,
      cooldownSecondsRemaining:
          cooldownSecondsRemaining ?? this.cooldownSecondsRemaining,
    );
  }
}
