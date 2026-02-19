import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';

import '../models/character.dart';
import '../models/character_action.dart';
import '../models/inventory_item.dart';
import '../models/user.dart';
import '../providers/auth_provider.dart';
import '../services/inventory_service.dart';
import '../services/poi_service.dart';

/// Inventory item IDs that can be "Used" from the inventory menu (match JS ItemsUsabledInMenu).
const _itemsUsableInMenu = <int>{
  1,  // CipherOfTheLaughingMonkey
  6,  // CortezsCutlass
  7,  // RustedMusket
  9,  // Dagger
  12, // Ale
  14, // WickedSpellbook
};

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
  String _sortMode = 'name';
  bool _filterUsable = false;
  bool _filterSellable = false;
  bool _filterOutfit = false;
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
    return _inventoryItems.firstWhere(
      (i) => i.id == id,
      orElse: () => const InventoryItem(
        id: 0,
        name: 'Unknown',
        imageUrl: '',
        flavorText: '',
        effectText: '',
      ),
    );
  }

  int _ownedQuantityForItemId(int itemId) {
    for (final owned in _ownedItems) {
      if (owned.inventoryItemId == itemId) return owned.quantity;
    }
    return 0;
  }

  bool _isUsableItem(int itemId) {
    return _itemsUsableInMenu.contains(itemId);
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

  bool _passesFilters(InventoryItem item) {
    if (_filterUsable && !_isUsableItem(item.id)) return false;
    if (_filterSellable && (item.sellValue ?? 0) <= 0) return false;
    if (_filterOutfit && !_isOutfitItem(item)) return false;
    return true;
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
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final user = context.watch<AuthProvider>().user;
    final gold = user?.gold ?? 0;
    final shopItems = widget.action.shopInventory ?? [];
    final filteredShopItems = shopItems.where((shopItem) {
      final item = _getItemById(shopItem.itemId);
      if (item == null) return false;
      return _passesFilters(item);
    }).toList();
    final sellableItems = _ownedItems.where((o) {
      if (o.quantity <= 0) return false;
      final item = _getItemById(o.inventoryItemId);
      return (item?.sellValue ?? 0) > 0;
    }).toList();
    final filteredSellableItems = sellableItems.where((o) {
      final item = _getItemById(o.inventoryItemId);
      if (item == null) return false;
      return _passesFilters(item);
    }).toList();
    final characterImageUrl =
        widget.character.dialogueImageUrl ?? widget.character.mapIconUrl;

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
              color: colorScheme.surface,
              borderRadius: BorderRadius.circular(16),
              elevation: 10,
              child: Container(
                width: 760,
                height: 680,
                padding: const EdgeInsets.all(24),
                child: Column(
                  children: [
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                "${widget.character.name}'s Shop",
                                style: theme.textTheme.titleLarge?.copyWith(
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                _activeTab == 'buy'
                                    ? 'Browse curated wares and curios.'
                                    : 'Trade in gear for a fair price.',
                                style: theme.textTheme.bodyMedium?.copyWith(
                                  color: colorScheme.onSurfaceVariant,
                                ),
                              ),
                            ],
                          ),
                        ),
                        Row(
                          children: [
                            if (user != null) ...[
                              _buildGoldPill(context, gold),
                              const SizedBox(width: 8),
                            ],
                            IconButton(
                              onPressed: widget.onClose,
                              icon: const Icon(Icons.close),
                            ),
                          ],
                        ),
                      ],
                    ),
                    if (characterImageUrl != null &&
                        characterImageUrl.isNotEmpty) ...[
                      const SizedBox(height: 16),
                      _buildShopHero(context, characterImageUrl),
                    ],
                    const SizedBox(height: 16),
                    _buildShopTabs(context),
                    if (_error != null)
                      _buildStatusBanner(
                        context,
                        message: _error!,
                        isError: true,
                      ),
                    if (_success != null)
                      _buildStatusBanner(
                        context,
                        message: _success!,
                        isError: false,
                      ),
                    const SizedBox(height: 8),
                    _buildFilterRow(context),
                    const SizedBox(height: 12),
                    Expanded(
                      child: _activeTab == 'buy'
                          ? _buildBuyList(context, filteredShopItems, user)
                          : _buildSellList(context, filteredSellableItems),
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

  Widget _buildGoldPill(BuildContext context, int gold) {
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
          Icon(
            Icons.monetization_on,
            size: 18,
            color: Colors.amber.shade700,
          ),
          const SizedBox(width: 6),
          Text(
            'GOLD',
            style: theme.textTheme.labelSmall?.copyWith(
              color: Colors.amber.shade800,
              fontWeight: FontWeight.w700,
              letterSpacing: 0.6,
            ),
          ),
          const SizedBox(width: 6),
          Text(
            '$gold',
            style: theme.textTheme.titleSmall?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildShopHero(BuildContext context, String imageUrl) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: theme.dividerColor),
      ),
      child: Row(
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(12),
            child: Image.network(
              imageUrl,
              width: 96,
              height: 96,
              fit: BoxFit.cover,
              errorBuilder: (_, __, ___) => Container(
                width: 96,
                height: 96,
                color: theme.colorScheme.surfaceVariant,
                child: const Icon(Icons.image),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  widget.character.name,
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  'Hand-picked goods and honest bargains. Take a look.',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildShopTabs(BuildContext context) {
    final theme = Theme.of(context);
    final activeColor = theme.colorScheme.primary;
    final inactiveColor = theme.colorScheme.onSurfaceVariant;
    return Container(
      padding: const EdgeInsets.all(4),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: theme.dividerColor),
      ),
      child: Row(
        children: [
          Expanded(
            child: _buildTabButton(
              context,
              label: 'Buy',
              icon: Icons.shopping_bag_outlined,
              isActive: _activeTab == 'buy',
              activeColor: activeColor,
              inactiveColor: inactiveColor,
              onTap: () => setState(() => _activeTab = 'buy'),
            ),
          ),
          Expanded(
            child: _buildTabButton(
              context,
              label: 'Sell',
              icon: Icons.sell_outlined,
              isActive: _activeTab == 'sell',
              activeColor: activeColor,
              inactiveColor: inactiveColor,
              onTap: () => setState(() => _activeTab = 'sell'),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTabButton(
    BuildContext context, {
    required String label,
    required IconData icon,
    required bool isActive,
    required Color activeColor,
    required Color inactiveColor,
    required VoidCallback onTap,
  }) {
    final theme = Theme.of(context);
    return AnimatedContainer(
      duration: const Duration(milliseconds: 200),
      curve: Curves.easeOut,
      decoration: BoxDecoration(
        color: isActive ? activeColor.withOpacity(0.15) : Colors.transparent,
        borderRadius: BorderRadius.circular(999),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(999),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 12),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                icon,
                size: 18,
                color: isActive ? activeColor : inactiveColor,
              ),
              const SizedBox(width: 6),
              Text(
                label,
                style: theme.textTheme.labelLarge?.copyWith(
                  fontWeight: FontWeight.w700,
                  color: isActive ? activeColor : inactiveColor,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFilterRow(BuildContext context) {
    final theme = Theme.of(context);
    return Row(
      children: [
        Expanded(
          child: Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              FilterChip(
                label: const Text('Usable'),
                selected: _filterUsable,
                onSelected: (value) => setState(() => _filterUsable = value),
              ),
              FilterChip(
                label: const Text('Sellable'),
                selected: _filterSellable,
                onSelected: (value) => setState(() => _filterSellable = value),
              ),
              FilterChip(
                label: const Text('Outfit'),
                selected: _filterOutfit,
                onSelected: (value) => setState(() => _filterOutfit = value),
              ),
            ],
          ),
        ),
        const SizedBox(width: 12),
        _buildSortMenu(context),
      ],
    );
  }

  Widget _buildSortMenu(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: theme.dividerColor),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<String>(
          value: _sortMode,
          icon: const Icon(Icons.sort),
          style: theme.textTheme.labelLarge,
          onChanged: (value) {
            if (value == null) return;
            setState(() => _sortMode = value);
          },
          items: const [
            DropdownMenuItem(
              value: 'name',
              child: Text('Name'),
            ),
            DropdownMenuItem(
              value: 'price',
              child: Text('Price'),
            ),
            DropdownMenuItem(
              value: 'owned',
              child: Text('Owned'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatusBanner(
    BuildContext context, {
    required String message,
    required bool isError,
  }) {
    final theme = Theme.of(context);
    final color = isError ? theme.colorScheme.error : Colors.green.shade700;
    final background = isError
        ? theme.colorScheme.errorContainer
        : Colors.green.shade100;
    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 200),
      transitionBuilder: (child, animation) => FadeTransition(
        opacity: animation,
        child: SizeTransition(
          sizeFactor: animation,
          axisAlignment: -1,
          child: child,
        ),
      ),
      child: Container(
        key: ValueKey(message),
        margin: const EdgeInsets.only(top: 8),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          color: background,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: color.withOpacity(0.35)),
        ),
        child: Row(
          children: [
            Icon(
              isError ? Icons.error_outline : Icons.check_circle_outline,
              size: 18,
              color: color,
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                message,
                style: theme.textTheme.bodyMedium?.copyWith(color: color),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBuyList(
    BuildContext context,
    List<ShopInventoryItem> shopItems,
    User? user,
  ) {
    if (shopItems.isEmpty) {
      return _buildEmptyState(
        context,
        title: 'No wares available',
        message: 'Check back later for new stock.',
        icon: Icons.store_outlined,
      );
    }
    final sorted = List<ShopInventoryItem>.from(shopItems)
      ..sort((a, b) {
        final itemA = _getItemById(a.itemId);
        final itemB = _getItemById(b.itemId);
        switch (_sortMode) {
          case 'price':
            return b.price.compareTo(a.price);
          case 'owned':
            final ownedA = itemA == null ? 0 : _ownedQuantityForItemId(itemA.id);
            final ownedB = itemB == null ? 0 : _ownedQuantityForItemId(itemB.id);
            return ownedB.compareTo(ownedA);
          case 'name':
          default:
            return (itemA?.name ?? '').compareTo(itemB?.name ?? '');
        }
      });
    return ListView.separated(
      padding: const EdgeInsets.only(bottom: 12),
      itemCount: sorted.length,
      separatorBuilder: (_, __) => const SizedBox(height: 12),
      itemBuilder: (_, i) {
        final shopItem = sorted[i];
        final item = _getItemById(shopItem.itemId);
        final canAfford = user != null && user.gold >= shopItem.price;
        return _buildShopItemCard(context, shopItem, item, canAfford);
      },
    );
  }

  Widget _buildSellList(
    BuildContext context,
    List<OwnedInventoryItem> sellableItems,
  ) {
    if (sellableItems.isEmpty) {
      return _buildEmptyState(
        context,
        title: 'Nothing to sell',
        message: 'Items you can sell will appear here.',
        icon: Icons.inventory_2_outlined,
      );
    }
    final sorted = List<OwnedInventoryItem>.from(sellableItems)
      ..sort((a, b) {
        final itemA = _getItemById(a.inventoryItemId);
        final itemB = _getItemById(b.inventoryItemId);
        switch (_sortMode) {
          case 'price':
            return (itemB?.sellValue ?? 0).compareTo(itemA?.sellValue ?? 0);
          case 'owned':
            return b.quantity.compareTo(a.quantity);
          case 'name':
          default:
            return (itemA?.name ?? '').compareTo(itemB?.name ?? '');
        }
      });
    return ListView.separated(
      padding: const EdgeInsets.only(bottom: 12),
      itemCount: sorted.length,
      separatorBuilder: (_, __) => const SizedBox(height: 12),
      itemBuilder: (_, i) {
        final owned = sorted[i];
        final item = _getItemById(owned.inventoryItemId);
        return _buildSellItemCard(context, owned, item);
      },
    );
  }

  Widget _buildEmptyState(
    BuildContext context, {
    required String title,
    required String message,
    required IconData icon,
  }) {
    final theme = Theme.of(context);
    return Center(
      child: Container(
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: theme.colorScheme.surfaceContainerHighest,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: theme.dividerColor),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 36, color: theme.colorScheme.onSurfaceVariant),
            const SizedBox(height: 12),
            Text(
              title,
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              message,
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildShopItemCard(
    BuildContext context,
    ShopInventoryItem shopItem,
    InventoryItem? item,
    bool canAfford,
  ) {
    final theme = Theme.of(context);
    final img = item?.imageUrl ?? '';
    final name = item?.name ?? 'Unknown';
    final flavor = item?.flavorText ?? '';
    final isPurchasing = _purchasingItemId == shopItem.itemId;
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: isPurchasing
              ? theme.colorScheme.primary.withOpacity(0.6)
              : theme.dividerColor,
        ),
        boxShadow: [
          if (isPurchasing)
            BoxShadow(
              color: theme.colorScheme.primary.withOpacity(0.15),
              blurRadius: 12,
              offset: const Offset(0, 6),
            ),
        ],
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildItemImage(context, img),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                if (flavor.isNotEmpty) ...[
                  const SizedBox(height: 4),
                  Text(
                    flavor,
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              _buildPricePill(context, 'Price', shopItem.price),
              const SizedBox(height: 8),
              AnimatedSwitcher(
                duration: const Duration(milliseconds: 200),
                child: isPurchasing
                    ? const SizedBox(
                        width: 110,
                        height: 40,
                        child: Center(
                          child: SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                        ),
                      )
                    : SizedBox(
                        width: 110,
                        height: 40,
                        child: FilledButton.icon(
                          onPressed: canAfford
                              ? () => _purchase(shopItem.itemId, shopItem.price)
                              : null,
                          icon: const Icon(Icons.shopping_bag_outlined, size: 18),
                          label: const Text('Buy'),
                        ),
                      ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildSellItemCard(
    BuildContext context,
    OwnedInventoryItem owned,
    InventoryItem? item,
  ) {
    final theme = Theme.of(context);
    final sellValue = item?.sellValue ?? 0;
    final img = item?.imageUrl ?? '';
    final name = item?.name ?? 'Unknown';
    final isSellingItem = _sellingItemId == owned.inventoryItemId;
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: isSellingItem
              ? theme.colorScheme.primary.withOpacity(0.6)
              : theme.dividerColor,
        ),
        boxShadow: [
          if (isSellingItem)
            BoxShadow(
              color: theme.colorScheme.primary.withOpacity(0.15),
              blurRadius: 12,
              offset: const Offset(0, 6),
            ),
        ],
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildItemImage(context, img),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  style: theme.textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Owned: ${owned.quantity}',
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              _buildPricePill(context, 'Value', sellValue),
              const SizedBox(height: 8),
              if (owned.quantity > 1)
                Wrap(
                  spacing: 6,
                  runSpacing: 6,
                  children: [
                    AnimatedSwitcher(
                      duration: const Duration(milliseconds: 200),
                      child: isSellingItem
                          ? const SizedBox(
                              width: 90,
                              height: 36,
                              child: Center(
                                child: SizedBox(
                                  width: 16,
                                  height: 16,
                                  child: CircularProgressIndicator(strokeWidth: 2),
                                ),
                              ),
                            )
                          : OutlinedButton(
                              onPressed: sellValue > 0
                                  ? () => _sell(
                                        owned.inventoryItemId,
                                        quantity: 1,
                                      )
                                  : null,
                              child: const Text('Sell 1'),
                            ),
                    ),
                    AnimatedSwitcher(
                      duration: const Duration(milliseconds: 200),
                      child: isSellingItem
                          ? const SizedBox(
                              width: 130,
                              height: 36,
                              child: Center(
                                child: SizedBox(
                                  width: 16,
                                  height: 16,
                                  child: CircularProgressIndicator(strokeWidth: 2),
                                ),
                              ),
                            )
                          : FilledButton(
                              onPressed: sellValue > 0
                                  ? () => _sell(
                                        owned.inventoryItemId,
                                        quantity: owned.quantity,
                                      )
                                  : null,
                              child: Text('Sell All (${owned.quantity})'),
                            ),
                    ),
                  ],
                )
              else
                AnimatedSwitcher(
                  duration: const Duration(milliseconds: 200),
                  child: isSellingItem
                      ? const SizedBox(
                          width: 100,
                          height: 36,
                          child: Center(
                            child: SizedBox(
                              width: 16,
                              height: 16,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            ),
                          ),
                        )
                      : FilledButton(
                          onPressed: sellValue > 0
                              ? () => _sell(
                                    owned.inventoryItemId,
                                    quantity: 1,
                                  )
                              : null,
                          child: const Text('Sell'),
                        ),
                ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildItemImage(BuildContext context, String imageUrl) {
    final theme = Theme.of(context);
    if (imageUrl.isEmpty) {
      return Container(
        width: 64,
        height: 64,
        decoration: BoxDecoration(
          color: theme.colorScheme.surfaceVariant,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: theme.dividerColor),
        ),
        child: Icon(
          Icons.inventory_2_outlined,
          color: theme.colorScheme.onSurfaceVariant,
        ),
      );
    }
    return ClipRRect(
      borderRadius: BorderRadius.circular(12),
      child: Image.network(
        imageUrl,
        width: 64,
        height: 64,
        fit: BoxFit.cover,
        errorBuilder: (_, __, ___) => Container(
          width: 64,
          height: 64,
          color: theme.colorScheme.surfaceVariant,
          child: const Icon(Icons.image),
        ),
      ),
    );
  }

  Widget _buildPricePill(BuildContext context, String label, int value) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: theme.dividerColor),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            Icons.monetization_on,
            size: 16,
            color: Colors.amber.shade700,
          ),
          const SizedBox(width: 6),
          Text(
            '$label $value',
            style: theme.textTheme.labelLarge?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}
