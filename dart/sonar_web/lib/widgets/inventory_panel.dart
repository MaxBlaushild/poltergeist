import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/inventory_item.dart';
import '../providers/auth_provider.dart';
import '../services/inventory_service.dart';

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
  List<InventoryItem> _items = [];
  List<OwnedInventoryItem> _owned = [];
  bool _loading = true;
  bool _using = false;
  String? _error;
  OwnedInventoryItem? _selected;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final svc = context.read<InventoryService>();
      final items = await svc.getInventoryItems();
      final owned = await svc.getOwnedInventoryItems();
      if (!mounted) return;
      setState(() {
        _items = items;
        _owned = owned.where((o) => o.quantity > 0).toList();
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

  InventoryItem? _itemFor(OwnedInventoryItem o) {
    for (final i in _items) {
      if (i.id == o.inventoryItemId) return i;
    }
    return null;
  }

  bool _isUsableInMenu(int inventoryItemId) {
    return _itemsUsableInMenu.contains(inventoryItemId);
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
      if (mounted) {
        setState(() {
          _using = false;
          _error = e.toString();
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
              onPressed: () => setState(() => _selected = null),
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

    return LayoutGrid(
      crossAxisCount: crossAxisCount,
      childAspectRatio: 1,
      children: List.generate(slots, (i) {
        if (i >= _owned.length) {
          return Container(
            decoration: BoxDecoration(
              border: Border.all(color: Theme.of(context).dividerColor),
              borderRadius: BorderRadius.circular(12),
            ),
          );
        }
        final o = _owned[i];
        final inv = _itemFor(o);
        if (inv == null) {
          return Container(
            decoration: BoxDecoration(
              border: Border.all(color: Theme.of(context).dividerColor),
              borderRadius: BorderRadius.circular(12),
            ),
          );
        }
        return InkWell(
          onTap: () => setState(() => _selected = o),
          borderRadius: BorderRadius.circular(12),
          child: Container(
            decoration: BoxDecoration(
              border: Border.all(color: Theme.of(context).dividerColor),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Stack(
              fit: StackFit.expand,
              children: [
                Padding(
                  padding: const EdgeInsets.all(8),
                  child: Image.network(
                    inv.imageUrl,
                    fit: BoxFit.contain,
                    errorBuilder: (_, __, ___) => Icon(
                      Icons.inventory_2_outlined,
                      size: 32,
                      color: Theme.of(context).colorScheme.onSurfaceVariant,
                    ),
                  ),
                ),
                Positioned(
                  right: 6,
                  bottom: 6,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: Colors.black87,
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      '${o.quantity}',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      }),
    );
  }

  Widget _buildDetail(
    BuildContext context,
    InventoryItem inv,
    OwnedInventoryItem owned,
  ) {
    final canUse = _isUsableInMenu(inv.id) && owned.quantity > 0;

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(12),
            child: AspectRatio(
              aspectRatio: 1.2,
              child: Image.network(
                inv.imageUrl,
                fit: BoxFit.contain,
                errorBuilder: (_, __, ___) => Container(
                  color: Theme.of(context).colorScheme.surfaceContainerHighest,
                  child: Icon(
                    Icons.inventory_2_outlined,
                    size: 64,
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(height: 16),
          if (inv.flavorText.isNotEmpty)
            Text(
              inv.flavorText,
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                  ),
            ),
          if (inv.flavorText.isNotEmpty) const SizedBox(height: 8),
          if (inv.effectText.isNotEmpty)
            Text(
              inv.effectText,
              style: Theme.of(context).textTheme.bodyLarge,
            ),
          if (inv.effectText.isNotEmpty) const SizedBox(height: 20),
          if (canUse)
            FilledButton(
              onPressed: _using ? null : () => _use(owned),
              child: Text(_using ? 'Using…' : 'Use'),
            ),
        ],
      ),
    );
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
