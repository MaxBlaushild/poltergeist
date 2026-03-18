import '../models/base_progression.dart';
import 'api_client.dart';

class BaseService {
  BaseService(this._api);

  final ApiClient _api;

  Future<BaseProgressionSnapshot> getMyBase() async {
    final data = await _api.get<Map<String, dynamic>>('/sonar/base/me');
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

  Future<BaseProgressionSnapshot> buildStructure(String key) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/base/structures/$key/build',
    );
    return BaseProgressionSnapshot.fromJson(data);
  }

  Future<BaseProgressionSnapshot> upgradeStructure(String key) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/base/structures/$key/upgrade',
    );
    return BaseProgressionSnapshot.fromJson(data);
  }
}
