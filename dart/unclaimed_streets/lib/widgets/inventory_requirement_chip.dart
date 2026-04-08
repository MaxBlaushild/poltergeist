import 'package:flutter/material.dart';

import '../models/inventory_item.dart';

class InventoryRequirementChip extends StatelessWidget {
  const InventoryRequirementChip({
    super.key,
    required this.item,
    required this.quantity,
    required this.ownedQuantity,
  });

  final InventoryItem item;
  final int quantity;
  final int ownedQuantity;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final hasEnough = ownedQuantity >= quantity;
    final color = hasEnough
        ? theme.colorScheme.primary
        : theme.colorScheme.error;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: color.withValues(alpha: 0.7)),
      ),
      child: Text(
        '${item.name}: $ownedQuantity/$quantity',
        style: theme.textTheme.bodySmall?.copyWith(
          color: color,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}
