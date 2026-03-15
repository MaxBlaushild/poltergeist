import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

import '../models/character_stats.dart';
import '../models/spell.dart';
import '../services/character_stats_service.dart';
import 'activity_feed_provider.dart';
import 'auth_provider.dart';

class AbilityUseResult {
  const AbilityUseResult({this.error, this.response = const {}});

  final String? error;
  final Map<String, dynamic> response;

  bool get isSuccess => error == null;
}

class CharacterStatsProvider with ChangeNotifier {
  static const int baseStatValue = 10;
  static const List<String> statKeys = [
    'strength',
    'dexterity',
    'constitution',
    'intelligence',
    'wisdom',
    'charisma',
  ];

  final CharacterStatsService _service;

  ActivityFeedProvider? _feed;
  String? _userId;
  bool _loading = false;
  CharacterStats? _stats;
  bool _refreshing = false;
  Set<String> _knownLevelUpIds = {};

  CharacterStatsProvider(this._service);

  bool get loading => _loading;
  int get level => _stats?.level ?? 1;
  int get unspentPoints => _stats?.unspentPoints ?? 0;
  int get maxHealth =>
      _stats?.maxHealth ??
      CharacterStats.deriveHealthFromConstitution(
        _defaultStats()['constitution'] ?? baseStatValue,
      );
  int get maxMana =>
      _stats?.maxMana ??
      CharacterStats.deriveManaFromMentalStats(
        _defaultStats()['intelligence'] ?? baseStatValue,
        _defaultStats()['wisdom'] ?? baseStatValue,
      );
  int get health => _stats?.health ?? maxHealth;
  int get mana => _stats?.mana ?? maxMana;
  bool get hasUnspentPoints => unspentPoints > 0;
  Map<String, int> get baseStats => _stats?.toMap() ?? _defaultStats();
  Map<String, int> get equipmentBonuses =>
      _stats?.bonusMap() ?? _defaultBonuses();
  Map<String, int> get statusBonuses =>
      _stats?.statusBonusMap() ?? _defaultBonuses();
  Map<String, int> get stats => _stats?.effectiveMap() ?? _defaultStats();
  List<CharacterProficiency> get proficiencies =>
      _stats?.proficiencies ?? const [];
  List<CharacterStatus> get statuses => _stats?.statuses ?? const [];
  List<Spell> get abilities => _stats?.spells ?? const [];
  List<Spell> get spells => abilities
      .where((spell) => spell.abilityType.toLowerCase() != 'technique')
      .toList(growable: false);
  List<Spell> get techniques => abilities
      .where((spell) => spell.abilityType.toLowerCase() == 'technique')
      .toList(growable: false);
  bool get hasProficiencies => proficiencies.isNotEmpty;
  bool get hasStatuses => statuses.isNotEmpty;

  void updateAuth(AuthProvider auth) {
    final newUserId = auth.user?.id;
    if (newUserId == _userId) return;
    _userId = newUserId;
    _stats = null;
    _knownLevelUpIds = {};
    if (newUserId == null) {
      _loading = false;
      notifyListeners();
      return;
    }
    unawaited(refresh());
  }

  void updateActivityFeed(ActivityFeedProvider feed) {
    if (_feed == feed) return;
    _feed?.removeListener(_onFeedChanged);
    _feed = feed;
    _feed?.addListener(_onFeedChanged);
    _syncKnownLevelUps();
  }

  Future<void> refresh({bool force = false, bool silent = false}) async {
    if (_refreshing || _userId == null) return;
    _refreshing = true;
    if (!silent && !_loading) {
      _loading = true;
      notifyListeners();
    }

    try {
      final stats = await _service.getStats();
      if (stats != null) {
        _stats = stats;
        if (silent) {
          notifyListeners();
        }
      }
    } finally {
      if (!silent) {
        _loading = false;
      }
      _refreshing = false;
      if (!silent) {
        notifyListeners();
      }
    }
  }

  Future<bool> applyAllocations(Map<String, int> allocations) async {
    if (_userId == null) return false;
    final filtered = <String, int>{};
    for (final entry in allocations.entries) {
      if (!statKeys.contains(entry.key)) continue;
      if (entry.value <= 0) continue;
      filtered[entry.key] = entry.value;
    }
    if (filtered.isEmpty) return false;

    final stats = await _service.applyAllocations(filtered);
    if (stats == null) return false;
    _stats = stats;
    notifyListeners();
    return true;
  }

  Future<String?> castSpell(
    String spellId, {
    String? targetUserId,
    String? targetMonsterId,
  }) async {
    final result = await castSpellDetailed(
      spellId,
      targetUserId: targetUserId,
      targetMonsterId: targetMonsterId,
    );
    return result.error;
  }

  Future<AbilityUseResult> castSpellDetailed(
    String spellId, {
    String? targetUserId,
    String? targetMonsterId,
    bool refreshAfterCast = true,
  }) async {
    if (_userId == null) {
      return const AbilityUseResult(
        error: 'You must be logged in to cast spells.',
      );
    }
    try {
      final response = await _service.castSpell(
        spellId,
        targetUserId: targetUserId,
        targetMonsterId: targetMonsterId,
      );
      if (refreshAfterCast) {
        await refresh(silent: true);
      }
      return AbilityUseResult(response: response);
    } catch (e) {
      debugPrint(
        '[combat][castSpellDetailed] spellId=$spellId targetUserId=$targetUserId targetMonsterId=$targetMonsterId error=$e',
      );
      if (e is DioException) {
        debugPrint(
          '[combat][castSpellDetailed] status=${e.response?.statusCode} data=${e.response?.data}',
        );
        final data = e.response?.data;
        if (data is Map<String, dynamic>) {
          final message = data['error'];
          if (message is String && message.trim().isNotEmpty) {
            return AbilityUseResult(error: message.trim());
          }
        }
      }
      return const AbilityUseResult(error: 'Failed to cast spell.');
    }
  }

