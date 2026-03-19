import '../models/base_progression.dart';
import 'api_client.dart';

class BaseService {
  BaseService(this._api);

  final ApiClient _api;

  Future<List<BaseResourceBalanceData>> getResourceBalances() async {
    final data = await _api.get<Map<String, dynamic>>('/sonar/base/resources');
    final rawBalances = data['balances'];
    if (rawBalances is! List) {
      return const <BaseResourceBalanceData>[];
    }
    return rawBalances
        .whereType<Map>()
        .map(
          (entry) => BaseResourceBalanceData.fromJson(
            Map<String, dynamic>.from(entry),
          ),
        )
        .toList(growable: false);
  }

  Future<BaseProgressionSnapshot> getMyBase() async {
    final data = await _api.get<Map<String, dynamic>>('/sonar/base/me');
    return BaseProgressionSnapshot.fromJson(data);
  }

  Future<BaseProgressionSnapshot> getBaseById(String baseId) async {
    final data = await _api.get<Map<String, dynamic>>(
      '/sonar/bases/$baseId/progression',
    );
    return BaseProgressionSnapshot.fromJson(data);
  }

  Future<BaseProgressionSnapshot> updateMyBaseDetails({
    required String name,
    required String description,
  }) async {
    final data = await _api.put<Map<String, dynamic>>(
      '/sonar/base/me',
      data: <String, dynamic>{'name': name, 'description': description},
    );
    return BaseProgressionSnapshot.fromJson(data);
  }

  Future<List<BaseStructureDefinitionData>> getCatalog() async {
    final data = await _api.get<Map<String, dynamic>>('/sonar/base/catalog');
    final raw = data['structures'];
    if (raw is! List) {
      return const <BaseStructureDefinitionData>[];
    }
    return raw
        .whereType<Map>()
        .map(
          (e) => BaseStructureDefinitionData.fromJson(
            Map<String, dynamic>.from(e),
          ),
        )
        .toList();
  }

  Future<BaseProgressionSnapshot> buildStructure(
    String key, {
    required int gridX,
    required int gridY,
  }) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/base/structures/$key/build',
      data: <String, dynamic>{'gridX': gridX, 'gridY': gridY},
    );
    return BaseProgressionSnapshot.fromJson(data);
  }

  Future<BaseProgressionSnapshot> upgradeStructure(String key) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/base/structures/$key/upgrade',
    );
    return BaseProgressionSnapshot.fromJson(data);
  }

  Future<BaseProgressionSnapshot> destroyStructure(String key) async {
    final data = await _api.delete<Map<String, dynamic>>(
      '/sonar/base/structures/$key',
    );
    return BaseProgressionSnapshot.fromJson(data);
  }

  Future<BaseProgressionSnapshot> moveRooms({
    required String anchorStructureKey,
    required List<String> structureKeys,
    required int targetGridX,
    required int targetGridY,
  }) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/base/layout/move',
      data: <String, dynamic>{
        'anchorStructureKey': anchorStructureKey,
        'structureKeys': structureKeys,
        'targetGridX': targetGridX,
        'targetGridY': targetGridY,
      },
    );
    return BaseProgressionSnapshot.fromJson(data);
  }
}
