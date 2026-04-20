import 'dart:async';
import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/character_stats.dart';
import '../models/equipment_item.dart';
import '../models/inventory_item.dart';
import '../models/monster.dart';
import '../models/spell.dart';
import '../models/user.dart';
import '../providers/auth_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/party_provider.dart';
import '../services/inventory_service.dart';
import '../services/poi_service.dart';
import '../utils/hand_attack_profile.dart';
import 'paper_texture.dart';

enum MonsterBattleOutcome { victory, defeat, escaped }

enum _BattleMenuView { root, spells, techniques, items }

class MonsterBattleResult {
  const MonsterBattleResult({
    required this.outcome,
    required this.playerHealthRemaining,
    required this.playerManaRemaining,
    this.rewardExperience = 0,
    this.rewardGold = 0,
    this.baseResourcesAwarded = const <Map<String, dynamic>>[],
    this.itemsAwarded = const <Map<String, dynamic>>[],
  });

  final MonsterBattleOutcome outcome;
  final int playerHealthRemaining;
  final int playerManaRemaining;
  final int rewardExperience;
  final int rewardGold;
  final List<Map<String, dynamic>> baseResourcesAwarded;
  final List<Map<String, dynamic>> itemsAwarded;
}

class _BattleItemChoice {
  const _BattleItemChoice({
    required this.ownedInventoryItemId,
    required this.name,
    required this.healthDelta,
    required this.manaDelta,
    required this.revivePartyMemberHealth,
    required this.reviveAllDownedPartyMembersHealth,
    required this.dealDamage,
    required this.dealDamageHits,
    required this.dealDamageAllEnemies,
    required this.dealDamageAllEnemiesHits,
    required this.quantity,
  });

  final String ownedInventoryItemId;
  final String name;
  final int healthDelta;
  final int manaDelta;
  final int revivePartyMemberHealth;
  final int reviveAllDownedPartyMembersHealth;
  final int dealDamage;
  final int dealDamageHits;
  final int dealDamageAllEnemies;
  final int dealDamageAllEnemiesHits;
  final int quantity;

  _BattleItemChoice copyWith({int? quantity}) => _BattleItemChoice(
    ownedInventoryItemId: ownedInventoryItemId,
    name: name,
    healthDelta: healthDelta,
    manaDelta: manaDelta,
    revivePartyMemberHealth: revivePartyMemberHealth,
    reviveAllDownedPartyMembersHealth: reviveAllDownedPartyMembersHealth,
    dealDamage: dealDamage,
    dealDamageHits: dealDamageHits,
    dealDamageAllEnemies: dealDamageAllEnemies,
    dealDamageAllEnemiesHits: dealDamageAllEnemiesHits,
    quantity: quantity ?? this.quantity,
  );
}

class _AbilitySection {
  const _AbilitySection({required this.title, required this.abilities});

  final String title;
  final List<Spell> abilities;
}

class _EncounterEnemyState {
  _EncounterEnemyState({
    required this.monster,
    required this.currentHealth,
    int? currentMana,
    List<MonsterStatus>? statuses,
    Map<String, DateTime>? cooldownExpiresAtByAbilityId,
  }) : currentMana = currentMana ?? math.max(0, monster.mana),
       statuses = statuses ?? List<MonsterStatus>.from(monster.statuses),
       cooldownExpiresAtByAbilityId =
           cooldownExpiresAtByAbilityId ?? <String, DateTime>{};

  final Monster monster;
  int currentHealth;
  int currentMana;
  List<MonsterStatus> statuses;
  final Map<String, DateTime> cooldownExpiresAtByAbilityId;

  int get maxHealth => math.max(1, monster.maxHealth);
  int get maxMana => math.max(0, monster.maxMana);
  bool get isDefeated => currentHealth <= 0;
}

class _CombatStatusVisual {
  const _CombatStatusVisual({
    required this.name,
    required this.description,
    required this.effect,
    required this.positive,
    required this.effectType,
    required this.damagePerTick,
    required this.healthPerTick,
    required this.manaPerTick,
  });

  final String name;
  final String description;
  final String effect;
  final bool positive;
  final String effectType;
  final int damagePerTick;
  final int healthPerTick;
  final int manaPerTick;

  factory _CombatStatusVisual.fromCharacterStatus(CharacterStatus status) {
    return _CombatStatusVisual(
      name: status.name,
      description: status.description,
      effect: status.effect,
      positive: status.positive,
      effectType: status.effectType,
      damagePerTick: status.damagePerTick,
      healthPerTick: status.healthPerTick,
      manaPerTick: status.manaPerTick,
    );
  }

  factory _CombatStatusVisual.fromMonsterStatus(MonsterStatus status) {
    return _CombatStatusVisual(
      name: status.name,
      description: status.description,
      effect: status.effect,
      positive: status.positive,
      effectType: status.effectType,
      damagePerTick: status.damagePerTick,
      healthPerTick: status.healthPerTick,
      manaPerTick: 0,
    );
  }
}

class _TurnOrderEntry {
  const _TurnOrderEntry({
    required this.iconUrl,
    required this.currentHealth,
    required this.maxHealth,
    required this.label,
    required this.fallbackIcon,
    this.userId,
    this.allyIndex,
    this.enemyIndex,
    this.isSelf = false,
  });

  final String iconUrl;
  final int currentHealth;
  final int maxHealth;
  final String label;
  final IconData fallbackIcon;
  final String? userId;
  final int? allyIndex;
  final int? enemyIndex;
  final bool isSelf;

  bool get isEnemy => enemyIndex != null;
  bool get isAlly => allyIndex != null;
  bool get isDefeated => currentHealth <= 0;
}

class _PartyAllyState {
  _PartyAllyState({
    required this.userId,
    required this.name,
    required this.iconUrl,
    required this.level,
    required this.currentHealth,
    required this.maxHealth,
    required this.currentMana,
    required this.maxMana,
    required this.isSelf,
  });

  final String userId;
  String name;
  String iconUrl;
  int level;
  int currentHealth;
  int maxHealth;
  int currentMana;
  int maxMana;
  final bool isSelf;
}

class MonsterBattleDialog extends StatefulWidget {
  const MonsterBattleDialog({
    super.key,
    required this.encounter,
    this.isPartyBattle = false,
    this.battleMonsterId,
    this.battleId,
  });

  final MonsterEncounter encounter;
  final bool isPartyBattle;
  final String? battleMonsterId;
  final String? battleId;

  @override
  State<MonsterBattleDialog> createState() => _MonsterBattleDialogState();
}

class _MonsterBattleDialogState extends State<MonsterBattleDialog> {
  static const Set<String> _handEquipmentSlots = {'dominant_hand', 'off_hand'};
  static const Duration _combatTurnDuration = Duration(seconds: 150);
  static const double _defaultBattleLogHeight = 138;
  static const double _compactBattleLogHeight = 54;
  static const double _rootCommandPanelHeight = 158;
  static const int _maxRecentAbilities = 6;
  static const String _favoriteAbilityPrefsKeyPrefix =
      'monster_battle.favorite_ability_ids';
  static const String _recentAbilityPrefsKeyPrefix =
      'monster_battle.recent_ability_ids';
  final math.Random _random = math.Random();
  final List<String> _battleLog = <String>[];

  late final String _playerName;
  late final int _playerLevel;
  late final Map<String, int> _playerStats;
  List<Spell> _spells = const [];
  List<Spell> _techniques = const [];
  late int _playerMaxHealth;
  late int _playerMaxMana;
  late int _playerHealth;
  late int _playerMana;
  List<CharacterStatus> _playerStatuses = const [];
  late final List<_EncounterEnemyState> _enemies;
  int _activeEnemyIndex = 0;
  int? _actingEnemyIndex;
  late final String _playerFrontSpriteUrl;
  late final String _playerBackSpriteUrl;
  late final bool _hasTrueBackSprite;
  List<_BattleItemChoice> _items = const [];
  List<EquippedItem> _equippedHandItems = const [];
  bool _equipmentLoaded = false;

  _BattleMenuView _menuView = _BattleMenuView.root;
  String? _selectedCommandKey;
  bool _loadingItems = false;
  bool _playerTurn = true;
  bool _busy = false;
  bool _battleOver = false;
  bool _endingBattle = false;
  double _monsterShakeDx = 0;
  double _playerShakeDx = 0;
  Color? _monsterFlashTint;
  Color? _playerFlashTint;
  String? _monsterFloatText;
  String? _playerFloatText;
  Color _monsterFloatColor = const Color(0xFF8C2F39);
  Color _playerFloatColor = const Color(0xFF8C2F39);
  int _monsterFxNonce = 0;
  int _playerFxNonce = 0;
  Timer? _partyBattleSyncTimer;
  String? _partyBattleMonsterId;
  String? _partyBattleId;
  bool _partyBattleSyncInFlight = false;
  String _selfUserId = '';
  int _activeAllyIndex = 0;
  final List<_PartyAllyState> _partyAllies = <_PartyAllyState>[];
  final Set<String> _activePartyParticipantIds = <String>{};
  final Map<String, DateTime> _partyAllyFetchedAt = <String, DateTime>{};
  List<_TurnOrderEntry> _partyTurnOrder = const [];
  int _partyTurnIndex = 0;
  bool _partyMonsterTurnInFlight = false;
  int _lastPartyMonsterHealthDeficit = -1;
  int _lastSeenPartyActionSequence = -1;
  int _pendingLocalDamage = 0;
  int _victoryRewardExperience = 0;
  int _victoryRewardGold = 0;
  bool _hasCachedVictoryRewards = false;
  List<Map<String, dynamic>> _victoryBaseResourcesAwarded =
      const <Map<String, dynamic>>[];
  List<Map<String, dynamic>> _victoryItemsAwarded =
      const <Map<String, dynamic>>[];
  bool _partySelfResourceSyncInFlight = false;
  bool _partySelfHealthIncreaseSyncAllowed = false;
  bool _pendingPartySelfHealthSync = false;
  bool _pendingPartySelfManaSync = false;
  final Map<String, DateTime> _techniqueCooldownExpiresAtById =
      <String, DateTime>{};
  final Set<String> _techniqueCooldownClearedById = <String>{};
  final Set<String> _favoriteAbilityIds = <String>{};
  List<String> _recentAbilityIds = <String>[];
  bool _battleLogCollapsed = true;

  bool _isBattleStatusNotFoundError(Object error) {
    if (error is DioException) {
      final code = error.response?.statusCode;
      return code == 404 || code == 410;
    }
    return false;
  }

  String _extractApiErrorMessage(Object error, {required String fallback}) {
    if (error is DioException) {
      final data = error.response?.data;
      if (data is Map) {
        final apiMessage = data['error']?.toString().trim() ?? '';
        if (apiMessage.isNotEmpty) {
          return apiMessage;
        }
      }
      final dioMessage = error.message?.trim() ?? '';
      if (dioMessage.isNotEmpty) {
        return dioMessage;
      }
    }
    final rawMessage = error.toString().trim();
    if (rawMessage.isNotEmpty &&
        rawMessage != 'Exception' &&
        rawMessage != 'Error') {
      return rawMessage;
    }
    return fallback;
  }

  Future<void> _finishAfterBattleStatusGone() async {
    if (_battleOver) return;
    if (_aliveEnemies.isEmpty) {
      await _finishBattle(
        MonsterBattleOutcome.victory,
        _enemies.length == 1
            ? 'Your party defeated ${_enemies.first.monster.name}!'
            : 'Your party won the battle.',
      );
      return;
    }
    if (_allPartyMembersDefeated()) {
      await _finishBattle(
        MonsterBattleOutcome.defeat,
        'Your party has been defeated.',
      );
      return;
    }
    if (widget.isPartyBattle) {
      await _finishBattle(
        MonsterBattleOutcome.victory,
        _enemies.length == 1
            ? 'Your party defeated ${_enemies.first.monster.name}!'
            : 'Your party won the battle.',
      );
      return;
    }
    await _finishBattle(MonsterBattleOutcome.escaped, 'Party battle ended.');
  }

