import 'dart:async';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import 'package:provider/provider.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/inventory_item.dart';
import '../models/outfit_generation.dart';
import '../providers/auth_provider.dart';
import '../services/media_service.dart';
import '../services/inventory_service.dart';
import '../utils/camera_capture.dart';
import '../constants/api_constants.dart';

/// Inventory item IDs that can be "Used" from the inventory menu (match JS ItemsUsabledInMenu).
const _itemsUsableInMenu = <int>{
  1,  // CipherOfTheLaughingMonkey
  6,  // CortezsCutlass
  7,  // RustedMusket
  9,  // Dagger
  12, // Ale
  14, // WickedSpellbook
};

class InventoryPanel extends StatefulWidget {
  const InventoryPanel({
    super.key,
    required this.onClose,
  });

  final VoidCallback onClose;

  @override
  State<InventoryPanel> createState() => _InventoryPanelState();
}

class _InventoryPanelState extends State<InventoryPanel> {
  static const String _dismissedOutfitStatusPrefsKey =
      'dismissed_outfit_statuses';
  List<InventoryItem> _items = [];
  List<OwnedInventoryItem> _owned = [];
  bool _loading = true;
  bool _using = false;
  String? _error;
  OwnedInventoryItem? _selected;
  OutfitGeneration? _outfitGeneration;
  bool _loadingOutfitStatus = false;
  String? _outfitError;
  Timer? _outfitPoller;
  final Set<String> _dismissedOutfitStatuses = {};

  @override
  void initState() {
    super.initState();
    _loadDismissedOutfitStatuses();
    _load();
  }

