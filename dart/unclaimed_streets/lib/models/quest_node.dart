import 'point_of_interest.dart';
import 'exposition.dart';
import 'quest_node_objective.dart';

class QuestNodePolygonPoint {
  final double latitude;
  final double longitude;

  const QuestNodePolygonPoint({
    required this.latitude,
    required this.longitude,
  });

  factory QuestNodePolygonPoint.fromJson(Map<String, dynamic> json) {
    return QuestNodePolygonPoint(
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
    );
  }
}

class QuestNode {
  static const submissionTypeText = 'text';
  static const submissionTypePhoto = 'photo';
  static const submissionTypeVideo = 'video';

  final String id;
  final int orderIndex;
  final PointOfInterest? pointOfInterest;
  final String objectiveText;
  final String? scenarioId;
  final String? expositionId;
  final String? monsterId;
  final String? monsterEncounterId;
  final String? challengeId;
  final List<QuestNodePolygonPoint> polygon;
  final QuestNodeObjective? objective;
  final Exposition? exposition;
  final String submissionType;

  const QuestNode({
    required this.id,
    required this.orderIndex,
    this.submissionType = submissionTypePhoto,
    this.pointOfInterest,
    this.objectiveText = '',
    this.scenarioId,
    this.expositionId,
    this.monsterId,
    this.monsterEncounterId,
    this.challengeId,
    this.polygon = const [],
    this.objective,
    this.exposition,
  });

  factory QuestNode.fromJson(Map<String, dynamic> json) {
    final rawSubmissionType = (json['submissionType'] as String?)
        ?.trim()
        .toLowerCase();
    final submissionType =
        (rawSubmissionType == submissionTypeText ||
            rawSubmissionType == submissionTypePhoto ||
            rawSubmissionType == submissionTypeVideo)
        ? rawSubmissionType!
        : submissionTypePhoto;
    return QuestNode(
      id: json['id'] as String? ?? '',
      orderIndex: (json['orderIndex'] as num?)?.toInt() ?? 0,
      submissionType: submissionType,
      pointOfInterest: json['pointOfInterest'] is Map<String, dynamic>
          ? PointOfInterest.fromJson(
              json['pointOfInterest'] as Map<String, dynamic>,
            )
          : null,
      objectiveText: json['objectiveText']?.toString() ?? '',
      scenarioId: json['scenarioId']?.toString(),
      expositionId: json['expositionId']?.toString(),
      monsterId: json['monsterId']?.toString(),
      monsterEncounterId: json['monsterEncounterId']?.toString(),
      challengeId: json['challengeId']?.toString(),
      polygon:
          (json['polygon'] as List<dynamic>?)
              ?.map(
                (e) =>
                    QuestNodePolygonPoint.fromJson(e as Map<String, dynamic>),
              )
              .toList() ??
          const [],
      objective: json['objective'] is Map<String, dynamic>
          ? QuestNodeObjective.fromJson(
              json['objective'] as Map<String, dynamic>,
            )
          : null,
      exposition: json['exposition'] is Map<String, dynamic>
          ? Exposition.fromJson(json['exposition'] as Map<String, dynamic>)
          : json['exposition'] is Map
          ? Exposition.fromJson(
              Map<String, dynamic>.from(json['exposition'] as Map),
            )
          : null,
    );
  }
}
