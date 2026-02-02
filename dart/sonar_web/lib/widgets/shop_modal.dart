import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';

import '../models/character.dart';
import '../models/character_action.dart';
import '../models/inventory_item.dart';
import '../providers/auth_provider.dart';
import '../services/inventory_service.dart';
import '../services/poi_service.dart';

class ShopModal extends StatefulWidget {
  const ShopModal({
    super.key,
    required this.character,
    required this.action,
    required this.onClose,
  });

  final Character character;
  final CharacterAction action;
  final VoidCallback onClose;

  @override
  State<ShopModal> createState() => _ShopModalState();
}

class _ShopModalState extends State<ShopModal> {
  final FocusNode _focusNode = FocusNode();
  String _activeTab = 'buy';
  int? _purchasingItemId;
  int? _sellingItemId;
  String? _error;
  String? _success;
  List<InventoryItem> _inventoryItems = [];
  List<OwnedInventoryItem> _ownedItems = [];

  @override
  void initState() {
    super.initState();
    _loadInventory();
  }

  @override
  void dispose() {
    _focusNode.dispose();
    super.dispose();
  }

  Future<void> _loadInventory() async {
    final svc = context.read<InventoryService>();
    final items = await svc.getInventoryItems();
    final owned = await svc.getOwnedInventoryItems();
    if (mounted) {
      setState(() {
        _inventoryItems = items;
        _ownedItems = owned;
      });
    }
  }

  InventoryItem? _getItemById(int id) {
    return _inventoryItems.firstWhere((i) => i.id == id, orElse: () => InventoryItem(
      id: 0,
      name: 'Unknown',
      imageUrl: '',
      flavorText: '',
      effectText: '',
    ));
  }

  Future<void> _purchase(int itemId, int price) async {
    final user = context.read<AuthProvider>().user;
    if (user == null) {
      setState(() => _error = 'You must be logged in to purchase items');
      return;
    }
    if (user.gold < price) {
      setState(() => _error = 'Insufficient gold');
      return;
    }
    setState(() {
      _purchasingItemId = itemId;
      _error = null;
      _success = null;
    });
    try {
      await context.read<PoiService>().purchaseFromShop(
            widget.action.id,
            itemId,
          );
      if (mounted) {
        await _loadInventory();
        await context.read<AuthProvider>().refresh();
        setState(() {
          _purchasingItemId = null;
          _success =
              'Purchased ${_getItemById(itemId)?.name ?? 'item'} for $price gold!';
        });
        Future.delayed(const Duration(seconds: 3), () {
          if (mounted) setState(() => _success = null);
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _purchasingItemId = null;
          _error = e.toString();
        });
      }
    }
  }

