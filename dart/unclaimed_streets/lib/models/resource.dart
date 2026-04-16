import 'inventory_item.dart';
import 'resource_type.dart';

class ResourceGatherRequirement {
  final String id;
  final String resourceId;
  final int minLevel;
  final int maxLevel;
  final int requiredInventoryItemId;
  final InventoryItem? requiredInventoryItem;

  const ResourceGatherRequirement({
    required this.id,
    required this.resourceId,
    required this.minLevel,
    required this.maxLevel,
    required this.requiredInventoryItemId,
    required this.requiredInventoryItem,
  });

  factory ResourceGatherRequirement.fromJson(Map<String, dynamic> json) {
    final rawRequiredItem = json['requiredInventoryItem'];
    return ResourceGatherRequirement(
      id: json['id']?.toString() ?? '',
      resourceId: json['resourceId']?.toString() ?? '',
      minLevel: (json['minLevel'] as num?)?.toInt() ?? 0,
      maxLevel: (json['maxLevel'] as num?)?.toInt() ?? 0,
      requiredInventoryItemId:
          (json['requiredInventoryItemId'] as num?)?.toInt() ?? 0,
      requiredInventoryItem: rawRequiredItem is Map<String, dynamic>
          ? InventoryItem.fromJson(rawRequiredItem)
          : null,
    );
  }
}

class ResourceNode {
  final String id;
  final String zoneId;
  final String resourceTypeId;
  final ResourceType? resourceType;
  final List<ResourceGatherRequirement> gatherRequirements;
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
    required this.gatherRequirements,
    required this.quantity,
    required this.latitude,
    required this.longitude,
    this.invalidated = false,
    this.gatheredByUser = false,
  });

  factory ResourceNode.fromJson(Map<String, dynamic> json) {
    final rawResourceType = json['resourceType'];
    final rawGatherRequirements = json['gatherRequirements'];
    return ResourceNode(
      id: json['id']?.toString() ?? '',
      zoneId: json['zoneId']?.toString() ?? '',
      resourceTypeId: json['resourceTypeId']?.toString() ?? '',
      resourceType: rawResourceType is Map<String, dynamic>
          ? ResourceType.fromJson(rawResourceType)
          : null,
      gatherRequirements:
          (rawGatherRequirements as List<dynamic>? ?? const <dynamic>[])
              .whereType<Map>()
              .map(
                (entry) => ResourceGatherRequirement.fromJson(
                  Map<String, dynamic>.from(entry),
                ),
              )
              .toList(),
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
      latitude: (json['latitude'] as num?)?.toDouble() ?? 0.0,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 0.0,
      invalidated: json['invalidated'] == true,
      gatheredByUser: json['gatheredByUser'] == true,
    );
  }
}
