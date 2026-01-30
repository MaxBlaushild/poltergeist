import '../models/inventory_item.dart';
import 'api_client.dart';

class InventoryService {
  final ApiClient _api;

  InventoryService(this._api);

  Future<List<InventoryItem>> getInventoryItems() async {
    try {
      final list = await _api.get<List<dynamic>>('/sonar/inventory-items');
      return list
          .map((e) => InventoryItem.fromJson(e as Map<String, dynamic>))
          .toList();
    } catch (_) {
      return [];
    }
  }

  Future<List<OwnedInventoryItem>> getOwnedInventoryItems() async {
    try {
      final list = await _api.get<List<dynamic>>('/sonar/ownedInventoryItems');
      return list
          .map((e) => OwnedInventoryItem.fromJson(e as Map<String, dynamic>))
          .toList();
    } catch (_) {
      return [];
    }
  }

  /// POST /sonar/inventory/:ownedInventoryItemID/use
  /// Optional [targetTeamId] for items that require a team target.
  Future<void> useItem(
    String ownedInventoryItemId, {
    String? targetTeamId,
  }) async {
    await _api.post<dynamic>(
      '/sonar/inventory/$ownedInventoryItemId/use',
      data: {
        if (targetTeamId != null && targetTeamId.isNotEmpty) 'targetTeamId': targetTeamId,
      },
    );
  }
}