  Future<void> _sell(
    int itemId, {
    int quantity = 1,
  }) async {
    final user = context.read<AuthProvider>().user;
    if (user == null) {
      setState(() => _error = 'You must be logged in to sell items');
      return;
    }
    final item = _getItemById(itemId);
    final sellValue = item?.sellValue ?? 0;
    if (sellValue == 0) {
      setState(() => _error = 'Item cannot be sold');
      return;
    }
    if (quantity <= 0) {
      setState(() => _error = 'Invalid quantity');
      return;
    }
    setState(() {
      _sellingItemId = itemId;
      _error = null;
      _success = null;
    });
    try {
      await context
          .read<PoiService>()
          .sellToShop(widget.action.id, itemId, quantity: quantity);
      if (mounted) {
        await _loadInventory();
        await context.read<AuthProvider>().refresh();
        final totalValue = sellValue * quantity;
        setState(() {
          _sellingItemId = null;
          _success =
              'Sold ${quantity}x ${item?.name ?? 'item'} for $totalValue gold!';
        });
        Future.delayed(const Duration(seconds: 3), () {
          if (mounted) setState(() => _success = null);
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _sellingItemId = null;
          _error = e.toString();
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final user = context.watch<AuthProvider>().user;
    final shopItems = widget.action.shopInventory ?? [];
    final sellableItems = _ownedItems.where((o) {
      if (o.quantity <= 0) return false;
      final item = _getItemById(o.inventoryItemId);
      return (item?.sellValue ?? 0) > 0;
    }).toList();
    final characterImageUrl =
        widget.character.dialogueImageUrl ?? widget.character.mapIconUrl;
    const activeTabColor = Color(0xFF007BFF);

    return Stack(
      children: [
        Positioned.fill(
          child: GestureDetector(
            onTap: widget.onClose,
            child: Container(color: Colors.black54),
          ),
        ),
        Center(
          child: Focus(
            autofocus: true,
            focusNode: _focusNode,
            onKeyEvent: (_, event) {
              if (event is KeyDownEvent &&
                  event.logicalKey == LogicalKeyboardKey.escape) {
                widget.onClose();
                return KeyEventResult.handled;
              }
              return KeyEventResult.ignored;
            },
            child: Material(
              color: Theme.of(context).colorScheme.surface,
              borderRadius: BorderRadius.circular(16),
              elevation: 10,
              child: Container(
                width: 700,
                height: 640,
                padding: const EdgeInsets.all(24),
                child: Column(
                  children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  "${widget.character.name}'s Shop",
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                Row(
                  children: [
                    if (user != null) ...[
                      const Icon(Icons.monetization_on, size: 20),
                      const SizedBox(width: 4),
                      Text('${user.gold}'),
                      const SizedBox(width: 16),
                    ],
                    IconButton(
                      onPressed: widget.onClose,
                      icon: const Icon(Icons.close),
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 16),
            if (characterImageUrl != null && characterImageUrl.isNotEmpty) ...[
              ClipRRect(
                borderRadius: BorderRadius.circular(12),
                child: Image.network(
                  characterImageUrl,
                  width: 140,
                  height: 140,
                  fit: BoxFit.cover,
                  errorBuilder: (_, __, ___) => const Icon(Icons.image),
                ),
              ),
              const SizedBox(height: 16),
            ],
            Container(
              decoration: BoxDecoration(
                border: Border(
                  bottom: BorderSide(color: Colors.grey.shade300, width: 2),
                ),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: InkWell(
                      onTap: () => setState(() => _activeTab = 'buy'),
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          vertical: 12,
                          horizontal: 16,
                        ),
                        decoration: BoxDecoration(
                          color: _activeTab == 'buy'
                              ? activeTabColor
                              : Colors.transparent,
                          border: Border(
                            bottom: BorderSide(
                              color: _activeTab == 'buy'
                                  ? activeTabColor
                                  : Colors.transparent,
                              width: 2,
                            ),
                          ),
                        ),
                        child: Text(
                          'Buy',
                          textAlign: TextAlign.center,
                          style: TextStyle(
                            color: _activeTab == 'buy'
                                ? Colors.white
                                : Colors.grey.shade800,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: InkWell(
                      onTap: () => setState(() => _activeTab = 'sell'),
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          vertical: 12,
                          horizontal: 16,
                        ),
                        decoration: BoxDecoration(
                          color: _activeTab == 'sell'
                              ? activeTabColor
                              : Colors.transparent,
                          border: Border(
                            bottom: BorderSide(
                              color: _activeTab == 'sell'
                                  ? activeTabColor
                                  : Colors.transparent,
                              width: 2,
                            ),
                          ),
                        ),
                        child: Text(
                          'Sell',
                          textAlign: TextAlign.center,
                          style: TextStyle(
                            color: _activeTab == 'sell'
                                ? Colors.white
                                : Colors.grey.shade800,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            if (_error != null)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(
                  _error!,
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ),
            if (_success != null)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(
                  _success!,
                  style: TextStyle(color: Colors.green.shade700),
                ),
              ),
            const SizedBox(height: 16),
            Expanded(
              child: _activeTab == 'buy'
                  ? shopItems.isEmpty
                      ? const Center(
                          child: Text('This shop has no items for sale.'),
                        )
                      : ListView.builder(
                          itemCount: shopItems.length,
                          itemBuilder: (_, i) {
                            final shopItem = shopItems[i];
                            final item = _getItemById(shopItem.itemId);
                            final canAfford = user != null && user.gold >= shopItem.price;
                            final img = item?.imageUrl ?? '';
                            final name = item?.name ?? 'Unknown';
                            final flavor = item?.flavorText ?? '';
                            return Card(
                              child: ListTile(
                                leading: img.isNotEmpty
                                    ? Image.network(
                                        img,
                                        width: 48,
                                        height: 48,
                                        fit: BoxFit.cover,
                                        errorBuilder: (_, __, ___) =>
                                            const Icon(Icons.image),
                                      )
                                    : const Icon(Icons.image),
                                title: Text(name),
                                subtitle: Text(flavor),
                                trailing: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  crossAxisAlignment: CrossAxisAlignment.end,
                                  children: [
                                    Text('💰 ${shopItem.price}'),
                                    FilledButton(
                                      onPressed: canAfford &&
                                              _purchasingItemId != shopItem.itemId
                                          ? () => _purchase(shopItem.itemId, shopItem.price)
                                          : null,
                                      child: Text(
                                        _purchasingItemId == shopItem.itemId
                                            ? '...'
                                            : 'Buy',
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            );
                          },
                        )
                  : sellableItems.isEmpty
                      ? const Center(
                          child: Text('You have no items that can be sold.'),
                        )
                      : ListView.builder(
                          itemCount: sellableItems.length,
                          itemBuilder: (_, i) {
                            final owned = sellableItems[i];
                            final item = _getItemById(owned.inventoryItemId);
                            final sellValue = item?.sellValue ?? 0;
                            final img = item?.imageUrl ?? '';
                            final name = item?.name ?? 'Unknown';
                            final isSellingItem =
                                _sellingItemId == owned.inventoryItemId;
                            return Card(
                              child: ListTile(
                                leading: img.isNotEmpty
                                    ? Image.network(
                                        img,
                                        width: 48,
                                        height: 48,
                                        fit: BoxFit.cover,
                                        errorBuilder: (_, __, ___) =>
                                            const Icon(Icons.image),
                                      )
                                    : const Icon(Icons.image),
                                title: Text(name),
                                subtitle: Text('Qty: ${owned.quantity}'),
                                trailing: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  crossAxisAlignment: CrossAxisAlignment.end,
                                  children: [
                                    Text('💰 $sellValue each'),
                                    const SizedBox(height: 6),
                                    if (owned.quantity > 1)
                                      Wrap(
                                        spacing: 6,
                                        children: [
                                          FilledButton(
                                            onPressed:
                                                sellValue > 0 && !isSellingItem
                                                    ? () => _sell(
                                                          owned.inventoryItemId,
                                                          quantity: 1,
                                                        )
                                                    : null,
                                            child: Text(
                                              isSellingItem ? '...' : 'Sell 1',
                                            ),
                                          ),
                                          FilledButton(
                                            onPressed:
                                                sellValue > 0 && !isSellingItem
                                                    ? () => _sell(
                                                          owned.inventoryItemId,
                                                          quantity: owned.quantity,
                                                        )
                                                    : null,
                                            child: Text(
                                              isSellingItem
                                                  ? '...'
                                                  : 'Sell All (${owned.quantity})',
                                            ),
                                          ),
                                        ],
                                      )
                                    else
                                      FilledButton(
                                        onPressed:
                                            sellValue > 0 && !isSellingItem
                                                ? () => _sell(
                                                      owned.inventoryItemId,
                                                      quantity: 1,
                                                    )
                                                : null,
                                        child: Text(
                                          isSellingItem ? '...' : 'Sell',
                                        ),
                                      ),
                                  ],
                                ),
                              ),
                            );
                          },
                        ),
            ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}
