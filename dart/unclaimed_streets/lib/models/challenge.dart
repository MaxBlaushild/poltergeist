import 'quest_node.dart';

class Challenge {
  final String id;
  final String zoneId;
  final String? pointOfInterestId;
  final double latitude;
  final double longitude;
  final List<QuestNodePolygonPoint> polygonPoints;
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
    this.polygonPoints = const [],
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

  bool get hasPolygon => polygonPoints.length >= 3;

  factory Challenge.fromJson(Map<String, dynamic> json) {
    final rawPolygonPoints = json['polygonPoints'] as List<dynamic>?;
    return Challenge(
      id: json['id']?.toString() ?? '',
      zoneId: json['zoneId']?.toString() ?? '',
      pointOfInterestId: json['pointOfInterestId']?.toString(),
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      polygonPoints:
          rawPolygonPoints
              ?.map((entry) {
                if (entry is List && entry.length >= 2) {
                  final lng = (entry[0] as num?)?.toDouble();
                  final lat = (entry[1] as num?)?.toDouble();
                  if (lat != null && lng != null) {
                    return QuestNodePolygonPoint(latitude: lat, longitude: lng);
                  }
                }
                if (entry is Map<String, dynamic>) {
                  return QuestNodePolygonPoint.fromJson(entry);
                }
                return null;
              })
              .whereType<QuestNodePolygonPoint>()
              .toList() ??
          const [],
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
