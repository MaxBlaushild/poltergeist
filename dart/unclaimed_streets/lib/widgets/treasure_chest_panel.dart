import 'dart:async';
import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../constants/gameplay_constants.dart';
import '../models/inventory_item.dart';
import '../models/spell.dart';
import '../models/treasure_chest.dart';
import '../providers/activity_feed_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/location_provider.dart';
import '../providers/user_level_provider.dart';
import '../services/inventory_service.dart';
import '../services/poi_service.dart';
import '../widgets/paper_texture.dart';

const _chestImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/inventory-items/1762314753387-0gdf0170kq5m.png';

const _openRadiusMeters = kTreasureChestUnlockRadiusMeters;
const _unlockMethodItem = 'item';
const _unlockMethodSpell = 'spell';

class TreasureChestPanel extends StatefulWidget {
  const TreasureChestPanel({
    super.key,
    required this.treasureChest,
    required this.onClose,
    this.onOpened,
  });

  final TreasureChest treasureChest;
  final VoidCallback onClose;
  final void Function(Map<String, dynamic> rewardData)? onOpened;

  @override
  State<TreasureChestPanel> createState() => _TreasureChestPanelState();
}

class _TreasureChestPanelState extends State<TreasureChestPanel> {
  bool _loading = false;
  String? _error;
  List<InventoryItem> _inventoryItems = [];
  List<OwnedInventoryItem> _ownedItems = [];

  @override
  void initState() {
    super.initState();
    _loadInventory();
  }

  Future<void> _loadInventory() async {
    try {
      final inv = context.read<InventoryService>();
      final items = await inv.getInventoryItems();
      final owned = await inv.getOwnedInventoryItems();
      if (mounted) {
        setState(() {
          _inventoryItems = items;
          _ownedItems = owned;
        });
      }
    } catch (_) {
      if (mounted) setState(() {});
    }
  }

