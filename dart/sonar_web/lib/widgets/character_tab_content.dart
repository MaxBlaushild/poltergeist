import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../models/character_stats.dart';
import '../models/equipment_item.dart';
import '../models/inventory_item.dart';
import '../models/user.dart';
import '../models/user_level.dart';
import '../providers/auth_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/user_level_provider.dart';
import '../services/inventory_service.dart';

class CharacterTabContent extends StatefulWidget {
  const CharacterTabContent({
    super.key,
    this.userOverride,
    this.statsOverride,
    this.userLevelOverride,
    this.readOnly = false,
  });

  final User? userOverride;
  final CharacterStats? statsOverride;
  final UserLevel? userLevelOverride;
  final bool readOnly;

  @override
  State<CharacterTabContent> createState() => _CharacterTabContentState();
}

class _CharacterTabContentState extends State<CharacterTabContent> {
  static const double _damageLabelColumnWidth = 120;
  static const Set<String> _handEquipmentSlots = {'dominant_hand', 'off_hand'};
  static const Map<String, String> _labels = {
    'strength': 'Strength',
    'dexterity': 'Dexterity',
    'constitution': 'Constitution',
    'intelligence': 'Intelligence',
    'wisdom': 'Wisdom',
    'charisma': 'Charisma',
  };

  String? _lastUserId;
  String? _lastEquipmentUserId;
  Map<String, int> _pending = {};
  List<EquippedItem> _equippedItems = const [];
  bool _loadingEquipment = false;
  final ScrollController _scrollController = ScrollController();
  bool _showTopFade = false;
  bool _showBottomFade = false;

  int get _pendingTotal =>
      _pending.values.where((value) => value > 0).fold(0, (a, b) => a + b);

