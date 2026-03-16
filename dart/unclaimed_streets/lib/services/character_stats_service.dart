import '../models/character_stats.dart';
import '../constants/api_constants.dart';
import 'api_client.dart';

class CharacterStatsService {
  final ApiClient _api;

  CharacterStatsService(this._api);

  Future<CharacterStats?> getStats() async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/sonar/character-stats',
      );
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

  Future<Map<String, dynamic>> castSpell(
    String spellId, {
    String? targetUserId,
    String? targetMonsterId,
  }) async {
    final payload = <String, dynamic>{};
    if (targetUserId != null && targetUserId.trim().isNotEmpty) {
      payload['targetUserId'] = targetUserId.trim();
    }
    if (targetMonsterId != null && targetMonsterId.trim().isNotEmpty) {
      payload['targetMonsterId'] = targetMonsterId.trim();
    }
    final data = await _api.post<Map<String, dynamic>>(
      ApiConstants.castSpellEndpoint(spellId),
      data: payload,
    );
    return data;
  }

  Future<Map<String, dynamic>> castTechnique(
    String techniqueId, {
    String? targetUserId,
    String? targetMonsterId,
  }) async {
    final payload = <String, dynamic>{};
    if (targetUserId != null && targetUserId.trim().isNotEmpty) {
      payload['targetUserId'] = targetUserId.trim();
    }
    if (targetMonsterId != null && targetMonsterId.trim().isNotEmpty) {
      payload['targetMonsterId'] = targetMonsterId.trim();
    }
    final data = await _api.post<Map<String, dynamic>>(
      ApiConstants.castTechniqueEndpoint(techniqueId),
      data: payload,
    );
    return data;
  }

  Future<CharacterStats?> adjustUserResources(
    String userId, {
    int healthDelta = 0,
    int manaDelta = 0,
    int? health,
    int? mana,
  }) async {
    if (healthDelta == 0 && manaDelta == 0 && health == null && mana == null) {
      return getStats();
    }
    try {
      final payload = <String, dynamic>{};
      if (health != null) {
        payload['health'] = health;
      } else if (healthDelta != 0) {
        payload['healthDelta'] = healthDelta;
      }
      if (mana != null) {
        payload['mana'] = mana;
      } else if (manaDelta != 0) {
        payload['manaDelta'] = manaDelta;
      }
      final data = await _api.post<Map<String, dynamic>>(
        '/sonar/admin/users/$userId/resources',
        data: payload,
      );
      return CharacterStats.fromJson(data);
    } catch (_) {
      return null;
    }
  }
}