  double _calculateDistance(
    double lat1,
    double lon1,
    double lat2,
    double lon2,
  ) {
    const R = 6371e3;
    final phi1 = lat1 * math.pi / 180;
    final phi2 = lat2 * math.pi / 180;
    final dPhi = (lat2 - lat1) * math.pi / 180;
    final dLambda = (lon2 - lon1) * math.pi / 180;
    final a =
        math.sin(dPhi / 2) * math.sin(dPhi / 2) +
        math.cos(phi1) *
            math.cos(phi2) *
            math.sin(dLambda / 2) *
            math.sin(dLambda / 2);
    final c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a));
    return R * c;
  }

  int _inventoryItemLockUnlockStrength(InventoryItem? item) {
    if (item == null) return 0;
    return item.unlockLocksStrength ?? item.unlockTier ?? 0;
  }

  int _spellLockUnlockStrength(Spell spell) {
    var best = 0;
    for (final effect in spell.effects) {
      if (effect.type.trim().toLowerCase() != Spell.effectTypeUnlockLocks) {
        continue;
      }
      if (effect.amount > best) {
        best = effect.amount;
      }
    }
    return best;
  }

  bool _hasUnlockCapability(List<Spell> abilities) {
    final t = widget.treasureChest.unlockTier;
    if (t == null) return true;
    for (final o in _ownedItems) {
      if (o.quantity <= 0) continue;
      final inv = _inventoryItemForId(o.inventoryItemId);
      if (_inventoryItemLockUnlockStrength(inv) >= t) {
        return true;
      }
    }
    for (final ability in abilities) {
      if (_spellLockUnlockStrength(ability) >= t) {
        return true;
      }
    }
    return false;
  }

  ({String text, bool disabled}) _buttonState(
    bool isWithinRange,
    List<Spell> abilities,
  ) {
    if (widget.treasureChest.openedByUser == true) {
      return (text: 'Already opened', disabled: true);
    }
    if (!isWithinRange) {
      return (text: 'Too far away', disabled: true);
    }
    if (widget.treasureChest.unlockTier == null) {
      return (text: 'Open Chest', disabled: false);
    }
    if (_hasUnlockCapability(abilities)) {
      return (text: 'Unlock', disabled: false);
    }
    return (text: 'Locked', disabled: true);
  }

  String _errorMessage(Object e) {
    if (e is DioException && e.response?.data is Map) {
      final d = e.response!.data as Map<String, dynamic>;
      final msg = d['error'] ?? d['message'];
      if (msg != null && msg.toString().isNotEmpty) return msg.toString();
    }
    return e.toString();
  }

  List<Map<String, dynamic>> _itemsAwardedFromResponse(
    Map<String, dynamic> response,
  ) {
    final rawItems = response['itemsAwarded'];
    if (rawItems is! List) {
      return const [];
    }
    return rawItems
        .whereType<Map>()
        .map((entry) => Map<String, dynamic>.from(entry))
        .where((entry) => ((entry['quantity'] as num?)?.toInt() ?? 0) > 0)
        .toList(growable: false);
  }

  Map<String, dynamic> _buildRewardModalData(Map<String, dynamic> response) {
    final rewardExperience =
        (response['rewardExperience'] as num?)?.toInt() ?? 0;
    final rewardGold =
        (response['rewardGold'] as num?)?.toInt() ??
        (response['goldAwarded'] as num?)?.toInt() ??
        0;
    final itemsAwarded = _itemsAwardedFromResponse(response);
    final baseResourcesAwarded =
        (response['baseResourcesAwarded'] as List<dynamic>?)
            ?.whereType<Map>()
            .map((entry) => Map<String, dynamic>.from(entry))
            .toList() ??
        const <Map<String, dynamic>>[];

    if (rewardExperience > 0 ||
        rewardGold > 0 ||
        itemsAwarded.isNotEmpty ||
        baseResourcesAwarded.isNotEmpty) {
      return {
        'rewardExperience': rewardExperience,
        'rewardGold': rewardGold,
        'goldAwarded': rewardGold,
        'baseResourcesAwarded': baseResourcesAwarded,
        'itemsAwarded': itemsAwarded,
      };
    }

    final fallbackItemsAwarded = <Map<String, dynamic>>[];
    for (final chestItem in widget.treasureChest.items) {
      if (chestItem.quantity <= 0) continue;
      final inv = _inventoryItemForId(chestItem.inventoryItemId);
      fallbackItemsAwarded.add({
        'id': chestItem.inventoryItemId,
        'name': inv?.name ?? 'Item #${chestItem.inventoryItemId}',
        'imageUrl': inv?.imageUrl ?? '',
        'quantity': chestItem.quantity,
      });
    }
    return {
      'rewardExperience': 0,
      'rewardGold': math.max(0, widget.treasureChest.gold ?? 0),
      'goldAwarded': math.max(0, widget.treasureChest.gold ?? 0),
      'baseResourcesAwarded': baseResourcesAwarded,
      'itemsAwarded': fallbackItemsAwarded,
    };
  }

  List<_ChestUnlockOption> _availableUnlockOptions(List<Spell> abilities) {
    final requiredStrength = widget.treasureChest.unlockTier;
    if (requiredStrength == null || requiredStrength <= 0) {
      return const [];
    }

    final spellOptions = <_ChestUnlockOption>[];
    for (final ability in abilities) {
      final strength = _spellLockUnlockStrength(ability);
      if (strength < requiredStrength) continue;
      spellOptions.add(
        _ChestUnlockOption(
          method: _unlockMethodSpell,
          title: ability.name.isNotEmpty ? ability.name : 'Unlock Skill',
          subtitle: 'Skill strength $strength',
          detail: ability.manaCost > 0 ? 'Uses ${ability.manaCost} mana' : null,
          spellId: ability.id,
          iconUrl: ability.iconUrl,
          consumesItem: false,
          strength: strength,
        ),
      );
    }
    spellOptions.sort((a, b) => a.strength.compareTo(b.strength));

    final itemOptions = <_ChestUnlockOption>[];
    for (final ownedItem in _ownedItems) {
      if (ownedItem.quantity <= 0) continue;
      final item = _inventoryItemForId(ownedItem.inventoryItemId);
      final strength = _inventoryItemLockUnlockStrength(item);
      if (strength < requiredStrength) continue;
      itemOptions.add(
        _ChestUnlockOption(
          method: _unlockMethodItem,
          title: item?.name.isNotEmpty == true ? item!.name : 'Key Item',
          subtitle: 'Key strength $strength',
          detail: ownedItem.quantity > 1 ? '${ownedItem.quantity} owned' : null,
          ownedInventoryItemId: ownedItem.id,
          iconUrl: item?.imageUrl ?? '',
          consumesItem: true,
          strength: strength,
        ),
      );
    }
    itemOptions.sort((a, b) => a.strength.compareTo(b.strength));

    return [...spellOptions, ...itemOptions];
  }

  Future<_ChestUnlockOption?> _selectUnlockOption(List<Spell> abilities) async {
    final options = _availableUnlockOptions(abilities);
    if (options.isEmpty) return null;

    return showModalBottomSheet<_ChestUnlockOption>(
      context: context,
      useSafeArea: true,
      backgroundColor: Theme.of(context).colorScheme.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (sheetContext) {
        final theme = Theme.of(sheetContext);
        final requiredStrength = widget.treasureChest.unlockTier ?? 0;
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Choose how to unlock',
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  'This chest needs lock strength $requiredStrength.',
                  style: theme.textTheme.bodyMedium,
                ),
                const SizedBox(height: 16),
                Flexible(
                  child: ListView.separated(
                    shrinkWrap: true,
                    itemCount: options.length,
                    separatorBuilder: (_, _) => const SizedBox(height: 10),
                    itemBuilder: (_, index) {
                      final option = options[index];
                      return Material(
                        color: theme.colorScheme.surfaceContainerHighest
                            .withOpacity(0.65),
                        borderRadius: BorderRadius.circular(16),
                        child: InkWell(
                          borderRadius: BorderRadius.circular(16),
                          onTap: () => Navigator.of(
                            sheetContext,
                          ).pop<_ChestUnlockOption>(option),
                          child: Padding(
                            padding: const EdgeInsets.all(12),
                            child: Row(
                              children: [
                                _UnlockOptionAvatar(option: option),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        option.title,
                                        style: theme.textTheme.titleSmall
                                            ?.copyWith(
                                              fontWeight: FontWeight.w700,
                                            ),
                                      ),
                                      const SizedBox(height: 4),
                                      Text(
                                        option.subtitle,
                                        style: theme.textTheme.bodyMedium,
                                      ),
                                      if (option.detail != null &&
                                          option.detail!.isNotEmpty) ...[
                                        const SizedBox(height: 4),
                                        Text(
                                          option.detail!,
                                          style: theme.textTheme.bodySmall,
                                        ),
                                      ],
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 12),
                                Column(
                                  crossAxisAlignment: CrossAxisAlignment.end,
                                  children: [
                                    _UnlockMethodChip(
                                      label: option.method == _unlockMethodSpell
                                          ? 'Skill'
                                          : 'Key',
                                    ),
                                    const SizedBox(height: 8),
                                    Icon(
                                      Icons.chevron_right,
                                      color: theme.colorScheme.onSurface
                                          .withOpacity(0.6),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ),
                        ),
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Future<void> _handlePrimaryAction(List<Spell> abilities) async {
    if (widget.treasureChest.unlockTier == null) {
      await _openChest();
      return;
    }
    final selection = await _selectUnlockOption(abilities);
    if (!mounted || selection == null) return;
    await _openChest(selection: selection);
  }

  Future<void> _openChest({_ChestUnlockOption? selection}) async {
    if (_loading) return;
    final loc = context.read<LocationProvider>().location;
    if (loc == null) {
      setState(() => _error = 'Location not available');
      return;
    }
    final distance = _calculateDistance(
      loc.latitude,
      loc.longitude,
      widget.treasureChest.latitude,
      widget.treasureChest.longitude,
    );
    if (distance > _openRadiusMeters) {
      setState(() => _error = 'Too far away (${distance.round()} m)');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final statsProvider = context.read<CharacterStatsProvider>();
      final previousLevel = statsProvider.level;
      final result = await context.read<PoiService>().openTreasureChest(
        widget.treasureChest.id,
        unlockMethod: selection?.method,
        ownedInventoryItemId: selection?.ownedInventoryItemId,
        spellId: selection?.spellId,
      );
      if (!mounted) return;
      final rewardData = _buildRewardModalData(result);
      await Future.wait([
        context.read<AuthProvider>().refresh(),
        statsProvider.refresh(silent: true),
        context.read<UserLevelProvider>().refresh(),
        context.read<ActivityFeedProvider>().refresh(),
      ]);
      if (!mounted) return;
      final modalData = {
        ...rewardData,
        'leveledUp': statsProvider.level > previousLevel,
        'newLevel': statsProvider.level,
        'levelsGained': math.max(0, statsProvider.level - previousLevel),
      };
      setState(() => _loading = false);
      widget.onClose();
      widget.onOpened?.call(modalData);
    } catch (e) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = _errorMessage(e);
        });
      }
    }
  }

  InventoryItem? _inventoryItemForId(int id) {
    for (final item in _inventoryItems) {
      if (item.id == id) return item;
    }
    return null;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final loc = context.watch<LocationProvider>().location;
    final abilities = context.watch<CharacterStatsProvider>().abilities;
    final distance = loc != null
        ? _calculateDistance(
            loc.latitude,
            loc.longitude,
            widget.treasureChest.latitude,
            widget.treasureChest.longitude,
          )
        : null;
    final isWithinRange = distance != null && distance <= _openRadiusMeters;
    final button = _buttonState(isWithinRange, abilities);
    final isOpened = widget.treasureChest.openedByUser == true;

    return DraggableScrollableSheet(
      initialChildSize: 0.9,
      minChildSize: 0.4,
      maxChildSize: 0.95,
      builder: (_, scrollController) => PaperSheet(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Treasure Chest',
                    style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  IconButton(
                    onPressed: widget.onClose,
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                controller: scrollController,
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        gradient: LinearGradient(
                          colors: [
                            theme.colorScheme.primary.withOpacity(0.12),
                            theme.colorScheme.secondary.withOpacity(0.18),
                          ],
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                        ),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                          color: theme.colorScheme.primary.withOpacity(0.25),
                        ),
                      ),
                      child: Column(
                        children: [
                          Stack(
                            alignment: Alignment.center,
                            children: [
                              Container(
                                width: 140,
                                height: 140,
                                decoration: BoxDecoration(
                                  shape: BoxShape.circle,
                                  color: theme.colorScheme.primary.withOpacity(
                                    0.08,
                                  ),
                                  boxShadow: [
                                    BoxShadow(
                                      color: theme.colorScheme.primary
                                          .withOpacity(0.35),
                                      blurRadius: 24,
                                      spreadRadius: 6,
                                    ),
                                  ],
                                ),
                              ),
                              ClipRRect(
                                borderRadius: BorderRadius.circular(16),
                                child: Image.network(
                                  _chestImageUrl,
                                  width: 128,
                                  height: 128,
                                  fit: BoxFit.contain,
                                  errorBuilder: (_, __, ___) => Container(
                                    width: 128,
                                    height: 128,
                                    color: Colors.grey.shade300,
                                    child: const Icon(
                                      Icons.inventory_2_outlined,
                                      size: 48,
                                    ),
                                  ),
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 12),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(
                                isOpened ? Icons.lock_open : Icons.lock_outline,
                                size: 18,
                                color: isOpened
                                    ? theme.colorScheme.primary
                                    : theme.colorScheme.onSurface.withOpacity(
                                        0.7,
                                      ),
                              ),
                              const SizedBox(width: 8),
                              Text(
                                isOpened
                                    ? 'Loot claimed'
                                    : 'Awaiting discovery',
                                style: theme.textTheme.bodyMedium?.copyWith(
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 16),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        if (distance != null)
                          _InfoChip(
                            icon: Icons.place_outlined,
                            label: '${distance.round()} m away',
                          ),
                        if (widget.treasureChest.unlockTier != null)
                          _InfoChip(
                            icon: Icons.vpn_key_outlined,
                            label:
                                'Lock Strength ${widget.treasureChest.unlockTier}',
                          ),
                      ],
                    ),
                    if (_error != null) ...[
                      const SizedBox(height: 12),
                      Text(
                        _error!,
                        style: TextStyle(
                          color: theme.colorScheme.error,
                          fontSize: 14,
                        ),
                      ),
                    ],
                    const SizedBox(height: 20),
                    FilledButton(
                      onPressed: button.disabled || _loading
                          ? null
                          : () => _handlePrimaryAction(abilities),
                      child: Text(_loading ? 'Opening…' : button.text),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ChestUnlockOption {
  const _ChestUnlockOption({
    required this.method,
    required this.title,
    required this.subtitle,
    required this.strength,
    this.detail,
    this.ownedInventoryItemId,
    this.spellId,
    this.iconUrl = '',
    this.consumesItem = false,
  });

  final String method;
  final String title;
  final String subtitle;
  final String? detail;
  final String? ownedInventoryItemId;
  final String? spellId;
  final String iconUrl;
  final bool consumesItem;
  final int strength;
}

class _UnlockOptionAvatar extends StatelessWidget {
  const _UnlockOptionAvatar({required this.option});

  final _ChestUnlockOption option;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final icon = option.method == _unlockMethodSpell
        ? Icons.auto_fix_high
        : Icons.vpn_key;
    return Container(
      width: 48,
      height: 48,
      decoration: BoxDecoration(
        color: theme.colorScheme.primary.withOpacity(0.12),
        borderRadius: BorderRadius.circular(14),
      ),
      clipBehavior: Clip.antiAlias,
      child: option.iconUrl.isNotEmpty
          ? Image.network(
              option.iconUrl,
              fit: BoxFit.cover,
              errorBuilder: (_, _, _) =>
                  Icon(icon, color: theme.colorScheme.primary),
            )
          : Icon(icon, color: theme.colorScheme.primary),
    );
  }
}

class _UnlockMethodChip extends StatelessWidget {
  const _UnlockMethodChip({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: theme.colorScheme.primary.withOpacity(0.1),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: theme.textTheme.labelMedium?.copyWith(
          color: theme.colorScheme.primary,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

class _InfoChip extends StatelessWidget {
  const _InfoChip({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceVariant.withOpacity(0.6),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: theme.colorScheme.outline.withOpacity(0.2)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16),
          const SizedBox(width: 6),
          Text(
            label,
            style: theme.textTheme.bodySmall?.copyWith(
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }
}
