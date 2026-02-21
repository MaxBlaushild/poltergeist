import 'dart:async';

import 'package:flutter/foundation.dart';

import '../models/character_stats.dart';
import '../services/character_stats_service.dart';
import 'activity_feed_provider.dart';
import 'auth_provider.dart';

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
  bool get hasUnspentPoints => unspentPoints > 0;
  Map<String, int> get baseStats => _stats?.toMap() ?? _defaultStats();
  Map<String, int> get equipmentBonuses =>
      _stats?.bonusMap() ?? _defaultBonuses();
  Map<String, int> get stats => _stats?.effectiveMap() ?? _defaultStats();
  List<CharacterProficiency> get proficiencies =>
      _stats?.proficiencies ?? const [];
  bool get hasProficiencies => proficiencies.isNotEmpty;

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

  Future<void> refresh({bool force = false}) async {
    if (_refreshing || _userId == null) return;
    _refreshing = true;
    if (!_loading) {
      _loading = true;
      notifyListeners();
    }

    try {
      final stats = await _service.getStats();
      if (stats != null) {
        _stats = stats;
      }
    } finally {
      _loading = false;
      _refreshing = false;
      notifyListeners();
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
