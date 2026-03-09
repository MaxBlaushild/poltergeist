class Challenge {
  final String id;
  final String zoneId;
  final String? pointOfInterestId;
  final double latitude;
  final double longitude;
  final String question;
  final String description;
  final String imageUrl;
  final String thumbnailUrl;
  final int reward;
  final int? inventoryItemId;
  final List<Map<String, dynamic>> itemChoiceRewards;
  final String submissionType;
  final int difficulty;
  final bool scaleWithUserLevel;
  final List<String> statTags;
  final String? proficiency;

  const Challenge({
    required this.id,
    required this.zoneId,
    this.pointOfInterestId,
    required this.latitude,
    required this.longitude,
    required this.question,
    this.description = '',
    this.imageUrl = '',
    this.thumbnailUrl = '',
    required this.reward,
    this.inventoryItemId,
    this.itemChoiceRewards = const [],
    this.submissionType = 'photo',
    this.difficulty = 0,
    this.scaleWithUserLevel = false,
    this.statTags = const [],
    this.proficiency,
  });

  factory Challenge.fromJson(Map<String, dynamic> json) {
    return Challenge(
      id: json['id']?.toString() ?? '',
      zoneId: json['zoneId']?.toString() ?? '',
      pointOfInterestId: json['pointOfInterestId']?.toString(),
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      question: json['question']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      imageUrl: json['imageUrl']?.toString() ?? '',
      thumbnailUrl: json['thumbnailUrl']?.toString() ?? '',
      reward: (json['reward'] as num?)?.toInt() ?? 0,
      inventoryItemId: (json['inventoryItemId'] as num?)?.toInt(),
      itemChoiceRewards:
          (json['itemChoiceRewards'] as List<dynamic>?)
              ?.whereType<Map>()
              .map((item) => Map<String, dynamic>.from(item))
              .toList() ??
          const [],
      submissionType:
          (json['submissionType']?.toString().trim().toLowerCase().isNotEmpty ??
              false)
          ? json['submissionType'].toString().trim().toLowerCase()
          : 'photo',
      difficulty: (json['difficulty'] as num?)?.toInt() ?? 0,
      scaleWithUserLevel: json['scaleWithUserLevel'] as bool? ?? false,
      statTags:
          (json['statTags'] as List<dynamic>?)
              ?.map((tag) => tag.toString())
              .where((tag) => tag.trim().isNotEmpty)
              .toList() ??
          const [],
      proficiency: json['proficiency']?.toString(),
    );
  }
}