  bool get _isReadOnly =>
      widget.readOnly ||
      widget.userOverride != null ||
      widget.statsOverride != null ||
      widget.userLevelOverride != null;

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_updateFades);
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final uid =
        widget.userOverride?.id ?? context.watch<AuthProvider>().user?.id;
    if (uid != _lastUserId) {
      _lastUserId = uid;
      _pending = {};
    }
    final unspent =
        widget.statsOverride?.unspentPoints ??
        context.watch<CharacterStatsProvider>().unspentPoints;
    if (unspent == 0 && _pending.isNotEmpty) {
      _pending = {};
    }
    if (_isReadOnly && _pending.isNotEmpty) {
      _pending = {};
    }
    _syncEquipment(uid);
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
  }

  @override
  void dispose() {
    _scrollController.removeListener(_updateFades);
    _scrollController.dispose();
    super.dispose();
  }

  void _bumpStat(String key, int delta, int remaining) {
    if (_isReadOnly) return;
    if (delta > 0 && remaining <= 0) return;
    final current = _pending[key] ?? 0;
    final next = (current + delta).clamp(0, 999);
    setState(() {
      if (next == 0) {
        _pending.remove(key);
      } else {
        _pending[key] = next;
      }
    });
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
  }

  Future<void> _confirmAllocations(CharacterStatsProvider stats) async {
    if (_isReadOnly || _pendingTotal == 0) return;
    final success = await stats.applyAllocations(_pending);
    if (!mounted) return;
    if (success) {
      setState(() => _pending = {});
      WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
      return;
    }
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Unable to apply stat points.')),
    );
  }

  bool get _canLoadLiveEquipment =>
      widget.userOverride == null &&
      widget.statsOverride == null &&
      widget.userLevelOverride == null;

  void _syncEquipment(String? uid) {
    if (!_canLoadLiveEquipment || uid == null || uid.isEmpty) {
      if (_lastEquipmentUserId != null ||
          _loadingEquipment ||
          _equippedItems.isNotEmpty) {
        setState(() {
          _lastEquipmentUserId = null;
          _loadingEquipment = false;
          _equippedItems = const [];
        });
      }
      return;
    }
    if (_lastEquipmentUserId == uid) return;
    _lastEquipmentUserId = uid;
    _loadEquipment(uid);
  }

  Future<void> _loadEquipment(String uid) async {
    setState(() {
      _loadingEquipment = true;
    });
    final equipment = await context.read<InventoryService>().getEquipment();
    if (!mounted) return;
    final currentUid =
        widget.userOverride?.id ?? context.read<AuthProvider>().user?.id;
    if (currentUid != uid || !_canLoadLiveEquipment) return;
    setState(() {
      _equippedItems = equipment;
      _loadingEquipment = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
    final auth = context.watch<AuthProvider>();
    final levels = context.watch<UserLevelProvider>();
    final statsProvider = context.watch<CharacterStatsProvider>();
    final user = widget.userOverride ?? auth.user;
    if (user == null) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: Text('Log in to see your character.'),
        ),
      );
    }
    final userLevel = widget.userLevelOverride ?? levels.userLevel;
    final levelLoading = widget.userLevelOverride != null
        ? false
        : levels.loading;
    final displayLevel =
        widget.userLevelOverride?.level ??
        widget.statsOverride?.level ??
        userLevel?.level ??
        statsProvider.level;
    final overrideStats = widget.statsOverride;
    final baseStats = overrideStats?.toMap() ?? statsProvider.baseStats;
    final bonusStats =
        overrideStats?.bonusMap() ?? statsProvider.equipmentBonuses;
    final effectiveConstitution =
        (baseStats['constitution'] ?? CharacterStatsProvider.baseStatValue) +
        (bonusStats['constitution'] ?? 0);
    final effectiveIntelligence =
        (baseStats['intelligence'] ?? CharacterStatsProvider.baseStatValue) +
        (bonusStats['intelligence'] ?? 0);
    final effectiveWisdom =
        (baseStats['wisdom'] ?? CharacterStatsProvider.baseStatValue) +
        (bonusStats['wisdom'] ?? 0);
    final health =
        overrideStats?.health ??
        CharacterStats.deriveHealthFromConstitution(effectiveConstitution);
    final mana =
        overrideStats?.mana ??
        CharacterStats.deriveManaFromMentalStats(
          effectiveIntelligence,
          effectiveWisdom,
        );
    final unspentPoints =
        overrideStats?.unspentPoints ?? statsProvider.unspentPoints;
    final hasUnspentPoints = unspentPoints > 0;
    final proficiencies =
        overrideStats?.proficiencies ?? statsProvider.proficiencies;
    final statuses = overrideStats?.statuses ?? statsProvider.statuses;
    final hasProficiencies = proficiencies.isNotEmpty;
    final canEdit = !_isReadOnly;

    void showProfileImage() {
      if (user.profilePictureUrl.isEmpty) return;
      showDialog<void>(
        context: context,
        barrierColor: Colors.black54,
        builder: (context) {
          final theme = Theme.of(context);
          return Dialog(
            backgroundColor: Colors.transparent,
            insetPadding: const EdgeInsets.all(24),
            child: Stack(
              alignment: Alignment.topRight,
              children: [
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.surface,
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: theme.colorScheme.outlineVariant),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.2),
                        blurRadius: 18,
                        offset: const Offset(0, 10),
                      ),
                    ],
                  ),
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(16),
                    child: Image.network(
                      user.profilePictureUrl,
                      width: 320,
                      height: 320,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Container(
                        width: 320,
                        height: 320,
                        color: theme.colorScheme.surfaceVariant,
                        child: const Icon(Icons.person, size: 96),
                      ),
                    ),
                  ),
                ),
                IconButton(
                  onPressed: () => Navigator.of(context).pop(),
                  icon: const Icon(Icons.close),
                  style: IconButton.styleFrom(
                    backgroundColor: theme.colorScheme.surfaceContainerHighest,
                    shape: const CircleBorder(),
                  ),
                ),
              ],
            ),
          );
        },
      );
    }

    return Stack(
      children: [
        SingleChildScrollView(
          controller: _scrollController,
          padding: const EdgeInsets.only(bottom: 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  GestureDetector(
                    onTap: user.profilePictureUrl.isNotEmpty
                        ? showProfileImage
                        : null,
                    child: Stack(
                      clipBehavior: Clip.none,
                      children: [
                        CircleAvatar(
                          radius: 28,
                          backgroundColor: Colors.grey.shade300,
                          backgroundImage: user.profilePictureUrl.isNotEmpty
                              ? NetworkImage(user.profilePictureUrl)
                              : null,
                          child: user.profilePictureUrl.isEmpty
                              ? const Icon(Icons.person)
                              : null,
                        ),
                        if (hasUnspentPoints)
                          Positioned(
                            right: -2,
                            top: -2,
                            child: Container(
                              padding: const EdgeInsets.all(3),
                              decoration: BoxDecoration(
                                color: const Color(0xFFFFD54F),
                                shape: BoxShape.circle,
                                border: Border.all(
                                  color: theme.colorScheme.surface,
                                  width: 1.2,
                                ),
                              ),
                              child: const Icon(
                                Icons.arrow_upward,
                                size: 12,
                                color: Colors.white,
                              ),
                            ),
                          ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisSize: MainAxisSize.min,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          user.username.isNotEmpty ? user.username : user.name,
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                        if (user.username.isNotEmpty &&
                            user.name != user.username)
                          Text(
                            user.name,
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              _buildResourceCard(context, health: health, mana: mana),
              const SizedBox(height: 16),
              _buildDamageCard(context),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surface,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: levelLoading
                    ? Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Text(
                            'Level',
                            style: theme.textTheme.titleSmall?.copyWith(
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                          const SizedBox(height: 10),
                          LinearProgressIndicator(
                            minHeight: 8,
                            color: theme.colorScheme.primary,
                            backgroundColor:
                                theme.colorScheme.surfaceContainerHighest,
                          ),
                        ],
                      )
                    : userLevel == null
                    ? Text(
                        'Level data unavailable right now.',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      )
                    : Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Row(
                            children: [
                              Text(
                                'Level ${userLevel.level}',
                                style: theme.textTheme.titleSmall?.copyWith(
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                              const Spacer(),
                              Text(
                                '${userLevel.experiencePointsOnLevel} / ${userLevel.experienceToNextLevel} XP',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  color: theme.colorScheme.onSurfaceVariant,
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 8),
                          ClipRRect(
                            borderRadius: BorderRadius.circular(999),
                            child: LinearProgressIndicator(
                              value: userLevel.experienceToNextLevel > 0
                                  ? (userLevel.experiencePointsOnLevel /
                                            userLevel.experienceToNextLevel)
                                        .clamp(0.0, 1.0)
                                  : 0.0,
                              minHeight: 8,
                              color: theme.colorScheme.primary,
                              backgroundColor:
                                  theme.colorScheme.surfaceContainerHighest,
                            ),
                          ),
                        ],
                      ),
              ),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surface,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      'Character stats',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        Text(
                          'Level $displayLevel',
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        const Spacer(),
                        Text(
                          'Unspent: $unspentPoints',
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: hasUnspentPoints
                                ? theme.colorScheme.primary
                                : theme.colorScheme.onSurfaceVariant,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                    if (hasUnspentPoints) ...[
                      const SizedBox(height: 6),
                      Text(
                        'Level up!',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: const Color(0xFFC58A00),
                          fontWeight: FontWeight.w700,
                          letterSpacing: 0.3,
                        ),
                      ),
                    ],
                    if (canEdit && _pendingTotal > 0) ...[
                      const SizedBox(height: 6),
                      Text(
                        'Remaining after pending: ${unspentPoints - _pendingTotal}',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                    const SizedBox(height: 12),
                    Column(
                      children: CharacterStatsProvider.statKeys.map((key) {
                        final label = _labels[key] ?? key;
                        final baseValue =
                            baseStats[key] ??
                            CharacterStatsProvider.baseStatValue;
                        final bonusValue = bonusStats[key] ?? 0;
                        final bonusLabel = bonusValue == 0
                            ? ''
                            : bonusValue > 0
                            ? '+$bonusValue'
                            : '$bonusValue';
                        final bonusColor = bonusValue == 0
                            ? theme.colorScheme.onSurfaceVariant
                            : bonusValue > 0
                            ? theme.colorScheme.primary
                            : theme.colorScheme.error;
                        final pendingValue = _pending[key] ?? 0;
                        final displayValue = baseValue + pendingValue;
                        final remaining = unspentPoints - _pendingTotal;
                        final canAdd = canEdit && remaining > 0;
                        final canRemove = canEdit && pendingValue > 0;
                        return Container(
                          margin: const EdgeInsets.only(bottom: 8),
                          padding: const EdgeInsets.symmetric(
                            horizontal: 10,
                            vertical: 8,
                          ),
                          decoration: BoxDecoration(
                            color: theme.colorScheme.surfaceContainerHighest,
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: theme.colorScheme.outlineVariant,
                            ),
                          ),
                          child: Row(
                            children: [
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      label,
                                      style: theme.textTheme.bodyMedium
                                          ?.copyWith(
                                            fontWeight: FontWeight.w700,
                                          ),
                                    ),
                                    if (pendingValue > 0)
                                      Text(
                                        '+$pendingValue pending',
                                        style: theme.textTheme.bodySmall
                                            ?.copyWith(
                                              color: theme.colorScheme.primary,
                                            ),
                                      ),
                                  ],
                                ),
                              ),
                              SizedBox(
                                width: 48,
                                child: Align(
                                  alignment: Alignment.centerRight,
                                  child: Text(
                                    bonusLabel,
                                    style: theme.textTheme.bodyMedium?.copyWith(
                                      color: bonusColor,
                                      fontWeight: FontWeight.w700,
                                    ),
                                  ),
                                ),
                              ),
                              const SizedBox(width: 6),
                              Text(
                                '$displayValue',
                                style: theme.textTheme.titleMedium?.copyWith(
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                              const SizedBox(width: 8),
                              IconButton(
                                visualDensity: VisualDensity.compact,
                                onPressed: canRemove
                                    ? () => _bumpStat(key, -1, remaining)
                                    : null,
                                icon: const Icon(Icons.remove_circle_outline),
                              ),
                              IconButton(
                                visualDensity: VisualDensity.compact,
                                onPressed: canAdd
                                    ? () => _bumpStat(key, 1, remaining)
                                    : null,
                                icon: const Icon(Icons.add_circle_outline),
                              ),
                            ],
                          ),
                        );
                      }).toList(),
                    ),
                    if (canEdit && _pendingTotal > 0) ...[
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          Expanded(
                            child: OutlinedButton(
                              onPressed: () => setState(() => _pending = {}),
                              child: const Text('Cancel'),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: FilledButton(
                              onPressed: () =>
                                  _confirmAllocations(statsProvider),
                              child: const Text('Confirm'),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(height: 16),
              _buildStatusesCard(context, statuses),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surface,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: theme.colorScheme.outlineVariant),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      'Proficiencies',
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: 8),
                    if (!hasProficiencies)
                      Text(
                        'No proficiencies yet. Complete quests to earn them.',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                    if (hasProficiencies)
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: proficiencies.map((proficiency) {
                          return Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 10,
                              vertical: 6,
                            ),
                            decoration: BoxDecoration(
                              color: theme.colorScheme.surfaceContainerHighest,
                              borderRadius: BorderRadius.circular(999),
                              border: Border.all(
                                color: theme.colorScheme.outlineVariant,
                              ),
                            ),
                            child: Text(
                              '${_toTitleCase(proficiency.proficiency)} · ${proficiency.level}',
                              style: theme.textTheme.bodySmall?.copyWith(
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          );
                        }).toList(),
                      ),
                  ],
                ),
              ),
              if (canEdit) ...[
                const SizedBox(height: 16),
                OutlinedButton(
                  onPressed: () {
                    Navigator.of(context).pop();
                    context.go('/logout');
                  },
                  child: const Text('Log out'),
                ),
              ],
            ],
          ),
        ),
        if (_showTopFade)
          Positioned(
            left: 0,
            right: 0,
            top: 0,
            child: IgnorePointer(
              child: Container(
                height: 18,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    colors: [
                      theme.colorScheme.surface,
                      theme.colorScheme.surface.withOpacity(0.0),
                    ],
                  ),
                ),
              ),
            ),
          ),
        if (_showBottomFade)
          Positioned(
            left: 0,
            right: 0,
            bottom: 0,
            child: IgnorePointer(
              child: Container(
                height: 20,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.bottomCenter,
                    end: Alignment.topCenter,
                    colors: [
                      theme.colorScheme.surface,
                      theme.colorScheme.surface.withOpacity(0.0),
                    ],
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }

  Widget _buildDamageCard(BuildContext context) {
    final theme = Theme.of(context);
    final handItems = _equippedItems
        .where((entry) => _handEquipmentSlots.contains(entry.slot))
        .map((entry) => entry.inventoryItem)
        .whereType<InventoryItem>()
        .toList();
    final weaponItems = handItems
        .where((item) => (item.damageMin ?? 0) > 0 && (item.damageMax ?? 0) > 0)
        .toList();
    final hasWeapon = weaponItems.isNotEmpty;
    final damageMin = hasWeapon
        ? weaponItems.fold<int>(0, (sum, item) => sum + (item.damageMin ?? 0))
        : 1;
    final damageMax = hasWeapon
        ? weaponItems.fold<int>(0, (sum, item) => sum + (item.damageMax ?? 0))
        : 2;
    final swipesPerAttack = hasWeapon
        ? weaponItems.fold<int>(
            0,
            (sum, item) =>
                sum +
                ((item.swipesPerAttack ?? 0) > 0 ? item.swipesPerAttack! : 1),
          )
        : 1;
    final attackSource = hasWeapon
        ? weaponItems.map((item) => item.name).join(' + ')
        : 'Unarmed';
    final spellDamageBonus = handItems.fold<int>(
      0,
      (sum, item) =>
          sum +
          ((item.spellDamageBonusPercent ?? 0) > 0
              ? item.spellDamageBonusPercent!
              : 0),
    );
    final blockPercentageRaw = handItems.fold<int>(
      0,
      (sum, item) =>
          sum + ((item.blockPercentage ?? 0) > 0 ? item.blockPercentage! : 0),
    );
    final blockPercentage = blockPercentageRaw > 100 ? 100 : blockPercentageRaw;
    final damageBlocked = handItems.fold<int>(
      0,
      (sum, item) =>
          sum + ((item.damageBlocked ?? 0) > 0 ? item.damageBlocked! : 0),
    );
    final hasBlockStats = blockPercentage > 0 || damageBlocked > 0;

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            'Damage',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          if (_loadingEquipment)
            LinearProgressIndicator(
              minHeight: 6,
              color: theme.colorScheme.primary,
              backgroundColor: theme.colorScheme.surfaceContainerHighest,
            )
          else ...[
            _buildDamageStatRow(
              context,
              label: 'Attack',
              value: '$damageMin-$damageMax',
            ),
            const SizedBox(height: 6),
            _buildDamageStatRow(
              context,
              label: 'Swipes / attack',
              value: '$swipesPerAttack',
            ),
            const SizedBox(height: 6),
            _buildDamageStatRow(context, label: 'Source', value: attackSource),
            const SizedBox(height: 6),
            _buildDamageStatRow(
              context,
              label: 'Spell bonus',
              value: '+$spellDamageBonus%',
            ),
            if (hasBlockStats) ...[
              const SizedBox(height: 6),
              _buildDamageStatRow(
                context,
                label: 'Block',
                value: blockPercentage > 0 && damageBlocked > 0
                    ? '$blockPercentage% up to $damageBlocked'
                    : blockPercentage > 0
                    ? '$blockPercentage%'
                    : 'up to $damageBlocked',
              ),
            ],
            if (!_canLoadLiveEquipment) ...[
              const SizedBox(height: 8),
              Text(
                'Equipment-specific values are only shown for your live character.',
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
            ],
          ],
        ],
      ),
    );
  }

  Widget _buildStatusesCard(
    BuildContext context,
    List<CharacterStatus> statuses,
  ) {
    final theme = Theme.of(context);
    final sortedStatuses = List<CharacterStatus>.from(statuses)
      ..sort((a, b) {
        final aTime = a.expiresAt;
        final bTime = b.expiresAt;
        if (aTime == null && bTime == null) return 0;
        if (aTime == null) return 1;
        if (bTime == null) return -1;
        return aTime.compareTo(bTime);
      });
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            'Statuses',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          if (sortedStatuses.isEmpty)
            Text(
              'No active statuses.',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          if (sortedStatuses.isNotEmpty)
            Column(
              children: sortedStatuses.map((status) {
                final effectText = status.effect.trim().isNotEmpty
                    ? status.effect.trim()
                    : _toTitleCase(status.effectType);
                final remainingText = _statusRemainingText(status.expiresAt);
                return Container(
                  width: double.infinity,
                  margin: const EdgeInsets.only(bottom: 8),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 8,
                  ),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.surfaceContainerHighest,
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: theme.colorScheme.outlineVariant),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: Text(
                              status.name,
                              style: theme.textTheme.bodyMedium?.copyWith(
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ),
                          Text(
                            remainingText,
                            style: theme.textTheme.labelSmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant,
                            ),
                          ),
                        ],
                      ),
                      if (status.description.trim().isNotEmpty) ...[
                        const SizedBox(height: 4),
                        Text(
                          status.description.trim(),
                          style: theme.textTheme.bodySmall,
                        ),
                      ],
                      if (effectText.isNotEmpty) ...[
                        const SizedBox(height: 4),
                        Text(
                          'Effect: $effectText',
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ],
                  ),
                );
              }).toList(),
            ),
        ],
      ),
    );
  }

  Widget _buildResourceCard(
    BuildContext context, {
    required int health,
    required int mana,
  }) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: theme.colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _buildResourceBar(
            context,
            label: 'Health',
            value: health,
            maxValue: health,
            icon: Icons.favorite,
            fillColor: const Color(0xFFB33939),
          ),
          const SizedBox(height: 8),
          _buildResourceBar(
            context,
            label: 'Mana',
            value: mana,
            maxValue: mana,
            icon: Icons.auto_fix_high,
            fillColor: const Color(0xFF1E6FA7),
          ),
        ],
      ),
    );
  }

  Widget _buildResourceBar(
    BuildContext context, {
    required String label,
    required int value,
    required int maxValue,
    required IconData icon,
    required Color fillColor,
  }) {
    final theme = Theme.of(context);
    final normalizedMax = maxValue <= 0 ? 1 : maxValue;
    final progress = (value / normalizedMax).clamp(0.0, 1.0);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, size: 16, color: fillColor),
            const SizedBox(width: 6),
            Text(
              label,
              style: theme.textTheme.bodySmall?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const Spacer(),
            Text(
              '$value / $normalizedMax',
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
        ),
        const SizedBox(height: 4),
        ClipRRect(
          borderRadius: BorderRadius.circular(999),
          child: LinearProgressIndicator(
            value: progress,
            minHeight: 8,
            color: fillColor,
            backgroundColor: theme.colorScheme.surfaceContainerHighest,
          ),
        ),
      ],
    );
  }

  Widget _buildDamageStatRow(
    BuildContext context, {
    required String label,
    required String value,
  }) {
    final theme = Theme.of(context);
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: _damageLabelColumnWidth,
          child: Text(
            label,
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
        Expanded(
          child: Text(
            value,
            textAlign: TextAlign.left,
            overflow: TextOverflow.ellipsis,
            style: theme.textTheme.bodyMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
        ),
      ],
    );
  }

  String _statusRemainingText(DateTime? expiresAt) {
    if (expiresAt == null) return 'Active';
    final remaining = expiresAt.toLocal().difference(DateTime.now());
    if (remaining.inSeconds <= 0) return 'Expired';
    if (remaining.inDays > 0) {
      final hours = remaining.inHours.remainder(24);
      if (hours > 0) {
        return '${remaining.inDays}d ${hours}h left';
      }
      return '${remaining.inDays}d left';
    }
    if (remaining.inHours > 0) {
      final minutes = remaining.inMinutes.remainder(60);
      if (minutes > 0) {
        return '${remaining.inHours}h ${minutes}m left';
      }
      return '${remaining.inHours}h left';
    }
    final minutes = remaining.inMinutes;
    if (minutes > 0) {
      return '${minutes}m left';
    }
    return '${remaining.inSeconds}s left';
  }

  String _toTitleCase(String value) {
    final normalized = value
        .trim()
        .replaceAll('_', ' ')
        .replaceAll('-', ' ')
        .toLowerCase();
    if (normalized.isEmpty) return value;
    return normalized
        .split(RegExp(r'\s+'))
        .where((part) => part.isNotEmpty)
        .map((part) => '${part[0].toUpperCase()}${part.substring(1)}')
        .join(' ');
  }

  void _updateFades() {
    if (!mounted || !_scrollController.hasClients) return;
    final position = _scrollController.position;
    final showTop = position.pixels > 0.5;
    final showBottom = position.pixels < position.maxScrollExtent - 0.5;
    if (showTop == _showTopFade && showBottom == _showBottomFade) return;
    setState(() {
      _showTopFade = showTop;
      _showBottomFade = showBottom;
    });
  }
}
