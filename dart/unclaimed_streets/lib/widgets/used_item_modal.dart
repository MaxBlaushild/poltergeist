import 'dart:async';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../providers/inventory_modal_provider.dart';
import 'cached_inventory_image.dart';

class UsedItemModal extends StatefulWidget {
  const UsedItemModal({super.key});

  @override
  State<UsedItemModal> createState() => _UsedItemModalState();
}

class _UsedItemModalState extends State<UsedItemModal> {
  Timer? _dismissTimer;
  String? _lastEventId;

  @override
  void dispose() {
    _dismissTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<InventoryModalProvider>(
      builder: (context, provider, _) {
        final rawItem = provider.usedItem;
        if (rawItem == null) {
          _lastEventId = null;
          _dismissTimer?.cancel();
          _dismissTimer = null;
          return const SizedBox.shrink();
        }

        final item = Map<String, dynamic>.from(rawItem);
        final eventId = item['eventId']?.toString().trim().isNotEmpty == true
            ? item['eventId']?.toString()
            : item['usedAt']?.toString();
        if (eventId != null && eventId != _lastEventId) {
          _lastEventId = eventId;
          _dismissTimer?.cancel();
          _dismissTimer = Timer(const Duration(seconds: 7), () {
            if (!mounted || provider.usedItem == null) {
              return;
            }
            provider.setUsedItem(null);
          });
        }

        final theme = Theme.of(context);
        final imageUrl = item['imageUrl']?.toString().trim() ?? '';
        final message = item['message']?.toString().trim() ?? '';
        final flavorText = item['flavorText']?.toString().trim() ?? '';
        final effectText = item['effectText']?.toString().trim() ?? '';
        final remainingQuantity =
            (item['remainingQuantity'] as num?)?.toInt() ?? 0;
        final consumedQuantity =
            (item['consumedQuantity'] as num?)?.toInt() ?? 1;
        final depleted = item['depleted'] == true || remainingQuantity <= 0;
        final effectSummary = _summaryLines(item);
        final learnedRecipes = _learnedRecipeNames(item);

        return Positioned.fill(
          child: Material(
            color: Colors.black45,
            child: InkWell(
              onTap: () => provider.setUsedItem(null),
              child: Center(
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 460),
                  child: Padding(
                    padding: const EdgeInsets.all(24),
                    child: InkWell(
                      onTap: () {},
                      borderRadius: BorderRadius.circular(24),
                      child: DecoratedBox(
                        decoration: BoxDecoration(
                          color: theme.colorScheme.surface,
                          borderRadius: BorderRadius.circular(24),
                          border: Border.all(
                            color: theme.colorScheme.outlineVariant,
                          ),
                          boxShadow: const [
                            BoxShadow(
                              color: Colors.black26,
                              blurRadius: 28,
                              offset: Offset(0, 18),
                            ),
                          ],
                        ),
                        child: Padding(
                          padding: const EdgeInsets.all(20),
                          child: SingleChildScrollView(
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          Text(
                                            'Item Used',
                                            style: theme.textTheme.labelLarge
                                                ?.copyWith(
                                                  color: theme
                                                      .colorScheme
                                                      .onSurfaceVariant,
                                                  letterSpacing: 0.4,
                                                ),
                                          ),
                                          const SizedBox(height: 4),
                                          Text(
                                            item['name']?.toString() ?? 'Item',
                                            style: theme.textTheme.headlineSmall
                                                ?.copyWith(
                                                  fontWeight: FontWeight.w700,
                                                ),
                                          ),
                                          if (message.isNotEmpty) ...[
                                            const SizedBox(height: 8),
                                            Text(
                                              _presentableMessage(message),
                                              style: theme.textTheme.bodyMedium
                                                  ?.copyWith(
                                                    color: theme
                                                        .colorScheme
                                                        .onSurfaceVariant,
                                                  ),
                                            ),
                                          ],
                                        ],
                                      ),
                                    ),
                                    IconButton(
                                      tooltip: 'Dismiss',
                                      onPressed: () =>
                                          provider.setUsedItem(null),
                                      icon: const Icon(Icons.close),
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 16),
                                Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Container(
                                      width: 120,
                                      height: 120,
                                      decoration: BoxDecoration(
                                        color: theme
                                            .colorScheme
                                            .surfaceContainerHighest,
                                        borderRadius: BorderRadius.circular(18),
                                        border: Border.all(
                                          color: theme.dividerColor,
                                        ),
                                      ),
                                      padding: const EdgeInsets.all(12),
                                      child: imageUrl.isEmpty
                                          ? Icon(
                                              Icons.inventory_2_outlined,
                                              size: 48,
                                              color: theme
                                                  .colorScheme
                                                  .onSurfaceVariant,
                                            )
                                          : CachedInventoryImage(
                                              imageUrl: imageUrl,
                                              fit: BoxFit.contain,
                                              cacheWidth: 512,
                                              errorBuilder: (context) => Icon(
                                                Icons.inventory_2_outlined,
                                                size: 48,
                                                color: theme
                                                    .colorScheme
                                                    .onSurfaceVariant,
                                              ),
                                            ),
                                    ),
                                    const SizedBox(width: 16),
                                    Expanded(
                                      child: Wrap(
                                        spacing: 8,
                                        runSpacing: 8,
                                        children: [
                                          _ReceiptPill(
                                            icon: Icons.check_circle_outline,
                                            label: 'Consumed $consumedQuantity',
                                          ),
                                          _ReceiptPill(
                                            icon: depleted
                                                ? Icons.remove_shopping_cart
                                                : Icons.inventory_2_outlined,
                                            label: depleted
                                                ? 'Stack depleted'
                                                : '$remainingQuantity remaining',
                                          ),
                                          if (learnedRecipes.isNotEmpty)
                                            _ReceiptPill(
                                              icon: Icons.menu_book_outlined,
                                              label:
                                                  '${learnedRecipes.length} recipe${learnedRecipes.length == 1 ? '' : 's'} learned',
                                            ),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                                if (effectSummary.isNotEmpty) ...[
                                  const SizedBox(height: 20),
                                  Text(
                                    'What happened',
                                    style: theme.textTheme.titleSmall?.copyWith(
                                      fontWeight: FontWeight.w700,
                                    ),
                                  ),
                                  const SizedBox(height: 10),
                                  ...effectSummary.map(
                                    (line) => Padding(
                                      padding: const EdgeInsets.only(bottom: 8),
                                      child: Row(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          Icon(
                                            Icons.auto_awesome,
                                            size: 18,
                                            color: theme.colorScheme.primary,
                                          ),
                                          const SizedBox(width: 8),
                                          Expanded(
                                            child: Text(
                                              line,
                                              style: theme.textTheme.bodyMedium,
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                  ),
                                ],
                                if (learnedRecipes.isNotEmpty) ...[
                                  const SizedBox(height: 12),
                                  Text(
                                    'Learned Recipes',
                                    style: theme.textTheme.titleSmall?.copyWith(
                                      fontWeight: FontWeight.w700,
                                    ),
                                  ),
                                  const SizedBox(height: 8),
                                  Wrap(
                                    spacing: 8,
                                    runSpacing: 8,
                                    children: learnedRecipes
                                        .map(
                                          (name) => _ReceiptPill(
                                            icon: Icons.menu_book_outlined,
                                            label: name,
                                          ),
                                        )
                                        .toList(growable: false),
                                  ),
                                ],
                                if (effectText.isNotEmpty) ...[
                                  const SizedBox(height: 16),
                                  Text(
                                    effectText,
                                    style: theme.textTheme.bodyMedium?.copyWith(
                                      fontWeight: FontWeight.w600,
                                    ),
                                  ),
                                ],
                                if (flavorText.isNotEmpty) ...[
                                  const SizedBox(height: 10),
                                  Text(
                                    flavorText,
                                    style: theme.textTheme.bodySmall?.copyWith(
                                      color: theme.colorScheme.onSurfaceVariant,
                                      fontStyle: FontStyle.italic,
                                      height: 1.35,
                                    ),
                                  ),
                                ],
                                const SizedBox(height: 20),
                                Align(
                                  alignment: Alignment.centerRight,
                                  child: FilledButton(
                                    onPressed: () => provider.setUsedItem(null),
                                    child: const Text('Continue'),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
        );
      },
    );
  }

  List<String> _summaryLines(Map<String, dynamic> item) {
    final raw = item['effectSummary'];
    if (raw is! List) {
      return const <String>[];
    }
    return raw
        .map((entry) => entry.toString().trim())
        .where((entry) => entry.isNotEmpty)
        .toList(growable: false);
  }

  List<String> _learnedRecipeNames(Map<String, dynamic> item) {
    final raw = item['learnedRecipes'];
    if (raw is! List) {
      return const <String>[];
    }
    return raw
        .whereType<Map>()
        .map((entry) => entry['itemName']?.toString().trim() ?? '')
        .where((entry) => entry.isNotEmpty)
        .toList(growable: false);
  }

  String _presentableMessage(String message) {
    if (message.isEmpty) {
      return message;
    }
    final normalized = message.trim();
    return normalized[0].toUpperCase() + normalized.substring(1);
  }
}

class _ReceiptPill extends StatelessWidget {
  const _ReceiptPill({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
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
          ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 220),
            child: Text(
              label,
              overflow: TextOverflow.ellipsis,
              style: theme.textTheme.labelMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
