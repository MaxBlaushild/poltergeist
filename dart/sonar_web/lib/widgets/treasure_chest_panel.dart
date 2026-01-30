import 'dart:math' as math;

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/inventory_item.dart';
import '../models/treasure_chest.dart';
import '../providers/location_provider.dart';
import '../services/inventory_service.dart';
import '../services/poi_service.dart';

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
    if (widget.treasureChest.openedByUser == true) {
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
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Treasure chest opened!')),
      );
      await Future.delayed(const Duration(milliseconds: 800));
      if (!mounted) return;
      widget.onClose();
    } catch (e) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = _errorMessage(e);
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
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

    return DraggableScrollableSheet(
      initialChildSize: 0.5,
      minChildSize: 0.3,
      maxChildSize: 0.7,
      builder: (_, scrollController) => Container(
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
        ),
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
                    Center(
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(12),
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
                    ),
                    const SizedBox(height: 16),
                    if (distance != null)
                      Text(
                        'Distance: ${distance.round()} m',
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                    if (widget.treasureChest.gold != null &&
                        widget.treasureChest.gold! > 0) ...[
                      const SizedBox(height: 4),
                      Text(
                        'Gold: ${widget.treasureChest.gold}',
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                    ],
                    if (widget.treasureChest.items.isNotEmpty) ...[
                      const SizedBox(height: 4),
                      Text(
                        'Items: ${widget.treasureChest.items.length}',
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                    ],
                    if (widget.treasureChest.unlockTier != null) ...[
                      const SizedBox(height: 4),
                      Text(
                        'Requires unlock tier: ${widget.treasureChest.unlockTier}',
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                    ],
                    if (_error != null) ...[
                      const SizedBox(height: 12),
                      Text(
                        _error!,
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.error,
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
