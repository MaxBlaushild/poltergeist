import 'package:dio/dio.dart';

import '../constants/api_constants.dart';
import '../models/character.dart';
import '../models/character_action.dart';
import '../models/challenge.dart';
import '../models/exposition.dart';
import '../models/monster.dart';
import '../models/base.dart';
import '../models/point_of_interest.dart';
import '../models/point_of_interest_discovery.dart';
import '../models/quest.dart';
import '../models/healing_fountain.dart';
import '../models/scenario.dart';
import '../models/treasure_chest.dart';
import '../models/tutorial.dart';
import '../models/user_zone_reputation.dart';
import '../models/zone.dart';
import 'api_client.dart';

class PartySubmissionStatus {
  const PartySubmissionStatus({
    required this.locked,
    this.status,
    this.submittedByUserId,
  });

  final bool locked;
  final String? status;
  final String? submittedByUserId;

  bool get isCompleted => (status ?? '').trim().toLowerCase() == 'completed';
  bool get isProcessing => (status ?? '').trim().toLowerCase() == 'processing';

  factory PartySubmissionStatus.fromJson(Map<String, dynamic> json) {
    return PartySubmissionStatus(
      locked: json['locked'] == true,
      status: json['status']?.toString(),
      submittedByUserId: json['submittedByUserId']?.toString(),
    );
  }
}

class PoiService {
  final ApiClient _api;

  PoiService(this._api);