  @override
  void dispose() {
    final selectedId = _selected?.id;
    final status = _outfitGeneration;
    if (selectedId != null &&
        status != null &&
        (status.isComplete || status.isFailed)) {
      _dismissedOutfitStatuses.add('$selectedId::${status.status}');
      _persistDismissedOutfitStatuses();
    }
    _outfitPoller?.cancel();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final authProvider = context.read<AuthProvider>();
      final svc = context.read<InventoryService>();
      final itemsFuture = svc.getInventoryItems();
      final ownedFuture = svc.getOwnedInventoryItems();
      try {
        await authProvider.refresh();
      } catch (_) {}
      final items = await itemsFuture;
      final owned = await ownedFuture;
      if (!mounted) return;
      setState(() {
        _items = items;
        _owned = owned.where((o) => o.quantity > 0).toList();
        if (_selected != null) {
          final stillOwned = _owned.any((o) => o.id == _selected!.id);
          if (!stillOwned) {
            _selected = null;
            _outfitGeneration = null;
            _outfitError = null;
            _outfitPoller?.cancel();
          }
        }
        _loading = false;
      });
    } catch (e) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = e.toString();
        });
      }
    }
  }

  Future<void> _loadDismissedOutfitStatuses() async {
    final prefs = await SharedPreferences.getInstance();
    final list = prefs.getStringList(_dismissedOutfitStatusPrefsKey) ?? [];
    if (!mounted) return;
    setState(() {
      _dismissedOutfitStatuses
        ..clear()
        ..addAll(list);
    });
  }

  Future<void> _persistDismissedOutfitStatuses() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setStringList(
      _dismissedOutfitStatusPrefsKey,
      _dismissedOutfitStatuses.toList(),
    );
  }

  InventoryItem? _itemFor(OwnedInventoryItem o) {
    for (final i in _items) {
      if (i.id == o.inventoryItemId) return i;
    }
    return null;
  }

  bool _isUsableInMenu(int inventoryItemId) {
    return _itemsUsableInMenu.contains(inventoryItemId);
  }

  bool _isOutfitItem(InventoryItem inv) {
    String normalize(String input) {
      final lower = input.toLowerCase();
      return lower.replaceAll(RegExp(r'[^a-z]'), '');
    }

    final name = normalize(inv.name);
    if (name.contains('outfit')) return true;

    final flavor = normalize(inv.flavorText);
    if (flavor.contains('outfit')) return true;

    final effect = normalize(inv.effectText);
    if (effect.contains('outfit')) return true;

    return false;
  }

  Future<void> _loadOutfitStatus(String ownedInventoryItemId, {bool silent = false}) async {
    if (!silent) {
      setState(() => _loadingOutfitStatus = true);
    }
    try {
      final status = await context
          .read<InventoryService>()
          .getOutfitGenerationStatus(ownedInventoryItemId);
      if (!mounted) return;
      bool clearedDismissal = false;
      setState(() {
        _outfitGeneration = status;
        _loadingOutfitStatus = false;
        _outfitError = status?.error;
        if (status != null && status.isPending) {
          _dismissedOutfitStatuses
              .removeWhere((key) => key.startsWith('$ownedInventoryItemId::'));
          clearedDismissal = true;
        }
      });
      if (clearedDismissal) {
        await _persistDismissedOutfitStatuses();
      }
      if (status != null && status.isPending) {
        _startOutfitPolling(ownedInventoryItemId);
      } else {
        _outfitPoller?.cancel();
      }
      if (status != null && status.isComplete) {
        await context.read<AuthProvider>().refresh();
        if (!mounted) return;
        await _load();
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loadingOutfitStatus = false;
        _outfitError = e.toString();
      });
    }
  }

  bool _isOutfitStatusDismissed(String ownedId, OutfitGeneration status) {
    if (!(status.isComplete || status.isFailed)) {
      return false;
    }
    return _dismissedOutfitStatuses.contains('$ownedId::${status.status}');
  }

  void _startOutfitPolling(String ownedInventoryItemId) {
    _outfitPoller?.cancel();
    _outfitPoller = Timer.periodic(const Duration(seconds: 5), (_) {
      _loadOutfitStatus(ownedInventoryItemId, silent: true);
    });
  }

  String _extensionFromMime(String? mimeType, String? name) {
    if (name != null && name.contains('.')) {
      return name.split('.').last.toLowerCase();
    }
    switch (mimeType) {
      case 'image/png':
        return 'png';
      case 'image/webp':
        return 'webp';
      default:
        return 'jpg';
    }
  }

  Future<void> _useOutfitItem(OwnedInventoryItem owned, InventoryItem inv) async {
    if (_using) return;
    setState(() {
      _using = true;
      _error = null;
      _outfitError = null;
    });
    try {
      final captured = await captureImageFromCamera(useFrontCamera: true);
      if (captured == null) {
        if (mounted) {
          setState(() => _using = false);
        }
        return;
      }
      final mediaService = context.read<MediaService>();
      final userId = context.read<AuthProvider>().user?.id ?? 'anonymous';
      final ext = _extensionFromMime(captured.mimeType, captured.name);
      final key =
          'selfies/$userId/${DateTime.now().millisecondsSinceEpoch}.$ext';
      final presigned = await mediaService.getPresignedUploadUrl(
        ApiConstants.crewProfileBucket,
        key,
      );
      if (presigned == null) {
        if (mounted) {
          setState(() {
            _using = false;
            _outfitError = 'Failed to prepare selfie upload.';
          });
        }
        return;
      }
      final ok = await mediaService.uploadToPresigned(
        presigned,
        Uint8List.fromList(captured.bytes),
        captured.mimeType ?? 'image/jpeg',
      );
      if (!ok) {
        if (mounted) {
          setState(() {
            _using = false;
            _outfitError = 'Failed to upload selfie.';
          });
        }
        return;
      }
      final selfieUrl = presigned.split('?').first;
      final status = await context.read<InventoryService>().useOutfitItem(
            owned.id,
            selfieUrl: selfieUrl,
          );
      if (!mounted) return;
      setState(() {
        _using = false;
        _outfitGeneration = status;
      });
      if (status.isPending) {
        _startOutfitPolling(owned.id);
      }
    } catch (e) {
      String message = e.toString();
      if (e is DioException) {
        final data = e.response?.data;
        if (data is Map<String, dynamic>) {
          final err = data['error'];
          if (err is String && err.trim().isNotEmpty) {
            message = err;
          }
        }
      }
      if (mounted) {
        setState(() {
          _using = false;
          _outfitGeneration = null;
          _outfitError = message;
        });
      }
    }
  }

  Future<void> _use(OwnedInventoryItem owned) async {
    if (_using) return;
    setState(() {
      _using = true;
      _error = null;
    });
    try {
      await context.read<InventoryService>().useItem(owned.id);
      if (!mounted) return;
      await context.read<AuthProvider>().refresh();
      if (!mounted) return;
      await _load();
      if (!mounted) return;
      widget.onClose();
    } catch (e) {
      String message = e.toString();
      if (e is DioException) {
        final data = e.response?.data;
        if (data is Map<String, dynamic>) {
          final err = data['error'];
          if (err is String && err.trim().isNotEmpty) {
            message = err;
          }
        }
      }
      if (message.toLowerCase().contains('outfit items require a selfie')) {
        final inv = _itemFor(owned);
        if (inv != null) {
          if (mounted) {
            setState(() {
              _using = false;
              _error = null;
            });
          }
          await _useOutfitItem(owned, inv);
          return;
        }
      }
      if (mounted) {
        setState(() {
          _using = false;
          _error = message;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final user = context.watch<AuthProvider>().user;
    final gold = user?.gold ?? 0;
    final showDetail = _selected != null;
    final inv = showDetail ? _itemFor(_selected!) : null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (showDetail)
          Align(
            alignment: Alignment.centerLeft,
            child: IconButton(
              icon: const Icon(Icons.arrow_back),
              onPressed: () async {
                _outfitPoller?.cancel();
                final selectedId = _selected?.id;
                final status = _outfitGeneration;
                if (selectedId != null &&
                    status != null &&
                    (status.isComplete || status.isFailed)) {
                  _dismissedOutfitStatuses.add('$selectedId::${status.status}');
                  await _persistDismissedOutfitStatuses();
                }
                setState(() {
                  _selected = null;
                  _outfitGeneration = null;
                  _outfitError = null;
                });
              },
            ),
          ),
        if (showDetail || user != null)
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              if (showDetail && inv != null)
                Expanded(
                  child: Text(
                    inv.name,
                    style: Theme.of(context).textTheme.titleLarge?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                  ),
                )
              else
                const SizedBox.shrink(),
              if (user != null)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.surfaceContainerHighest,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: Colors.amber.shade400,
                    width: 1.5,
                  ),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.monetization_on, size: 20, color: Colors.amber.shade700),
                    const SizedBox(width: 6),
                    Text(
                      'GOLD',
                      style: Theme.of(context).textTheme.labelMedium?.copyWith(
                            color: Colors.amber.shade800,
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                    const SizedBox(width: 6),
                    Text(
                      '$gold',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        if (_error != null) ...[
          const SizedBox(height: 8),
          Text(
            _error!,
            style: TextStyle(color: Theme.of(context).colorScheme.error),
          ),
        ],
        const SizedBox(height: 16),
        Expanded(
          child: _loading
              ? const Center(
                  child: Padding(
                    padding: EdgeInsets.all(24),
                    child: CircularProgressIndicator(),
                  ),
                )
              : showDetail && inv != null
                  ? _buildDetail(context, inv, _selected!)
                  : _buildGrid(context),
        ),
      ],
    );
  }

  Widget _buildGrid(BuildContext context) {
    const slots = 12;
    const crossAxisCount = 3;
    final theme = Theme.of(context);
    final filled = _owned.length.clamp(0, slots);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'Inventory',
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            Text(
              '$filled / $slots slots',
              style: theme.textTheme.labelLarge?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Expanded(
          child: LayoutGrid(
            crossAxisCount: crossAxisCount,
            childAspectRatio: 1,
            children: List.generate(slots, (i) {
              if (i >= _owned.length) {
                return _buildEmptySlot(context);
              }
              final o = _owned[i];
              final inv = _itemFor(o);
              if (inv == null) {
                return _buildEmptySlot(context);
              }
              return _buildFilledSlot(context, inv, o);
            }),
          ),
        ),
      ],
    );
  }

  Widget _buildEmptySlot(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest.withOpacity(0.4),
        border: Border.all(color: theme.dividerColor),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.inventory_2_outlined,
              size: 28,
              color: theme.colorScheme.onSurfaceVariant,
            ),
            const SizedBox(height: 6),
            Text(
              'Empty',
              style: theme.textTheme.labelSmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFilledSlot(
    BuildContext context,
    InventoryItem inv,
    OwnedInventoryItem owned,
  ) {
    final theme = Theme.of(context);
    return InkWell(
      onTap: () {
        setState(() {
          _selected = owned;
          _outfitGeneration = null;
          _outfitError = null;
        });
        if (_isOutfitItem(inv)) {
          _loadOutfitStatus(owned.id);
        } else {
          _outfitPoller?.cancel();
        }
      },
      borderRadius: BorderRadius.circular(12),
      child: Container(
        decoration: BoxDecoration(
          color: theme.colorScheme.surfaceContainerHighest,
          border: Border.all(color: theme.dividerColor),
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(
              color: theme.shadowColor.withOpacity(0.08),
              blurRadius: 10,
              offset: const Offset(0, 6),
            ),
          ],
        ),
        child: Stack(
          fit: StackFit.expand,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(10, 10, 10, 30),
              child: Image.network(
                inv.imageUrl,
                fit: BoxFit.contain,
                errorBuilder: (_, __, ___) => Icon(
                  Icons.inventory_2_outlined,
                  size: 32,
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
            ),
            Positioned(
              top: 6,
              right: 6,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.black87,
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  '${owned.quantity}',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
            Positioned(
              left: 0,
              right: 0,
              bottom: 0,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
                decoration: BoxDecoration(
                  color: theme.colorScheme.surface.withOpacity(0.92),
                  borderRadius: const BorderRadius.only(
                    bottomLeft: Radius.circular(12),
                    bottomRight: Radius.circular(12),
                  ),
                  border: Border(
                    top: BorderSide(color: theme.dividerColor),
                  ),
                ),
                child: Text(
                  inv.name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.labelMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDetail(
    BuildContext context,
    InventoryItem inv,
    OwnedInventoryItem owned,
  ) {
    final theme = Theme.of(context);
    final isOutfit = _isOutfitItem(inv);
    final outfitStatus = _outfitGeneration;
    final outfitPending = isOutfit && (outfitStatus?.isPending ?? false);
    final canUse =
        owned.quantity > 0 && (isOutfit || _isUsableInMenu(inv.id)) && !outfitPending;
    final hasDetails = inv.flavorText.isNotEmpty || inv.effectText.isNotEmpty;

    final metaChips = <Widget>[
      _buildMetaChip(
        context,
        icon: Icons.inventory_2_outlined,
        label: 'Owned ${owned.quantity}',
      ),
      if (inv.sellValue != null)
        _buildMetaChip(
          context,
          icon: Icons.sell_outlined,
          label: 'Sell ${inv.sellValue}',
        ),
      if (inv.unlockTier != null)
        _buildMetaChip(
          context,
          icon: Icons.lock_open,
          label: 'Tier ${inv.unlockTier}',
        ),
      if (isOutfit)
        _buildMetaChip(
          context,
          icon: Icons.auto_awesome,
          label: 'Outfit item',
        ),
    ];

    Widget actionArea = const SizedBox.shrink();
    if (isOutfit) {
      actionArea = Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _buildOutfitStatus(context, outfitStatus),
          const SizedBox(height: 12),
          FilledButton(
            onPressed: _using || outfitPending ? null : () => _useOutfitItem(owned, inv),
            child: Text(
              outfitPending
                  ? 'Generating…'
                  : _using
                      ? 'Using…'
                      : (outfitStatus?.isFailed == true ? 'Try Again' : 'Take Selfie'),
            ),
          ),
        ],
      );
    } else if (canUse) {
      actionArea = FilledButton(
        onPressed: _using ? null : () => _use(owned),
        child: Text(_using ? 'Using…' : 'Use'),
      );
    }

    return SingleChildScrollView(
      child: LayoutBuilder(
        builder: (context, constraints) {
          final isWide = constraints.maxWidth >= 520;
          final image = _buildDetailImage(context, inv);
          final metaRow = Wrap(
            spacing: 8,
            runSpacing: 8,
            children: metaChips,
          );
          final infoColumn = Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              metaRow,
              if (actionArea is! SizedBox) const SizedBox(height: 16),
              if (actionArea is! SizedBox) actionArea,
            ],
          );

          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (isWide)
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(flex: 5, child: image),
                    const SizedBox(width: 16),
                    Expanded(flex: 7, child: infoColumn),
                  ],
                )
              else ...[
                image,
                const SizedBox(height: 16),
                infoColumn,
              ],
              if (hasDetails) const SizedBox(height: 16),
              if (inv.flavorText.isNotEmpty)
                _buildDetailSection(
                  context,
                  title: 'Story',
                  child: Text(
                    inv.flavorText,
                    style: theme.textTheme.bodyLarge,
                  ),
                ),
              if (inv.flavorText.isNotEmpty && inv.effectText.isNotEmpty)
                const SizedBox(height: 12),
              if (inv.effectText.isNotEmpty)
                _buildDetailSection(
                  context,
                  title: 'Effect',
                  child: Text(
                    inv.effectText,
                    style: theme.textTheme.bodyLarge,
                  ),
                ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildDetailImage(BuildContext context, InventoryItem inv) {
    final theme = Theme.of(context);
    return Container(
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: theme.dividerColor),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(12),
        child: AspectRatio(
          aspectRatio: 1,
          child: Image.network(
            inv.imageUrl,
            fit: BoxFit.contain,
            errorBuilder: (_, __, ___) => Center(
              child: Icon(
                Icons.inventory_2_outlined,
                size: 64,
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildMetaChip(
    BuildContext context, {
    required IconData icon,
    required String label,
  }) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: theme.dividerColor),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16, color: theme.colorScheme.onSurfaceVariant),
          const SizedBox(width: 6),
          Text(
            label,
            style: theme.textTheme.labelLarge?.copyWith(
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailSection(
    BuildContext context, {
    required String title,
    required Widget child,
  }) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: theme.dividerColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title.toUpperCase(),
            style: theme.textTheme.labelLarge?.copyWith(
              letterSpacing: 0.6,
              fontWeight: FontWeight.w700,
              color: theme.colorScheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(height: 8),
          child,
        ],
      ),
    );
  }

  Widget _buildOutfitStatus(BuildContext context, OutfitGeneration? status) {
    final theme = Theme.of(context);
    final surface = theme.colorScheme.surfaceVariant;
    final textStyle = theme.textTheme.bodyMedium;
    final selectedId = _selected?.id;
    if (_loadingOutfitStatus) {
      return Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: surface,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            const SizedBox(
              width: 20,
              height: 20,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
            const SizedBox(width: 12),
            Text('Checking portrait status…', style: textStyle),
          ],
        ),
      );
    }

    if (status == null) {
      return Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: surface,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Use this outfit to craft a personalized portrait from your selfie.',
              style: textStyle,
            ),
            if (_outfitError != null) ...[
              const SizedBox(height: 6),
              Text(
                _outfitError!,
                style: textStyle?.copyWith(color: theme.colorScheme.error),
              ),
            ],
          ],
        ),
      );
    }

    if (selectedId != null && _isOutfitStatusDismissed(selectedId, status)) {
      return const SizedBox.shrink();
    }

    if (status.isPending) {
      return Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: surface,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('The dungeon artist is at work…', style: textStyle),
            const SizedBox(height: 8),
            const LinearProgressIndicator(),
          ],
        ),
      );
    }

    if (status.isFailed) {
      return Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: surface,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Portrait generation failed.', style: textStyle),
            if (status.error != null) ...[
              const SizedBox(height: 6),
              Text(
                status.error!,
                style: textStyle?.copyWith(color: theme.colorScheme.error),
              ),
            ],
          ],
        ),
      );
    }

    if (status.isComplete) {
      return Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: surface,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Text(
          'Your new portrait is ready and equipped.',
          style: textStyle,
        ),
      );
    }

    return const SizedBox.shrink();
  }
}

/// Simple grid with fixed cross-axis count and aspect ratio.
class LayoutGrid extends StatelessWidget {
  const LayoutGrid({
    super.key,
    required this.crossAxisCount,
    required this.childAspectRatio,
    required this.children,
  });

  final int crossAxisCount;
  final double childAspectRatio;
  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    return GridView.count(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      crossAxisCount: crossAxisCount,
      mainAxisSpacing: 8,
      crossAxisSpacing: 8,
      childAspectRatio: childAspectRatio,
      children: children,
    );
  }
}
