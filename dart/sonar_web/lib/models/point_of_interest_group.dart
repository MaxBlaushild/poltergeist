import 'point_of_interest.dart';

class PointOfInterestGroup {
  final String id;
  final String name;
  final String description;
  final String? imageUrl;
  final List<PointOfInterest> pointsOfInterest;

  const PointOfInterestGroup({
    required this.id,
    required this.name,
    required this.description,
    this.imageUrl,
    this.pointsOfInterest = const [],
  });

  factory PointOfInterestGroup.fromJson(Map<String, dynamic> json) {
    return PointOfInterestGroup(
      id: json['id'] as String,
      name: json['name'] as String? ?? '',
      description: json['description'] as String? ?? '',
      imageUrl: json['imageUrl'] as String?,
      pointsOfInterest: (json['pointsOfInterest'] as List<dynamic>?)
              ?.map((e) => PointOfInterest.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}