  Future<List<Character>> getCharacters() async {
    final list = await _api.get<List<dynamic>>('/sonar/characters');
    return list
        .map((e) => Character.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<BasePin>> getVisibleBases() async {
    final list = await _api.get<List<dynamic>>('/sonar/bases');
    return list
        .map((e) => BasePin.fromJson(e as Map<String, dynamic>))
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

  Future<List<Exposition>> getExpositionsForZone(String zoneId) async {
    final list = await _api.get<List<dynamic>>(
      '/sonar/zones/$zoneId/expositions',
    );
    return list
        .map((e) => Exposition.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<HealingFountain>> getHealingFountainsForZone(
    String zoneId,
  ) async {
    final list = await _api.get<List<dynamic>>(
      '/sonar/zones/$zoneId/healing-fountains',
    );
    return list
        .map((e) => HealingFountain.fromJson(e as Map<String, dynamic>))
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

  Future<Exposition?> getExpositionById(String expositionId) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/sonar/expositions/$expositionId',
      );
      return Exposition.fromJson(data);
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

  Future<List<MonsterEncounter>> getMonsterEncountersForZone(
    String zoneId,
  ) async {
    final list = await _api.get<List<dynamic>>(
      '/sonar/zones/$zoneId/monster-encounters',
    );
    return list
        .map((e) => MonsterEncounter.fromJson(e as Map<String, dynamic>))
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

  Future<MonsterEncounter?> getMonsterEncounterById(String encounterId) async {
    try {
      final data = await _api.get<Map<String, dynamic>>(
        '/sonar/monster-encounters/$encounterId',
      );
      return MonsterEncounter.fromJson(data);
    } catch (_) {
      return null;
    }
  }

  Future<Map<String, dynamic>> spawnNearbyScenarioAndMonster() async {
    final raw = await _api.post<dynamic>(
      '/sonar/settings/spawn-nearby-content',
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
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

  Future<Map<String, dynamic>> submitChallenge(
    String challengeId, {
    String? textSubmission,
    String? imageSubmissionUrl,
    String? videoSubmissionUrl,
  }) async {
    final raw = await _api.post<dynamic>(
      '/sonar/challenges/$challengeId/submit',
      data: {
        if (textSubmission != null) 'textSubmission': textSubmission,
        if (imageSubmissionUrl != null)
          'imageSubmissionUrl': imageSubmissionUrl,
        if (videoSubmissionUrl != null)
          'videoSubmissionUrl': videoSubmissionUrl,
      },
    );
    final map = raw is Map
        ? Map<String, dynamic>.from(raw)
        : <String, dynamic>{};
    return map;
  }

  Future<Map<String, dynamic>> startMonsterBattle(
    String monsterId, {
    String? monsterEncounterId,
  }) async {
    final raw = await _api.post<dynamic>(
      ApiConstants.monsterBattleStartEndpoint(monsterId),
      data: {
        if (monsterEncounterId != null && monsterEncounterId.trim().isNotEmpty)
          'monsterEncounterId': monsterEncounterId.trim(),
      },
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }

  Future<Map<String, dynamic>> getMonsterBattleStatus(String monsterId) async {
    final raw = await _api.get<Map<String, dynamic>>(
      ApiConstants.monsterBattleStatusEndpoint(monsterId),
    );
    return raw;
  }

  Future<Map<String, dynamic>> endMonsterBattle(
    String monsterId, {
    String? outcome,
  }) async {
    final raw = await _api.post<dynamic>(
      ApiConstants.monsterBattleEndEndpoint(monsterId),
      data: outcome == null ? null : {'outcome': outcome},
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }

  Future<Map<String, dynamic>> applyMonsterBattleDamage(
    String monsterId,
    int damage, {
    Map<String, dynamic>? action,
  }) async {
    final raw = await _api.post<dynamic>(
      ApiConstants.monsterBattleDamageEndpoint(monsterId),
      data: {
        'damage': damage,
        if (action != null && action.isNotEmpty) 'action': action,
      },
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }

  Future<Map<String, dynamic>> advanceMonsterBattleTurn(
    String monsterId,
  ) async {
    final raw = await _api.post<dynamic>(
      ApiConstants.monsterBattleTurnEndpoint(monsterId),
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }

  Future<TutorialStatus?> getTutorialStatus() async {
    try {
      final raw = await _api.get<Map<String, dynamic>>(
        '/sonar/tutorial/status',
      );
      return TutorialStatus.fromJson(raw);
    } catch (_) {
      return null;
    }
  }

  Future<Scenario?> activateTutorial({bool force = false}) async {
    final raw = await _api.post<Map<String, dynamic>>(
      '/sonar/tutorial/activate',
      data: {'force': force},
    );
    return Scenario.fromJson(raw);
  }

  Future<TutorialStatus> resetTutorial() async {
    final raw = await _api.post<Map<String, dynamic>>('/sonar/tutorial/reset');
    return TutorialStatus.fromJson(raw);
  }

  Future<TutorialStatus?> advanceTutorial(String action) async {
    try {
      final raw = await _api.post<Map<String, dynamic>>(
        '/sonar/tutorial/advance',
        data: {'action': action},
      );
      return TutorialStatus.fromJson(raw);
    } catch (_) {
      return null;
    }
  }

  static String extractApiErrorMessage(Object error, String fallback) {
    if (error is DioException) {
      final responseData = error.response?.data;
      if (responseData is Map) {
        final message = responseData['error']?.toString().trim() ?? '';
        if (message.isNotEmpty) {
          return message;
        }
      }
      final message = error.message?.trim() ?? '';
      if (message.isNotEmpty) {
        return message;
      }
    }
    return fallback;
  }

  Future<Map<String, dynamic>> escapeMonsterBattle(String monsterId) async {
    final raw = await _api.post<dynamic>(
      ApiConstants.monsterBattleEscapeEndpoint(monsterId),
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }

  Future<Map<String, dynamic>> getMonsterBattleStatusById(
    String battleId,
  ) async {
    final raw = await _api.get<Map<String, dynamic>>(
      ApiConstants.monsterBattleStatusByIdEndpoint(battleId),
    );
    return raw;
  }

  Future<Map<String, dynamic>> applyMonsterBattleDamageById(
    String battleId,
    int damage, {
    Map<String, dynamic>? action,
  }) async {
    final raw = await _api.post<dynamic>(
      ApiConstants.monsterBattleDamageByIdEndpoint(battleId),
      data: {
        'damage': damage,
        if (action != null && action.isNotEmpty) 'action': action,
      },
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }

  Future<Map<String, dynamic>> getUserCharacterProfile(String userId) async {
    final raw = await _api.get<Map<String, dynamic>>(
      ApiConstants.userCharacterEndpoint(userId),
    );
    return raw;
  }

  Future<PartySubmissionStatus> getPartySubmissionStatus({
    required String contentType,
    required String contentId,
  }) async {
    final raw = await _api.get<Map<String, dynamic>>(
      ApiConstants.partySubmissionStatusEndpoint,
      params: {'contentType': contentType, 'contentId': contentId},
    );
    return PartySubmissionStatus.fromJson(raw);
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

  Future<Map<String, dynamic>> openTreasureChest(
    String chestId, {
    String? unlockMethod,
    String? ownedInventoryItemId,
    String? spellId,
  }) async {
    return await _api.post<Map<String, dynamic>>(
      '/sonar/treasure-chests/$chestId/open',
      data: {
        if (unlockMethod != null && unlockMethod.trim().isNotEmpty)
          'unlockMethod': unlockMethod.trim(),
        if (ownedInventoryItemId != null &&
            ownedInventoryItemId.trim().isNotEmpty)
          'ownedInventoryItemId': ownedInventoryItemId.trim(),
        if (spellId != null && spellId.trim().isNotEmpty)
          'spellId': spellId.trim(),
      },
    );
  }

  Future<Map<String, dynamic>> useHealingFountain(String fountainId) async {
    final raw = await _api.post<dynamic>(
      '/sonar/healing-fountains/$fountainId/use',
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }

  Future<Map<String, dynamic>> unlockHealingFountain(String fountainId) async {
    final raw = await _api.post<dynamic>(
      '/sonar/healing-fountains/$fountainId/unlock',
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
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

  Future<ExpositionPerformResult> performExposition(String expositionId) async {
    final data = await _api.post<Map<String, dynamic>>(
      '/sonar/expositions/$expositionId/perform',
    );
    return ExpositionPerformResult.fromJson(data);
  }

  Future<Map<String, dynamic>> chooseScenarioRewardItem(
    String scenarioId, {
    required int inventoryItemId,
  }) async {
    final raw = await _api.post<dynamic>(
      ApiConstants.chooseScenarioRewardItemEndpoint(scenarioId),
      data: {'inventoryItemId': inventoryItemId},
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }

  Future<Map<String, dynamic>> chooseChallengeRewardItem(
    String challengeId, {
    required int inventoryItemId,
  }) async {
    final raw = await _api.post<dynamic>(
      ApiConstants.chooseChallengeRewardItemEndpoint(challengeId),
      data: {'inventoryItemId': inventoryItemId},
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
  }

  /// POST /sonar/pointOfInterest/unlock — unlock a POI when user is within 200m.
  /// [userId] must be set so the discovery is stored for the user and persists.
  Future<Map<String, dynamic>> unlockPointOfInterest(
    String pointOfInterestId,
    double lat,
    double lng, {
    String? userId,
  }) async {
    final raw = await _api.post<dynamic>(
      '/sonar/pointOfInterest/unlock',
      data: {
        'pointOfInterestId': pointOfInterestId,
        'lat': lat.toString(),
        'lng': lng.toString(),
        if (userId != null && userId.isNotEmpty) 'userId': userId,
      },
    );
    return raw is Map ? Map<String, dynamic>.from(raw) : <String, dynamic>{};
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
