import 'dart:async';
import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/inventory_item.dart';
import '../models/treasure_chest.dart';
import '../providers/location_provider.dart';
import '../services/inventory_service.dart';
import '../services/poi_service.dart';
import '../widgets/paper_texture.dart';

const _chestImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/inventory-items/1762314753387-0gdf0170kq5m.png';

const _openRadiusMeters = 10.0;

class TreasureChestPanel extends StatefulWidget {
  const TreasureChestPanel({
    super.key,
    required this.treasureChest,
    required this.onClose,
  });

  final TreasureChest treasureChest;
  final VoidCallback onClose;

  @override
  State<TreasureChestPanel> createState() => _TreasureChestPanelState();
}

class _TreasureChestPanelState extends State<TreasureChestPanel> {
  bool _loading = false;
  bool _opened = false;
  bool _showCongrats = false;
  String? _error;
  List<InventoryItem> _inventoryItems = [];
  List<OwnedInventoryItem> _ownedItems = [];
  Timer? _congratsTimer;

  @override
  void initState() {
    super.initState();
    _loadInventory();
  }

  @override
  void dispose() {
    _congratsTimer?.cancel();
    super.dispose();
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

  double _calculateDistance(double lat1, double lon1, double lat2, double lon2) {
    const R = 6371e3;
    final phi1 = lat1 * math.pi / 180;
    final phi2 = lat2 * math.pi / 180;
    final dPhi = (lat2 - lat1) * math.pi / 180;
    final dLambda = (lon2 - lon1) * math.pi / 180;
    final a = math.sin(dPhi / 2) * math.sin(dPhi / 2) +
        math.cos(phi1) * math.cos(phi2) *
            math.sin(dLambda / 2) * math.sin(dLambda / 2);
    final c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a));
    return R * c;
  }

  bool get _hasUnlockItem {
    final t = widget.treasureChest.unlockTier;
    if (t == null) return true;
    for (final o in _ownedItems) {
      if (o.quantity <= 0) continue;
      InventoryItem? inv;
      for (final i in _inventoryItems) {
        if (i.id == o.inventoryItemId) {
          inv = i;
          break;
        }
      }
      if (inv != null &&
          inv.unlockTier != null &&
          inv.unlockTier! >= t) {
        return true;
      }
    }
    return false;
  }

  ({String text, bool disabled}) _buttonState(bool isWithinRange) {
    if (widget.treasureChest.openedByUser == true || _opened) {
      return (text: 'Already opened', disabled: true);
    }
    if (!isWithinRange) {
      return (text: 'Too far away', disabled: true);
    }
    if (widget.treasureChest.unlockTier == null) {
      return (text: 'Open Chest', disabled: false);
    }
    if (_hasUnlockItem) {
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

  Future<void> _openChest() async {
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
      await context.read<PoiService>().openTreasureChest(widget.treasureChest.id);
      if (!mounted) return;
      await _loadInventory();
      if (!mounted) return;
      _congratsTimer?.cancel();
      setState(() {
        _loading = false;
        _opened = true;
        _showCongrats = true;
      });
      _congratsTimer = Timer(const Duration(seconds: 2), () {
        if (!mounted) return;
        setState(() => _showCongrats = false);
      });
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

  List<Widget> _buildLootTiles(ThemeData theme) {
    final tiles = <Widget>[];
    final gold = widget.treasureChest.gold ?? 0;
    if (gold > 0) {
      tiles.add(
        _RewardTile(
          icon: Icons.paid_rounded,
          title: '${gold} gold',
          subtitle: 'Added to your purse',
          color: const Color(0xFFF5C542),
        ),
      );
    }
    for (final chestItem in widget.treasureChest.items) {
      if (chestItem.quantity <= 0) continue;
      final inv = _inventoryItemForId(chestItem.inventoryItemId);
      final name = inv?.name ?? 'Item #${chestItem.inventoryItemId}';
      final imageUrl = inv?.imageUrl ?? '';
      tiles.add(
        _RewardTile(
          icon: Icons.auto_awesome,
          title: '$name ×${chestItem.quantity}',
          subtitle: inv?.effectText.isNotEmpty == true
              ? inv!.effectText
              : (inv?.flavorText.isNotEmpty == true ? inv!.flavorText : 'Added to inventory'),
          imageUrl: imageUrl.isNotEmpty ? imageUrl : null,
          color: theme.colorScheme.primary,
        ),
      );
    }
    if (tiles.isEmpty) {
      tiles.add(
        _RewardTile(
          icon: Icons.savings_outlined,
          title: 'No loot this time',
          subtitle: 'Better luck on the next chest.',
          color: theme.colorScheme.tertiary,
        ),
      );
    }
    return tiles;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final loc = context.watch<LocationProvider>().location;
    final distance = loc != null
        ? _calculateDistance(
            loc.latitude,
            loc.longitude,
            widget.treasureChest.latitude,
            widget.treasureChest.longitude,
          )
        : null;
    final isWithinRange = distance != null && distance <= _openRadiusMeters;
    final button = _buttonState(isWithinRange);
    final isOpened = widget.treasureChest.openedByUser == true || _opened;

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
                                  color: theme.colorScheme.primary.withOpacity(0.08),
                                  boxShadow: [
                                    BoxShadow(
                                      color: theme.colorScheme.primary.withOpacity(0.35),
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
                                    child: const Icon(Icons.inventory_2_outlined, size: 48),
                                  ),
                                ),
                              ),
                              if (_showCongrats)
                                Positioned(
                                  top: 0,
                                  child: AnimatedScale(
                                    duration: const Duration(milliseconds: 250),
                                    scale: _showCongrats ? 1 : 0.9,
                                    child: AnimatedOpacity(
                                      duration: const Duration(milliseconds: 250),
                                      opacity: _showCongrats ? 1 : 0,
                                      child: Container(
                                        padding: const EdgeInsets.symmetric(
                                          horizontal: 12,
                                          vertical: 6,
                                        ),
                                        decoration: BoxDecoration(
                                          color: const Color(0xFFF5C542),
                                          borderRadius: BorderRadius.circular(999),
                                          boxShadow: [
                                            BoxShadow(
                                              color: Colors.black.withOpacity(0.15),
                                              blurRadius: 12,
                                            ),
                                          ],
                                        ),
                                        child: Row(
                                          mainAxisSize: MainAxisSize.min,
                                          children: const [
                                            Icon(Icons.auto_awesome, size: 16),
                                            SizedBox(width: 6),
                                            Text(
                                              'Chest opened!',
                                              style: TextStyle(
                                                fontWeight: FontWeight.bold,
                                              ),
                                            ),
                                          ],
                                        ),
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
                                    : theme.colorScheme.onSurface.withOpacity(0.7),
                              ),
                              const SizedBox(width: 8),
                              Text(
                                isOpened ? 'Loot claimed' : 'Awaiting discovery',
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
                                'Unlock tier ${widget.treasureChest.unlockTier}',
                          ),
                        if (widget.treasureChest.gold != null &&
                            widget.treasureChest.gold! > 0)
                          _InfoChip(
                            icon: Icons.paid,
                            label: '${widget.treasureChest.gold} gold',
                          ),
                        if (widget.treasureChest.items.isNotEmpty)
                          _InfoChip(
                            icon: Icons.inventory_2_outlined,
                            label:
                                '${widget.treasureChest.items.length} item types',
                          ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'Rewards',
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 8),
                    ..._buildLootTiles(theme),
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
                      onPressed:
                          button.disabled || _loading ? null : _openChest,
                      child: Text(
                        _loading ? 'Opening…' : button.text,
                      ),
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

class _InfoChip extends StatelessWidget {
  const _InfoChip({
    required this.icon,
    required this.label,
  });

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
        border: Border.all(
          color: theme.colorScheme.outline.withOpacity(0.2),
        ),
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

class _RewardTile extends StatelessWidget {
  const _RewardTile({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.color,
    this.imageUrl,
  });

  final IconData icon;
  final String title;
  final String subtitle;
  final Color color;
  final String? imageUrl;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceVariant.withOpacity(0.35),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: theme.colorScheme.outline.withOpacity(0.15),
        ),
      ),
      child: Row(
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: color.withOpacity(0.15),
              borderRadius: BorderRadius.circular(12),
            ),
            child: imageUrl != null
                ? ClipRRect(
                    borderRadius: BorderRadius.circular(12),
                    child: Image.network(
                      imageUrl!,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Icon(icon, color: color),
                    ),
                  )
                : Icon(icon, color: color),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  subtitle,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: theme.colorScheme.onSurface.withOpacity(0.7),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
