import '../models/character.dart';
import '../models/character_action.dart';
import '../models/challenge.dart';
import '../models/monster.dart';
import '../models/point_of_interest.dart';
import '../models/point_of_interest_discovery.dart';
import '../models/quest.dart';
import '../models/scenario.dart';
import '../models/treasure_chest.dart';
import '../models/user_zone_reputation.dart';
import '../models/zone.dart';
import 'api_client.dart';

class PoiService {
  final ApiClient _api;

  PoiService(this._api);

  Future<List<Character>> getCharacters() async {
    final list = await _api.get<List<dynamic>>('/sonar/characters');
    return list
        .map((e) => Character.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Quest?> getQuestById(String questId) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/sonar/quests/$questId',
      );
      return Quest.fromJson(data);
    } catch (_) {
      return null;
    }
  }

  Future<List<TreasureChest>> getTreasureChestsForZone(String zoneId) async {
    final list = await _api.get<List<dynamic>>(
      '/sonar/zones/$zoneId/treasure-chests',
    );
    return list
        .map((e) => TreasureChest.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<Scenario>> getScenariosForZone(String zoneId) async {
    final list = await _api.get<List<dynamic>>(
      '/sonar/zones/$zoneId/scenarios',
    );
    return list
        .map((e) => Scenario.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Scenario?> getScenarioById(String scenarioId) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/sonar/scenarios/$scenarioId',
      );
      return Scenario.fromJson(data);
    } catch (_) {
      return null;
    }
  }

  Future<List<Monster>> getMonstersForZone(String zoneId) async {
    final list = await _api.get<List<dynamic>>('/sonar/zones/$zoneId/monsters');
    return list
        .map((e) => Monster.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Monster?> getMonsterById(String monsterId) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/sonar/monsters/$monsterId',
      );
      return Monster.fromJson(data);
    } catch (_) {
      return null;
    }
  }

  Future<List<Challenge>> getChallengesForZone(String zoneId) async {
    final list = await _api.get<List<dynamic>>(
      '/sonar/zones/$zoneId/challenges',
    );
    return list
        .map((e) => Challenge.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Challenge?> getChallengeById(String challengeId) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/sonar/challenges/$challengeId',
      );
      return Challenge.fromJson(data);
    } catch (_) {
      return null;
    }
  }

  Future<List<Zone>> getZones() async {
    final list = await _api.get<List<dynamic>>('/sonar/zones');
    final zones = <Zone>[];
    for (var i = 0; i < list.length; i++) {
      try {
        zones.add(Zone.fromJson(list[i] as Map<String, dynamic>));
      } catch (_) {}
    }
    return zones;
  }

  Future<List<PointOfInterest>> getPointsOfInterest() async {
    final list = await _api.get<List<dynamic>>('/sonar/pointsOfInterest');
    return list
        .map((e) => PointOfInterest.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// GET /sonar/pointsOfInterest/discoveries — user's POI discoveries.
  Future<List<PointOfInterestDiscovery>> getDiscoveries() async {
    try {
      final list = await _api.get<List<dynamic>>(
        '/sonar/pointsOfInterest/discoveries',
      );
      return list
          .map(
            (e) => PointOfInterestDiscovery.fromJson(e as Map<String, dynamic>),
          )
          .toList();
    } catch (_) {
      return [];
    }
  }

  /// POST /sonar/pointsOfInterest/group/:id — create a POI under a group.
  Future<void> createPointOfInterestForGroup(
    String groupId, {
    required String name,
    required String description,
    required String imageUrl,
    required String latitude,
    required String longitude,
    required String clue,
    int? unlockTier,
  }) async {
    await _api.post<dynamic>(
      '/sonar/pointsOfInterest/group/$groupId',
      data: {
        'name': name,
        'description': description,
        'imageUrl': imageUrl,
        'latitude': latitude,
        'longitude': longitude,
        'clue': clue,
        if (unlockTier != null) 'unlockTier': unlockTier,
      },
    );
  }

  Future<List<CharacterAction>> getCharacterActions(String characterId) async {
    final list = await _api.get<List<dynamic>>(
      '/sonar/characters/$characterId/actions',
    );
    return list
        .map((e) => CharacterAction.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<void> acceptQuest({
    required String characterId,
    required String questId,
  }) async {
    await _api.post<dynamic>(
      '/sonar/quests/accept',
      data: {'characterId': characterId, 'questId': questId},
    );
  }

  Future<Map<String, dynamic>> purchaseFromShop(
    String actionId,
    int itemId,
  ) async {
    return await _api.post<Map<String, dynamic>>(
      '/sonar/character-actions/$actionId/purchase',
      data: {'itemId': itemId, 'quantity': 1},
    );
  }

  Future<Map<String, dynamic>> sellToShop(
    String actionId,
    int itemId, {
    int quantity = 1,
  }) async {
    return await _api.post<Map<String, dynamic>>(
      '/sonar/character-actions/$actionId/sell',
      data: {'itemId': itemId, 'quantity': quantity},
    );
  }

  Future<Map<String, dynamic>> openTreasureChest(String chestId) async {
    return await _api.post<Map<String, dynamic>>(
      '/sonar/treasure-chests/$chestId/open',
    );
  }

  Future<ScenarioPerformResult> performScenario(
    String scenarioId, {
    String? scenarioOptionId,
    String? responseText,
  }) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/scenarios/$scenarioId/perform',
      data: {
        if (scenarioOptionId != null && scenarioOptionId.isNotEmpty)
          'scenarioOptionId': scenarioOptionId,
        if (responseText != null && responseText.trim().isNotEmpty)
          'responseText': responseText.trim(),
      },
    );
    return ScenarioPerformResult.fromJson(data);
  }

  /// POST /sonar/pointOfInterest/unlock — unlock a POI when user is within 200m.
  /// [userId] must be set so the discovery is stored for the user and persists.
  Future<void> unlockPointOfInterest(
    String pointOfInterestId,
    double lat,
    double lng, {
    String? userId,
  }) async {
    await _api.post<dynamic>(
      '/sonar/pointOfInterest/unlock',
      data: {
        'pointOfInterestId': pointOfInterestId,
        'lat': lat.toString(),
        'lng': lng.toString(),
        if (userId != null && userId.isNotEmpty) 'userId': userId,
      },
    );
  }

  Future<UserZoneReputation?> getUserZoneReputation(String zoneId) async {
    try {
      final json = await _api.get<Map<String, dynamic>>(
        '/sonar/zones/$zoneId/reputation',
        skipAuthError: true,
      );
      return UserZoneReputation.fromJson(json);
    } catch (_) {
      return null;
    }
  }
}
