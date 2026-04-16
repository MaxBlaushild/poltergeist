import 'resource_type.dart';

class ResourceNode {
  final String id;
  final String zoneId;
  final String resourceTypeId;
  final ResourceType? resourceType;
  final int quantity;
  final double latitude;
  final double longitude;
  final bool invalidated;
  final bool gatheredByUser;

  const ResourceNode({
    required this.id,
    required this.zoneId,
    required this.resourceTypeId,
    required this.resourceType,
    required this.quantity,
    required this.latitude,
    required this.longitude,
    this.invalidated = false,
    this.gatheredByUser = false,
  });

  factory ResourceNode.fromJson(Map<String, dynamic> json) {
    final rawResourceType = json['resourceType'];
    return ResourceNode(
      id: json['id']?.toString() ?? '',
      zoneId: json['zoneId']?.toString() ?? '',
      resourceTypeId: json['resourceTypeId']?.toString() ?? '',
      resourceType: rawResourceType is Map<String, dynamic>
          ? ResourceType.fromJson(rawResourceType)
          : null,
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      invalidated: json['invalidated'] == true,
      gatheredByUser: json['gatheredByUser'] == true,
    );
  }
}
