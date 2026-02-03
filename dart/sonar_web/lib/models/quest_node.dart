import 'point_of_interest.dart';
import 'quest_node_challenge.dart';

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
  final String id;
  final int orderIndex;
  final PointOfInterest? pointOfInterest;
  final List<QuestNodePolygonPoint> polygon;
  final List<QuestNodeChallenge> challenges;

  const QuestNode({
    required this.id,
    required this.orderIndex,
    this.pointOfInterest,
    this.polygon = const [],
    this.challenges = const [],
  });

  factory QuestNode.fromJson(Map<String, dynamic> json) {
    return QuestNode(
      id: json['id'] as String? ?? '',
      orderIndex: (json['orderIndex'] as num?)?.toInt() ?? 0,
      pointOfInterest: json['pointOfInterest'] is Map<String, dynamic>
          ? PointOfInterest.fromJson(json['pointOfInterest'] as Map<String, dynamic>)
          : null,
      polygon: (json['polygon'] as List<dynamic>?)
              ?.map((e) => QuestNodePolygonPoint.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      challenges: (json['challenges'] as List<dynamic>?)
              ?.map((e) => QuestNodeChallenge.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );
  }
}