  Future<String?> castTechnique(
    String techniqueId, {
    String? targetUserId,
    String? targetMonsterId,
  }) async {
    final result = await castTechniqueDetailed(
      techniqueId,
      targetUserId: targetUserId,
      targetMonsterId: targetMonsterId,
    );
    return result.error;
  }

  Future<AbilityUseResult> castTechniqueDetailed(
    String techniqueId, {
    String? targetUserId,
    String? targetMonsterId,
    bool refreshAfterCast = true,
  }) async {
    if (_userId == null) {
      return const AbilityUseResult(
        error: 'You must be logged in to use techniques.',
      );
    }
    try {
      final response = await _service.castTechnique(
        techniqueId,
        targetUserId: targetUserId,
        targetMonsterId: targetMonsterId,
      );
      if (refreshAfterCast) {
        await refresh(silent: true);
      }
      return AbilityUseResult(response: response);
    } catch (e) {
      debugPrint(
        '[combat][castTechniqueDetailed] techniqueId=$techniqueId targetUserId=$targetUserId targetMonsterId=$targetMonsterId error=$e',
      );
      if (e is DioException) {
        debugPrint(
          '[combat][castTechniqueDetailed] status=${e.response?.statusCode} data=${e.response?.data}',
        );
        final data = e.response?.data;
        if (data is Map<String, dynamic>) {
          final message = data['error'];
          if (message is String && message.trim().isNotEmpty) {
            return AbilityUseResult(error: message.trim());
          }
        }
      }
      return const AbilityUseResult(error: 'Failed to use technique.');
    }
  }

  Future<bool> setHealthToOne() => setHealthTo(1);

  Future<bool> setHealthAndManaTo({
    required int health,
    required int mana,
    bool refreshBeforeAdjust = true,
  }) {
    return setCombatResources(
      health: health,
      mana: mana,
      refreshBeforeAdjust: refreshBeforeAdjust,
    );
  }

  Future<bool> setCombatResources({
    int? health,
    int? mana,
    bool refreshBeforeAdjust = true,
  }) async {
    if (_userId == null) return false;
    if (refreshBeforeAdjust || _stats == null) {
      await refresh(silent: true);
    }

    final current = _stats;
    if (current == null) return false;

    final targetHealth =
        (health ?? current.health).clamp(0, current.maxHealth).toInt();
    final targetMana = (mana ?? current.mana).clamp(0, current.maxMana).toInt();
    final healthDelta = targetHealth - current.health;
    final manaDelta = targetMana - current.mana;
    if (healthDelta == 0 && manaDelta == 0) return true;
    debugPrint(
      '[combat][setHealthAndManaTo] currentHealth=${current.health} '
      'targetHealth=$targetHealth currentMana=${current.mana} '
      'targetMana=$targetMana healthDelta=$healthDelta manaDelta=$manaDelta',
    );

    final updated = await _service.adjustUserResources(
      _userId!,
      healthDelta: healthDelta,
      manaDelta: manaDelta,
    );
    if (updated != null) {
      _stats = updated;
      notifyListeners();
      return true;
    }

    _stats = CharacterStats(
      strength: current.strength,
      dexterity: current.dexterity,
      constitution: current.constitution,
      intelligence: current.intelligence,
      wisdom: current.wisdom,
      charisma: current.charisma,
      health: targetHealth,
      maxHealth: current.maxHealth,
      mana: targetMana,
      maxMana: current.maxMana,
      equipmentBonuses: current.equipmentBonuses,
      statusBonuses: current.statusBonuses,
      unspentPoints: current.unspentPoints,
      level: current.level,
      proficiencies: current.proficiencies,
      statuses: current.statuses,
      spells: current.spells,
    );
    notifyListeners();
    return true;
  }

  Future<bool> setHealthTo(int value) async {
    final targetHealth = value < 1 ? 1 : value;
    return setHealthAndManaTo(health: targetHealth, mana: mana);
  }

  @override
  void dispose() {
    _feed?.removeListener(_onFeedChanged);
    super.dispose();
  }

  void _onFeedChanged() {
    if (_feed == null) return;
    final levelUps = _feed!.activities
        .where((activity) => activity.activityType == 'level_up')
        .toList();
    if (levelUps.isEmpty) return;
    var hasNew = false;
    for (final activity in levelUps) {
      if (_knownLevelUpIds.add(activity.id)) {
        hasNew = true;
      }
    }
    if (hasNew) {
      unawaited(refresh());
    }
  }

  void _syncKnownLevelUps() {
    final feed = _feed;
    if (feed == null) return;
    _knownLevelUpIds = feed.activities
        .where((activity) => activity.activityType == 'level_up')
        .map((activity) => activity.id)
        .toSet();
  }

  static Map<String, int> _defaultStats() => {
    for (final key in statKeys) key: baseStatValue,
  };

  static Map<String, int> _defaultBonuses() => {
    for (final key in statKeys) key: 0,
  };
}
