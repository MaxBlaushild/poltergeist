import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../providers/inventory_modal_provider.dart';

class UsedItemModal extends StatefulWidget {
  const UsedItemModal({super.key});

  @override
  State<UsedItemModal> createState() => _UsedItemModalState();
}

class _UsedItemModalState extends State<UsedItemModal> {
  String? _lastItemId;

  @override
  Widget build(BuildContext context) {
    return Consumer<InventoryModalProvider>(
      builder: (context, provider, _) {
        final item = provider.usedItem;
        if (item == null) {
          _lastItemId = null;
          return const SizedBox.shrink();
        }

        final itemId = item['id']?.toString() ?? item['name']?.toString();
        if (itemId != _lastItemId) {
          _lastItemId = itemId;
          Future.delayed(const Duration(seconds: 3), () {
            if (mounted && provider.usedItem != null) {
              provider.setUsedItem(null);
            }
          });
        }

        return Dialog(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  'You used a ${item['name'] ?? 'item'}!',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 16),
                if (item['imageUrl'] != null)
                  Image.network(
                    item['imageUrl'] as String,
                    width: 200,
                    height: 200,
                    fit: BoxFit.cover,
                    errorBuilder: (_, __, ___) => const Icon(Icons.image, size: 200),
                  ),
              ],
            ),
          ),
        );
      },
    );
  }
}
