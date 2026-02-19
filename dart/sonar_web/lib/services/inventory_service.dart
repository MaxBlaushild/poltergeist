import '../models/inventory_item.dart';
import '../models/outfit_generation.dart';
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

  Future<OutfitGeneration?> getOutfitGenerationStatus(String ownedInventoryItemId) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/sonar/inventory/$ownedInventoryItemId/outfit-generation',
      );
      return OutfitGeneration.fromJson(data);
    } catch (_) {
      return null;
    }
  }

  Future<OutfitGeneration> useOutfitItem(
    String ownedInventoryItemId, {
    required String selfieUrl,
  }) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/inventory/$ownedInventoryItemId/use-outfit',
      data: {'selfieUrl': selfieUrl},
    );
    return OutfitGeneration.fromJson(data);
  }
}
