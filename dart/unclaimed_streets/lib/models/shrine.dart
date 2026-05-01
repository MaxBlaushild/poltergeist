class Shrine {
  const Shrine({
    required this.id,
    required this.shrineTemplateId,
    required this.name,
    required this.description,
    required this.blessingName,
    required this.effectDescription,
    required this.effectKind,
    required this.baseMagnitude,
    required this.zoneId,
    required this.latitude,
    required this.longitude,
    required this.cooldownSeconds,
    this.availableNow = true,
    this.lastUsedAt,
    this.nextAvailableAt,
    this.cooldownSecondsRemaining = 0,
    this.mapMarkerUrl = '',
  });

  final String id;
  final String shrineTemplateId;
  final String name;
  final String description;
  final String blessingName;
  final String effectDescription;
  final String effectKind;
  final int baseMagnitude;
  final String zoneId;
  final double latitude;
  final double longitude;
  final int cooldownSeconds;
  final bool availableNow;
  final DateTime? lastUsedAt;
  final DateTime? nextAvailableAt;
  final int cooldownSecondsRemaining;
  final String mapMarkerUrl;

  factory Shrine.fromJson(Map<String, dynamic> json) {
    DateTime? parseDateTime(dynamic raw) {
      if (raw == null) return null;
      final text = raw.toString().trim();
      if (text.isEmpty) return null;
      return DateTime.tryParse(text)?.toLocal();
    }

    return Shrine(
      id: json['id']?.toString() ?? '',
      shrineTemplateId: json['shrineTemplateId']?.toString() ?? '',
      name: json['name']?.toString() ?? 'Shrine',
      description: json['description']?.toString() ?? '',
      blessingName: json['blessingName']?.toString() ?? '',
      effectDescription: json['effectDescription']?.toString() ?? '',
      effectKind: json['effectKind']?.toString() ?? '',
      baseMagnitude: (json['baseMagnitude'] as num?)?.toInt() ?? 0,
      zoneId: json['zoneId']?.toString() ?? '',
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      cooldownSeconds: (json['cooldownSeconds'] as num?)?.toInt() ?? 0,
      availableNow: json['availableNow'] as bool? ?? true,
      lastUsedAt: parseDateTime(json['lastUsedAt']),
      nextAvailableAt: parseDateTime(json['nextAvailableAt']),
      cooldownSecondsRemaining:
          (json['cooldownSecondsRemaining'] as num?)?.toInt() ?? 0,
      mapMarkerUrl: json['mapMarkerUrl']?.toString() ?? '',
    );
  }

  Shrine copyWith({
    bool? availableNow,
    DateTime? lastUsedAt,
    DateTime? nextAvailableAt,
    int? cooldownSecondsRemaining,
    String? mapMarkerUrl,
  }) {
    return Shrine(
      id: id,
      shrineTemplateId: shrineTemplateId,
      name: name,
      description: description,
      blessingName: blessingName,
      effectDescription: effectDescription,
      effectKind: effectKind,
      baseMagnitude: baseMagnitude,
      zoneId: zoneId,
      latitude: latitude,
      longitude: longitude,
      cooldownSeconds: cooldownSeconds,
      availableNow: availableNow ?? this.availableNow,
      lastUsedAt: lastUsedAt ?? this.lastUsedAt,
      nextAvailableAt: nextAvailableAt ?? this.nextAvailableAt,
      cooldownSecondsRemaining:
          cooldownSecondsRemaining ?? this.cooldownSecondsRemaining,
      mapMarkerUrl: mapMarkerUrl ?? this.mapMarkerUrl,
    );
  }
}
