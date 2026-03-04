import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/inventory_item.dart';
import '../models/monster.dart';
import '../models/spell.dart';
import '../providers/auth_provider.dart';
import '../providers/character_stats_provider.dart';
import '../services/inventory_service.dart';
import 'paper_texture.dart';

enum MonsterBattleOutcome { victory, defeat }

enum _BattleMenuView { root, spells, techniques, items }

class MonsterBattleResult {
  const MonsterBattleResult({
    required this.outcome,
    required this.playerHealthRemaining,
  });

  final MonsterBattleOutcome outcome;
  final int playerHealthRemaining;
}

class _BattleItemChoice {
  const _BattleItemChoice({
    required this.ownedInventoryItemId,
    required this.name,
    required this.healthDelta,
    required this.manaDelta,
    required this.quantity,
  });

  final String ownedInventoryItemId;
  final String name;
  final int healthDelta;
  final int manaDelta;
  final int quantity;

  _BattleItemChoice copyWith({int? quantity}) => _BattleItemChoice(
    ownedInventoryItemId: ownedInventoryItemId,
    name: name,
    healthDelta: healthDelta,
    manaDelta: manaDelta,
    quantity: quantity ?? this.quantity,
  );
}

class _EncounterEnemyState {
  _EncounterEnemyState({required this.monster, required this.currentHealth});

  final Monster monster;
  int currentHealth;

  int get maxHealth => math.max(1, monster.maxHealth);
  bool get isDefeated => currentHealth <= 0;
}

class MonsterBattleDialog extends StatefulWidget {
  const MonsterBattleDialog({super.key, required this.encounter});

  final MonsterEncounter encounter;

  @override
  State<MonsterBattleDialog> createState() => _MonsterBattleDialogState();
}

class _MonsterBattleDialogState extends State<MonsterBattleDialog> {
  final math.Random _random = math.Random();
  final List<String> _battleLog = <String>[];

  late final String _playerName;
  late final int _playerLevel;
  late final Map<String, int> _playerStats;
  late final List<Spell> _spells;
  late final List<Spell> _techniques;
  late final int _playerMaxHealth;
  late final int _playerMaxMana;
  late int _playerHealth;
  late int _playerMana;
  late final List<_EncounterEnemyState> _enemies;
  int _activeEnemyIndex = 0;
  late final String _playerFrontSpriteUrl;
  late final String _playerBackSpriteUrl;
  late final bool _hasTrueBackSprite;
  List<_BattleItemChoice> _items = const [];

  _BattleMenuView _menuView = _BattleMenuView.root;
  String? _selectedCommandKey;
  bool _loadingItems = false;
  bool _playerTurn = true;
  bool _busy = false;
  bool _battleOver = false;
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

