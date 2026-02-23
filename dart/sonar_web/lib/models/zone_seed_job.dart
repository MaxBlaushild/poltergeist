class ZoneSeedJob {
  final String id;
  final String zoneId;
  final String status;
  final String? errorMessage;
  final int placeCount;
  final int characterCount;
  final int questCount;
  final DateTime? createdAt;
  final DateTime? updatedAt;
  final ZoneSeedDraft? draft;

  const ZoneSeedJob({
    required this.id,
    required this.zoneId,
    required this.status,
    this.errorMessage,
    required this.placeCount,
    required this.characterCount,
    required this.questCount,
    this.createdAt,
    this.updatedAt,
    this.draft,
  });

  factory ZoneSeedJob.fromJson(Map<String, dynamic> json) {
    return ZoneSeedJob(
      id: json['id'] as String,
      zoneId: json['zoneId'] as String,
      status: json['status'] as String? ?? '',
      errorMessage: json['errorMessage'] as String?,
      placeCount: (json['placeCount'] as num?)?.toInt() ?? 0,
      characterCount: (json['characterCount'] as num?)?.toInt() ?? 0,
      questCount: (json['questCount'] as num?)?.toInt() ?? 0,
      createdAt: _parseDate(json['createdAt']),
      updatedAt: _parseDate(json['updatedAt']),
      draft: json['draft'] is Map<String, dynamic>
          ? ZoneSeedDraft.fromJson(json['draft'] as Map<String, dynamic>)
          : null,
    );
  }

  static DateTime? _parseDate(dynamic value) {
    if (value is String && value.isNotEmpty) {
      return DateTime.tryParse(value);
    }
    return null;
  }
}

class ZoneSeedDraft {
  final String? fantasyName;
  final String? zoneDescription;
  final List<ZoneSeedPointOfInterestDraft> pointsOfInterest;
  final List<ZoneSeedCharacterDraft> characters;
  final List<ZoneSeedQuestDraft> quests;

  const ZoneSeedDraft({
    this.fantasyName,
    this.zoneDescription,
    required this.pointsOfInterest,
    required this.characters,
    required this.quests,
  });

  factory ZoneSeedDraft.fromJson(Map<String, dynamic> json) {
    return ZoneSeedDraft(
      fantasyName: json['fantasyName'] as String?,
      zoneDescription: json['zoneDescription'] as String?,
      pointsOfInterest: (json['pointsOfInterest'] as List<dynamic>? ?? [])
          .whereType<Map<String, dynamic>>()
          .map(ZoneSeedPointOfInterestDraft.fromJson)
          .toList(),
      characters: (json['characters'] as List<dynamic>? ?? [])
          .whereType<Map<String, dynamic>>()
          .map(ZoneSeedCharacterDraft.fromJson)
          .toList(),
      quests: (json['quests'] as List<dynamic>? ?? [])
          .whereType<Map<String, dynamic>>()
          .map(ZoneSeedQuestDraft.fromJson)
          .toList(),
    );
  }
}

class ZoneSeedPointOfInterestDraft {
  final String draftId;
  final String placeId;
  final String name;
  final String? address;
  final List<String> types;
  final double latitude;
  final double longitude;
  final double? rating;
  final int? userRatingCount;
  final String? editorialSummary;

  const ZoneSeedPointOfInterestDraft({
    required this.draftId,
    required this.placeId,
    required this.name,
    this.address,
    required this.types,
    required this.latitude,
    required this.longitude,
    this.rating,
    this.userRatingCount,
    this.editorialSummary,
  });

  factory ZoneSeedPointOfInterestDraft.fromJson(Map<String, dynamic> json) {
    return ZoneSeedPointOfInterestDraft(
      draftId: json['draftId'] as String? ?? '',
      placeId: json['placeId'] as String? ?? '',
      name: json['name'] as String? ?? '',
      address: json['address'] as String?,
      types: (json['types'] as List<dynamic>? ?? []).map((e) => e.toString()).toList(),
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0,
      rating: (json['rating'] as num?)?.toDouble(),
      userRatingCount: (json['userRatingCount'] as num?)?.toInt(),
      editorialSummary: json['editorialSummary'] as String?,
    );
  }
}

class ZoneSeedCharacterDraft {
  final String draftId;
  final String name;
  final String description;
  final String placeId;

  const ZoneSeedCharacterDraft({
    required this.draftId,
    required this.name,
    required this.description,
    required this.placeId,
  });

  factory ZoneSeedCharacterDraft.fromJson(Map<String, dynamic> json) {
    return ZoneSeedCharacterDraft(
      draftId: json['draftId'] as String? ?? '',
      name: json['name'] as String? ?? '',
      description: json['description'] as String? ?? '',
      placeId: json['placeId'] as String? ?? '',
    );
  }
}

class ZoneSeedQuestDraft {
  final String draftId;
  final String name;
  final String description;
  final List<String> acceptanceDialogue;
  final String placeId;
  final String questGiverDraftId;
  final String? challengeQuestion;
  final int gold;

  const ZoneSeedQuestDraft({
    required this.draftId,
    required this.name,
    required this.description,
    required this.acceptanceDialogue,
    required this.placeId,
    required this.questGiverDraftId,
    this.challengeQuestion,
    required this.gold,
  });

  factory ZoneSeedQuestDraft.fromJson(Map<String, dynamic> json) {
    return ZoneSeedQuestDraft(
      draftId: json['draftId'] as String? ?? '',
      name: json['name'] as String? ?? '',
      description: json['description'] as String? ?? '',
      acceptanceDialogue: (json['acceptanceDialogue'] as List<dynamic>? ?? [])
          .map((e) => e.toString())
          .toList(),
      placeId: json['placeId'] as String? ?? '',
      questGiverDraftId: json['questGiverDraftId'] as String? ?? '',
      challengeQuestion: json['challengeQuestion'] as String?,
      gold: (json['gold'] as num?)?.toInt() ?? 0,
    );
  }
}