  @override
  void initState() {
    super.initState();
    final statsProvider = context.read<CharacterStatsProvider>();
    final authProvider = context.read<AuthProvider>();
    final user = authProvider.user;
    _selfUserId = (user?.id ?? '').trim();
    _playerName = user?.username.trim().isNotEmpty == true
        ? '@${user!.username.trim()}'
        : (user?.name.trim().isNotEmpty == true ? user!.name.trim() : 'You');
    _playerLevel = statsProvider.level;
    _playerStats = statsProvider.stats;
    _spells = statsProvider.spells;
    _techniques = statsProvider.techniques;
    _syncTechniqueCooldownsFromAbilities();
    _playerMaxHealth = math.max(1, statsProvider.maxHealth);
    _playerMaxMana = math.max(0, statsProvider.maxMana);
    _playerHealth = math.max(0, statsProvider.health);
    _playerMana = math.max(0, statsProvider.mana);
    _playerStatuses = List<CharacterStatus>.from(statsProvider.statuses);
    final encounterMonsters = widget.encounter.monsters.isNotEmpty
        ? widget.encounter.monsters
        : widget.encounter.members
              .map((member) => member.monster)
              .where((monster) => monster.id.isNotEmpty)
              .toList(growable: false);
    _enemies = encounterMonsters
        .take(9)
        .map(
          (monster) => _EncounterEnemyState(
            monster: monster,
            currentHealth: math.max(0, monster.health),
            currentMana: math.max(0, monster.mana),
          ),
        )
        .toList(growable: false);
    if (_enemies.isEmpty) {
      _enemies.add(
        _EncounterEnemyState(
          monster: const Monster(
            id: '',
            name: 'Unknown Monster',
            zoneId: '',
            latitude: 0,
            longitude: 0,
          ),
          currentHealth: 1,
          currentMana: 0,
        ),
      );
    }
    _activeEnemyIndex = 0;
    _ensureSelectedEnemyIsAlive();
    _actingEnemyIndex = null;
    _playerFrontSpriteUrl = (user?.profilePictureUrl ?? '').trim();
    _playerBackSpriteUrl = (user?.backProfilePictureUrl ?? '').trim();
    _hasTrueBackSprite = _playerBackSpriteUrl.isNotEmpty;
    _selectedCommandKey = 'root:Attack';
    if (_enemies.length == 1) {
      _battleLog.add(
        '${widget.encounter.encounterTypeLabel}: ${_enemies.first.monster.name} appears!',
      );
    } else {
      _battleLog.add(
        '${widget.encounter.encounterTypeLabel}: a hostile group of ${_enemies.length} monsters appears!',
      );
    }
    _battleLog.add('Choose a command.');
    if (widget.isPartyBattle) {
      _playerTurn = false;
      _ensureSelfPartyAlly();
      _seedPartyAlliesFromPartyProvider();
      final battleId = (widget.battleId ?? '').trim();
      if (battleId.isNotEmpty) {
        _partyBattleId = battleId;
      }
      final monsterId = (widget.battleMonsterId ?? _enemies.first.monster.id)
          .trim();
      if (monsterId.isNotEmpty) {
        _partyBattleMonsterId = monsterId;
        _partyBattleSyncTimer = Timer.periodic(const Duration(seconds: 1), (_) {
          unawaited(_syncPartyBattleState());
        });
        unawaited(_syncPartyBattleState());
      }
    }
    if (!widget.isPartyBattle && _playerTurn) {
      _beginSelfTurn();
    }
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      unawaited(_refreshBattleAbilities());
    });
    unawaited(_ensureEquipmentLoaded());
    unawaited(_loadItemChoices());
    unawaited(_loadAbilityPickerPrefs());
  }

  @override
  void dispose() {
    _partyBattleSyncTimer?.cancel();
    _partyBattleSyncTimer = null;
    super.dispose();
  }

  bool get _canAct =>
      _playerTurn && !_busy && !_battleOver && _playerHealth > 0;

  Future<void> _refreshBattleAbilities() async {
    final statsProvider = context.read<CharacterStatsProvider>();
    await statsProvider.refresh(silent: true);
    if (!mounted) return;
    setState(() {
      _spells = statsProvider.spells;
      _techniques = statsProvider.techniques;
      _playerStatuses = List<CharacterStatus>.from(statsProvider.statuses);
      _syncTechniqueCooldownsFromAbilities();
    });
  }

  String _humanizeToken(String value) {
    final normalized = value.trim().replaceAll('_', ' ');
    if (normalized.isEmpty) return '';
    return normalized
        .split(RegExp(r'\s+'))
        .where((part) => part.isNotEmpty)
        .map(
          (part) =>
              '${part[0].toUpperCase()}${part.substring(1).toLowerCase()}',
        )
        .join(' ');
  }

  String _statusDisplayName(String name, String effectType) {
    final trimmed = name.trim();
    if (trimmed.isNotEmpty) return trimmed;
    return _humanizeToken(effectType);
  }

  int _playerStatusHealthTick(CharacterStatus status) {
    final effectType = status.effectType.trim().toLowerCase();
    if (effectType == 'damage_over_time') {
      return status.damagePerTick > 0 ? -status.damagePerTick : 0;
    }
    if (effectType == 'health_over_time') {
      return status.healthPerTick;
    }
    return 0;
  }

  int _playerStatusManaTick(CharacterStatus status) {
    final effectType = status.effectType.trim().toLowerCase();
    if (effectType == 'mana_over_time') {
      return status.manaPerTick;
    }
    return 0;
  }

  int _monsterStatusHealthTick(MonsterStatus status) {
    final effectType = status.effectType.trim().toLowerCase();
    if (effectType == 'damage_over_time') {
      return status.damagePerTick > 0 ? -status.damagePerTick : 0;
    }
    if (effectType == 'health_over_time') {
      return status.healthPerTick;
    }
    return 0;
  }

  List<MonsterStatus> _parseMonsterStatuses(dynamic rawStatuses) {
    final statuses = <MonsterStatus>[];
    if (rawStatuses is! List) return statuses;
    for (final raw in rawStatuses) {
      if (raw is Map<String, dynamic>) {
        statuses.add(MonsterStatus.fromJson(raw));
      } else if (raw is Map) {
        statuses.add(MonsterStatus.fromJson(Map<String, dynamic>.from(raw)));
      }
    }
    return statuses;
  }

  List<String> _parseRemovedStatusNames(dynamic rawStatuses) {
    if (rawStatuses is! List) return const [];
    return rawStatuses
        .map((status) => status?.toString().trim() ?? '')
        .where((status) => status.isNotEmpty)
        .toList(growable: false);
  }

  void _pruneExpiredStatuses() {
    final now = DateTime.now();
    _playerStatuses = _playerStatuses.where((status) {
      final expiresAt = status.expiresAt;
      return expiresAt == null || expiresAt.isAfter(now);
    }).toList();
    for (final enemy in _enemies) {
      enemy.statuses = enemy.statuses.where((status) {
        final expiresAt = status.expiresAt;
        return expiresAt == null || expiresAt.isAfter(now);
      }).toList();
    }
  }

  void _syncEnemyStatusesFromPayload(Map<String, dynamic>? payload) {
    if (payload == null || payload.isEmpty) return;

    void syncFromMap(Map<String, dynamic> source) {
      final rawMonster = source['monster'];
      if (rawMonster is Map<String, dynamic>) {
        final monster = Monster.fromJson(rawMonster);
        final index = _enemies.indexWhere(
          (enemy) => enemy.monster.id == monster.id,
        );
        if (index >= 0) {
          _enemies[index].statuses = List<MonsterStatus>.from(monster.statuses);
        }
      } else if (rawMonster is Map) {
        final monster = Monster.fromJson(Map<String, dynamic>.from(rawMonster));
        final index = _enemies.indexWhere(
          (enemy) => enemy.monster.id == monster.id,
        );
        if (index >= 0) {
          _enemies[index].statuses = List<MonsterStatus>.from(monster.statuses);
        }
      }

      final rawMonsters = source['monsters'];
      if (rawMonsters is List) {
        for (final raw in rawMonsters) {
          if (raw is Map<String, dynamic>) {
            final monster = Monster.fromJson(raw);
            final index = _enemies.indexWhere(
              (enemy) => enemy.monster.id == monster.id,
            );
            if (index >= 0) {
              _enemies[index].statuses = List<MonsterStatus>.from(
                monster.statuses,
              );
            }
          } else if (raw is Map) {
            final monster = Monster.fromJson(Map<String, dynamic>.from(raw));
            final index = _enemies.indexWhere(
              (enemy) => enemy.monster.id == monster.id,
            );
            if (index >= 0) {
              _enemies[index].statuses = List<MonsterStatus>.from(
                monster.statuses,
              );
            }
          }
        }
      }

      final rawMembers = source['members'];
      if (rawMembers is List) {
        for (final raw in rawMembers) {
          final mapped = raw is Map<String, dynamic>
              ? raw
              : (raw is Map ? Map<String, dynamic>.from(raw) : null);
          if (mapped == null) continue;
          final rawMemberMonster = mapped['monster'];
          if (rawMemberMonster is Map<String, dynamic>) {
            final monster = Monster.fromJson(rawMemberMonster);
            final index = _enemies.indexWhere(
              (enemy) => enemy.monster.id == monster.id,
            );
            if (index >= 0) {
              _enemies[index].statuses = List<MonsterStatus>.from(
                monster.statuses,
              );
            }
          } else if (rawMemberMonster is Map) {
            final monster = Monster.fromJson(
              Map<String, dynamic>.from(rawMemberMonster),
            );
            final index = _enemies.indexWhere(
              (enemy) => enemy.monster.id == monster.id,
            );
            if (index >= 0) {
              _enemies[index].statuses = List<MonsterStatus>.from(
                monster.statuses,
              );
            }
          }
        }
      }
    }

    syncFromMap(payload);
    final nested = payload['battleDetail'];
    if (nested is Map<String, dynamic>) {
      syncFromMap(nested);
    } else if (nested is Map) {
      syncFromMap(Map<String, dynamic>.from(nested));
    }
  }

  void _updatePartyTurnOrderFromPayload(Map<String, dynamic>? payload) {
    if (!widget.isPartyBattle || payload == null || payload.isEmpty) return;
    final detailRaw = payload['battleDetail'];
    if (detailRaw is Map<String, dynamic>) {
      _updatePartyTurnOrderFromStatus(detailRaw);
      return;
    }
    if (detailRaw is Map) {
      _updatePartyTurnOrderFromStatus(Map<String, dynamic>.from(detailRaw));
      return;
    }
    _updatePartyTurnOrderFromStatus(payload);
  }

  void _syncPartyMonsterHealthFromPayload(Map<String, dynamic>? payload) {
    if (!widget.isPartyBattle || payload == null || payload.isEmpty) return;
    final detailRaw = payload['battleDetail'];
    final detail = detailRaw is Map<String, dynamic>
        ? detailRaw
        : (detailRaw is Map ? Map<String, dynamic>.from(detailRaw) : payload);
    final battleRaw = detail['battle'];
    final battle = battleRaw is Map<String, dynamic>
        ? battleRaw
        : (battleRaw is Map ? Map<String, dynamic>.from(battleRaw) : null);
    if (battle == null || _enemies.isEmpty) return;
    final deficit = _parseIntValue(
      battle['monsterHealthDeficit'],
      fallback: _lastPartyMonsterHealthDeficit,
    );
    if (deficit < 0) return;
    _lastPartyMonsterHealthDeficit = deficit;
    final primaryEnemy = _enemies.first;
    primaryEnemy.currentHealth = math.max(0, primaryEnemy.maxHealth - deficit);
    _ensureSelectedEnemyIsAlive();
  }

  void _applyMonsterStatusChanges(
    Map<String, dynamic> response, {
    int? targetEnemyIndex,
  }) {
    final appliedRaw = response['monsterStatusesApplied'];
    final applied = _parseMonsterStatuses(appliedRaw);
    final removedNames = _parseRemovedStatusNames(
      response['monsterStatusesRemoved'],
    );
    if (applied.isEmpty && removedNames.isEmpty) return;

    final resolvedTargetIndex = _resolveTargetEnemyIndex(targetEnemyIndex);
    if (resolvedTargetIndex == null) return;

    final enemy = _enemies[resolvedTargetIndex];
    final now = DateTime.now();
    final appliedMaps = appliedRaw is List
        ? appliedRaw
              .map(
                (raw) => raw is Map<String, dynamic>
                    ? raw
                    : (raw is Map ? Map<String, dynamic>.from(raw) : null),
              )
              .whereType<Map<String, dynamic>>()
              .toList(growable: false)
        : const <Map<String, dynamic>>[];

    for (final status in applied) {
      final matching = appliedMaps.firstWhere(
        (raw) => (raw['name']?.toString() ?? '').trim() == status.name.trim(),
        orElse: () => const <String, dynamic>{},
      );
      final durationSeconds = matching['durationSeconds'] is num
          ? (matching['durationSeconds'] as num).toInt()
          : 0;
      final expiresAt = durationSeconds > 0
          ? now.add(Duration(seconds: durationSeconds))
          : null;

      enemy.statuses.removeWhere(
        (existing) =>
            existing.name.trim().toLowerCase() ==
            status.name.trim().toLowerCase(),
      );
      enemy.statuses = [
        ...enemy.statuses,
        MonsterStatus(
          id: status.id,
          name: status.name,
          description: status.description,
          effect: status.effect,
          positive: status.positive,
          effectType: status.effectType,
          damagePerTick: status.damagePerTick,
          healthPerTick: status.healthPerTick,
          startedAt: now,
          expiresAt: expiresAt,
          lastTickAt: now,
        ),
      ];
      _battleLog.add(
        '${enemy.monster.name} gains ${_statusDisplayName(status.name, status.effectType)}.',
      );
    }

    if (removedNames.isNotEmpty) {
      enemy.statuses.removeWhere(
        (status) => removedNames.any(
          (removed) =>
              removed.trim().toLowerCase() == status.name.trim().toLowerCase(),
        ),
      );
      for (final removed in removedNames) {
        _battleLog.add(
          '${removed.trim()} is removed from ${enemy.monster.name}.',
        );
      }
    }
  }

  void _applyLocalBattleStatusTicks() {
    _pruneExpiredStatuses();
    final mutateResources = !widget.isPartyBattle;

    for (final status in _playerStatuses) {
      final statusName = _statusDisplayName(status.name, status.effectType);
      final healthDelta = _playerStatusHealthTick(status);
      final manaDelta = _playerStatusManaTick(status);
      if (healthDelta != 0) {
        if (mutateResources) {
          _playerHealth = (_playerHealth + healthDelta)
              .clamp(0, _playerMaxHealth)
              .toInt();
        }
        if (healthDelta < 0) {
          _battleLog.add(
            '$statusName deals ${healthDelta.abs()} damage to you.',
          );
        } else {
          _battleLog.add('$statusName restores $healthDelta HP to you.');
        }
      }
      if (manaDelta != 0) {
        if (mutateResources) {
          _playerMana = (_playerMana + manaDelta)
              .clamp(0, _playerMaxMana)
              .toInt();
        }
        if (manaDelta < 0) {
          _battleLog.add('$statusName drains ${manaDelta.abs()} MP from you.');
        } else {
          _battleLog.add('$statusName restores $manaDelta MP to you.');
        }
      }
    }
    _syncSelfAllyFromLocalResources();

    for (final enemy in _enemies) {
      if (enemy.isDefeated) continue;
      for (final status in enemy.statuses) {
        final delta = _monsterStatusHealthTick(status);
        if (delta == 0) continue;
        if (mutateResources) {
          enemy.currentHealth = (enemy.currentHealth + delta)
              .clamp(0, enemy.maxHealth)
              .toInt();
        }
        final statusName = _statusDisplayName(status.name, status.effectType);
        if (delta < 0) {
          _battleLog.add(
            '$statusName deals ${delta.abs()} damage to ${enemy.monster.name}.',
          );
        } else {
          _battleLog.add(
            '$statusName restores $delta HP to ${enemy.monster.name}.',
          );
        }
      }
    }
    _ensureSelectedEnemyIsAlive();
  }

  void _shiftCombatStatusTimersByTurn() {
    DateTime? shiftExpiry(DateTime? expiresAt) {
      if (expiresAt == null) return null;
      return expiresAt.subtract(_combatTurnDuration);
    }

    _playerStatuses = _playerStatuses
        .map(
          (status) => CharacterStatus(
            id: status.id,
            name: status.name,
            description: status.description,
            effect: status.effect,
            positive: status.positive,
            effectType: status.effectType,
            damagePerTick: status.damagePerTick,
            healthPerTick: status.healthPerTick,
            manaPerTick: status.manaPerTick,
            startedAt: status.startedAt,
            expiresAt: shiftExpiry(status.expiresAt),
          ),
        )
        .toList();

    for (final enemy in _enemies) {
      enemy.statuses = enemy.statuses
          .map(
            (status) => MonsterStatus(
              id: status.id,
              name: status.name,
              description: status.description,
              effect: status.effect,
              positive: status.positive,
              effectType: status.effectType,
              damagePerTick: status.damagePerTick,
              healthPerTick: status.healthPerTick,
              startedAt: status.startedAt,
              expiresAt: shiftExpiry(status.expiresAt),
              lastTickAt: status.lastTickAt,
            ),
          )
          .toList();
    }

    _pruneExpiredStatuses();
  }

  void _applyTurnSyncResults({
    Map<String, dynamic>? setupResponse,
    Map<String, dynamic>? turnResponse,
    int? targetEnemyIndex,
    bool refreshPlayerStatusesFromProvider = false,
  }) {
    if (refreshPlayerStatusesFromProvider) {
      final statsProvider = context.read<CharacterStatsProvider>();
      _playerStatuses = List<CharacterStatus>.from(statsProvider.statuses);
      _playerHealth = math.max(0, statsProvider.health);
      _playerMana = math.max(0, statsProvider.mana);
    }

    if (setupResponse != null && setupResponse.isNotEmpty) {
      _cacheVictoryRewardsFromPayload(setupResponse);
      _markSeenPartyActionSequenceFromPayload(setupResponse);
      _updatePartyTurnOrderFromPayload(setupResponse);
      _applyParticipantResourcesFromPayload(setupResponse);
      _syncPartyMonsterHealthFromPayload(setupResponse);
      _applyMonsterStatusChanges(
        setupResponse,
        targetEnemyIndex: targetEnemyIndex,
      );
      _syncEnemyStatusesFromPayload(setupResponse);
    }
    if (turnResponse != null && turnResponse.isNotEmpty) {
      _cacheVictoryRewardsFromPayload(turnResponse);
      _markSeenPartyActionSequenceFromPayload(turnResponse);
      _updatePartyTurnOrderFromPayload(turnResponse);
      _syncPartyMonsterHealthFromPayload(turnResponse);
      _syncEnemyStatusesFromPayload(turnResponse);
      _applyLocalBattleStatusTicks();
      _shiftCombatStatusTimersByTurn();
    }
    if (widget.isPartyBattle) {
      _refreshPartyTurnFlag();
    }
  }

  void _applyParticipantResourcesFromPayload(Map<String, dynamic> payload) {
    final rawSnapshots = payload['participantResources'];
    if (rawSnapshots is! List) {
      return;
    }
    for (final raw in rawSnapshots) {
      final snapshot = raw is Map<String, dynamic>
          ? raw
          : (raw is Map ? Map<String, dynamic>.from(raw) : null);
      if (snapshot == null) {
        continue;
      }
      final userId = (snapshot['userId']?.toString() ?? '').trim();
      if (userId.isEmpty) {
        continue;
      }
      final health = _parseIntValue(snapshot['health']);
      final maxHealth = math.max(1, _parseIntValue(snapshot['maxHealth']));
      final mana = _parseIntValue(snapshot['mana']);
      final maxMana = math.max(0, _parseIntValue(snapshot['maxMana']));
      final nextHealth = health.clamp(0, maxHealth).toInt();
      var downedLogLabel = '';
      var shouldLogDowned = false;
      if (userId == _selfUserId) {
        shouldLogDowned = _playerHealth > 0 && nextHealth <= 0;
        downedLogLabel = 'You';
        _playerHealth = nextHealth;
        _playerMaxHealth = maxHealth;
        _playerMana = mana.clamp(0, maxMana).toInt();
        _playerMaxMana = maxMana;
        _syncSelfAllyFromLocalResources();
      }
      final allyIndex = _partyAllies.indexWhere(
        (ally) => ally.userId == userId,
      );
      if (allyIndex < 0) {
        if (shouldLogDowned) {
          _battleLog.add('$downedLogLabel are down.');
        }
        continue;
      }
      final ally = _partyAllies[allyIndex];
      if (!shouldLogDowned && ally.currentHealth > 0 && nextHealth <= 0) {
        shouldLogDowned = true;
        downedLogLabel = ally.isSelf ? 'You' : ally.name;
      }
      ally.maxHealth = maxHealth;
      ally.currentHealth = nextHealth;
      ally.maxMana = maxMana;
      ally.currentMana = mana.clamp(0, maxMana).toInt();
      if (shouldLogDowned) {
        if (ally.isSelf) {
          _battleLog.add('$downedLogLabel are down.');
        } else {
          _battleLog.add('$downedLogLabel is down.');
        }
      }
    }
  }

  void _applyMonsterActionPayload(Map<String, dynamic>? payload) {
    if (payload == null || payload.isEmpty) {
      return;
    }
    _applyParticipantResourcesFromPayload(payload);

    final rawAction = payload['monsterAction'];
    final action = rawAction is Map<String, dynamic>
        ? rawAction
        : (rawAction is Map ? Map<String, dynamic>.from(rawAction) : null);
    if (action == null || action.isEmpty) {
      return;
    }

    final actorName = (action['actorMonsterName']?.toString() ?? '').trim();
    final abilityName = (action['abilityName']?.toString() ?? '').trim();
    final damage = _parseIntValue(action['damage']);
    final heal = _parseIntValue(action['heal']);
    final targetUserId = (action['targetUserId']?.toString() ?? '').trim();
    final targetUserIds =
        ((action['targetUserIds'] as List<dynamic>?) ?? const [])
            .map((entry) => entry?.toString().trim() ?? '')
            .where((entry) => entry.isNotEmpty)
            .toList(growable: false);

    String targetLabel() {
      if (targetUserIds.length > 1) {
        return 'the party';
      }
      if (targetUserId == _selfUserId || targetUserIds.contains(_selfUserId)) {
        return 'you';
      }
      if (targetUserId.isNotEmpty) {
        final ally = _partyAllies.where(
          (entry) => entry.userId == targetUserId,
        );
        if (ally.isNotEmpty) {
          return ally.first.name;
        }
      }
      return 'the party';
    }

    final actor = actorName.isNotEmpty ? actorName : 'The monster';
    final actionType = (action['actionType']?.toString() ?? '').trim();
    final statusesApplied =
        ((action['userStatusesApplied'] as List<dynamic>?) ?? const []).length;

    if (actionType == 'attack' && damage > 0) {
      _battleLog.add('$actor attacks ${targetLabel()} for $damage damage.');
    } else if (abilityName.isNotEmpty && damage > 0) {
      _battleLog.add(
        '$actor uses $abilityName on ${targetLabel()} for $damage damage.',
      );
    } else if (abilityName.isNotEmpty && heal > 0) {
      _battleLog.add('$actor uses $abilityName and restores $heal HP.');
    } else if (abilityName.isNotEmpty && statusesApplied > 0) {
      _battleLog.add('$actor uses $abilityName on ${targetLabel()}.');
    } else if (abilityName.isNotEmpty) {
      _battleLog.add('$actor uses $abilityName.');
    }

    if ((targetUserId == _selfUserId || targetUserIds.contains(_selfUserId)) &&
        damage > 0) {
      unawaited(
        _playSpriteFx(targetMonster: false, amount: damage, healing: false),
      );
    }
    if (heal > 0) {
      unawaited(
        _playSpriteFx(targetMonster: true, amount: heal, healing: true),
      );
    }
  }

  List<_EncounterEnemyState> get _aliveEnemies =>
      _enemies.where((enemy) => !enemy.isDefeated).toList(growable: false);

  int? _firstAliveEnemyIndex() {
    for (var i = 0; i < _enemies.length; i++) {
      if (!_enemies[i].isDefeated) return i;
    }
    return null;
  }

  int? _resolveTargetEnemyIndex([int? preferredIndex]) {
    if (preferredIndex != null &&
        preferredIndex >= 0 &&
        preferredIndex < _enemies.length &&
        !_enemies[preferredIndex].isDefeated) {
      return preferredIndex;
    }
    return _firstAliveEnemyIndex();
  }

  void _ensureSelectedEnemyIsAlive() {
    final resolved = _resolveTargetEnemyIndex(_activeEnemyIndex);
    if (resolved != null) {
      _activeEnemyIndex = resolved;
    }
  }

  _EncounterEnemyState? get _activeEnemy {
    final resolved = _resolveTargetEnemyIndex(_activeEnemyIndex);
    if (resolved == null) return null;
    return _enemies[resolved];
  }

  bool _allPartyMembersDefeated() {
    if (!widget.isPartyBattle) {
      return _playerHealth <= 0;
    }
    if (_partyAllies.isEmpty) {
      return _playerHealth <= 0;
    }
    final trackedUserIDs = <String>{};
    if (_activePartyParticipantIds.isNotEmpty) {
      trackedUserIDs.addAll(_activePartyParticipantIds);
    } else {
      for (final ally in _partyAllies) {
        trackedUserIDs.add(ally.userId);
      }
    }
    if (trackedUserIDs.isEmpty && _selfUserId.isNotEmpty) {
      trackedUserIDs.add(_selfUserId);
    }
    var foundTrackedParticipant = false;
    for (final ally in _partyAllies) {
      if (!trackedUserIDs.contains(ally.userId)) {
        continue;
      }
      foundTrackedParticipant = true;
      if (ally.currentHealth > 0) return false;
    }
    if (!foundTrackedParticipant) {
      return _playerHealth <= 0;
    }
    return true;
  }

  bool _isTurnEntryAlive(_TurnOrderEntry entry) {
    if (entry.isEnemy) {
      final index = entry.enemyIndex;
      if (index == null || index < 0 || index >= _enemies.length) return false;
      return !_enemies[index].isDefeated;
    }
    final userId = (entry.userId ?? '').trim();
    if (userId.isEmpty) return false;
    final allyIndex = _partyAllies.indexWhere((ally) => ally.userId == userId);
    if (allyIndex < 0) return false;
    return _partyAllies[allyIndex].currentHealth > 0;
  }

  _TurnOrderEntry? _currentPartyTurnEntry() {
    if (!widget.isPartyBattle || _partyTurnOrder.isEmpty) return null;
    final total = _partyTurnOrder.length;
    if (_partyTurnIndex < 0 || _partyTurnIndex >= total) {
      _partyTurnIndex = 0;
    }
    for (var i = 0; i < total; i++) {
      final idx = (_partyTurnIndex + i) % total;
      final candidate = _partyTurnOrder[idx];
      if (_isTurnEntryAlive(candidate)) {
        _partyTurnIndex = idx;
        return candidate;
      }
    }
    return null;
  }

  bool _sameTurnEntry(_TurnOrderEntry a, _TurnOrderEntry b) {
    if (a.isEnemy && b.isEnemy) {
      return a.enemyIndex == b.enemyIndex;
    }
    if (a.isAlly && b.isAlly) {
      return (a.userId ?? '') == (b.userId ?? '');
    }
    return false;
  }

  bool _isPartySelfTurn() {
    final current = _currentPartyTurnEntry();
    return current?.isSelf == true;
  }

  bool _isPartyMonsterTurn() {
    final current = _currentPartyTurnEntry();
    return current?.isEnemy == true;
  }

  void _beginSelfTurn() {
    _pruneExpiredTechniqueCooldowns();
  }

  void _pruneExpiredTechniqueCooldowns() {
    final now = DateTime.now();
    final expiredIds = <String>[];
    _techniqueCooldownExpiresAtById.forEach((id, expiresAt) {
      if (!expiresAt.isAfter(now)) {
        expiredIds.add(id);
      }
    });
    for (final id in expiredIds) {
      _techniqueCooldownExpiresAtById.remove(id);
      _techniqueCooldownClearedById.add(id);
    }
  }

  void _advanceLocalTechniqueCooldownsForAction({String? excludeTechniqueId}) {
    final now = DateTime.now();
    final next = <String, DateTime>{};
    for (final entry in _techniqueCooldownExpiresAtById.entries) {
      if (excludeTechniqueId != null && entry.key == excludeTechniqueId) {
        next[entry.key] = entry.value;
        continue;
      }
      final shifted = entry.value.subtract(_combatTurnDuration);
      if (shifted.isAfter(now)) {
        next[entry.key] = shifted;
      } else {
        _techniqueCooldownClearedById.add(entry.key);
      }
    }
    _techniqueCooldownExpiresAtById
      ..clear()
      ..addAll(next);
    _pruneExpiredTechniqueCooldowns();
  }

  int _cooldownTurnsRemainingFromExpiry(
    DateTime? expiresAt, {
    int fallbackTurns = 0,
  }) {
    if (expiresAt == null) {
      return math.max(0, fallbackTurns);
    }
    final remainingSeconds = expiresAt.difference(DateTime.now()).inSeconds;
    if (remainingSeconds <= 0) return 0;
    return (remainingSeconds + _combatTurnDuration.inSeconds - 1) ~/
        _combatTurnDuration.inSeconds;
  }

  int _techniqueCooldownRemaining(Spell technique) {
    _pruneExpiredTechniqueCooldowns();
    final id = technique.id.trim();
    if (id.isEmpty) return 0;
    if (_techniqueCooldownClearedById.contains(id)) {
      return 0;
    }
    final override = _techniqueCooldownExpiresAtById[id];
    if (override != null) {
      return _cooldownTurnsRemainingFromExpiry(override);
    }
    return _cooldownTurnsRemainingFromExpiry(
      technique.cooldownExpiresAt,
      fallbackTurns: technique.cooldownTurnsRemaining,
    );
  }

  bool _isTechniqueOnCooldown(Spell technique) =>
      _techniqueCooldownRemaining(technique) > 0;

  void _setTechniqueCooldown(Spell technique) {
    final id = technique.id.trim();
    if (id.isEmpty) return;
    final cooldownTurns = math.max(
      0,
      technique.cooldownTurnsRemaining > 0
          ? technique.cooldownTurnsRemaining
          : technique.cooldownTurns,
    );
    if (cooldownTurns <= 0) {
      _techniqueCooldownExpiresAtById.remove(id);
      _techniqueCooldownClearedById.add(id);
      return;
    }
    _techniqueCooldownClearedById.remove(id);
    _techniqueCooldownExpiresAtById[id] =
        technique.cooldownExpiresAt ??
        DateTime.now().add(_combatTurnDuration * cooldownTurns);
  }

  void _setTechniqueCooldownFromResponse(
    Spell technique,
    Map<String, dynamic> response,
  ) {
    if (!_isTechnique(technique)) return;
    final id = technique.id.trim();
    if (id.isEmpty) return;

    final secondsRaw = response['cooldownSecondsRemaining'];
    final turnsRaw = response['cooldownTurnsRemaining'];
    final seconds = secondsRaw is num ? secondsRaw.toInt() : 0;
    final turns = turnsRaw is num ? turnsRaw.toInt() : 0;
    if (seconds <= 0 && turns <= 0) {
      _techniqueCooldownExpiresAtById.remove(id);
      _techniqueCooldownClearedById.add(id);
      return;
    }

    final remaining = seconds > 0
        ? Duration(seconds: seconds)
        : _combatTurnDuration * turns;
    _techniqueCooldownClearedById.remove(id);
    _techniqueCooldownExpiresAtById[id] = DateTime.now().add(remaining);
  }

  void _syncTechniqueCooldownsFromAbilities() {
    final now = DateTime.now();
    final next = <String, DateTime>{};
    final nextCleared = <String>{..._techniqueCooldownClearedById};
    for (final technique in _techniques) {
      final id = technique.id.trim();
      if (id.isEmpty) continue;
      final remaining = math.max(0, technique.cooldownTurnsRemaining);
      if (remaining <= 0) {
        nextCleared.add(id);
        continue;
      }
      if (nextCleared.contains(id)) {
        continue;
      }
      final serverExpiry =
          technique.cooldownExpiresAt ??
          now.add(_combatTurnDuration * remaining);
      final localExpiry = _techniqueCooldownExpiresAtById[id];
      if (localExpiry != null && localExpiry.isBefore(serverExpiry)) {
        next[id] = localExpiry;
      } else {
        next[id] = serverExpiry;
      }
    }
    _techniqueCooldownExpiresAtById
      ..clear()
      ..addAll(next);
    _techniqueCooldownClearedById
      ..clear()
      ..addAll(nextCleared);
    _pruneExpiredTechniqueCooldowns();
  }

  void _refreshPartyTurnFlag({bool announce = true}) {
    if (!widget.isPartyBattle) return;
    final wasSelfTurn = _playerTurn;
    final isSelfTurn = !_battleOver && _isPartySelfTurn() && _playerHealth > 0;
    _playerTurn = isSelfTurn;
    if (!wasSelfTurn && isSelfTurn) {
      _beginSelfTurn();
    }
    if (announce && !wasSelfTurn && isSelfTurn && !_busy) {
      _battleLog.add('Your turn. Choose a command.');
    }
  }

  Future<void> _runPartyMonsterTurn() async {
    if (!widget.isPartyBattle || _battleOver || _partyMonsterTurnInFlight) {
      return;
    }
    final currentTurn = _currentPartyTurnEntry();
    if (currentTurn == null || !currentTurn.isEnemy) {
      return;
    }
    final enemyIndex =
        currentTurn.enemyIndex ??
        _resolveTargetEnemyIndex(_activeEnemyIndex) ??
        0;
    if (enemyIndex < 0 || enemyIndex >= _enemies.length) {
      return;
    }
    final enemy = _enemies[enemyIndex];
    if (enemy.isDefeated) {
      await _syncPartyBattleState();
      return;
    }

    _partyMonsterTurnInFlight = true;
    setState(() {
      _busy = true;
      _actingEnemyIndex = enemyIndex;
    });
    await Future<void>.delayed(const Duration(milliseconds: 320));
    if (!mounted || _battleOver) {
      _partyMonsterTurnInFlight = false;
      return;
    }
    Map<String, dynamic>? turnResponse;
    final monsterId = (_partyBattleMonsterId ?? '').trim();
    if (monsterId.isNotEmpty) {
      try {
        final poiService = context.read<PoiService>();
        turnResponse = await poiService.advanceMonsterBattleTurn(monsterId);
      } catch (_) {
        turnResponse = null;
      }
    }
    setState(() {
      _actingEnemyIndex = null;
      _busy = false;
      _selectedCommandKey = 'root:Attack';
      _applyMonsterActionPayload(turnResponse);
      _applyTurnSyncResults(turnResponse: turnResponse);
      if (_playerHealth <= 0) {
        _battleLog.add('You are down. Waiting for your party...');
      }
    });
    if (!_battleOver && _allPartyMembersDefeated()) {
      _partyMonsterTurnInFlight = false;
      await _finishBattle(
        MonsterBattleOutcome.defeat,
        'Your party has been defeated.',
      );
      return;
    }
    _partyMonsterTurnInFlight = false;
  }

  int _parseIntValue(dynamic raw, {int fallback = 0}) {
    if (raw is num) return raw.toInt();
    return int.tryParse(raw?.toString() ?? '') ?? fallback;
  }

  bool _parseBoolValue(dynamic raw, {bool fallback = false}) {
    if (raw is bool) return raw;
    final normalized = (raw?.toString() ?? '').trim().toLowerCase();
    if (normalized == 'true') return true;
    if (normalized == 'false') return false;
    return fallback;
  }

  Map<String, dynamic>? _battleMapFromPayload(Map<String, dynamic>? payload) {
    if (payload == null || payload.isEmpty) return null;
    final directBattleRaw = payload['battle'];
    if (directBattleRaw is Map<String, dynamic>) {
      return directBattleRaw;
    }
    if (directBattleRaw is Map) {
      return Map<String, dynamic>.from(directBattleRaw);
    }
    final detailRaw = payload['battleDetail'];
    final detail = detailRaw is Map<String, dynamic>
        ? detailRaw
        : (detailRaw is Map ? Map<String, dynamic>.from(detailRaw) : null);
    if (detail == null) return null;
    final nestedBattleRaw = detail['battle'];
    if (nestedBattleRaw is Map<String, dynamic>) {
      return nestedBattleRaw;
    }
    if (nestedBattleRaw is Map) {
      return Map<String, dynamic>.from(nestedBattleRaw);
    }
    return null;
  }

  void _markSeenPartyActionSequenceFromPayload(Map<String, dynamic>? payload) {
    final battle = _battleMapFromPayload(payload);
    if (battle == null || battle.isEmpty) return;
    final sequence = _parseIntValue(battle['lastActionSequence']);
    if (sequence > _lastSeenPartyActionSequence) {
      _lastSeenPartyActionSequence = sequence;
    }
  }

  bool _partyHasLivingParticipantInPayload(Map<String, dynamic>? payload) {
    if (!widget.isPartyBattle || payload == null || payload.isEmpty) {
      return _playerHealth > 0;
    }
    final detailRaw = payload['battleDetail'];
    final detail = detailRaw is Map<String, dynamic>
        ? detailRaw
        : (detailRaw is Map ? Map<String, dynamic>.from(detailRaw) : payload);
    final rawSnapshots = detail['participantResources'];
    if (rawSnapshots is! List) {
      return !_allPartyMembersDefeated();
    }
    for (final raw in rawSnapshots) {
      final snapshot = raw is Map<String, dynamic>
          ? raw
          : (raw is Map ? Map<String, dynamic>.from(raw) : null);
      if (snapshot == null) continue;
      final health = _parseIntValue(snapshot['health']);
      if (health > 0) {
        return true;
      }
    }
    return false;
  }

  void _cacheVictoryRewardsFromPayload(Map<String, dynamic>? payload) {
    if (payload == null || payload.isEmpty) return;
    final detailRaw = payload['battleDetail'];
    final detail = detailRaw is Map<String, dynamic>
        ? detailRaw
        : (detailRaw is Map ? Map<String, dynamic>.from(detailRaw) : payload);
    final rawRewards = detail['participantRewards'];
    if (rawRewards is! List) return;
    for (final raw in rawRewards) {
      final reward = raw is Map<String, dynamic>
          ? raw
          : (raw is Map ? Map<String, dynamic>.from(raw) : null);
      if (reward == null) continue;
      final userId = (reward['userId']?.toString() ?? '').trim();
      if (userId != _selfUserId) continue;
      final rawItems = reward['itemsAwarded'];
      final items = rawItems is List
          ? rawItems
                .whereType<Map>()
                .map((entry) => Map<String, dynamic>.from(entry))
                .toList(growable: false)
          : const <Map<String, dynamic>>[];
      final rawBaseResources = reward['baseResourcesAwarded'];
      final baseResources = rawBaseResources is List
          ? rawBaseResources
                .whereType<Map>()
                .map((entry) => Map<String, dynamic>.from(entry))
                .toList(growable: false)
          : const <Map<String, dynamic>>[];
      _victoryRewardExperience = _parseIntValue(reward['rewardExperience']);
      _victoryRewardGold = _parseIntValue(reward['rewardGold']);
      _hasCachedVictoryRewards = true;
      _victoryBaseResourcesAwarded = baseResources;
      _victoryItemsAwarded = items;
      return;
    }
  }

  List<Map<String, dynamic>> _fallbackVictoryItemsAwarded() {
    final itemTotals = <int, Map<String, dynamic>>{};
    for (final reward in widget.encounter.itemRewards) {
      final quantity = reward.quantity > 0 ? reward.quantity : 1;
      final entry = itemTotals.putIfAbsent(reward.inventoryItemId, () {
        return <String, dynamic>{
          'id': reward.inventoryItemId,
          'name': reward.inventoryItemName.isNotEmpty
              ? reward.inventoryItemName
              : 'Item #${reward.inventoryItemId}',
          'imageUrl': reward.inventoryItemImageUrl,
          'quantity': 0,
        };
      });
      entry['quantity'] = (entry['quantity'] as int) + quantity;
      if ((entry['imageUrl'] as String).isEmpty &&
          reward.inventoryItemImageUrl.isNotEmpty) {
        entry['imageUrl'] = reward.inventoryItemImageUrl;
      }
    }
    return itemTotals.values
        .map((entry) => Map<String, dynamic>.from(entry))
        .toList(growable: false);
  }

  String _formatPartyBattleActionLog(Map<String, dynamic> action) {
    final actorName = (action['actorName']?.toString() ?? '').trim();
    if (actorName.isEmpty) return '';
    final abilityName = (action['abilityName']?.toString() ?? '').trim();
    final actionType = (action['actionType']?.toString() ?? '').trim();
    final targetName = (action['targetName']?.toString() ?? '').trim();
    final targetUserId = (action['targetUserId']?.toString() ?? '').trim();
    final targetsAllEnemies = _parseBoolValue(action['targetsAllEnemies']);
    final damage = _parseIntValue(action['damage']);
    final heal = _parseIntValue(action['heal']);
    final statusesApplied = _parseIntValue(action['statusesApplied']);
    final statusesRemoved = _parseIntValue(action['statusesRemoved']);
    String resolvedTargetName() {
      if (targetsAllEnemies) return 'all enemies';
      if (targetName.isNotEmpty) return targetName;
      if (targetUserId == _selfUserId && _selfUserId.isNotEmpty) return 'you';
      if (targetUserId.isNotEmpty) {
        final allyIndex = _partyAllies.indexWhere(
          (ally) => ally.userId == targetUserId,
        );
        if (allyIndex >= 0) {
          return _partyAllies[allyIndex].name;
        }
      }
      return 'the enemy';
    }

    final resolvedTarget = resolvedTargetName();

    if (damage > 0) {
      if (abilityName.isNotEmpty) {
        return '$actorName uses $abilityName on $resolvedTarget for $damage damage.';
      }
      return '$actorName attacks $resolvedTarget for $damage damage.';
    }
    if (heal > 0) {
      if (abilityName.isNotEmpty) {
        if (targetName.isNotEmpty) {
          return '$actorName uses $abilityName on $targetName and restores $heal HP.';
        }
        return '$actorName uses $abilityName and restores $heal HP.';
      }
      if (targetName.isNotEmpty) {
        return '$actorName restores $heal HP to $targetName.';
      }
      return '$actorName restores $heal HP.';
    }
    if (abilityName.isNotEmpty) {
      if (targetsAllEnemies) {
        return '$actorName uses $abilityName on all enemies.';
      }
      if (targetName.isNotEmpty) {
        return '$actorName uses $abilityName on $targetName.';
      }
      if (statusesApplied > 0 || statusesRemoved > 0) {
        return '$actorName uses $abilityName.';
      }
    }
    if (actionType == 'attack') {
      return '$actorName attacks.';
    }
    return '';
  }

  bool _consumePartyLastActionFromPayload(Map<String, dynamic>? payload) {
    final battle = _battleMapFromPayload(payload);
    if (battle == null || battle.isEmpty) return false;
    final sequence = _parseIntValue(battle['lastActionSequence']);
    if (_lastSeenPartyActionSequence < 0) {
      _lastSeenPartyActionSequence = math.max(0, sequence);
      return false;
    }
    if (sequence <= 0) return false;
    if (sequence <= _lastSeenPartyActionSequence) return false;
    _lastSeenPartyActionSequence = sequence;

    final rawAction = battle['lastAction'];
    final action = rawAction is Map<String, dynamic>
        ? rawAction
        : (rawAction is Map ? Map<String, dynamic>.from(rawAction) : null);
    if (action == null || action.isEmpty) return true;

    final actorUserId = (action['actorUserId']?.toString() ?? '').trim();
    if (_selfUserId.isNotEmpty && actorUserId == _selfUserId) {
      return true;
    }

    final message = _formatPartyBattleActionLog(action);
    if (message.isNotEmpty) {
      _battleLog.add(message);
    }
    return true;
  }

  String _userDisplayNameFromRaw(Map<String, dynamic> raw) {
    final username = (raw['username']?.toString() ?? '').trim();
    if (username.isNotEmpty) return '@$username';
    final name = (raw['name']?.toString() ?? '').trim();
    if (name.isNotEmpty) return name;
    return 'Party Member';
  }

  String _shortTurnLabel(String value) {
    final trimmed = value.trim();
    if (trimmed.isEmpty) return '?';
    return trimmed.length <= 10 ? trimmed : '${trimmed.substring(0, 9)}…';
  }

  _PartyAllyState? get _activeAlly {
    if (!widget.isPartyBattle || _partyAllies.isEmpty) return null;
    if (_activeAllyIndex < 0 || _activeAllyIndex >= _partyAllies.length) {
      _activeAllyIndex = 0;
    }
    return _partyAllies[_activeAllyIndex];
  }

  _PartyAllyState? get _selfPartyAlly {
    if (!widget.isPartyBattle || _selfUserId.isEmpty) return null;
    final index = _partyAllies.indexWhere((ally) => ally.userId == _selfUserId);
    if (index < 0) return null;
    return _partyAllies[index];
  }

  List<_PartyAllyState> _supportTargetOptions() {
    final options = <_PartyAllyState>[];
    final seenUserIds = <String>{};

    void addOption(_PartyAllyState ally) {
      final userId = ally.userId.trim();
      if (userId.isEmpty || !seenUserIds.add(userId)) return;
      options.add(ally);
    }

    if (widget.isPartyBattle && _partyAllies.isNotEmpty) {
      for (final ally in _partyAllies) {
        addOption(ally);
      }
      return options;
    }

    if (_selfUserId.isNotEmpty) {
      addOption(
        _PartyAllyState(
          userId: _selfUserId,
          name: _playerName,
          iconUrl: _playerFrontSpriteUrl,
          level: _playerLevel,
          currentHealth: _playerHealth,
          maxHealth: _playerMaxHealth,
          currentMana: _playerMana,
          maxMana: _playerMaxMana,
          isSelf: true,
        ),
      );
    }

    final party = context.read<PartyProvider>().party;
    if (party == null) {
      return options;
    }
    final users = <User>[party.leader, ...party.members];
    for (final user in users) {
      final userId = user.id.trim();
      if (userId.isEmpty) continue;
      final isSelf = userId == _selfUserId;
      addOption(
        _PartyAllyState(
          userId: userId,
          name: user.username.trim().isNotEmpty
              ? '@${user.username.trim()}'
              : (user.name.trim().isNotEmpty
                    ? user.name.trim()
                    : 'Party Member'),
          iconUrl: user.profilePictureUrl.trim(),
          level: isSelf ? _playerLevel : 1,
          currentHealth: isSelf ? _playerHealth : 0,
          maxHealth: isSelf ? _playerMaxHealth : 1,
          currentMana: isSelf ? _playerMana : 0,
          maxMana: isSelf ? _playerMaxMana : 0,
          isSelf: isSelf,
        ),
      );
    }
    return options;
  }

  void _ensureSelfPartyAlly() {
    if (!widget.isPartyBattle) return;
    if (_selfUserId.isEmpty) return;
    final existing = _partyAllies.indexWhere(
      (ally) => ally.userId == _selfUserId,
    );
    if (existing >= 0) {
      final self = _partyAllies[existing];
      self.name = _playerName;
      self.iconUrl = _playerFrontSpriteUrl;
      self.level = _playerLevel;
      self.maxHealth = _playerMaxHealth;
      self.currentHealth = _playerHealth;
      self.maxMana = _playerMaxMana;
      self.currentMana = _playerMana;
      return;
    }
    _partyAllies.insert(
      0,
      _PartyAllyState(
        userId: _selfUserId,
        name: _playerName,
        iconUrl: _playerFrontSpriteUrl,
        level: _playerLevel,
        currentHealth: _playerHealth,
        maxHealth: _playerMaxHealth,
        currentMana: _playerMana,
        maxMana: _playerMaxMana,
        isSelf: true,
      ),
    );
    _activeAllyIndex = 0;
  }

  void _seedPartyAlliesFromPartyProvider() {
    if (!widget.isPartyBattle) return;
    final provider = context.read<PartyProvider>();
    final party = provider.party;
    if (party == null) return;

    final users = <Map<String, dynamic>>[
      party.leader.toJson(),
      ...party.members.map((member) => member.toJson()),
    ];
    for (final raw in users) {
      final id = (raw['id']?.toString() ?? '').trim();
      if (id.isEmpty) continue;
      final existing = _partyAllies.indexWhere((ally) => ally.userId == id);
      if (existing < 0) {
        _partyAllies.add(
          _PartyAllyState(
            userId: id,
            name: _userDisplayNameFromRaw(raw),
            iconUrl: (raw['profilePictureUrl']?.toString() ?? '').trim(),
            level: id == _selfUserId ? _playerLevel : 1,
            currentHealth: id == _selfUserId ? _playerHealth : 1,
            maxHealth: id == _selfUserId ? _playerMaxHealth : 1,
            currentMana: id == _selfUserId ? _playerMana : 0,
            maxMana: id == _selfUserId ? _playerMaxMana : 0,
            isSelf: id == _selfUserId,
          ),
        );
        continue;
      }
      if (id == _selfUserId) continue;
      final ally = _partyAllies[existing];
      ally.name = _userDisplayNameFromRaw(raw);
      final icon = (raw['profilePictureUrl']?.toString() ?? '').trim();
      if (icon.isNotEmpty) {
        ally.iconUrl = icon;
      }
    }
  }

  void _syncSelfAllyFromLocalResources() {
    if (!widget.isPartyBattle || _selfUserId.isEmpty) return;
    final existing = _partyAllies.indexWhere(
      (ally) => ally.userId == _selfUserId,
    );
    if (existing < 0) return;
    final self = _partyAllies[existing];
    self.level = _playerLevel;
    self.name = _playerName;
    self.iconUrl = _playerFrontSpriteUrl;
    self.currentHealth = _playerHealth;
    self.maxHealth = _playerMaxHealth;
    self.currentMana = _playerMana;
    self.maxMana = _playerMaxMana;
  }

  void _allowPartySelfHealthIncreaseSync() {
    if (!widget.isPartyBattle) return;
    _partySelfHealthIncreaseSyncAllowed = true;
  }

  void _applySelfHealsFromResponse(Map<String, dynamic> response) {
    final healsRaw = response['heals'];
    if (healsRaw is! List) return;
    for (final raw in healsRaw) {
      final heal = raw is Map<String, dynamic>
          ? raw
          : (raw is Map ? Map<String, dynamic>.from(raw) : null);
      if (heal == null) continue;
      final userId = (heal['userId']?.toString() ?? '').trim();
      if (userId.isEmpty) continue;
      final maxHealth = math.max(
        1,
        _parseIntValue(
          heal['maxHealth'],
          fallback: userId == _selfUserId ? _playerMaxHealth : 1,
        ),
      );
      final nextHealth = _parseIntValue(
        heal['health'],
        fallback: userId == _selfUserId ? _playerHealth : 0,
      ).clamp(0, maxHealth).toInt();
      if (userId == _selfUserId) {
        if (widget.isPartyBattle && nextHealth > _playerHealth) {
          _allowPartySelfHealthIncreaseSync();
        }
        _playerHealth = nextHealth;
        _playerMaxHealth = maxHealth;
        _syncSelfAllyFromLocalResources();
        continue;
      }
      final allyIndex = _partyAllies.indexWhere(
        (ally) => ally.userId == userId,
      );
      if (allyIndex < 0) continue;
      final ally = _partyAllies[allyIndex];
      ally.maxHealth = maxHealth;
      ally.currentHealth = nextHealth;
    }
  }

  void _requestPartySelfResourceSync({
    bool syncHealth = true,
    bool syncMana = true,
  }) {
    if (!widget.isPartyBattle || _battleOver || _endingBattle) return;
    if (syncHealth) {
      _pendingPartySelfHealthSync = true;
    }
    if (syncMana) {
      _pendingPartySelfManaSync = true;
    }
    if (_partySelfResourceSyncInFlight) {
      debugPrint(
        '[combat][_requestPartySelfResourceSync] queued '
        'syncHealth=$syncHealth syncMana=$syncMana '
        'pendingHealth=$_pendingPartySelfHealthSync '
        'pendingMana=$_pendingPartySelfManaSync',
      );
      return;
    }
    unawaited(_syncPartySelfResourcesToBackend());
  }

  Future<void> _syncPartySelfResourcesToBackend() async {
    if (!widget.isPartyBattle ||
        _selfUserId.isEmpty ||
        !mounted ||
        _battleOver ||
        _endingBattle) {
      return;
    }
    if (_partySelfResourceSyncInFlight) {
      return;
    }
    final statsProvider = context.read<CharacterStatsProvider>();
    _partySelfResourceSyncInFlight = true;
    try {
      while (mounted &&
          widget.isPartyBattle &&
          !_battleOver &&
          !_endingBattle &&
          (_pendingPartySelfHealthSync || _pendingPartySelfManaSync)) {
        final syncHealth = _pendingPartySelfHealthSync;
        final syncMana = _pendingPartySelfManaSync;
        _pendingPartySelfHealthSync = false;
        _pendingPartySelfManaSync = false;
        if (!syncHealth && !syncMana) {
          continue;
        }
        await statsProvider.refresh(silent: true);
        if (!mounted || !widget.isPartyBattle || _battleOver || _endingBattle) {
          return;
        }
        final backendHealth = statsProvider.health.clamp(0, _playerMaxHealth);
        final backendMana = statsProvider.mana.clamp(0, _playerMaxMana);
        var targetHealth = _playerHealth.clamp(0, _playerMaxHealth).toInt();
        var targetMana = _playerMana.clamp(0, _playerMaxMana).toInt();
        final allowHealthIncrease = _partySelfHealthIncreaseSyncAllowed;
        if (!syncHealth) {
          targetHealth = backendHealth;
        } else if (!allowHealthIncrease && targetHealth > backendHealth) {
          debugPrint(
            '[combat][_syncPartySelfResourcesToBackend] suppressingHealthIncrease '
            'localHealth=$targetHealth backendHealth=$backendHealth '
            'localMana=$targetMana backendMana=$backendMana',
          );
          targetHealth = backendHealth;
          if (mounted) {
            setState(() {
              _playerHealth = targetHealth;
              _syncSelfAllyFromLocalResources();
            });
          } else {
            _playerHealth = targetHealth;
            _syncSelfAllyFromLocalResources();
          }
        }
        if (!syncMana) {
          targetMana = backendMana;
        }
        debugPrint(
          '[combat][_syncPartySelfResourcesToBackend] syncing '
          'localHealth=$targetHealth backendHealth=$backendHealth '
          'localMana=$targetMana backendMana=$backendMana '
          'syncHealth=$syncHealth syncMana=$syncMana '
          'allowHealthIncrease=$allowHealthIncrease',
        );
        if (!mounted || !widget.isPartyBattle || _battleOver || _endingBattle) {
          return;
        }
        await statsProvider.setCombatResources(
          health: syncHealth ? targetHealth : null,
          mana: syncMana ? targetMana : null,
          refreshBeforeAdjust: false,
        );
        _partySelfHealthIncreaseSyncAllowed = false;
      }
    } catch (error, stackTrace) {
      debugPrint(
        '[combat][_syncPartySelfResourcesToBackend] error=$error\n$stackTrace',
      );
      // Keep combat flow local if backend resource sync fails.
    } finally {
      _partySelfResourceSyncInFlight = false;
      if ((_pendingPartySelfHealthSync || _pendingPartySelfManaSync) &&
          mounted &&
          widget.isPartyBattle) {
        unawaited(_syncPartySelfResourcesToBackend());
      }
    }
  }

  Future<void> _endSharedPartyBattleOnServer() async {
    if (!widget.isPartyBattle) return;
    final monsterId = (_partyBattleMonsterId ?? '').trim();
    if (monsterId.isEmpty) return;
    try {
      final poiService = context.read<PoiService>();
      await poiService.endMonsterBattle(
        monsterId,
        outcome: MonsterBattleOutcome.defeat.name,
      );
    } catch (_) {
      // Best-effort: battle may already be ended by another participant.
    }
  }

  Future<void> _syncPartyParticipantsFromStatus(
    Map<String, dynamic> status,
  ) async {
    if (!widget.isPartyBattle) return;
    _ensureSelfPartyAlly();
    _seedPartyAlliesFromPartyProvider();

    final participantsRaw = status['participants'];
    if (participantsRaw is! List) return;
    final activeParticipantIDs = <String>{};
    final now = DateTime.now();
    final idsToRefresh = <String>[];
    for (final raw in participantsRaw) {
      final mapped = raw is Map<String, dynamic>
          ? raw
          : (raw is Map
                ? Map<String, dynamic>.from(raw)
                : const <String, dynamic>{});
      final userId = (mapped['userId']?.toString() ?? '').trim();
      if (userId.isEmpty || userId == _selfUserId) continue;
      activeParticipantIDs.add(userId);
      final existing = _partyAllies.indexWhere((ally) => ally.userId == userId);
      if (existing < 0) {
        _partyAllies.add(
          _PartyAllyState(
            userId: userId,
            name: 'Party Member',
            iconUrl: '',
            level: 1,
            currentHealth: 1,
            maxHealth: 1,
            currentMana: 0,
            maxMana: 0,
            isSelf: false,
          ),
        );
      }
      final fetchedAt = _partyAllyFetchedAt[userId];
      if (fetchedAt == null ||
          now.difference(fetchedAt) >= const Duration(seconds: 2)) {
        idsToRefresh.add(userId);
      }
    }
    if (_selfUserId.isNotEmpty) {
      activeParticipantIDs.add(_selfUserId);
      _partyAllyFetchedAt[_selfUserId] = now;
    }

    final uniqueRefreshIDs = idsToRefresh.toSet().toList(growable: false);
    if (uniqueRefreshIDs.isEmpty) {
      if (!mounted) return;
      setState(() {
        _activePartyParticipantIds
          ..clear()
          ..addAll(activeParticipantIDs);
        for (final ally in _partyAllies) {
          if (ally.isSelf) continue;
          if (_activePartyParticipantIds.contains(ally.userId)) continue;
          ally.currentHealth = 0;
          ally.currentMana = 0;
        }
      });
      return;
    }
    final poiService = context.read<PoiService>();
    final profiles = await Future.wait(
      uniqueRefreshIDs.map((userId) async {
        try {
          final profile = await poiService.getUserCharacterProfile(userId);
          return <String, dynamic>{'userId': userId, 'profile': profile};
        } catch (_) {
          return <String, dynamic>{'userId': userId, 'profile': null};
        }
      }),
    );
    if (!mounted) return;
    setState(() {
      _activePartyParticipantIds
        ..clear()
        ..addAll(activeParticipantIDs);
      for (final entry in profiles) {
        final userId = (entry['userId']?.toString() ?? '').trim();
        if (userId.isEmpty) continue;
        _partyAllyFetchedAt[userId] = now;
        final profileRaw = entry['profile'];
        if (profileRaw is Map<String, dynamic>) {
          _upsertPartyAllyFromProfile(profileRaw);
        } else if (profileRaw is Map) {
          _upsertPartyAllyFromProfile(Map<String, dynamic>.from(profileRaw));
        }
      }
      for (final ally in _partyAllies) {
        if (ally.isSelf) continue;
        if (_activePartyParticipantIds.contains(ally.userId)) continue;
        ally.currentHealth = 0;
        ally.currentMana = 0;
      }
    });
  }

  void _upsertPartyAllyFromProfile(Map<String, dynamic> profile) {
    final userRaw = profile['user'];
    final statsRaw = profile['stats'];
    if (userRaw is! Map && userRaw is! Map<String, dynamic>) return;
    if (statsRaw is! Map && statsRaw is! Map<String, dynamic>) return;
    final user = userRaw is Map<String, dynamic>
        ? userRaw
        : Map<String, dynamic>.from(userRaw as Map);
    final stats = statsRaw is Map<String, dynamic>
        ? statsRaw
        : Map<String, dynamic>.from(statsRaw as Map);

    final userId = (user['id']?.toString() ?? '').trim();
    if (userId.isEmpty) return;
    final existing = _partyAllies.indexWhere((ally) => ally.userId == userId);
    if (existing < 0) return;

    final ally = _partyAllies[existing];
    final prevHealth = ally.currentHealth;
    final prevMana = ally.currentMana;
    final isInitialSnapshot =
        !ally.isSelf &&
        ally.level <= 1 &&
        ally.maxHealth <= 1 &&
        ally.currentHealth <= 1 &&
        ally.maxMana == 0 &&
        ally.currentMana == 0;
    ally.name = _userDisplayNameFromRaw(user);
    final icon = (user['profilePictureUrl']?.toString() ?? '').trim();
    if (icon.isNotEmpty) {
      ally.iconUrl = icon;
    }
    ally.level = _parseIntValue(stats['level'], fallback: ally.level);
    ally.maxHealth = math.max(
      1,
      _parseIntValue(stats['maxHealth'], fallback: ally.maxHealth),
    );
    ally.currentHealth = _parseIntValue(
      stats['health'],
      fallback: ally.currentHealth,
    ).clamp(0, ally.maxHealth).toInt();
    ally.maxMana = math.max(
      0,
      _parseIntValue(stats['maxMana'], fallback: ally.maxMana),
    );
    ally.currentMana = _parseIntValue(
      stats['mana'],
      fallback: ally.currentMana,
    ).clamp(0, ally.maxMana).toInt();

    if (ally.isSelf) return;
    if (isInitialSnapshot) return;
    if (prevHealth != ally.currentHealth) {
      final delta = ally.currentHealth - prevHealth;
      if (delta > 0) {
        _battleLog.add('${ally.name} recovers $delta HP.');
      } else {
        _battleLog.add(
          '${ally.name} takes ${delta.abs()} damage from battle effects.',
        );
      }
    }
    if (prevMana != ally.currentMana) {
      final delta = ally.currentMana - prevMana;
      if (delta > 0) {
        _battleLog.add('${ally.name} recovers $delta MP.');
      } else {
        _battleLog.add('${ally.name} spends ${delta.abs()} MP.');
      }
    }
  }

  void _updatePartyTurnOrderFromStatus(Map<String, dynamic> status) {
    if (!widget.isPartyBattle) return;
    final turnOrderRaw = status['turnOrder'];
    if (turnOrderRaw is! List || turnOrderRaw.isEmpty) return;
    final battleRaw = status['battle'];
    final battle = battleRaw is Map<String, dynamic>
        ? battleRaw
        : (battleRaw is Map ? Map<String, dynamic>.from(battleRaw) : null);
    final serverTurnIndex = _parseIntValue(battle?['turnIndex'], fallback: 0);

    final entries = <_TurnOrderEntry>[];
    for (final raw in turnOrderRaw) {
      final mapped = raw is Map<String, dynamic>
          ? raw
          : (raw is Map
                ? Map<String, dynamic>.from(raw)
                : const <String, dynamic>{});
      final entityType = (mapped['entityType']?.toString() ?? '')
          .trim()
          .toLowerCase();
      if (entityType == 'user') {
        final userId = (mapped['userId']?.toString() ?? '').trim();
        if (userId.isEmpty) continue;
        final allyIndex = _partyAllies.indexWhere(
          (ally) => ally.userId == userId,
        );
        if (allyIndex < 0) continue;
        final ally = _partyAllies[allyIndex];
        entries.add(
          _TurnOrderEntry(
            iconUrl: ally.iconUrl,
            currentHealth: ally.currentHealth,
            maxHealth: ally.maxHealth,
            label: _shortTurnLabel(ally.name),
            fallbackIcon: Icons.person,
            userId: userId,
            allyIndex: allyIndex,
            isSelf: ally.isSelf,
          ),
        );
        continue;
      }
      if (entityType == 'monster') {
        final monsterId = (mapped['monsterId']?.toString() ?? '').trim();
        var enemyIndex = -1;
        if (monsterId.isNotEmpty) {
          enemyIndex = _enemies.indexWhere(
            (enemy) => enemy.monster.id == monsterId,
          );
        }
        if (enemyIndex < 0) {
          enemyIndex = _resolveTargetEnemyIndex(_activeEnemyIndex) ?? 0;
        }
        if (enemyIndex < 0 || enemyIndex >= _enemies.length) continue;
        final enemy = _enemies[enemyIndex];
        entries.add(
          _TurnOrderEntry(
            iconUrl: enemy.monster.thumbnailUrl.isNotEmpty
                ? enemy.monster.thumbnailUrl
                : enemy.monster.imageUrl,
            currentHealth: enemy.currentHealth,
            maxHealth: enemy.maxHealth,
            label: _shortTurnLabel(enemy.monster.name),
            fallbackIcon: Icons.pets,
            enemyIndex: enemyIndex,
          ),
        );
      }
    }
    if (entries.isNotEmpty) {
      _partyTurnOrder = entries;
      _partyTurnIndex = serverTurnIndex.clamp(0, entries.length - 1).toInt();
    }
  }

  void _replaceEnemySnapshot(
    int index,
    Monster monster, {
    int? currentHealth,
    int? currentMana,
  }) {
    if (index < 0 || index >= _enemies.length) return;
    final existing = _enemies[index];
    _enemies[index] = _EncounterEnemyState(
      monster: monster,
      currentHealth: (currentHealth ?? existing.currentHealth).clamp(
        0,
        monster.maxHealth,
      ),
      currentMana: (currentMana ?? existing.currentMana).clamp(
        0,
        monster.maxMana,
      ),
      statuses: monster.statuses.isNotEmpty
          ? List<MonsterStatus>.from(monster.statuses)
          : List<MonsterStatus>.from(existing.statuses),
      cooldownExpiresAtByAbilityId: Map<String, DateTime>.from(
        existing.cooldownExpiresAtByAbilityId,
      ),
    );
  }

  Future<void> _syncPartyBattleState() async {
    if (!widget.isPartyBattle || _battleOver || _partyBattleSyncInFlight) {
      return;
    }
    final battleId = (_partyBattleId ?? '').trim();
    final monsterId = _partyBattleMonsterId;
    if ((monsterId == null || monsterId.isEmpty) && battleId.isEmpty) return;
    _partyBattleSyncInFlight = true;
    try {
      final poiService = context.read<PoiService>();
      final status = battleId.isNotEmpty
          ? await poiService.getMonsterBattleStatusById(battleId)
          : await poiService.getMonsterBattleStatus(monsterId!);
      _cacheVictoryRewardsFromPayload(status);
      if (!mounted || _battleOver) return;
      await _syncPartyParticipantsFromStatus(status);
      if (!mounted || _battleOver) return;
      _syncEnemyStatusesFromPayload(status);
      final battle = _battleMapFromPayload(status);
      final deficit = _parseIntValue(battle?['monsterHealthDeficit']);
      final endedAtRaw = (battle?['endedAt']?.toString() ?? '').trim();
      final partyHasLivingParticipant = _partyHasLivingParticipantInPayload(
        status,
      );
      if (_enemies.isNotEmpty) {
        final primaryEnemy = _enemies.first;
        var nextHealth = math.max(0, primaryEnemy.maxHealth - deficit);
        Monster? authoritativeMonster;
        if (endedAtRaw.isEmpty && nextHealth <= 0) {
          final refreshedMonsterId = primaryEnemy.monster.id.trim();
          if (refreshedMonsterId.isNotEmpty) {
            authoritativeMonster = await poiService.getMonsterById(
              refreshedMonsterId,
            );
            if (authoritativeMonster != null) {
              nextHealth = math.max(
                0,
                authoritativeMonster.maxHealth - deficit,
              );
              debugPrint(
                '[combat][_syncPartyBattleState] authoritative monster refresh '
                'monster=${authoritativeMonster.id} '
                'maxHealth=${authoritativeMonster.maxHealth} '
                'deficit=$deficit nextHealth=$nextHealth',
              );
            }
          }
        }
        setState(() {
          _applyParticipantResourcesFromPayload(status);
          _syncSelfAllyFromLocalResources();
          _updatePartyTurnOrderFromStatus(status);
          final consumedLastAction = _consumePartyLastActionFromPayload(status);
          if (_lastPartyMonsterHealthDeficit >= 0 &&
              deficit > _lastPartyMonsterHealthDeficit) {
            final rawDelta = deficit - _lastPartyMonsterHealthDeficit;
            final attributedToLocal = math.min(_pendingLocalDamage, rawDelta);
            _pendingLocalDamage -= attributedToLocal;
            final remoteDelta = rawDelta - attributedToLocal;
            if (remoteDelta > 0 && !consumedLastAction) {
              _battleLog.add(
                'A party member hits ${primaryEnemy.monster.name} for $remoteDelta damage.',
              );
            }
          }
          _lastPartyMonsterHealthDeficit = deficit;
          if (authoritativeMonster != null &&
              (authoritativeMonster.maxHealth != primaryEnemy.maxHealth ||
                  authoritativeMonster.level != primaryEnemy.monster.level ||
                  authoritativeMonster.name != primaryEnemy.monster.name)) {
            _replaceEnemySnapshot(
              0,
              authoritativeMonster,
              currentHealth: nextHealth,
              currentMana: authoritativeMonster.mana,
            );
            _ensureSelectedEnemyIsAlive();
          } else if (primaryEnemy.currentHealth != nextHealth) {
            primaryEnemy.currentHealth = nextHealth;
            _ensureSelectedEnemyIsAlive();
          }
          _updatePartyTurnOrderFromStatus(status);
          _syncSelfAllyFromLocalResources();
          if (_activeAllyIndex < 0 || _activeAllyIndex >= _partyAllies.length) {
            _activeAllyIndex = 0;
          }
          _refreshPartyTurnFlag();
        });
        if (endedAtRaw.isNotEmpty && !_battleOver) {
          final didWin =
              _aliveEnemies.isEmpty ||
              (widget.isPartyBattle && partyHasLivingParticipant);
          final didLose = widget.isPartyBattle
              ? !partyHasLivingParticipant
              : _playerHealth <= 0;
          if (didWin) {
            await _finishBattle(
              MonsterBattleOutcome.victory,
              _enemies.length == 1
                  ? 'Your party defeated ${_enemies.first.monster.name}!'
                  : 'Your party won the battle.',
            );
          } else if (didLose) {
            await _finishBattle(
              MonsterBattleOutcome.defeat,
              'Your party has been defeated.',
            );
          } else {
            await _finishBattle(
              MonsterBattleOutcome.escaped,
              'Party battle ended.',
            );
          }
          return;
        }
        if (_allPartyMembersDefeated() && !_battleOver) {
          await _finishBattle(
            MonsterBattleOutcome.defeat,
            'Your party has been defeated.',
          );
          return;
        }
        if (nextHealth > 0 &&
            _isPartyMonsterTurn() &&
            !_partyMonsterTurnInFlight &&
            !_busy) {
          await _runPartyMonsterTurn();
          if (_battleOver) return;
        }
      }
    } catch (error, stackTrace) {
      debugPrint('[combat][_syncPartyBattleState] error=$error\n$stackTrace');
      if (!mounted || _battleOver) return;
      if (_isBattleStatusNotFoundError(error)) {
        await _finishAfterBattleStatusGone();
      }
    } finally {
      _partyBattleSyncInFlight = false;
    }
  }

  Future<Map<String, dynamic>?> _reportPartyBattleDamage(
    int damage, {
    Map<String, dynamic>? action,
  }) async {
    if (!widget.isPartyBattle || damage <= 0 || _battleOver) return null;
    final battleId = (_partyBattleId ?? '').trim();
    final monsterId = _partyBattleMonsterId;
    if ((monsterId == null || monsterId.isEmpty) && battleId.isEmpty) {
      return null;
    }
    try {
      _pendingLocalDamage += damage;
      final poiService = context.read<PoiService>();
      late final Map<String, dynamic> response;
      if (battleId.isNotEmpty) {
        response = await poiService.applyMonsterBattleDamageById(
          battleId,
          damage,
          action: action,
        );
      } else {
        response = await poiService.applyMonsterBattleDamage(
          monsterId!,
          damage,
          action: action,
        );
      }
      _cacheVictoryRewardsFromPayload(response);
      try {
        await _syncPartyBattleState();
      } catch (_) {
        // Keep the successful action response even if the follow-up sync fails.
      }
      return response;
    } catch (error, stackTrace) {
      debugPrint(
        '[combat][_reportPartyBattleDamage] damage=$damage battleId=$battleId '
        'monsterId=$monsterId error=$error\n$stackTrace',
      );
      // Keep combat moving locally if the sync request fails.
      return null;
    }
  }

  Future<Map<String, dynamic>?> _reportSoloBattleDamage(int damage) async {
    if (widget.isPartyBattle || damage <= 0 || _battleOver) return null;
    final monsterId = (widget.battleMonsterId ?? _activeEnemy?.monster.id ?? '')
        .trim();
    if (monsterId.isEmpty) return null;
    try {
      final poiService = context.read<PoiService>();
      final response = await poiService.applyMonsterBattleDamage(
        monsterId,
        damage,
      );
      _cacheVictoryRewardsFromPayload(response);
      return response;
    } catch (_) {
      // Keep local combat responsive if the server sync fails.
      return null;
    }
  }

  String get _monsterSpriteUrl {
    final enemy = _activeEnemy;
    if (enemy == null) return '';
    return enemy.monster.thumbnailUrl.isNotEmpty
        ? enemy.monster.thumbnailUrl
        : enemy.monster.imageUrl;
  }

  String get _playerSpriteUrl =>
      _hasTrueBackSprite ? _playerBackSpriteUrl : _playerFrontSpriteUrl;

  String _spriteForAlly(_PartyAllyState? ally) {
    if (ally == null) return _playerSpriteUrl;
    if (ally.isSelf) return _playerSpriteUrl;
    return ally.iconUrl;
  }

  List<_TurnOrderEntry> _currentTurnOrder() {
    if (widget.isPartyBattle && _partyTurnOrder.isNotEmpty) {
      final current = _currentPartyTurnEntry();
      if (current == null) {
        return _partyTurnOrder;
      }
      final currentIndex = _partyTurnOrder.indexWhere(
        (entry) => _sameTurnEntry(entry, current),
      );
      if (currentIndex <= 0) {
        return _partyTurnOrder;
      }
      return <_TurnOrderEntry>[
        ..._partyTurnOrder.sublist(currentIndex),
        ..._partyTurnOrder.sublist(0, currentIndex),
      ];
    }
    final selfAllyIndex = widget.isPartyBattle
        ? _partyAllies.indexWhere((ally) => ally.isSelf)
        : -1;
    final entries = <_TurnOrderEntry>[
      _TurnOrderEntry(
        iconUrl: _playerFrontSpriteUrl,
        currentHealth: _playerHealth,
        maxHealth: _playerMaxHealth,
        label: _shortTurnLabel(_playerName),
        fallbackIcon: Icons.person,
        userId: _selfUserId,
        allyIndex: selfAllyIndex >= 0 ? selfAllyIndex : null,
        isSelf: true,
      ),
    ];
    for (var i = 0; i < _enemies.length; i++) {
      final enemy = _enemies[i];
      if (enemy.isDefeated) continue;
      entries.add(
        _TurnOrderEntry(
          iconUrl: enemy.monster.thumbnailUrl.isNotEmpty
              ? enemy.monster.thumbnailUrl
              : enemy.monster.imageUrl,
          currentHealth: enemy.currentHealth,
          maxHealth: enemy.maxHealth,
          label: _shortTurnLabel(enemy.monster.name),
          fallbackIcon: Icons.pets,
          enemyIndex: i,
        ),
      );
    }
    if (entries.length <= 1) return entries;

    var currentIndex = 0;
    if (!_playerTurn && _actingEnemyIndex != null) {
      currentIndex = entries.indexWhere(
        (entry) => entry.enemyIndex == _actingEnemyIndex,
      );
      if (currentIndex < 0) currentIndex = 0;
    }
    if (currentIndex <= 0) return entries;
    return <_TurnOrderEntry>[
      ...entries.sublist(currentIndex),
      ...entries.sublist(0, currentIndex),
    ];
  }

  bool _isCurrentTurnEntry(_TurnOrderEntry entry) {
    if (widget.isPartyBattle) {
      final current = _currentPartyTurnEntry();
      if (current == null) return false;
      if (entry.isEnemy && current.isEnemy) {
        return entry.enemyIndex == current.enemyIndex;
      }
      if (entry.isAlly && current.isAlly) {
        return (entry.userId ?? '') == (current.userId ?? '');
      }
      return false;
    }
    if (_playerTurn) return entry.isSelf;
    if (!entry.isEnemy) return false;
    return entry.enemyIndex != null && entry.enemyIndex == _actingEnemyIndex;
  }

  void _shiftActiveEnemy(int direction) {
    final aliveIndexes = <int>[];
    for (var i = 0; i < _enemies.length; i++) {
      if (!_enemies[i].isDefeated) {
        aliveIndexes.add(i);
      }
    }
    if (aliveIndexes.length <= 1) {
      _ensureSelectedEnemyIsAlive();
      return;
    }

    final currentIndex =
        _resolveTargetEnemyIndex(_activeEnemyIndex) ?? aliveIndexes.first;
    final currentPosition = aliveIndexes.indexOf(currentIndex);
    final safePosition = currentPosition < 0 ? 0 : currentPosition;
    final shiftedPosition = (safePosition + direction) % aliveIndexes.length;
    setState(() {
      _activeEnemyIndex = aliveIndexes[shiftedPosition];
    });
  }

  void _shiftActiveAlly(int direction) {
    if (!widget.isPartyBattle || _partyAllies.length <= 1) return;
    final safeLength = _partyAllies.length;
    final shifted = (_activeAllyIndex + direction) % safeLength;
    setState(() {
      _activeAllyIndex = shifted < 0 ? shifted + safeLength : shifted;
    });
  }

  int _rollDamage(int minDamage, int maxDamage) {
    final minValue = math.max(1, minDamage);
    final maxValue = math.max(minValue, maxDamage);
    if (maxValue <= minValue) return minValue;
    return minValue + _random.nextInt(maxValue - minValue + 1);
  }

  bool _isTechnique(Spell ability) =>
      ability.abilityType.toLowerCase() == 'technique';

  bool _isDamageEffect(String effectType) {
    final normalized = effectType.trim().toLowerCase();
    return normalized.contains('damage') ||
        normalized.contains('harm') ||
        normalized.contains('attack');
  }

  bool _isAllEnemiesDamageEffect(String effectType) {
    final normalized = effectType.trim().toLowerCase();
    return normalized == 'deal_damage_all_enemies' ||
        normalized.contains('all_enemies') ||
        normalized.contains('all enemies') ||
        normalized.contains('aoe') ||
        normalized.contains('area_damage');
  }

  bool _isHealingEffect(String effectType) {
    final normalized = effectType.trim().toLowerCase();
    return normalized.contains('restore_life') ||
        normalized.contains('heal') ||
        normalized.contains('revive');
  }

  int _monsterAbilityTierLevel(Monster monster) {
    final monsterLevel = math.max(1, monster.level);
    final playerLevel = math.max(1, _playerLevel);
    return math.min(monsterLevel, playerLevel + 1);
  }

  int _monsterAbilityDamageDampenerPercent(int abilityLevel) {
    if (abilityLevel <= 5) return 55;
    if (abilityLevel <= 10) return 70;
    if (abilityLevel <= 15) return 85;
    return 100;
  }

  int _applyMonsterAbilityDamageDampener(int damage, int abilityLevel) {
    if (damage <= 0) return 0;
    final dampenerPercent = _monsterAbilityDamageDampenerPercent(abilityLevel);
    if (dampenerPercent >= 100) return damage;
    return math.max(1, ((damage * dampenerPercent) / 100).round());
  }

  int _monsterAbilityDamageCapPercent(Spell ability, int abilityLevel) {
    if (abilityLevel > 10) return 0;
    final targetsAllEnemies = ability.effects.any(
      (effect) => _isAllEnemiesDamageEffect(effect.type),
    );
    final isBossBurst =
        widget.encounter.isBossEncounter && ability.cooldownTurns >= 2;
    if (abilityLevel <= 4) {
      if (targetsAllEnemies) return 12;
      if (isBossBurst) return 35;
      return 25;
    }
    if (targetsAllEnemies) return 15;
    if (isBossBurst) return 40;
    return 30;
  }

  int _capMonsterAbilityDamageAgainstPlayer(
    Monster monster,
    Spell ability,
    int damage,
  ) {
    if (damage <= 0 || _playerMaxHealth <= 0) return math.max(0, damage);
    final capPercent = _monsterAbilityDamageCapPercent(
      ability,
      _monsterAbilityTierLevel(monster),
    );
    if (capPercent <= 0) return damage;
    final maxDamage = math.max(1, (_playerMaxHealth * capPercent) ~/ 100);
    return math.min(damage, maxDamage);
  }

  int _damageHits(int hits, {required bool hasDamage}) {
    if (!hasDamage) return 0;
    return hits > 0 ? hits : 1;
  }

  int _strengthDamageBonus() {
    final strength =
        _playerStats['strength'] ?? CharacterStatsProvider.baseStatValue;
    return math.max(
      0,
      ((strength - CharacterStatsProvider.baseStatValue) / 2).floor(),
    );
  }

  bool _isPhysicalTechnique(Spell ability) => _isTechnique(ability);

  int _playerAttackDamage() {
    final attackProfile = buildHandAttackProfile(_equippedHandItems);
    if (attackProfile.hasWeapon) {
      var totalDamage = 0;
      for (final contribution in attackProfile.contributions) {
        for (var i = 0; i < contribution.swipesPerAttack; i++) {
          totalDamage += _rollDamage(
            contribution.damageMin,
            contribution.damageMax,
          );
        }
      }
      return math.max(1, totalDamage);
    }

    final strength =
        _playerStats['strength'] ?? CharacterStatsProvider.baseStatValue;
    final dexterity =
        _playerStats['dexterity'] ?? CharacterStatsProvider.baseStatValue;
    final strengthBonus = _strengthDamageBonus();
    final minDamage = math.max<int>(
      1,
      _playerLevel + (strength / 4).floor() + strengthBonus,
    );
    final maxDamage = math.max<int>(
      minDamage,
      minDamage + math.max<int>(1, (dexterity / 3).floor()) + strengthBonus,
    );
    return _rollDamage(minDamage, maxDamage);
  }

  Future<void> _ensureEquipmentLoaded() async {
    if (_equipmentLoaded) return;
    final inventoryService = context.read<InventoryService>();
    final equipment = await inventoryService.getEquipment();
    if (!mounted) return;
    setState(() {
      _equippedHandItems = equipment
          .where((entry) => _handEquipmentSlots.contains(entry.slot))
          .where((entry) => entry.inventoryItem != null)
          .toList(growable: false);
      _equipmentLoaded = true;
    });
  }

  int _monsterAttackDamage(Monster monster) {
    final swipes = math.max(1, monster.attackSwipesPerAttack);
    var totalDamage = 0;
    for (var i = 0; i < swipes; i++) {
      totalDamage += _rollDamage(
        monster.attackDamageMin,
        monster.attackDamageMax,
      );
    }
    return math.max(1, totalDamage);
  }

  int _monsterAbilityDamage(Monster monster, Spell ability) {
    final damageEffects = ability.effects
        .where((effect) => _isDamageEffect(effect.type))
        .toList(growable: false);
    if (damageEffects.isEmpty) {
      return 0;
    }
    final explicitDamage = damageEffects.fold<int>(
      0,
      (sum, effect) =>
          sum +
          math.max(0, effect.amount) *
              _damageHits(effect.hits, hasDamage: effect.amount > 0),
    );
    final techniqueBonus = _isTechnique(ability)
        ? math.max(0, (monster.strength - 10) ~/ 2)
        : 0;
    final rawDamage = math.max<int>(
      1,
      explicitDamage + math.max<int>(0, monster.level ~/ 3) + techniqueBonus,
    );
    return _applyMonsterAbilityDamageDampener(
      rawDamage,
      _monsterAbilityTierLevel(monster),
    );
  }

  int _monsterAbilityHealing(Spell ability) {
    return ability.effects
        .where((effect) => _isHealingEffect(effect.type))
        .fold<int>(0, (sum, effect) => sum + math.max(0, effect.amount));
  }

  void _pruneExpiredMonsterAbilityCooldowns(_EncounterEnemyState enemy) {
    final now = DateTime.now();
    final expiredIds = <String>[];
    enemy.cooldownExpiresAtByAbilityId.forEach((abilityId, expiresAt) {
      if (!expiresAt.isAfter(now)) {
        expiredIds.add(abilityId);
      }
    });
    for (final abilityId in expiredIds) {
      enemy.cooldownExpiresAtByAbilityId.remove(abilityId);
    }
  }

  int _monsterAbilityCooldownRemaining(
    _EncounterEnemyState enemy,
    Spell ability,
  ) {
    _pruneExpiredMonsterAbilityCooldowns(enemy);
    final abilityId = ability.id.trim();
    if (abilityId.isEmpty) return 0;
    final expiresAt = enemy.cooldownExpiresAtByAbilityId[abilityId];
    return _cooldownTurnsRemainingFromExpiry(expiresAt);
  }

  bool _monsterCanUseAbility(_EncounterEnemyState enemy, Spell ability) {
    if (_monsterAbilityCooldownRemaining(enemy, ability) > 0) {
      return false;
    }
    final manaCost = _isTechnique(ability) ? 0 : math.max(0, ability.manaCost);
    return manaCost <= enemy.currentMana;
  }

  void _consumeMonsterAbilityResources(
    _EncounterEnemyState enemy,
    Spell ability,
  ) {
    final manaCost = _isTechnique(ability) ? 0 : math.max(0, ability.manaCost);
    if (manaCost > 0) {
      enemy.currentMana = math.max(0, enemy.currentMana - manaCost);
    }
    final cooldownTurns = math.max(0, ability.cooldownTurns);
    final abilityId = ability.id.trim();
    if (cooldownTurns <= 0 || abilityId.isEmpty) {
      return;
    }
    enemy.cooldownExpiresAtByAbilityId[abilityId] = DateTime.now().add(
      _combatTurnDuration * cooldownTurns,
    );
  }

  Spell? _pickMonsterAbility(
    _EncounterEnemyState enemy,
    Monster monster,
    int currentHealth,
    int maxHealth,
  ) {
    final abilities = <Spell>[...monster.spells, ...monster.techniques]
        .where((ability) => _monsterCanUseAbility(enemy, ability))
        .toList(growable: false);
    if (abilities.isEmpty) {
      return null;
    }
    final support = abilities
        .where((ability) => _monsterAbilityHealing(ability) > 0)
        .toList(growable: false);
    final offense = abilities
        .where((ability) => _monsterAbilityDamage(monster, ability) > 0)
        .toList(growable: false);
    final healthRatio = maxHealth <= 0
        ? 1.0
        : currentHealth / math.max(1, maxHealth);

    if (healthRatio <= 0.45 && support.isNotEmpty) {
      support.sort(
        (left, right) => _monsterAbilityHealing(
          right,
        ).compareTo(_monsterAbilityHealing(left)),
      );
      return support.first;
    }
    if (offense.isNotEmpty && _random.nextInt(100) < 55) {
      offense.sort(
        (left, right) => _monsterAbilityDamage(
          monster,
          right,
        ).compareTo(_monsterAbilityDamage(monster, left)),
      );
      return offense.first;
    }
    if (support.isNotEmpty && offense.isEmpty) {
      return support.first;
    }
    if (offense.isNotEmpty) {
      return offense[_random.nextInt(offense.length)];
    }
    return null;
  }

  bool _applySoloMonsterAbility(_EncounterEnemyState enemy) {
    final ability = _pickMonsterAbility(
      enemy,
      enemy.monster,
      enemy.currentHealth,
      enemy.maxHealth,
    );
    if (ability == null) {
      return false;
    }

    final healAmount = _monsterAbilityHealing(ability);
    if (healAmount > 0 && enemy.currentHealth < enemy.maxHealth) {
      setState(() {
        _consumeMonsterAbilityResources(enemy, ability);
        enemy.currentHealth = (enemy.currentHealth + healAmount)
            .clamp(0, enemy.maxHealth)
            .toInt();
        _battleLog.add(
          '${enemy.monster.name} uses ${ability.name} and restores $healAmount HP.',
        );
      });
      unawaited(
        _playSpriteFx(targetMonster: true, amount: healAmount, healing: true),
      );
      return true;
    }

    final damage = _capMonsterAbilityDamageAgainstPlayer(
      enemy.monster,
      ability,
      _monsterAbilityDamage(enemy.monster, ability),
    );
    if (damage <= 0) {
      return false;
    }
    setState(() {
      _consumeMonsterAbilityResources(enemy, ability);
      _playerHealth = math.max(0, _playerHealth - damage);
      _syncSelfAllyFromLocalResources();
      _battleLog.add(
        '${enemy.monster.name} uses ${ability.name} for $damage damage.',
      );
    });
    unawaited(
      _playSpriteFx(targetMonster: false, amount: damage, healing: false),
    );
    return true;
  }

  int _abilityDamage(Spell ability) {
    final damageEffects = ability.effects
        .where((effect) => _isDamageEffect(effect.type))
        .toList(growable: false);
    final healEffects = ability.effects
        .where((effect) => _isHealingEffect(effect.type))
        .toList(growable: false);

    final explicitDamage = damageEffects.fold<int>(
      0,
      (sum, effect) =>
          sum +
          math.max(0, effect.amount) *
              _damageHits(effect.hits, hasDamage: effect.amount > 0),
    );
    if (damageEffects.isNotEmpty) {
      final strengthBonus = _isPhysicalTechnique(ability)
          ? _strengthDamageBonus()
          : 0;
      return explicitDamage +
          math.max<int>(0, _playerLevel ~/ 3) +
          strengthBonus;
    }
    if (healEffects.isNotEmpty) {
      return 0;
    }

    if (_isPhysicalTechnique(ability)) {
      return _playerAttackDamage() + _strengthDamageBonus();
    }

    final intelligence =
        _playerStats['intelligence'] ?? CharacterStatsProvider.baseStatValue;
    final wisdom =
        _playerStats['wisdom'] ?? CharacterStatsProvider.baseStatValue;
    final minDamage = math.max<int>(
      1,
      _playerLevel + ((intelligence + wisdom) ~/ 6),
    );
    final maxDamage = math.max<int>(
      minDamage,
      minDamage + math.max<int>(2, ability.manaCost),
    );
    return _rollDamage(minDamage, maxDamage);
  }

  int _explicitAbilityDamageByTargeting(
    Spell ability, {
    required bool allEnemies,
  }) {
    final damageEffects = ability.effects
        .where((effect) {
          if (!_isDamageEffect(effect.type)) return false;
          final isAllEnemiesEffect = _isAllEnemiesDamageEffect(effect.type);
          return allEnemies ? isAllEnemiesEffect : !isAllEnemiesEffect;
        })
        .toList(growable: false);
    if (damageEffects.isEmpty) return 0;

    final explicitDamage = damageEffects.fold<int>(
      0,
      (sum, effect) =>
          sum +
          math.max(0, effect.amount) *
              _damageHits(effect.hits, hasDamage: effect.amount > 0),
    );
    final strengthBonus = _isPhysicalTechnique(ability)
        ? _strengthDamageBonus()
        : 0;
    return explicitDamage + math.max<int>(0, _playerLevel ~/ 3) + strengthBonus;
  }

  int _abilityHealing(Spell ability) {
    return ability.effects
        .where((effect) => _isHealingEffect(effect.type))
        .fold<int>(0, (sum, effect) => sum + math.max(0, effect.amount));
  }

  Future<void> _loadItemChoices() async {
    setState(() {
      _loadingItems = true;
    });
    try {
      final inventoryService = context.read<InventoryService>();
      final ownedFuture = inventoryService.getOwnedInventoryItems();
      final itemsFuture = inventoryService.getInventoryItems();
      final owned = await ownedFuture;
      final allItems = await itemsFuture;
      if (!mounted) return;

      final itemById = <int, InventoryItem>{
        for (final item in allItems) item.id: item,
      };
      final choices = <_BattleItemChoice>[];
      for (final entry in owned) {
        if (entry.quantity <= 0) continue;
        final item = itemById[entry.inventoryItemId];
        if (item == null) continue;
        if (item.consumeHealthDelta == 0 &&
            item.consumeManaDelta == 0 &&
            item.consumeDealDamage == 0 &&
            item.consumeDealDamageAllEnemies == 0 &&
            item.consumeRevivePartyMemberHealth <= 0 &&
            item.consumeReviveAllDownedPartyMembersHealth <= 0) {
          continue;
        }
        choices.add(
          _BattleItemChoice(
            ownedInventoryItemId: entry.id,
            name: item.name,
            healthDelta: item.consumeHealthDelta,
            manaDelta: item.consumeManaDelta,
            revivePartyMemberHealth: item.consumeRevivePartyMemberHealth,
            reviveAllDownedPartyMembersHealth:
                item.consumeReviveAllDownedPartyMembersHealth,
            dealDamage: item.consumeDealDamage,
            dealDamageHits: _damageHits(
              item.consumeDealDamageHits,
              hasDamage: item.consumeDealDamage > 0,
            ),
            dealDamageAllEnemies: item.consumeDealDamageAllEnemies,
            dealDamageAllEnemiesHits: _damageHits(
              item.consumeDealDamageAllEnemiesHits,
              hasDamage: item.consumeDealDamageAllEnemies > 0,
            ),
            quantity: entry.quantity,
          ),
        );
      }

      setState(() {
        _items = choices;
        _loadingItems = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _items = const [];
        _loadingItems = false;
      });
    }
  }

  void _openMenu(_BattleMenuView view, {String? selectedCommandKey}) {
    setState(() {
      _menuView = view;
      if (selectedCommandKey != null) {
        _selectedCommandKey = selectedCommandKey;
      }
    });
  }

  bool get _isAbilityPickerView =>
      _menuView == _BattleMenuView.spells ||
      _menuView == _BattleMenuView.techniques;

  bool get _showCompactBattleLog => _battleLogCollapsed;

  String _abilityPrefsKey(String prefix) {
    final scopedUserId = _selfUserId.trim();
    if (scopedUserId.isEmpty) return prefix;
    return '$prefix.$scopedUserId';
  }

  String _abilityStorageKey(Spell ability) {
    final id = ability.id.trim();
    if (id.isNotEmpty) return id;
    final type = ability.abilityType.trim().toLowerCase();
    final name = ability.name.trim().toLowerCase();
    return '$type:$name';
  }

  Future<void> _loadAbilityPickerPrefs() async {
    final prefs = await SharedPreferences.getInstance();
    final favoriteIds =
        prefs.getStringList(_abilityPrefsKey(_favoriteAbilityPrefsKeyPrefix)) ??
        const <String>[];
    final recentIds =
        prefs.getStringList(_abilityPrefsKey(_recentAbilityPrefsKeyPrefix)) ??
        const <String>[];
    if (!mounted) return;
    setState(() {
      _favoriteAbilityIds
        ..clear()
        ..addAll(
          favoriteIds
              .map((entry) => entry.trim())
              .where((entry) => entry.isNotEmpty),
        );
      _recentAbilityIds = recentIds
          .map((entry) => entry.trim())
          .where((entry) => entry.isNotEmpty)
          .take(_maxRecentAbilities)
          .toList(growable: true);
    });
  }

  Future<void> _saveFavoriteAbilityPrefs() async {
    final prefs = await SharedPreferences.getInstance();
    final favorites = _favoriteAbilityIds.toList(growable: false)..sort();
    await prefs.setStringList(
      _abilityPrefsKey(_favoriteAbilityPrefsKeyPrefix),
      favorites,
    );
  }

  Future<void> _saveRecentAbilityPrefs() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setStringList(
      _abilityPrefsKey(_recentAbilityPrefsKeyPrefix),
      _recentAbilityIds,
    );
  }

  void _toggleFavoriteAbility(Spell ability) {
    final key = _abilityStorageKey(ability);
    if (key.isEmpty) return;
    setState(() {
      if (_favoriteAbilityIds.contains(key)) {
        _favoriteAbilityIds.remove(key);
      } else {
        _favoriteAbilityIds.add(key);
      }
    });
    unawaited(_saveFavoriteAbilityPrefs());
  }

  void _recordRecentAbility(Spell ability) {
    final key = _abilityStorageKey(ability);
    if (key.isEmpty) return;
    if (!mounted) return;
    setState(() {
      _recentAbilityIds = <String>[
        key,
        ..._recentAbilityIds.where((entry) => entry != key),
      ].take(_maxRecentAbilities).toList(growable: true);
    });
    unawaited(_saveRecentAbilityPrefs());
  }

  bool _isFavoriteAbility(Spell ability) =>
      _favoriteAbilityIds.contains(_abilityStorageKey(ability));

  bool _canUseAbilityNow(Spell ability) {
    if (_isTechnique(ability)) {
      return !_isTechniqueOnCooldown(ability);
    }
    return math.max(0, ability.manaCost) <= _playerMana;
  }

  double _commandPanelHeightFor(double viewportHeight) {
    if (!_isAbilityPickerView) {
      return _rootCommandPanelHeight;
    }
    final targetFraction = _showCompactBattleLog ? 0.38 : 0.3;
    final minHeight = _showCompactBattleLog ? 280.0 : 228.0;
    final maxHeight = _showCompactBattleLog ? 380.0 : 320.0;
    return (viewportHeight * targetFraction)
        .clamp(minHeight, maxHeight)
        .toDouble();
  }

  double get _battleLogHeight =>
      _showCompactBattleLog ? _compactBattleLogHeight : _defaultBattleLogHeight;

  Future<void> _playSpriteFx({
    required bool targetMonster,
    required int amount,
    required bool healing,
  }) async {
    if (amount <= 0 || !mounted) return;

    final flashTint = healing
        ? const Color(0x6647A86C)
        : const Color(0x66AA3A49);
    final floatText = '${healing ? '+' : '-'}$amount';
    final floatColor = healing
        ? const Color(0xFF2B7A4B)
        : const Color(0xFF8C2F39);

    if (targetMonster) {
      _monsterFxNonce += 1;
      final nonce = _monsterFxNonce;
      setState(() {
        _monsterFlashTint = flashTint;
        _monsterFloatText = floatText;
        _monsterFloatColor = floatColor;
      });

      if (!healing) {
        for (final offset in const [8.0, -8.0, 6.0, -6.0, 0.0]) {
          if (!mounted || _monsterFxNonce != nonce) return;
          setState(() {
            _monsterShakeDx = offset;
          });
          await Future<void>.delayed(const Duration(milliseconds: 55));
        }
      }

      await Future<void>.delayed(const Duration(milliseconds: 500));
      if (!mounted || _monsterFxNonce != nonce) return;
      setState(() {
        _monsterFlashTint = null;
        _monsterFloatText = null;
        _monsterShakeDx = 0;
      });
      return;
    }

    _playerFxNonce += 1;
    final nonce = _playerFxNonce;
    setState(() {
      _playerFlashTint = flashTint;
      _playerFloatText = floatText;
      _playerFloatColor = floatColor;
    });

    if (!healing) {
      for (final offset in const [8.0, -8.0, 6.0, -6.0, 0.0]) {
        if (!mounted || _playerFxNonce != nonce) return;
        setState(() {
          _playerShakeDx = offset;
        });
        await Future<void>.delayed(const Duration(milliseconds: 55));
      }
    }

    await Future<void>.delayed(const Duration(milliseconds: 500));
    if (!mounted || _playerFxNonce != nonce) return;
    setState(() {
      _playerFlashTint = null;
      _playerFloatText = null;
      _playerShakeDx = 0;
    });
  }

  Future<void> _resolvePlayerAction({
    required String message,
    required int damageToMonster,
    int damageToAllEnemies = 0,
    int? targetEnemyIndex,
    int playerHealthDelta = 0,
    int playerManaDelta = 0,
    int? allyTargetIndex,
    int allyHealthDelta = 0,
    int allyManaDelta = 0,
    bool? syncPartyHealth,
    bool? syncPartyMana,
    Map<String, dynamic>? partyActionPayload,
    Future<void> Function(Map<String, dynamic>? damageResponse)?
    onAfterTurnSynced,
  }) async {
    if (!_canAct) return;
    try {
      final resolvedTargetIndex = _resolveTargetEnemyIndex(targetEnemyIndex);
      final targetEnemy = resolvedTargetIndex == null
          ? null
          : _enemies[resolvedTargetIndex];
      var targetName = 'the enemy';
      if (damageToAllEnemies > 0) {
        for (final enemy in _enemies) {
          if (enemy.isDefeated) continue;
          enemy.currentHealth = math.max(
            0,
            enemy.currentHealth - damageToAllEnemies,
          );
        }
      }
      if (targetEnemy != null && !targetEnemy.isDefeated) {
        targetName = targetEnemy.monster.name;
        if (damageToMonster > 0) {
          targetEnemy.currentHealth = math.max(
            0,
            targetEnemy.currentHealth - damageToMonster,
          );
        }
      }
      _ensureSelectedEnemyIsAlive();

      setState(() {
        _busy = true;
        _menuView = _BattleMenuView.root;
        _selectedCommandKey = 'root:Attack';
        _playerHealth = (_playerHealth + playerHealthDelta)
            .clamp(0, _playerMaxHealth)
            .toInt();
        _playerMana = (_playerMana + playerManaDelta)
            .clamp(0, _playerMaxMana)
            .toInt();
        if (widget.isPartyBattle &&
            allyTargetIndex != null &&
            allyTargetIndex >= 0 &&
            allyTargetIndex < _partyAllies.length &&
            (allyHealthDelta != 0 || allyManaDelta != 0)) {
          final ally = _partyAllies[allyTargetIndex];
          ally.currentHealth = (ally.currentHealth + allyHealthDelta)
              .clamp(0, ally.maxHealth)
              .toInt();
          ally.currentMana = (ally.currentMana + allyManaDelta)
              .clamp(0, ally.maxMana)
              .toInt();
          if (ally.isSelf) {
            _playerHealth = ally.currentHealth;
            _playerMana = ally.currentMana;
          }
        }
        _syncSelfAllyFromLocalResources();
        if (damageToMonster > 0 && damageToAllEnemies <= 0) {
          _battleLog.add('$message (Target: $targetName).');
        } else {
          _battleLog.add(message);
        }
      });
      if (widget.isPartyBattle &&
          (playerHealthDelta > 0 ||
              (allyTargetIndex != null &&
                  allyTargetIndex >= 0 &&
                  allyTargetIndex < _partyAllies.length &&
                  _partyAllies[allyTargetIndex].isSelf &&
                  allyHealthDelta > 0))) {
        _allowPartySelfHealthIncreaseSync();
      }
      final shouldSyncPartyHealth =
          syncPartyHealth ??
          (playerHealthDelta != 0 ||
              (allyTargetIndex != null &&
                  allyTargetIndex >= 0 &&
                  allyTargetIndex < _partyAllies.length &&
                  _partyAllies[allyTargetIndex].isSelf &&
                  allyHealthDelta != 0));
      final shouldSyncPartyMana =
          syncPartyMana ??
          (playerManaDelta != 0 ||
              (allyTargetIndex != null &&
                  allyTargetIndex >= 0 &&
                  allyTargetIndex < _partyAllies.length &&
                  _partyAllies[allyTargetIndex].isSelf &&
                  allyManaDelta != 0));
      if (widget.isPartyBattle &&
          (shouldSyncPartyHealth || shouldSyncPartyMana)) {
        _requestPartySelfResourceSync(
          syncHealth: shouldSyncPartyHealth,
          syncMana: shouldSyncPartyMana,
        );
      }
      final sharedDamage =
          (math.max(0, damageToMonster) + math.max(0, damageToAllEnemies))
              .toInt();
      Map<String, dynamic>? damageResponse;
      if (sharedDamage > 0) {
        if (widget.isPartyBattle) {
          damageResponse = await _reportPartyBattleDamage(
            sharedDamage,
            action: partyActionPayload,
          );
        } else {
          damageResponse = await _reportSoloBattleDamage(sharedDamage);
        }
        if (!mounted || _battleOver) return;
      }
      if (onAfterTurnSynced != null) {
        await onAfterTurnSynced(damageResponse);
        if (!mounted || _battleOver) return;
      }
      if (_playerHealth <= 0) {
        if (widget.isPartyBattle && !_allPartyMembersDefeated()) {
          setState(() {
            _busy = false;
            _actingEnemyIndex = null;
            _playerTurn = false;
            _selectedCommandKey = 'root:Attack';
            _battleLog.add('You are down. Waiting for your party...');
          });
          return;
        }
        await _finishBattle(
          MonsterBattleOutcome.defeat,
          widget.isPartyBattle
              ? 'Your party has been defeated.'
              : 'You were defeated by battle effects.',
        );
        return;
      }
      final monsterFxAmount = math.max(damageToMonster, damageToAllEnemies);
      if (monsterFxAmount > 0) {
        unawaited(
          _playSpriteFx(
            targetMonster: true,
            amount: monsterFxAmount,
            healing: false,
          ),
        );
      }
      if (playerHealthDelta != 0) {
        unawaited(
          _playSpriteFx(
            targetMonster: false,
            amount: playerHealthDelta.abs(),
            healing: playerHealthDelta > 0,
          ),
        );
      }
      if (widget.isPartyBattle &&
          allyTargetIndex != null &&
          allyTargetIndex >= 0 &&
          allyTargetIndex < _partyAllies.length &&
          _partyAllies[allyTargetIndex].isSelf &&
          allyHealthDelta != 0) {
        unawaited(
          _playSpriteFx(
            targetMonster: false,
            amount: allyHealthDelta.abs(),
            healing: allyHealthDelta > 0,
          ),
        );
      }

      if (!widget.isPartyBattle && _aliveEnemies.isEmpty) {
        await _finishBattle(
          MonsterBattleOutcome.victory,
          _enemies.length == 1
              ? 'You defeated ${_enemies.first.monster.name}!'
              : 'You defeated the entire encounter!',
        );
        return;
      }
      if (widget.isPartyBattle) {
        if (_aliveEnemies.isEmpty) {
          await _syncPartyBattleState();
          if (_battleOver) return;
        }
        setState(() {
          _busy = false;
          _actingEnemyIndex = null;
          _selectedCommandKey = 'root:Attack';
        });
        if (_isPartyMonsterTurn() && !_partyMonsterTurnInFlight && !_busy) {
          await _runPartyMonsterTurn();
        }
        return;
      }

      setState(() {
        _playerTurn = false;
        _actingEnemyIndex = _firstAliveEnemyIndex();
      });
      await _monsterTurn();
    } catch (error, stackTrace) {
      debugPrint(
        '[combat][_resolvePlayerAction] party=${widget.isPartyBattle} '
        'damageToMonster=$damageToMonster damageToAllEnemies=$damageToAllEnemies '
        'playerTurn=$_playerTurn busy=$_busy battleOver=$_battleOver '
        'turnIndex=$_partyTurnIndex turnOrder=${_partyTurnOrder.length} '
        'error=$error\n$stackTrace',
      );
      if (!mounted || _battleOver) return;
      if (widget.isPartyBattle) {
        try {
          await _syncPartyBattleState();
        } catch (_) {
          // Fall through to local unlock if sync recovery fails.
        }
      }
      if (!mounted || _battleOver) return;
      setState(() {
        _busy = false;
        _actingEnemyIndex = null;
        _selectedCommandKey = 'root:Attack';
        _battleLog.add(
          _extractApiErrorMessage(
            error,
            fallback: 'Combat sync hiccuped. Try your action again.',
          ),
        );
        if (!widget.isPartyBattle) {
          _playerTurn = true;
        } else {
          _refreshPartyTurnFlag(announce: false);
        }
      });
    }
  }

  Future<void> _monsterTurn() async {
    await Future<void>.delayed(const Duration(milliseconds: 420));
    if (!mounted || _battleOver) return;
    setState(() {
      _actingEnemyIndex = _firstAliveEnemyIndex();
    });
    for (var i = 0; i < _enemies.length; i++) {
      final enemy = _enemies[i];
      if (enemy.isDefeated) continue;

      setState(() {
        _actingEnemyIndex = i;
      });

      final usedAbility = _applySoloMonsterAbility(enemy);
      if (!usedAbility) {
        final damage = _monsterAttackDamage(enemy.monster);
        final weaponName =
            enemy.monster.weaponInventoryItemName.trim().isNotEmpty
            ? enemy.monster.weaponInventoryItemName.trim()
            : 'its weapon';

        setState(() {
          _playerHealth = math.max(0, _playerHealth - damage);
          _syncSelfAllyFromLocalResources();
          _battleLog.add(
            '${enemy.monster.name} attacks with $weaponName for $damage damage.',
          );
        });
        unawaited(
          _playSpriteFx(targetMonster: false, amount: damage, healing: false),
        );
      }

      if (_playerHealth <= 0) {
        if (widget.isPartyBattle && !_allPartyMembersDefeated()) {
          setState(() {
            _actingEnemyIndex = null;
            _busy = false;
            _playerTurn = false;
            _selectedCommandKey = 'root:Attack';
            _battleLog.add('You are down. Waiting for your party...');
          });
          return;
        }
        await _finishBattle(
          MonsterBattleOutcome.defeat,
          widget.isPartyBattle
              ? 'Your party has been defeated.'
              : 'You were defeated by ${enemy.monster.name}.',
        );
        return;
      }

      await Future<void>.delayed(const Duration(milliseconds: 250));
      if (!mounted || _battleOver) return;
    }

    setState(() {
      final wasPlayerTurn = _playerTurn;
      _actingEnemyIndex = null;
      _busy = false;
      _playerTurn = _playerHealth > 0;
      if (!wasPlayerTurn && _playerTurn) {
        _beginSelfTurn();
      }
      _ensureSelectedEnemyIsAlive();
      _selectedCommandKey = 'root:Attack';
      _battleLog.add(
        _playerHealth > 0
            ? 'Your turn. Choose a command.'
            : 'You are down. Waiting for your party...',
      );
    });
  }

  Future<void> _finishBattle(
    MonsterBattleOutcome outcome,
    String summary,
  ) async {
    if (_battleOver || _endingBattle) return;
    final statsProvider = widget.isPartyBattle
        ? context.read<CharacterStatsProvider>()
        : null;
    _endingBattle = true;
    _pendingPartySelfHealthSync = false;
    _pendingPartySelfManaSync = false;
    _partySelfHealthIncreaseSyncAllowed = false;
    if (widget.isPartyBattle && outcome == MonsterBattleOutcome.defeat) {
      await _endSharedPartyBattleOnServer();
    }
    _partyBattleSyncTimer?.cancel();
    _partyBattleSyncTimer = null;
    setState(() {
      _battleOver = true;
      _busy = false;
      _battleLog.add(summary);
    });
    final useCachedVictoryRewards =
        outcome == MonsterBattleOutcome.victory && _hasCachedVictoryRewards;
    final postBattleHealthRemaining = math.max(1, _playerHealth);
    final rewardExperience = useCachedVictoryRewards
        ? _victoryRewardExperience
        : widget.isPartyBattle
        ? _victoryRewardExperience
        : widget.encounter.totalRewardExperience;
    final rewardGold = useCachedVictoryRewards
        ? _victoryRewardGold
        : widget.isPartyBattle
        ? _victoryRewardGold
        : widget.encounter.totalRewardGold;
    final baseResourcesAwarded = useCachedVictoryRewards
        ? List<Map<String, dynamic>>.from(_victoryBaseResourcesAwarded)
        : widget.isPartyBattle
        ? List<Map<String, dynamic>>.from(_victoryBaseResourcesAwarded)
        : const <Map<String, dynamic>>[];
    final itemsAwarded = useCachedVictoryRewards
        ? List<Map<String, dynamic>>.from(_victoryItemsAwarded)
        : widget.isPartyBattle
        ? List<Map<String, dynamic>>.from(_victoryItemsAwarded)
        : _fallbackVictoryItemsAwarded();
    debugPrint(
      '[monster-rewards][client][finish] '
      'party=${widget.isPartyBattle} outcome=$outcome '
      'encounter=${widget.encounter.id} '
      'rewardExperience=$rewardExperience rewardGold=$rewardGold '
      'itemsAwarded=${itemsAwarded.length}',
    );
    if (widget.isPartyBattle &&
        outcome != MonsterBattleOutcome.defeat &&
        _selfUserId.isNotEmpty &&
        _playerHealth != postBattleHealthRemaining) {
      await statsProvider!.setCombatResources(
        health: postBattleHealthRemaining,
        mana: _playerMana,
        refreshBeforeAdjust: false,
      );
      _playerHealth = postBattleHealthRemaining;
      _syncSelfAllyFromLocalResources();
    }
    await Future<void>.delayed(const Duration(milliseconds: 650));
    if (!mounted) return;
    Navigator.of(context).pop(
      MonsterBattleResult(
        outcome: outcome,
        playerHealthRemaining: postBattleHealthRemaining,
        playerManaRemaining: _playerMana,
        rewardExperience: outcome == MonsterBattleOutcome.victory
            ? rewardExperience
            : 0,
        rewardGold: outcome == MonsterBattleOutcome.victory ? rewardGold : 0,
        baseResourcesAwarded: outcome == MonsterBattleOutcome.victory
            ? baseResourcesAwarded
            : const <Map<String, dynamic>>[],
        itemsAwarded: outcome == MonsterBattleOutcome.victory
            ? itemsAwarded
            : const <Map<String, dynamic>>[],
      ),
    );
  }

  Future<int?> _pickSingleTargetEnemyIndex({
    required String actionLabel,
  }) async {
    final aliveEntries = _enemies
        .asMap()
        .entries
        .where((entry) => !entry.value.isDefeated)
        .toList(growable: false);
    if (aliveEntries.isEmpty) return null;
    if (aliveEntries.length == 1) return aliveEntries.first.key;

    final initialTarget = _resolveTargetEnemyIndex(_activeEnemyIndex);
    final selected = await showDialog<int>(
      context: context,
      builder: (dialogContext) {
        final theme = Theme.of(dialogContext);
        return AlertDialog(
          title: Text('Choose target for $actionLabel'),
          content: SizedBox(
            width: 360,
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxHeight: 320),
              child: ListView.separated(
                shrinkWrap: true,
                itemCount: aliveEntries.length,
                separatorBuilder: (context, index) => const SizedBox(height: 8),
                itemBuilder: (context, index) {
                  final entry = aliveEntries[index];
                  final enemyIndex = entry.key;
                  final enemy = entry.value;
                  final isSelected = enemyIndex == initialTarget;
                  final imageUrl = enemy.monster.thumbnailUrl.isNotEmpty
                      ? enemy.monster.thumbnailUrl
                      : enemy.monster.imageUrl;
                  return InkWell(
                    onTap: () => Navigator.of(dialogContext).pop(enemyIndex),
                    borderRadius: BorderRadius.circular(10),
                    child: Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: isSelected
                            ? const Color(0xFFB5872F).withValues(alpha: 0.14)
                            : theme.colorScheme.surfaceContainerHighest
                                  .withValues(alpha: 0.45),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(
                          color: isSelected
                              ? const Color(0xFFB5872F)
                              : theme.colorScheme.outline.withValues(
                                  alpha: 0.55,
                                ),
                          width: isSelected ? 1.8 : 1.2,
                        ),
                      ),
                      child: Row(
                        children: [
                          SizedBox(
                            width: 40,
                            height: 40,
                            child: ClipRRect(
                              borderRadius: BorderRadius.circular(8),
                              child: imageUrl.isNotEmpty
                                  ? Image.network(
                                      imageUrl,
                                      fit: BoxFit.cover,
                                      errorBuilder:
                                          (context, error, stackTrace) => Icon(
                                            Icons.pets,
                                            size: 22,
                                            color: theme
                                                .colorScheme
                                                .onSurfaceVariant,
                                          ),
                                    )
                                  : Icon(
                                      Icons.pets,
                                      size: 22,
                                      color: theme.colorScheme.onSurfaceVariant,
                                    ),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Text(
                              enemy.monster.name,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: theme.textTheme.titleSmall?.copyWith(
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Text(
                            '${enemy.currentHealth}/${enemy.maxHealth}',
                            style: theme.textTheme.labelMedium?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(dialogContext).pop(),
              child: const Text('Cancel'),
            ),
          ],
        );
      },
    );
    if (!mounted || selected == null) return selected;
    setState(() {
      _activeEnemyIndex = selected;
    });
    return selected;
  }

  Future<_PartyAllyState?> _pickSingleTargetAlly({
    required String actionLabel,
  }) async {
    final options = _supportTargetOptions();
    if (options.isEmpty) return null;
    if (options.length == 1) return options.first;

    final selected = await showDialog<int>(
      context: context,
      builder: (dialogContext) {
        final theme = Theme.of(dialogContext);
        return AlertDialog(
          title: Text('Choose party member for $actionLabel'),
          content: SizedBox(
            width: 360,
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxHeight: 320),
              child: ListView.separated(
                shrinkWrap: true,
                itemCount: options.length,
                separatorBuilder: (context, index) => const SizedBox(height: 8),
                itemBuilder: (context, index) {
                  final ally = options[index];
                  return InkWell(
                    onTap: () => Navigator.of(dialogContext).pop(index),
                    borderRadius: BorderRadius.circular(10),
                    child: Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.surfaceContainerHighest
                            .withValues(alpha: 0.45),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(
                          color: theme.colorScheme.outline.withValues(
                            alpha: 0.55,
                          ),
                          width: 1.2,
                        ),
                      ),
                      child: Row(
                        children: [
                          SizedBox(
                            width: 40,
                            height: 40,
                            child: ClipRRect(
                              borderRadius: BorderRadius.circular(8),
                              child: ally.iconUrl.isNotEmpty
                                  ? Image.network(
                                      ally.iconUrl,
                                      fit: BoxFit.cover,
                                      errorBuilder:
                                          (context, error, stackTrace) => Icon(
                                            Icons.person,
                                            size: 22,
                                            color: theme
                                                .colorScheme
                                                .onSurfaceVariant,
                                          ),
                                    )
                                  : Icon(
                                      Icons.person,
                                      size: 22,
                                      color: theme.colorScheme.onSurfaceVariant,
                                    ),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: Text(
                              ally.name,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: theme.textTheme.titleSmall?.copyWith(
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ),
                          if (ally.maxHealth > 1 || ally.currentHealth > 0) ...[
                            const SizedBox(width: 10),
                            Text(
                              '${ally.currentHealth}/${ally.maxHealth}',
                              style: theme.textTheme.labelMedium?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant,
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(dialogContext).pop(),
              child: const Text('Cancel'),
            ),
          ],
        );
      },
    );
    if (!mounted || selected == null) return null;
    return options[selected];
  }

  Future<void> _attack() async {
    if (!_canAct) return;
    await _ensureEquipmentLoaded();
    if (!mounted || !_canAct) return;
    int? targetIndex;
    if (_aliveEnemies.length > 1) {
      targetIndex = await _pickSingleTargetEnemyIndex(actionLabel: 'Attack');
      if (!mounted || !_canAct || targetIndex == null) return;
    } else {
      targetIndex = _resolveTargetEnemyIndex(_activeEnemyIndex);
    }
    if (targetIndex == null) return;
    final target = _enemies[targetIndex];
    final damage = _playerAttackDamage();
    await _resolvePlayerAction(
      message: '$_playerName attacks ${target.monster.name} for $damage damage',
      damageToMonster: damage,
      targetEnemyIndex: targetIndex,
      partyActionPayload: const <String, dynamic>{'actionType': 'attack'},
      onAfterTurnSynced: (damageResponse) async {
        if (damageResponse != null && damageResponse.isNotEmpty) {
          _advanceLocalTechniqueCooldownsForAction();
        }
        setState(() {
          _applyTurnSyncResults(
            turnResponse: damageResponse,
            targetEnemyIndex: targetIndex,
          );
        });
      },
    );
  }

  Future<void> _useAbility(Spell ability) async {
    if (!_canAct) return;
    final statsProvider = context.read<CharacterStatsProvider>();
    await _ensureEquipmentLoaded();
    if (!mounted || !_canAct) return;
    if (_isTechnique(ability) && _isTechniqueOnCooldown(ability)) {
      final remaining = _techniqueCooldownRemaining(ability);
      setState(() {
        _battleLog.add(
          '${ability.name} is on cooldown for $remaining more turn${remaining == 1 ? '' : 's'}.',
        );
        _menuView = _BattleMenuView.root;
      });
      return;
    }
    final manaCost = _isTechnique(ability) ? 0 : math.max(0, ability.manaCost);
    if (manaCost > _playerMana) {
      setState(() {
        _battleLog.add('Not enough mana for ${ability.name}.');
        _menuView = _BattleMenuView.root;
      });
      return;
    }

    final singleTargetDamage = _explicitAbilityDamageByTargeting(
      ability,
      allEnemies: false,
    );
    final allEnemiesDamage = _explicitAbilityDamageByTargeting(
      ability,
      allEnemies: true,
    );
    final hasStructuredEffects = ability.effects.isNotEmpty;
    final fallbackDamage = singleTargetDamage == 0 && allEnemiesDamage == 0
        ? (hasStructuredEffects ? 0 : _abilityDamage(ability))
        : 0;
    final damage = singleTargetDamage > 0 ? singleTargetDamage : fallbackDamage;
    final healing = _abilityHealing(ability);
    final isSupportAbility = damage <= 0 && allEnemiesDamage <= 0;
    final requiresSupportTarget = ability.effects.any((effect) {
      final type = effect.type.trim().toLowerCase();
      return type == 'restore_life_party_member' ||
          type == 'revive_party_member';
    });
    final parts = <String>['$_playerName uses ${ability.name}'];
    if (damage > 0) parts.add('dealing $damage damage');
    if (allEnemiesDamage > 0) {
      parts.add('dealing $allEnemiesDamage damage to all enemies');
    }
    if (healing > 0) {
      parts.add('restoring $healing HP');
    }
    int? targetIndex;
    int? allyTargetIndex;
    _PartyAllyState? selectedSupportTarget;
    String? allyTargetName;
    final resolvedBattleMonsterId = () {
      if (widget.isPartyBattle) {
        final partyMonsterId = (_partyBattleMonsterId ?? '').trim();
        if (partyMonsterId.isNotEmpty) return partyMonsterId;
      }
      final dialogMonsterId =
          (widget.battleMonsterId ?? _activeEnemy?.monster.id ?? '').trim();
      if (dialogMonsterId.isNotEmpty) return dialogMonsterId;
      return '';
    }();
    if (damage > 0) {
      if (allEnemiesDamage == 0 && _aliveEnemies.length > 1) {
        targetIndex = await _pickSingleTargetEnemyIndex(
          actionLabel: ability.name,
        );
        if (!mounted || !_canAct || targetIndex == null) return;
      } else {
        targetIndex = _resolveTargetEnemyIndex(_activeEnemyIndex);
      }
    }
    if (isSupportAbility) {
      if (requiresSupportTarget) {
        if (widget.isPartyBattle) {
          selectedSupportTarget = await _pickSingleTargetAlly(
            actionLabel: ability.name,
          );
          if (!mounted || !_canAct || selectedSupportTarget == null) return;
        } else {
          for (final ally in _supportTargetOptions()) {
            if (ally.isSelf) {
              selectedSupportTarget = ally;
              break;
            }
          }
          if (selectedSupportTarget == null) {
            setState(() {
              _battleLog.add('Unable to target yourself with ${ability.name}.');
              _menuView = _BattleMenuView.root;
            });
            return;
          }
        }
        if (widget.isPartyBattle) {
          allyTargetIndex = _partyAllies.indexWhere(
            (ally) => ally.userId == selectedSupportTarget!.userId,
          );
          if (allyTargetIndex < 0) {
            allyTargetIndex = null;
          }
        }
      }
      if (selectedSupportTarget != null) {
        allyTargetName = selectedSupportTarget.name;
        parts.add('on $allyTargetName');
      }
      final targetUserId = requiresSupportTarget
          ? selectedSupportTarget?.userId ?? ''
          : '';
      if (ability.id.trim().isNotEmpty) {
        final healthBeforeCast = _playerHealth;
        final manaBeforeCast = _playerMana;
        final result = _isTechnique(ability)
            ? await statsProvider.castTechniqueDetailed(
                ability.id,
                targetUserId: targetUserId.isNotEmpty ? targetUserId : null,
                targetMonsterId: resolvedBattleMonsterId,
                refreshAfterCast: !widget.isPartyBattle,
              )
            : await statsProvider.castSpellDetailed(
                ability.id,
                targetUserId: targetUserId.isNotEmpty ? targetUserId : null,
                targetMonsterId: resolvedBattleMonsterId,
                refreshAfterCast: !widget.isPartyBattle,
              );
        if (!result.isSuccess && (result.error?.trim().isNotEmpty ?? false)) {
          setState(() {
            _battleLog.add(result.error!.trim());
            _menuView = _BattleMenuView.root;
          });
          return;
        }
        _advanceLocalTechniqueCooldownsForAction(
          excludeTechniqueId: _isTechnique(ability) ? ability.id.trim() : null,
        );
        _playerMana = _parseIntValue(
          result.response['currentMana'],
          fallback: _playerMana,
        ).clamp(0, _playerMaxMana).toInt();
        _applySelfHealsFromResponse(result.response);
        if (!widget.isPartyBattle) {
          _playerStatuses = List<CharacterStatus>.from(statsProvider.statuses);
        }
        if (widget.isPartyBattle) {
          debugPrint(
            '[combat][_useAbility][support] ability=${ability.name} '
            'healthBefore=$healthBeforeCast healthAfter=$_playerHealth '
            'manaBefore=$manaBeforeCast manaAfter=$_playerMana '
            'response=${result.response}',
          );
        }
        final didSelfHealthChange = _playerHealth != healthBeforeCast;
        final didManaChange = _playerMana != manaBeforeCast;
        _spells = statsProvider.spells;
        _techniques = statsProvider.techniques;
        _syncTechniqueCooldownsFromAbilities();
        _setTechniqueCooldownFromResponse(ability, result.response);
        _syncSelfAllyFromLocalResources();
        _recordRecentAbility(ability);
        await _resolvePlayerAction(
          message: parts.join(', '),
          damageToMonster: damage,
          damageToAllEnemies: allEnemiesDamage,
          targetEnemyIndex: targetIndex,
          playerHealthDelta: 0,
          playerManaDelta: 0,
          allyTargetIndex: allyTargetIndex,
          allyHealthDelta: 0,
          syncPartyHealth: didSelfHealthChange,
          syncPartyMana: didManaChange,
          onAfterTurnSynced: (damageResponse) async {
            final hasSeparateTurnResponse =
                damageResponse != null && damageResponse.isNotEmpty;
            setState(() {
              _applyTurnSyncResults(
                setupResponse: result.response,
                turnResponse: hasSeparateTurnResponse ? damageResponse : null,
                targetEnemyIndex: targetIndex,
                refreshPlayerStatusesFromProvider: false,
              );
            });
            if (widget.isPartyBattle && !hasSeparateTurnResponse) {
              await _syncPartyBattleState();
            }
          },
        );
        return;
      }
    }
    if (!isSupportAbility && ability.id.trim().isNotEmpty) {
      final healthBeforeCast = _playerHealth;
      final manaBeforeCast = _playerMana;
      final result = _isTechnique(ability)
          ? await statsProvider.castTechniqueDetailed(
              ability.id,
              targetMonsterId: resolvedBattleMonsterId,
              refreshAfterCast: !widget.isPartyBattle,
            )
          : await statsProvider.castSpellDetailed(
              ability.id,
              targetMonsterId: resolvedBattleMonsterId,
              refreshAfterCast: !widget.isPartyBattle,
            );
      if (!result.isSuccess && (result.error?.trim().isNotEmpty ?? false)) {
        setState(() {
          _battleLog.add(result.error!.trim());
          _menuView = _BattleMenuView.root;
        });
        return;
      }
      _advanceLocalTechniqueCooldownsForAction(
        excludeTechniqueId: _isTechnique(ability) ? ability.id.trim() : null,
      );
      _playerMana = _parseIntValue(
        result.response['currentMana'],
        fallback: _playerMana,
      ).clamp(0, _playerMaxMana).toInt();
      _applySelfHealsFromResponse(result.response);
      if (!widget.isPartyBattle) {
        _playerStatuses = List<CharacterStatus>.from(statsProvider.statuses);
      }
      if (widget.isPartyBattle) {
        debugPrint(
          '[combat][_useAbility][offense] ability=${ability.name} '
          'healthBefore=$healthBeforeCast healthAfter=$_playerHealth '
          'manaBefore=$manaBeforeCast manaAfter=$_playerMana '
          'response=${result.response}',
        );
      }
      final didSelfHealthChange = _playerHealth != healthBeforeCast;
      final didManaChange = _playerMana != manaBeforeCast;
      _spells = statsProvider.spells;
      _techniques = statsProvider.techniques;
      _syncTechniqueCooldownsFromAbilities();
      _setTechniqueCooldownFromResponse(ability, result.response);
      _syncSelfAllyFromLocalResources();
      _recordRecentAbility(ability);
      await _resolvePlayerAction(
        message: parts.join(', '),
        damageToMonster: damage,
        damageToAllEnemies: allEnemiesDamage,
        targetEnemyIndex: targetIndex,
        playerHealthDelta: 0,
        playerManaDelta: 0,
        allyTargetIndex: allyTargetIndex,
        allyHealthDelta: 0,
        syncPartyHealth: didSelfHealthChange,
        syncPartyMana: didManaChange,
        partyActionPayload: <String, dynamic>{
          'actionType': 'ability',
          'abilityId': ability.id,
          'abilityName': ability.name,
          'abilityType': ability.abilityType,
          'targetsAllEnemies': allEnemiesDamage > 0,
          if (didSelfHealthChange)
            'heal': math.max(0, _playerHealth - healthBeforeCast),
        },
        onAfterTurnSynced: (damageResponse) async {
          final hasSeparateTurnResponse =
              damageResponse != null && damageResponse.isNotEmpty;
          setState(() {
            _applyTurnSyncResults(
              setupResponse: result.response,
              turnResponse: hasSeparateTurnResponse ? damageResponse : null,
              targetEnemyIndex: targetIndex,
              refreshPlayerStatusesFromProvider: false,
            );
          });
          if (widget.isPartyBattle && !hasSeparateTurnResponse) {
            await _syncPartyBattleState();
          }
        },
      );
      return;
    } else if (_isTechnique(ability)) {
      _setTechniqueCooldown(ability);
    }
    _recordRecentAbility(ability);
    await _resolvePlayerAction(
      message: parts.join(', '),
      damageToMonster: damage,
      damageToAllEnemies: allEnemiesDamage,
      targetEnemyIndex: targetIndex,
      playerHealthDelta: isSupportAbility ? 0 : healing,
      playerManaDelta: isSupportAbility ? 0 : -manaCost,
      allyTargetIndex: allyTargetIndex,
      allyHealthDelta: isSupportAbility ? healing : 0,
    );
  }

  Future<void> _useItem(_BattleItemChoice item) async {
    if (!_canAct) return;
    if (item.quantity <= 0) {
      setState(() {
        _battleLog.add('${item.name} is out of stock.');
        _menuView = _BattleMenuView.root;
      });
      return;
    }

    final parts = <String>['$_playerName uses ${item.name}'];
    final singleTargetDamage = item.dealDamage * item.dealDamageHits;
    final allEnemiesDamage =
        item.dealDamageAllEnemies * item.dealDamageAllEnemiesHits;
    final canTargetAlly =
        widget.isPartyBattle &&
        _partyAllies.length > 1 &&
        (item.healthDelta > 0 ||
            item.manaDelta > 0 ||
            item.revivePartyMemberHealth > 0);
    int? targetIndex;
    if (singleTargetDamage > 0) {
      if (allEnemiesDamage == 0 && _aliveEnemies.length > 1) {
        targetIndex = await _pickSingleTargetEnemyIndex(actionLabel: item.name);
        if (!mounted || !_canAct || targetIndex == null) return;
      } else {
        targetIndex = _resolveTargetEnemyIndex(_activeEnemyIndex);
      }
      if (item.dealDamageHits > 1) {
        parts.add(
          'dealing ${item.dealDamage} damage ${item.dealDamageHits} times',
        );
      } else {
        parts.add('dealing $singleTargetDamage damage');
      }
    }
    if (allEnemiesDamage > 0) {
      if (item.dealDamageAllEnemiesHits > 1) {
        parts.add(
          'dealing ${item.dealDamageAllEnemies} damage to all enemies ${item.dealDamageAllEnemiesHits} times',
        );
      } else {
        parts.add('dealing $allEnemiesDamage damage to all enemies');
      }
    }
    int? allyTargetIndex;
    if (canTargetAlly) {
      final selectedAlly = await _pickSingleTargetAlly(actionLabel: item.name);
      if (!mounted || !_canAct || selectedAlly == null) return;
      allyTargetIndex = _partyAllies.indexWhere(
        (ally) => ally.userId == selectedAlly.userId,
      );
      if (allyTargetIndex < 0) {
        return;
      }
      parts.add('on ${_partyAllies[allyTargetIndex].name}');
    } else if (widget.isPartyBattle &&
        item.revivePartyMemberHealth > 0 &&
        _partyAllies.isNotEmpty) {
      allyTargetIndex = 0;
    }
    final itemTargetUserId =
        (widget.isPartyBattle &&
            allyTargetIndex != null &&
            allyTargetIndex >= 0 &&
            allyTargetIndex < _partyAllies.length)
        ? _partyAllies[allyTargetIndex].userId
        : null;

    try {
      final inventoryService = context.read<InventoryService>();
      await inventoryService.useItem(
        item.ownedInventoryItemId,
        targetUserId: itemTargetUserId,
      );
    } catch (error) {
      if (!mounted) return;
      setState(() {
        _battleLog.add(
          _extractApiErrorMessage(
            error,
            fallback: 'Failed to use ${item.name}.',
          ),
        );
        _menuView = _BattleMenuView.root;
      });
      return;
    }

    final index = _items.indexWhere(
      (entry) => entry.ownedInventoryItemId == item.ownedInventoryItemId,
    );
    if (index >= 0) {
      final next = item.quantity - 1;
      setState(() {
        _items = [
          ..._items.sublist(0, index),
          _items[index].copyWith(quantity: next),
          ..._items.sublist(index + 1),
        ];
      });
    }
    if (item.healthDelta != 0) {
      final hpLabel = item.healthDelta > 0
          ? 'restoring ${item.healthDelta} HP'
          : 'losing ${item.healthDelta.abs()} HP';
      parts.add(hpLabel);
    }
    if (item.manaDelta != 0) {
      final manaLabel = item.manaDelta > 0
          ? 'restoring ${item.manaDelta} MP'
          : 'losing ${item.manaDelta.abs()} MP';
      parts.add(manaLabel);
    }
    var playerHealthDelta = allyTargetIndex == null ? item.healthDelta : 0;
    var allyHealthDelta = allyTargetIndex != null ? item.healthDelta : 0;
    if (item.revivePartyMemberHealth > 0 &&
        item.reviveAllDownedPartyMembersHealth <= 0 &&
        allyTargetIndex != null &&
        allyTargetIndex >= 0 &&
        allyTargetIndex < _partyAllies.length) {
      final ally = _partyAllies[allyTargetIndex];
      if (ally.currentHealth <= 0) {
        final reviveTo = item.revivePartyMemberHealth
            .clamp(0, ally.maxHealth)
            .toInt();
        if (reviveTo > 0) {
          parts.add('reviving ${ally.name} to $reviveTo HP');
          if (allyTargetIndex == 0 && ally.isSelf) {
            playerHealthDelta += reviveTo;
          } else {
            allyHealthDelta += reviveTo;
          }
        }
      }
    }
    if (item.reviveAllDownedPartyMembersHealth > 0 &&
        widget.isPartyBattle &&
        _partyAllies.isNotEmpty) {
      final revivedNames = <String>[];
      final reviveTo = item.reviveAllDownedPartyMembersHealth;
      for (final ally in _partyAllies) {
        if (ally.currentHealth > 0) continue;
        final revivedHp = reviveTo.clamp(0, ally.maxHealth).toInt();
        if (revivedHp <= 0) continue;
        ally.currentHealth = revivedHp;
        if (ally.isSelf) {
          _playerHealth = revivedHp;
        }
        revivedNames.add(ally.name);
      }
      if (revivedNames.isNotEmpty) {
        parts.add(
          'reviving all downed party members to '
          '${item.reviveAllDownedPartyMembersHealth} HP',
        );
        _syncSelfAllyFromLocalResources();
      }
    }
    await _resolvePlayerAction(
      message: '${parts.join(', ')}.',
      damageToMonster: singleTargetDamage,
      damageToAllEnemies: allEnemiesDamage,
      targetEnemyIndex: targetIndex,
      playerHealthDelta: playerHealthDelta,
      playerManaDelta: allyTargetIndex == null ? item.manaDelta : 0,
      allyTargetIndex: allyTargetIndex,
      allyHealthDelta: allyHealthDelta,
      allyManaDelta: allyTargetIndex != null ? item.manaDelta : 0,
      partyActionPayload: <String, dynamic>{
        'actionType': 'item',
        'abilityName': item.name,
        'abilityType': 'item',
        'targetsAllEnemies': allEnemiesDamage > 0,
        if (playerHealthDelta > 0 || allyHealthDelta > 0)
          'heal': math.max(playerHealthDelta, allyHealthDelta),
      },
      onAfterTurnSynced: (damageResponse) async {
        if (damageResponse == null || damageResponse.isEmpty) return;
        setState(() {
          _applyTurnSyncResults(
            turnResponse: damageResponse,
            targetEnemyIndex: targetIndex,
          );
        });
      },
    );
  }

  Future<void> _escape() async {
    if (!_canAct) return;
    if (widget.isPartyBattle) {
      final monsterId = (_partyBattleMonsterId ?? '').trim();
      if (monsterId.isNotEmpty) {
        try {
          final poiService = context.read<PoiService>();
          await poiService.escapeMonsterBattle(monsterId);
        } catch (_) {
          // Best-effort: local escape should still close the dialog.
        }
      }
    }
    await _finishBattle(
      MonsterBattleOutcome.escaped,
      widget.isPartyBattle
          ? 'You escaped. Your party remains engaged in the fight.'
          : 'You escaped from the encounter.',
    );
  }

  Widget _buildHealthBar({
    required ThemeData theme,
    required String label,
    required int current,
    required int max,
    required Color color,
  }) {
    final safeMax = math.max(1, max);
    final value = (current / safeMax).clamp(0.0, 1.0);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            Text(
              label,
              style: theme.textTheme.labelSmall?.copyWith(
                fontWeight: FontWeight.w700,
                letterSpacing: 0.4,
              ),
            ),
            const Spacer(),
            Text(
              '$current/$safeMax',
              textAlign: TextAlign.right,
              style: theme.textTheme.labelMedium?.copyWith(
                fontFeatures: const [FontFeature.tabularFigures()],
              ),
            ),
          ],
        ),
        const SizedBox(height: 2),
        ClipRRect(
          borderRadius: BorderRadius.circular(999),
          child: LinearProgressIndicator(
            value: value,
            minHeight: 8,
            backgroundColor: color.withValues(alpha: 0.2),
            valueColor: AlwaysStoppedAnimation<Color>(color),
          ),
        ),
      ],
    );
  }

  String _combatStatusTickText(_CombatStatusVisual status) {
    final parts = <String>[];
    if (status.damagePerTick > 0) {
      parts.add('-${status.damagePerTick} HP/turn');
    }
    if (status.healthPerTick != 0) {
      final prefix = status.healthPerTick > 0 ? '+' : '';
      parts.add('$prefix${status.healthPerTick} HP/turn');
    }
    if (status.manaPerTick != 0) {
      final prefix = status.manaPerTick > 0 ? '+' : '';
      parts.add('$prefix${status.manaPerTick} MP/turn');
    }
    return parts.join('  ');
  }

  Widget _buildCombatStatusWrap(
    ThemeData theme, {
    required List<_CombatStatusVisual> statuses,
  }) {
    if (statuses.isEmpty) {
      return Text(
        'No active effects.',
        style: theme.textTheme.labelSmall?.copyWith(
          color: theme.colorScheme.onSurfaceVariant,
        ),
      );
    }

    return Wrap(
      spacing: 6,
      runSpacing: 6,
      children: statuses
          .map((status) {
            final accentColor = status.positive
                ? const Color(0xFF2B7A4B)
                : const Color(0xFF8C2F39);
            final tickText = _combatStatusTickText(status);
            return Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
              decoration: BoxDecoration(
                color: accentColor.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: accentColor.withValues(alpha: 0.35)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    _statusDisplayName(status.name, status.effectType),
                    style: theme.textTheme.labelMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                      color: accentColor,
                    ),
                  ),
                  if (tickText.isNotEmpty)
                    Text(
                      tickText,
                      style: theme.textTheme.labelSmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                ],
              ),
            );
          })
          .toList(growable: false),
    );
  }

  Widget _buildStatusPanel({
    required ThemeData theme,
    required String name,
    required int level,
    required int currentHp,
    required int maxHp,
    int? currentMana,
    int? maxMana,
    List<_CombatStatusVisual> statuses = const [],
  }) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: theme.colorScheme.outline, width: 2),
        boxShadow: const [
          BoxShadow(
            color: Color(0x33000000),
            blurRadius: 8,
            offset: Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
              Text('Lv.$level', style: theme.textTheme.labelLarge),
            ],
          ),
          const SizedBox(height: 10),
          _buildHealthBar(
            theme: theme,
            label: 'HP',
            current: currentHp,
            max: maxHp,
            color: const Color(0xFF8C2F39),
          ),
          if (currentMana != null && maxMana != null) ...[
            const SizedBox(height: 8),
            _buildHealthBar(
              theme: theme,
              label: 'MP',
              current: currentMana,
              max: maxMana,
              color: const Color(0xFF355C7D),
            ),
          ],
          const SizedBox(height: 8),
          _buildCombatStatusWrap(theme, statuses: statuses),
        ],
      ),
    );
  }

  Widget _buildTurnOrderStrip(ThemeData theme) {
    final entries = _currentTurnOrder();
    return Container(
      padding: const EdgeInsets.fromLTRB(10, 8, 10, 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: theme.colorScheme.outline, width: 2),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: entries
                  .map((entry) {
                    final isCurrent = _isCurrentTurnEntry(entry);
                    final isSelectedEnemy =
                        entry.enemyIndex != null &&
                        entry.enemyIndex == _activeEnemyIndex &&
                        !entry.isDefeated;
                    final isSelectedAlly =
                        entry.allyIndex != null &&
                        entry.allyIndex == _activeAllyIndex;
                    final onTap = _battleOver
                        ? null
                        : () {
                            setState(() {
                              if (entry.enemyIndex != null &&
                                  !entry.isDefeated) {
                                _activeEnemyIndex = entry.enemyIndex!;
                              } else if (entry.allyIndex != null) {
                                _activeAllyIndex = entry.allyIndex!;
                              }
                            });
                          };
                    return InkWell(
                      onTap: onTap,
                      borderRadius: BorderRadius.circular(10),
                      child: Container(
                        width: 72,
                        margin: const EdgeInsets.only(right: 8),
                        padding: const EdgeInsets.fromLTRB(4, 4, 4, 4),
                        decoration: BoxDecoration(
                          color: isCurrent
                              ? const Color(0xFFB5872F).withValues(alpha: 0.16)
                              : isSelectedEnemy
                              ? const Color(0xFFB5872F).withValues(alpha: 0.1)
                              : isSelectedAlly
                              ? const Color(0xFF4E9B7D).withValues(alpha: 0.16)
                              : theme.colorScheme.surfaceContainerHighest
                                    .withValues(alpha: 0.45),
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(
                            color: isCurrent || isSelectedEnemy
                                ? const Color(0xFFB5872F)
                                : isSelectedAlly
                                ? const Color(0xFF4E9B7D)
                                : theme.colorScheme.outline.withValues(
                                    alpha: 0.55,
                                  ),
                            width: isCurrent
                                ? 2
                                : (isSelectedEnemy || isSelectedAlly
                                      ? 1.8
                                      : 1.1),
                          ),
                        ),
                        child: Column(
                          children: [
                            SizedBox(
                              width: 38,
                              height: 38,
                              child: ClipRRect(
                                borderRadius: BorderRadius.circular(8),
                                child: entry.iconUrl.isNotEmpty
                                    ? Image.network(
                                        entry.iconUrl,
                                        fit: BoxFit.cover,
                                        color: entry.isDefeated
                                            ? theme.colorScheme.onSurface
                                                  .withValues(alpha: 0.55)
                                            : null,
                                        colorBlendMode: entry.isDefeated
                                            ? BlendMode.saturation
                                            : null,
                                        errorBuilder:
                                            (context, error, stackTrace) =>
                                                Icon(
                                                  entry.fallbackIcon,
                                                  size: 24,
                                                ),
                                      )
                                    : Icon(entry.fallbackIcon, size: 24),
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              entry.label,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: theme.textTheme.labelSmall?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant,
                                fontSize: 10,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              '${entry.currentHealth}/${entry.maxHealth}',
                              style: theme.textTheme.labelSmall?.copyWith(
                                color: theme.colorScheme.onSurfaceVariant,
                                fontSize: 10,
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  })
                  .toList(growable: false),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSpriteBox({
    required ThemeData theme,
    required String imageUrl,
    required IconData fallbackIcon,
    required double size,
    Color? flashTint,
    bool flipHorizontally = false,
  }) {
    final child = imageUrl.isNotEmpty
        ? Image.network(
            imageUrl,
            fit: BoxFit.contain,
            errorBuilder: (context, error, stackTrace) =>
                Icon(fallbackIcon, size: 84),
          )
        : Icon(fallbackIcon, size: 84);

    return SizedBox(
      width: size,
      height: size,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: theme.colorScheme.surface.withValues(alpha: 0.9),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: theme.colorScheme.outline.withValues(alpha: 0.6),
            width: 2,
          ),
        ),
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Stack(
            fit: StackFit.expand,
            children: [
              Center(
                child: flipHorizontally
                    ? Transform.flip(flipX: true, child: child)
                    : child,
              ),
              if (flashTint != null)
                DecoratedBox(
                  decoration: BoxDecoration(
                    color: flashTint,
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFloatingFxLabel({
    required String text,
    required Color color,
    required int nonce,
  }) {
    return TweenAnimationBuilder<double>(
      key: ValueKey<String>('fx-$nonce-$text'),
      tween: Tween<double>(begin: 0, end: 1),
      duration: const Duration(milliseconds: 520),
      builder: (context, value, child) {
        final opacity = (1 - (value * 0.95)).clamp(0.0, 1.0);
        final yOffset = -26.0 * value;
        return Transform.translate(
          offset: Offset(0, yOffset),
          child: Opacity(opacity: opacity, child: child),
        );
      },
      child: Text(
        text,
        style: Theme.of(context).textTheme.titleMedium?.copyWith(
          color: color,
          fontWeight: FontWeight.w800,
          shadows: const [
            Shadow(
              blurRadius: 3,
              color: Color(0x66000000),
              offset: Offset(0, 1),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSpriteStage({
    required ThemeData theme,
    required String imageUrl,
    required IconData fallbackIcon,
    required double size,
    required double shakeDx,
    required Color? flashTint,
    required String? floatingText,
    required Color floatingTextColor,
    required int fxNonce,
    bool flipHorizontally = false,
  }) {
    final stageWidth = size + 28;
    final stageHeight = size + 34;
    return SizedBox(
      width: stageWidth,
      height: stageHeight,
      child: Stack(
        clipBehavior: Clip.none,
        alignment: Alignment.bottomCenter,
        children: [
          Positioned(
            bottom: 2,
            child: Container(
              width: size * 0.72,
              height: size * 0.16,
              decoration: BoxDecoration(
                color: theme.colorScheme.onSurface.withValues(alpha: 0.16),
                borderRadius: BorderRadius.circular(999),
              ),
            ),
          ),
          Positioned(
            bottom: 10,
            child: Transform.translate(
              offset: Offset(shakeDx, 0),
              child: _buildSpriteBox(
                theme: theme,
                imageUrl: imageUrl,
                fallbackIcon: fallbackIcon,
                size: size,
                flashTint: flashTint,
                flipHorizontally: flipHorizontally,
              ),
            ),
          ),
          if (floatingText != null)
            Positioned(
              top: -4,
              child: _buildFloatingFxLabel(
                text: floatingText,
                color: floatingTextColor,
                nonce: fxNonce,
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildEnemyCycleChevron({
    required ThemeData theme,
    required IconData icon,
    required VoidCallback? onPressed,
  }) {
    final enabled = onPressed != null;
    return Container(
      width: 30,
      height: 30,
      decoration: BoxDecoration(
        color: enabled
            ? theme.colorScheme.surface.withValues(alpha: 0.9)
            : theme.colorScheme.surfaceContainerHighest.withValues(alpha: 0.88),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(
          color: enabled
              ? const Color(0xFFB5872F).withValues(alpha: 0.8)
              : theme.colorScheme.outline.withValues(alpha: 0.55),
          width: enabled ? 1.6 : 1.2,
        ),
        boxShadow: const [
          BoxShadow(
            color: Color(0x33000000),
            blurRadius: 5,
            offset: Offset(0, 2),
          ),
        ],
      ),
      child: IconButton(
        onPressed: onPressed,
        icon: Icon(
          icon,
          size: 18,
          color: enabled
              ? theme.colorScheme.onSurface
              : theme.colorScheme.onSurfaceVariant.withValues(alpha: 0.6),
        ),
        padding: EdgeInsets.zero,
        splashRadius: 16,
        visualDensity: VisualDensity.compact,
      ),
    );
  }

  Widget _buildEnemyStageWithChevrons({
    required ThemeData theme,
    required double spriteSize,
  }) {
    final canCycle = !_battleOver && _aliveEnemies.length > 1;
    final stageWidth = spriteSize + 28;
    final stageHeight = spriteSize + 34;
    return SizedBox(
      width: stageWidth,
      height: stageHeight,
      child: Stack(
        alignment: Alignment.center,
        clipBehavior: Clip.none,
        children: [
          _buildSpriteStage(
            theme: theme,
            imageUrl: _monsterSpriteUrl,
            fallbackIcon: Icons.pets,
            size: spriteSize,
            shakeDx: _monsterShakeDx,
            flashTint: _monsterFlashTint,
            floatingText: _monsterFloatText,
            floatingTextColor: _monsterFloatColor,
            fxNonce: _monsterFxNonce,
            flipHorizontally: true,
          ),
          Positioned(
            left: 2,
            child: _buildEnemyCycleChevron(
              theme: theme,
              icon: Icons.chevron_left,
              onPressed: canCycle ? () => _shiftActiveEnemy(-1) : null,
            ),
          ),
          Positioned(
            right: 2,
            child: _buildEnemyCycleChevron(
              theme: theme,
              icon: Icons.chevron_right,
              onPressed: canCycle ? () => _shiftActiveEnemy(1) : null,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAllyStageWithChevrons({
    required ThemeData theme,
    required double spriteSize,
    required _PartyAllyState? activeAlly,
  }) {
    final canCycle =
        !_battleOver && widget.isPartyBattle && _partyAllies.length > 1;
    final stageWidth = spriteSize + 28;
    final stageHeight = spriteSize + 34;
    final spriteUrl = _spriteForAlly(activeAlly);
    final shouldFlip = activeAlly?.isSelf == true
        ? (!_hasTrueBackSprite && spriteUrl.isNotEmpty)
        : false;
    return SizedBox(
      width: stageWidth,
      height: stageHeight,
      child: Stack(
        alignment: Alignment.center,
        clipBehavior: Clip.none,
        children: [
          _buildSpriteStage(
            theme: theme,
            imageUrl: spriteUrl,
            fallbackIcon: Icons.person,
            size: spriteSize,
            shakeDx: _playerShakeDx,
            flashTint: _playerFlashTint,
            floatingText: _playerFloatText,
            floatingTextColor: _playerFloatColor,
            fxNonce: _playerFxNonce,
            flipHorizontally: shouldFlip,
          ),
          Positioned(
            left: 2,
            child: _buildEnemyCycleChevron(
              theme: theme,
              icon: Icons.chevron_left,
              onPressed: canCycle ? () => _shiftActiveAlly(-1) : null,
            ),
          ),
          Positioned(
            right: 2,
            child: _buildEnemyCycleChevron(
              theme: theme,
              icon: Icons.chevron_right,
              onPressed: canCycle ? () => _shiftActiveAlly(1) : null,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBattlefield(ThemeData theme) {
    final activeEnemy = _activeEnemy;
    final activeEnemyName = activeEnemy?.monster.name ?? widget.encounter.name;
    final activeEnemyLevel = activeEnemy?.monster.level ?? 1;
    final activeEnemyHp = activeEnemy?.currentHealth ?? 0;
    final activeEnemyMaxHp = activeEnemy?.maxHealth ?? 1;
    final activeEnemyStatuses = activeEnemy == null
        ? const <_CombatStatusVisual>[]
        : activeEnemy.statuses
              .map(_CombatStatusVisual.fromMonsterStatus)
              .toList(growable: false);
    final activeAlly = widget.isPartyBattle
        ? (_activeAlly ?? _selfPartyAlly)
        : _activeAlly;
    final allyName = activeAlly?.name ?? _playerName;
    final allyLevel = activeAlly?.level ?? _playerLevel;
    final allyCurrentHp = activeAlly?.currentHealth ?? _playerHealth;
    final allyMaxHp = activeAlly?.maxHealth ?? _playerMaxHealth;
    final allyCurrentMana = activeAlly?.currentMana ?? _playerMana;
    final allyMaxMana = activeAlly?.maxMana ?? _playerMaxMana;
    final playerStatuses = (activeAlly?.isSelf ?? true)
        ? _playerStatuses
              .map(_CombatStatusVisual.fromCharacterStatus)
              .toList(growable: false)
        : const <_CombatStatusVisual>[];
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(12),
        gradient: const LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [Color(0xFFEDE2C4), Color(0xFFD7C39F)],
        ),
        border: Border.all(color: theme.colorScheme.outline, width: 2),
      ),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final battlefieldWidth = constraints.maxWidth;
          final spriteSize = (battlefieldWidth * 0.38).clamp(145.0, 235.0);
          var statusWidth = (battlefieldWidth * 0.45).clamp(160.0, 300.0);
          final minGap = 24.0;
          final maxStatusWidth = battlefieldWidth - spriteSize - minGap;
          if (statusWidth > maxStatusWidth) {
            statusWidth = maxStatusWidth.clamp(140.0, 300.0);
          }

          return Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SizedBox(
                      width: statusWidth,
                      child: _buildStatusPanel(
                        theme: theme,
                        name: activeEnemyName,
                        level: activeEnemyLevel,
                        currentHp: activeEnemyHp,
                        maxHp: activeEnemyMaxHp,
                        statuses: activeEnemyStatuses,
                      ),
                    ),
                    const Spacer(),
                    _buildEnemyStageWithChevrons(
                      theme: theme,
                      spriteSize: spriteSize,
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    _buildAllyStageWithChevrons(
                      theme: theme,
                      spriteSize: spriteSize,
                      activeAlly: activeAlly,
                    ),
                    const Spacer(),
                    SizedBox(
                      width: statusWidth,
                      child: _buildStatusPanel(
                        theme: theme,
                        name: allyName,
                        level: allyLevel,
                        currentHp: allyCurrentHp,
                        maxHp: allyMaxHp,
                        currentMana: allyCurrentMana,
                        maxMana: allyMaxMana,
                        statuses: playerStatuses,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  Iterable<Spell> _recentAbilitiesInOrder(List<Spell> abilities) sync* {
    for (final recentId in _recentAbilityIds) {
      for (final ability in abilities) {
        if (_abilityStorageKey(ability) == recentId) {
          yield ability;
          break;
        }
      }
    }
  }

  List<_AbilitySection> _abilitySections(
    List<Spell> abilities, {
    required String unavailableTitle,
  }) {
    final seen = <String>{};

    List<Spell> takeUnique(Iterable<Spell> candidates) {
      final next = <Spell>[];
      for (final ability in candidates) {
        final key = _abilityStorageKey(ability);
        if (key.isEmpty || seen.contains(key)) continue;
        seen.add(key);
        next.add(ability);
      }
      return next;
    }

    final favorites = takeUnique(
      abilities.where((ability) => _isFavoriteAbility(ability)),
    );
    final recent = takeUnique(_recentAbilitiesInOrder(abilities));
    final usable = takeUnique(
      abilities.where((ability) => _canUseAbilityNow(ability)),
    );
    final unavailable = takeUnique(
      abilities.where((ability) => !_canUseAbilityNow(ability)),
    );

    final sections = <_AbilitySection>[];
    if (favorites.isNotEmpty) {
      sections.add(_AbilitySection(title: 'Favorites', abilities: favorites));
    }
    if (recent.isNotEmpty) {
      sections.add(_AbilitySection(title: 'Recent', abilities: recent));
    }
    if (usable.isNotEmpty) {
      sections.add(_AbilitySection(title: 'Usable Now', abilities: usable));
    }
    if (unavailable.isNotEmpty) {
      sections.add(
        _AbilitySection(title: unavailableTitle, abilities: unavailable),
      );
    }
    return sections;
  }

  String _abilityCommandKey(
    Spell ability, {
    required List<Spell> source,
    required String prefix,
  }) {
    final key = _abilityStorageKey(ability);
    var index = -1;
    if (key.isNotEmpty) {
      index = source.indexWhere((entry) => _abilityStorageKey(entry) == key);
    }
    if (index < 0) {
      index = source.indexOf(ability);
    }
    if (index < 0) {
      index = 0;
    }
    return '$prefix:$index';
  }

  String _defaultAbilityCommandKey(
    List<Spell> abilities, {
    required String prefix,
    required String unavailableTitle,
  }) {
    final sections = _abilitySections(
      abilities,
      unavailableTitle: unavailableTitle,
    );
    for (final section in sections) {
      if (section.abilities.isNotEmpty) {
        return _abilityCommandKey(
          section.abilities.first,
          source: abilities,
          prefix: prefix,
        );
      }
    }
    return '$prefix:0';
  }

  String _abilityRoleLabel(Spell ability) {
    final effectTypes = ability.effects
        .map((effect) => effect.type.trim().toLowerCase())
        .where((type) => type.isNotEmpty)
        .toSet();
    if (effectTypes.any((type) => type.contains('all_enemies'))) {
      return 'AOE';
    }
    if (effectTypes.any((type) => type.contains('restore'))) {
      return 'Heal';
    }
    if (effectTypes.any((type) => type.contains('revive'))) {
      return 'Revive';
    }
    if (effectTypes.any(
      (type) => type.contains('status') || type.contains('detrimental'),
    )) {
      return 'Status';
    }
    if (effectTypes.any((type) => type.contains('damage'))) {
      return 'Damage';
    }
    return _isTechnique(ability) ? 'Technique' : 'Spell';
  }

  String _abilitySubtitle(Spell ability) {
    final parts = <String>['Lv.${math.max(1, ability.abilityLevel)}'];
    final school = ability.schoolOfMagic.trim();
    if (school.isNotEmpty) {
      parts.add(_humanizeToken(school));
    }
    final role = _abilityRoleLabel(ability);
    if (role.isNotEmpty) {
      parts.add(role);
    }
    final summary = ability.effectText.trim().isNotEmpty
        ? ability.effectText.trim()
        : ability.description.trim();
    if (summary.isNotEmpty) {
      parts.add(summary);
    }
    return parts.join(' | ');
  }

  String? _abilityBadgeLabel(Spell ability) {
    if (_isTechnique(ability)) {
      final remaining = _techniqueCooldownRemaining(ability);
      if (remaining > 0) {
        return '$remaining turn${remaining == 1 ? '' : 's'}';
      }
      final totalCooldown = math.max(0, ability.cooldownTurns);
      if (totalCooldown > 0) {
        return 'Ready';
      }
      return null;
    }
    final manaCost = math.max(0, ability.manaCost);
    if (manaCost <= 0) {
      return 'Free';
    }
    if (manaCost > _playerMana) {
      final shortage = manaCost - _playerMana;
      return 'Need $shortage MP';
    }
    return '$manaCost MP';
  }

  Color _abilityBadgeColor(ThemeData theme, Spell ability) {
    if (_isTechnique(ability)) {
      return _isTechniqueOnCooldown(ability)
          ? const Color(0xFF8C2F39)
          : const Color(0xFF2F6B3D);
    }
    return math.max(0, ability.manaCost) > _playerMana
        ? const Color(0xFF8C2F39)
        : const Color(0xFF355C7D);
  }

  double? _abilityBadgeProgress(Spell ability) {
    if (!_isTechnique(ability)) return null;
    final remaining = _techniqueCooldownRemaining(ability);
    final totalCooldown = math.max(0, ability.cooldownTurns);
    if (totalCooldown <= 0) return null;
    return (totalCooldown - remaining) / totalCooldown;
  }

  Widget _buildSectionHeader(
    ThemeData theme, {
    required String title,
    required int count,
  }) {
    return Padding(
      padding: const EdgeInsets.only(top: 2),
      child: Row(
        children: [
          Text(
            title,
            style: theme.textTheme.labelLarge?.copyWith(
              fontWeight: FontWeight.w800,
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(width: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
            decoration: BoxDecoration(
              color: theme.colorScheme.surfaceContainerHighest.withValues(
                alpha: 0.9,
              ),
              borderRadius: BorderRadius.circular(999),
            ),
            child: Text(
              '$count',
              style: theme.textTheme.labelSmall?.copyWith(
                fontWeight: FontWeight.w700,
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMenuShell({required String title, required Widget child}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            TextButton(
              onPressed: _busy
                  ? null
                  : () => _openMenu(
                      _BattleMenuView.root,
                      selectedCommandKey: 'root:Attack',
                    ),
              child: const Text('Back'),
            ),
            Expanded(
              child: Text(
                title,
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ),
          ],
        ),
        const SizedBox(height: 4),
        Expanded(child: child),
      ],
    );
  }

  Widget _buildAbilityRow({
    required ThemeData theme,
    required Spell ability,
    required List<Spell> source,
    required String prefix,
  }) {
    final isFavorite = _isFavoriteAbility(ability);
    final enabled = _canAct && _canUseAbilityNow(ability);
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          child: _buildCommandButton(
            context: context,
            label: ability.name,
            commandKey: _abilityCommandKey(
              ability,
              source: source,
              prefix: prefix,
            ),
            onPressed: enabled ? () => _useAbility(ability) : null,
            subtitle: _abilitySubtitle(ability),
            subtitleMaxLines: 2,
            badgeLabel: _abilityBadgeLabel(ability),
            badgeColor: _abilityBadgeColor(theme, ability),
            progress: _abilityBadgeProgress(ability),
          ),
        ),
        const SizedBox(width: 6),
        Tooltip(
          message: isFavorite ? 'Remove favorite' : 'Add favorite',
          child: IconButton(
            onPressed: () => _toggleFavoriteAbility(ability),
            icon: Icon(
              isFavorite ? Icons.star : Icons.star_border,
              color: isFavorite
                  ? const Color(0xFFB5872F)
                  : theme.colorScheme.onSurfaceVariant,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildAbilityChoiceList({
    required ThemeData theme,
    required String title,
    required String prefix,
    required List<Spell> abilities,
    required String unavailableTitle,
  }) {
    final sections = _abilitySections(
      abilities,
      unavailableTitle: unavailableTitle,
    );
    if (sections.isEmpty) {
      return _buildMenuShell(
        title: title,
        child: const Center(child: Text('No options available.')),
      );
    }

    final sectionWidgets = <Widget>[];
    for (var sectionIndex = 0; sectionIndex < sections.length; sectionIndex++) {
      final section = sections[sectionIndex];
      if (sectionIndex > 0) {
        sectionWidgets.add(const SizedBox(height: 12));
      }
      sectionWidgets.add(
        _buildSectionHeader(
          theme,
          title: section.title,
          count: section.abilities.length,
        ),
      );
      sectionWidgets.add(const SizedBox(height: 8));
      for (
        var abilityIndex = 0;
        abilityIndex < section.abilities.length;
        abilityIndex++
      ) {
        if (abilityIndex > 0) {
          sectionWidgets.add(const SizedBox(height: 6));
        }
        sectionWidgets.add(
          _buildAbilityRow(
            theme: theme,
            ability: section.abilities[abilityIndex],
            source: abilities,
            prefix: prefix,
          ),
        );
      }
    }

    return _buildMenuShell(
      title: title,
      child: ListView(children: sectionWidgets),
    );
  }

  Widget _buildLogPanel(ThemeData theme, {required bool compact}) {
    String chipText;
    Color chipColor;
    if (_battleOver) {
      chipText = 'Battle Over';
      chipColor = theme.colorScheme.secondary;
    } else if (widget.isPartyBattle) {
      final turn = _currentPartyTurnEntry();
      if (turn == null) {
        chipText = 'Syncing Turn';
        chipColor = theme.colorScheme.primary;
      } else if (turn.isSelf) {
        chipText = 'Your Turn';
        chipColor = const Color(0xFF2B7A4B);
      } else if (turn.isEnemy) {
        chipText = 'Monster Turn';
        chipColor = const Color(0xFF8C2F39);
      } else {
        chipText = 'Party Turn';
        chipColor = theme.colorScheme.primary;
      }
    } else if (_playerTurn) {
      chipText = 'Your Turn';
      chipColor = const Color(0xFF2B7A4B);
    } else {
      chipText = 'Enemy Turn';
      chipColor = const Color(0xFF8C2F39);
    }

    final latestMessage = _battleLog.isEmpty ? '...' : _battleLog.last;

    Widget buildTurnStateButton({required bool collapsed}) {
      return OutlinedButton.icon(
        onPressed: () {
          setState(() {
            _battleLogCollapsed = !_battleLogCollapsed;
          });
        },
        icon: Icon(collapsed ? Icons.expand_more : Icons.expand_less, size: 18),
        label: Text(chipText, maxLines: 1, overflow: TextOverflow.ellipsis),
        style: OutlinedButton.styleFrom(
          minimumSize: const Size(0, 34),
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          tapTargetSize: MaterialTapTargetSize.shrinkWrap,
          visualDensity: VisualDensity.compact,
          foregroundColor: chipColor,
          backgroundColor: chipColor.withValues(alpha: 0.12),
          side: BorderSide(color: chipColor.withValues(alpha: 0.45)),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(999),
          ),
          textStyle: theme.textTheme.labelMedium?.copyWith(
            fontWeight: FontWeight.w700,
          ),
        ),
      );
    }

    return Container(
      width: double.infinity,
      height: _battleLogHeight,
      padding: EdgeInsets.symmetric(horizontal: 12, vertical: compact ? 8 : 12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: theme.colorScheme.outline, width: 2),
      ),
      child: compact
          ? Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Expanded(
                  child: Text(
                    latestMessage,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: theme.textTheme.bodyMedium,
                  ),
                ),
                const SizedBox(width: 10),
                buildTurnStateButton(collapsed: true),
              ],
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(
                      'Combat Log',
                      style: theme.textTheme.labelLarge?.copyWith(
                        fontWeight: FontWeight.w800,
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                    const Spacer(),
                    buildTurnStateButton(collapsed: false),
                  ],
                ),
                const SizedBox(height: 6),
                Expanded(
                  child: _battleLog.isEmpty
                      ? Text('...', style: theme.textTheme.bodyMedium)
                      : ListView.builder(
                          reverse: true,
                          itemCount: _battleLog.length,
                          itemBuilder: (context, index) {
                            final message =
                                _battleLog[_battleLog.length - 1 - index];
                            return Padding(
                              padding: const EdgeInsets.only(bottom: 4),
                              child: Text(
                                message,
                                style: theme.textTheme.bodyMedium,
                              ),
                            );
                          },
                        ),
                ),
              ],
            ),
    );
  }

  Widget _buildCommandButton({
    required BuildContext context,
    required String label,
    required String commandKey,
    required VoidCallback? onPressed,
    String? subtitle,
    String? badgeLabel,
    Color? badgeColor,
    double? progress,
    int labelMaxLines = 1,
    int subtitleMaxLines = 1,
  }) {
    final selected = _selectedCommandKey == commandKey;
    final theme = Theme.of(context);
    final normalizedProgress = progress?.clamp(0.0, 1.0).toDouble();
    final resolvedBadgeColor = badgeColor ?? theme.colorScheme.primary;
    return OutlinedButton(
      onPressed: onPressed == null
          ? null
          : () {
              setState(() {
                _selectedCommandKey = commandKey;
              });
              onPressed();
            },
      style: OutlinedButton.styleFrom(
        padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 10),
        backgroundColor: selected
            ? theme.colorScheme.primary.withValues(alpha: 0.12)
            : Colors.transparent,
        side: BorderSide(
          color: selected
              ? theme.colorScheme.primary
              : theme.colorScheme.outline,
          width: selected ? 2.0 : 1.4,
        ),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  label,
                  maxLines: labelMaxLines,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (badgeLabel != null && badgeLabel.trim().isNotEmpty) ...[
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 3,
                  ),
                  decoration: BoxDecoration(
                    color: resolvedBadgeColor.withValues(alpha: 0.14),
                    borderRadius: BorderRadius.circular(999),
                    border: Border.all(
                      color: resolvedBadgeColor.withValues(alpha: 0.4),
                    ),
                  ),
                  child: Text(
                    badgeLabel,
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: resolvedBadgeColor,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
              ],
            ],
          ),
          if (subtitle != null && subtitle.trim().isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              subtitle,
              maxLines: subtitleMaxLines,
              overflow: TextOverflow.ellipsis,
              style: theme.textTheme.labelSmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ],
          if (normalizedProgress != null) ...[
            const SizedBox(height: 6),
            ClipRRect(
              borderRadius: BorderRadius.circular(999),
              child: LinearProgressIndicator(
                value: normalizedProgress,
                minHeight: 5,
                backgroundColor: theme.colorScheme.surfaceContainerHighest
                    .withValues(alpha: 0.9),
                valueColor: AlwaysStoppedAnimation<Color>(resolvedBadgeColor),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildCommandMenu(ThemeData theme) {
    switch (_menuView) {
      case _BattleMenuView.root:
        return GridView.count(
          physics: const NeverScrollableScrollPhysics(),
          crossAxisCount: 3,
          mainAxisSpacing: 8,
          crossAxisSpacing: 8,
          childAspectRatio: 2.1,
          children: [
            _buildCommandButton(
              context: context,
              label: 'Attack',
              commandKey: 'root:Attack',
              onPressed: _canAct ? _attack : null,
            ),
            _buildCommandButton(
              context: context,
              label: 'Spell',
              commandKey: 'root:Spell',
              onPressed: _canAct
                  ? () => _openMenu(
                      _BattleMenuView.spells,
                      selectedCommandKey: _defaultAbilityCommandKey(
                        _spells,
                        prefix: 'spell',
                        unavailableTitle: 'Need More Mana',
                      ),
                    )
                  : null,
            ),
            _buildCommandButton(
              context: context,
              label: 'Technique',
              commandKey: 'root:Technique',
              onPressed: _canAct
                  ? () => _openMenu(
                      _BattleMenuView.techniques,
                      selectedCommandKey: _defaultAbilityCommandKey(
                        _techniques,
                        prefix: 'technique',
                        unavailableTitle: 'Cooling Down',
                      ),
                    )
                  : null,
            ),
            _buildCommandButton(
              context: context,
              label: _loadingItems ? 'Item (Loading...)' : 'Item',
              commandKey: 'root:Item',
              onPressed: _canAct && !_loadingItems
                  ? () => _openMenu(
                      _BattleMenuView.items,
                      selectedCommandKey: 'item:0',
                    )
                  : null,
            ),
            _buildCommandButton(
              context: context,
              label: 'Escape',
              commandKey: 'root:Escape',
              onPressed: _canAct ? _escape : null,
            ),
          ],
        );
      case _BattleMenuView.spells:
        return _buildAbilityChoiceList(
          theme: theme,
          title: 'Spells',
          prefix: 'spell',
          abilities: _spells,
          unavailableTitle: 'Need More Mana',
        );
      case _BattleMenuView.techniques:
        return _buildAbilityChoiceList(
          theme: theme,
          title: 'Techniques',
          prefix: 'technique',
          abilities: _techniques,
          unavailableTitle: 'Cooling Down',
        );
      case _BattleMenuView.items:
        final availableItems = _items
            .where((item) => item.quantity > 0)
            .toList(growable: false);
        final itemChildren = availableItems
            .asMap()
            .entries
            .map((entry) {
              final index = entry.key;
              final item = entry.value;
              final damageLabel = item.dealDamage > 0
                  ? ' | DMG ${item.dealDamage}x${item.dealDamageHits}'
                  : '';
              final allEnemiesDamageLabel = item.dealDamageAllEnemies > 0
                  ? ' | AOE ${item.dealDamageAllEnemies}x${item.dealDamageAllEnemiesHits}'
                  : '';
              return _buildCommandButton(
                context: context,
                label:
                    '${item.name} x${item.quantity}${item.healthDelta != 0 ? ' | HP ${item.healthDelta > 0 ? '+' : ''}${item.healthDelta}' : ''}${item.manaDelta != 0 ? ' | MP ${item.manaDelta > 0 ? '+' : ''}${item.manaDelta}' : ''}$damageLabel$allEnemiesDamageLabel',
                commandKey: 'item:$index',
                onPressed: _canAct ? () => _useItem(item) : null,
              );
            })
            .toList(growable: false);
        return _buildMenuShell(
          title: 'Items',
          child: itemChildren.isEmpty
              ? const Center(child: Text('No options available.'))
              : ListView.separated(
                  itemCount: itemChildren.length,
                  separatorBuilder: (context, index) =>
                      const SizedBox(height: 6),
                  itemBuilder: (context, index) => itemChildren[index],
                ),
        );
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return PopScope(
      canPop: false,
      child: SizedBox.expand(
        child: Material(
          color: Colors.transparent,
          child: PaperSheet(
            child: SafeArea(
              child: LayoutBuilder(
                builder: (context, constraints) {
                  final commandPanelHeight = _commandPanelHeightFor(
                    constraints.maxHeight,
                  );
                  return SingleChildScrollView(
                    padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
                    child: ConstrainedBox(
                      constraints: BoxConstraints(
                        minHeight: constraints.maxHeight,
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          _buildTurnOrderStrip(theme),
                          const SizedBox(height: 8),
                          _buildBattlefield(theme),
                          const SizedBox(height: 8),
                          _buildLogPanel(theme, compact: _showCompactBattleLog),
                          const SizedBox(height: 8),
                          Container(
                            height: commandPanelHeight,
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: theme.colorScheme.surface,
                              borderRadius: BorderRadius.circular(10),
                              border: Border.all(
                                color: theme.colorScheme.outline,
                                width: 2,
                              ),
                            ),
                            child: _buildCommandMenu(theme),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),
          ),
        ),
      ),
    );
  }
}