  @override
  void initState() {
    super.initState();
    final statsProvider = context.read<CharacterStatsProvider>();
    final authProvider = context.read<AuthProvider>();
    final user = authProvider.user;
    _playerName = user?.username.trim().isNotEmpty == true
        ? '@${user!.username.trim()}'
        : (user?.name.trim().isNotEmpty == true ? user!.name.trim() : 'You');
    _playerLevel = statsProvider.level;
    _playerStats = statsProvider.stats;
    _spells = statsProvider.spells;
    _techniques = statsProvider.techniques;
    _playerMaxHealth = math.max(1, statsProvider.maxHealth);
    _playerMaxMana = math.max(0, statsProvider.maxMana);
    _playerHealth = math.max(1, statsProvider.health);
    _playerMana = math.max(0, statsProvider.mana);
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
            currentHealth: math.max(1, monster.maxHealth),
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
        ),
      );
    }
    _activeEnemyIndex = 0;
    _playerFrontSpriteUrl = (user?.profilePictureUrl ?? '').trim();
    _playerBackSpriteUrl = (user?.backProfilePictureUrl ?? '').trim();
    _hasTrueBackSprite = _playerBackSpriteUrl.isNotEmpty;
    _selectedCommandKey = 'root:Attack';
    if (_enemies.length == 1) {
      _battleLog.add('A wild ${_enemies.first.monster.name} appears!');
    } else {
      _battleLog.add('A hostile group of ${_enemies.length} monsters appears!');
    }
    _battleLog.add('Choose a command.');
    unawaited(_loadItemChoices());
  }

  bool get _canAct => _playerTurn && !_busy && !_battleOver;

  List<_EncounterEnemyState> get _aliveEnemies =>
      _enemies.where((enemy) => !enemy.isDefeated).toList(growable: false);

  _EncounterEnemyState? get _activeEnemy {
    if (_activeEnemyIndex >= 0 &&
        _activeEnemyIndex < _enemies.length &&
        !_enemies[_activeEnemyIndex].isDefeated) {
      return _enemies[_activeEnemyIndex];
    }
    for (var i = 0; i < _enemies.length; i++) {
      if (!_enemies[i].isDefeated) {
        _activeEnemyIndex = i;
        return _enemies[i];
      }
    }
    return null;
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

  bool _isHealingEffect(String effectType) {
    final normalized = effectType.trim().toLowerCase();
    return normalized.contains('restore_life') || normalized.contains('heal');
  }

  int _playerAttackDamage() {
    final strength =
        _playerStats['strength'] ?? CharacterStatsProvider.baseStatValue;
    final dexterity =
        _playerStats['dexterity'] ?? CharacterStatsProvider.baseStatValue;
    final minDamage = math.max<int>(1, _playerLevel + (strength / 4).floor());
    final maxDamage = math.max<int>(
      minDamage,
      minDamage + math.max<int>(1, (dexterity / 3).floor()),
    );
    return _rollDamage(minDamage, maxDamage);
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

  int _abilityDamage(Spell ability) {
    final damageEffects = ability.effects
        .where((effect) => _isDamageEffect(effect.type))
        .toList(growable: false);
    final healEffects = ability.effects
        .where((effect) => _isHealingEffect(effect.type))
        .toList(growable: false);

    final explicitDamage = damageEffects.fold<int>(
      0,
      (sum, effect) => sum + math.max(0, effect.amount),
    );
    if (damageEffects.isNotEmpty) {
      return explicitDamage + math.max(0, _playerLevel ~/ 3);
    }
    if (healEffects.isNotEmpty) {
      return 0;
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
        if (item.consumeHealthDelta == 0 && item.consumeManaDelta == 0) {
          continue;
        }
        choices.add(
          _BattleItemChoice(
            ownedInventoryItemId: entry.id,
            name: item.name,
            healthDelta: item.consumeHealthDelta,
            manaDelta: item.consumeManaDelta,
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
    int playerHealthDelta = 0,
    int playerManaDelta = 0,
  }) async {
    if (!_canAct) return;

    final targetEnemy = _activeEnemy;
    var targetName = 'the enemy';
    if (targetEnemy != null) {
      targetName = targetEnemy.monster.name;
      if (damageToMonster > 0) {
        targetEnemy.currentHealth = math.max(
          0,
          targetEnemy.currentHealth - damageToMonster,
        );
      }
    }

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
      if (damageToMonster > 0) {
        _battleLog.add('$message (Target: $targetName)');
      } else {
        _battleLog.add(message);
      }
    });
    if (damageToMonster > 0) {
      unawaited(
        _playSpriteFx(
          targetMonster: true,
          amount: damageToMonster,
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

    if (_aliveEnemies.isEmpty) {
      await _finishBattle(
        MonsterBattleOutcome.victory,
        _enemies.length == 1
            ? 'You defeated ${_enemies.first.monster.name}!'
            : 'You defeated the entire encounter!',
      );
      return;
    }

    setState(() {
      _playerTurn = false;
    });
    await _monsterTurn();
  }

  Future<void> _monsterTurn() async {
    await Future<void>.delayed(const Duration(milliseconds: 420));
    if (!mounted || _battleOver) return;
    for (var i = 0; i < _enemies.length; i++) {
      final enemy = _enemies[i];
      if (enemy.isDefeated) continue;

      final damage = _monsterAttackDamage(enemy.monster);
      final weaponName = enemy.monster.weaponInventoryItemName.trim().isNotEmpty
          ? enemy.monster.weaponInventoryItemName.trim()
          : 'its weapon';

      setState(() {
        _activeEnemyIndex = i;
        _playerHealth = math.max(0, _playerHealth - damage);
        _battleLog.add(
          '${enemy.monster.name} attacks with $weaponName for $damage damage.',
        );
      });
      unawaited(
        _playSpriteFx(targetMonster: false, amount: damage, healing: false),
      );

      if (_playerHealth <= 0) {
        await _finishBattle(
          MonsterBattleOutcome.defeat,
          'You were defeated by ${enemy.monster.name}.',
        );
        return;
      }

      await Future<void>.delayed(const Duration(milliseconds: 250));
      if (!mounted || _battleOver) return;
    }

    setState(() {
      _busy = false;
      _playerTurn = true;
      _selectedCommandKey = 'root:Attack';
      _battleLog.add('Your turn. Choose a command.');
    });
  }

  Future<void> _finishBattle(
    MonsterBattleOutcome outcome,
    String summary,
  ) async {
    if (_battleOver) return;
    setState(() {
      _battleOver = true;
      _busy = false;
      _battleLog.add(summary);
    });
    await Future<void>.delayed(const Duration(milliseconds: 650));
    if (!mounted) return;
    Navigator.of(context).pop(
      MonsterBattleResult(
        outcome: outcome,
        playerHealthRemaining: _playerHealth,
      ),
    );
  }

  Future<void> _attack() async {
    if (!_canAct) return;
    final damage = _playerAttackDamage();
    final target = _activeEnemy;
    await _resolvePlayerAction(
      message: target == null
          ? '$_playerName attacks for $damage damage.'
          : '$_playerName attacks ${target.monster.name} for $damage damage.',
      damageToMonster: damage,
    );
  }

  Future<void> _useAbility(Spell ability) async {
    if (!_canAct) return;
    final manaCost = _isTechnique(ability) ? 0 : math.max(0, ability.manaCost);
    if (manaCost > _playerMana) {
      setState(() {
        _battleLog.add('Not enough mana for ${ability.name}.');
        _menuView = _BattleMenuView.root;
      });
      return;
    }

    final damage = _abilityDamage(ability);
    final healing = _abilityHealing(ability);
    final parts = <String>['$_playerName uses ${ability.name}'];
    if (damage > 0) parts.add('dealing $damage damage');
    if (healing > 0) parts.add('restoring $healing HP');
    await _resolvePlayerAction(
      message: '${parts.join(', ')}.',
      damageToMonster: damage,
      playerHealthDelta: healing,
      playerManaDelta: -manaCost,
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

    final parts = <String>['$_playerName uses ${item.name}'];
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
    await _resolvePlayerAction(
      message: '${parts.join(', ')}.',
      damageToMonster: 0,
      playerHealthDelta: item.healthDelta,
      playerManaDelta: item.manaDelta,
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

  Widget _buildStatusPanel({
    required ThemeData theme,
    required String name,
    required int level,
    required int currentHp,
    required int maxHp,
    int? currentMana,
    int? maxMana,
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
        ],
      ),
    );
  }

  Widget _buildEnemyRoster(ThemeData theme) {
    if (_enemies.length <= 1) {
      return const SizedBox.shrink();
    }
    final aliveCount = _aliveEnemies.length;
    return Container(
      margin: const EdgeInsets.only(top: 8),
      padding: const EdgeInsets.fromLTRB(10, 8, 10, 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface.withValues(alpha: 0.9),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(
          color: theme.colorScheme.outline.withValues(alpha: 0.5),
          width: 1.2,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Enemies Remaining: $aliveCount/${_enemies.length}',
            style: theme.textTheme.labelMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 6),
          ..._enemies.map((enemy) {
            final isActive = identical(enemy, _activeEnemy);
            return Padding(
              padding: const EdgeInsets.only(bottom: 4),
              child: Text(
                '${enemy.isDefeated ? 'X' : '-'} ${enemy.monster.name} (${enemy.currentHealth}/${enemy.maxHealth} HP)',
                style: theme.textTheme.bodySmall?.copyWith(
                  fontWeight: isActive ? FontWeight.w700 : FontWeight.w400,
                  color: enemy.isDefeated
                      ? theme.colorScheme.onSurface.withValues(alpha: 0.55)
                      : null,
                ),
              ),
            );
          }),
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

  Widget _buildBattlefield(ThemeData theme) {
    final activeEnemy = _activeEnemy;
    final activeEnemyName = activeEnemy?.monster.name ?? widget.encounter.name;
    final activeEnemyLevel = activeEnemy?.monster.level ?? 1;
    final activeEnemyHp = activeEnemy?.currentHealth ?? 0;
    final activeEnemyMaxHp = activeEnemy?.maxHealth ?? 1;
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
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          _buildStatusPanel(
                            theme: theme,
                            name: activeEnemyName,
                            level: activeEnemyLevel,
                            currentHp: activeEnemyHp,
                            maxHp: activeEnemyMaxHp,
                          ),
                          _buildEnemyRoster(theme),
                        ],
                      ),
                    ),
                    const Spacer(),
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
                  ],
                ),
                const SizedBox(height: 8),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    _buildSpriteStage(
                      theme: theme,
                      imageUrl: _playerSpriteUrl,
                      fallbackIcon: Icons.person,
                      size: spriteSize,
                      shakeDx: _playerShakeDx,
                      flashTint: _playerFlashTint,
                      floatingText: _playerFloatText,
                      floatingTextColor: _playerFloatColor,
                      fxNonce: _playerFxNonce,
                      flipHorizontally:
                          !_hasTrueBackSprite && _playerSpriteUrl.isNotEmpty,
                    ),
                    const Spacer(),
                    SizedBox(
                      width: statusWidth,
                      child: _buildStatusPanel(
                        theme: theme,
                        name: _playerName,
                        level: _playerLevel,
                        currentHp: _playerHealth,
                        maxHp: _playerMaxHealth,
                        currentMana: _playerMana,
                        maxMana: _playerMaxMana,
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

  Widget _buildLogPanel(ThemeData theme) {
    final (chipText, chipColor) = _battleOver
        ? ('Battle Over', theme.colorScheme.secondary)
        : _playerTurn
        ? ('Your Turn', const Color(0xFF2B7A4B))
        : ('Enemy Turn', const Color(0xFF8C2F39));

    return Container(
      width: double.infinity,
      height: 138,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: theme.colorScheme.outline, width: 2),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Align(
            alignment: Alignment.centerRight,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: chipColor.withValues(alpha: 0.14),
                borderRadius: BorderRadius.circular(999),
                border: Border.all(color: chipColor.withValues(alpha: 0.5)),
              ),
              child: Text(
                chipText,
                style: theme.textTheme.labelMedium?.copyWith(
                  color: chipColor,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
          ),
          const SizedBox(height: 6),
          Expanded(
            child: _battleLog.isEmpty
                ? Text('...', style: theme.textTheme.bodyMedium)
                : ListView.builder(
                    reverse: true,
                    itemCount: _battleLog.length,
                    itemBuilder: (context, index) {
                      final message = _battleLog[_battleLog.length - 1 - index];
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 4),
                        child: Text(message, style: theme.textTheme.bodyMedium),
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
  }) {
    final selected = _selectedCommandKey == commandKey;
    final theme = Theme.of(context);
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
      child: Text(label),
    );
  }

  Widget _buildChoiceList({
    required String title,
    required List<Widget> children,
  }) {
    return SizedBox(
      height: 138,
      child: Column(
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
              Text(title, style: Theme.of(context).textTheme.titleMedium),
            ],
          ),
          const SizedBox(height: 4),
          Expanded(
            child: children.isEmpty
                ? const Center(child: Text('No options available.'))
                : ListView.separated(
                    itemCount: children.length,
                    separatorBuilder: (context, index) =>
                        const SizedBox(height: 6),
                    itemBuilder: (context, index) => children[index],
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildCommandMenu(ThemeData theme) {
    switch (_menuView) {
      case _BattleMenuView.root:
        return SizedBox(
          height: 138,
          child: GridView.count(
            physics: const NeverScrollableScrollPhysics(),
            crossAxisCount: 2,
            mainAxisSpacing: 8,
            crossAxisSpacing: 8,
            childAspectRatio: 3.2,
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
                        selectedCommandKey: 'spell:0',
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
                        selectedCommandKey: 'technique:0',
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
            ],
          ),
        );
      case _BattleMenuView.spells:
        return _buildChoiceList(
          title: 'Spells',
          children: _spells
              .asMap()
              .entries
              .map((entry) {
                final index = entry.key;
                final spell = entry.value;
                return _buildCommandButton(
                  context: context,
                  label: '${spell.name} (${spell.manaCost} MP)',
                  commandKey: 'spell:$index',
                  onPressed: _canAct ? () => _useAbility(spell) : null,
                );
              })
              .toList(growable: false),
        );
      case _BattleMenuView.techniques:
        return _buildChoiceList(
          title: 'Techniques',
          children: _techniques
              .asMap()
              .entries
              .map((entry) {
                final index = entry.key;
                final technique = entry.value;
                return _buildCommandButton(
                  context: context,
                  label: technique.name,
                  commandKey: 'technique:$index',
                  onPressed: _canAct ? () => _useAbility(technique) : null,
                );
              })
              .toList(growable: false),
        );
      case _BattleMenuView.items:
        final availableItems = _items
            .where((item) => item.quantity > 0)
            .toList(growable: false);
        return _buildChoiceList(
          title: 'Items',
          children: availableItems
              .asMap()
              .entries
              .map((entry) {
                final index = entry.key;
                final item = entry.value;
                return _buildCommandButton(
                  context: context,
                  label:
                      '${item.name} x${item.quantity}${item.healthDelta != 0 ? ' | HP ${item.healthDelta > 0 ? '+' : ''}${item.healthDelta}' : ''}${item.manaDelta != 0 ? ' | MP ${item.manaDelta > 0 ? '+' : ''}${item.manaDelta}' : ''}',
                  commandKey: 'item:$index',
                  onPressed: _canAct ? () => _useItem(item) : null,
                );
              })
              .toList(growable: false),
        );
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return PopScope(
      canPop: false,
      child: Dialog.fullscreen(
        backgroundColor: Colors.transparent,
        child: PaperSheet(
          child: SafeArea(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    'Battle',
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 8),
                  _buildBattlefield(theme),
                  const SizedBox(height: 8),
                  _buildLogPanel(theme),
                  const SizedBox(height: 8),
                  Container(
                    height: 158,
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
          ),
        ),
      ),
    );
  }
}
