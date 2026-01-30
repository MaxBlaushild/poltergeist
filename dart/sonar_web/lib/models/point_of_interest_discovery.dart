/// User/team discovery of a POI. See GET /sonar/pointsOfInterest/discoveries.
class PointOfInterestDiscovery {
  final String id;
  final String pointOfInterestId;
  final String? userId;
  final String? teamId;

  const PointOfInterestDiscovery({
    required this.id,
    required this.pointOfInterestId,
    this.userId,
    this.teamId,
  });

  factory PointOfInterestDiscovery.fromJson(Map<String, dynamic> json) {
    return PointOfInterestDiscovery(
      id: json['id']?.toString() ?? '',
      pointOfInterestId: json['pointOfInterestId']?.toString() ?? '',
      userId: json['userId']?.toString(),
      teamId: json['teamId']?.toString(),
    );
  }
}

/// True if any [discoveries] row matches [pointOfInterestId] and [entityId]
/// (userId or teamId). Use currentUser.id as entityId for single-player.
bool hasDiscoveredPointOfInterest(
  String pointOfInterestId,
  String entityId,
  List<PointOfInterestDiscovery> discoveries,
) {
  if (entityId.isEmpty) return false;
  return discoveries.any((d) {
    if (d.pointOfInterestId != pointOfInterestId) return false;
    return d.userId == entityId || d.teamId == entityId;
  });
}
