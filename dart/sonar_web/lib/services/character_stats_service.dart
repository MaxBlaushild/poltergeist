import '../models/character_stats.dart';
import 'api_client.dart';

class CharacterStatsService {
  final ApiClient _api;

  CharacterStatsService(this._api);

  Future<CharacterStats?> getStats() async {
    try {
      final data = await _api.get<Map<String, dynamic>>('/sonar/character-stats');
      return CharacterStats.fromJson(data);
    } catch (_) {
      return null;
    }
  }

  Future<CharacterStats?> applyAllocations(Map<String, int> allocations) async {
    if (allocations.isEmpty) return null;
    try {
      final data = await _api.put<Map<String, dynamic>>(
        '/sonar/character-stats/allocate',
        data: {'allocations': allocations},
      );
      return CharacterStats.fromJson(data);
    } catch (_) {
      return null;
    }
  }
}
