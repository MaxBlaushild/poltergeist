import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../providers/inventory_modal_provider.dart';

class NewItemModal extends StatelessWidget {
  const NewItemModal({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<InventoryModalProvider>(
      builder: (context, provider, _) {
        final item = provider.presentedItem;
        if (item == null) return const SizedBox.shrink();

        return Dialog(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: Text(
                        'You got a ${item['name'] ?? 'item'}!',
                        style: Theme.of(context).textTheme.titleLarge,
                      ),
                    ),
                    IconButton(
                      onPressed: () => provider.setPresentedItem(null),
                      icon: const Icon(Icons.close),
                    ),
                  ],
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
                if (item['flavorText'] != null) ...[
                  const SizedBox(height: 16),
                  Text(
                    item['flavorText'] as String,
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                  ),
                ],
                if (item['effectText'] != null) ...[
                  const SizedBox(height: 8),
                  Text(item['effectText'] as String),
                ],
                const SizedBox(height: 16),
                FilledButton(
                  onPressed: () => provider.setPresentedItem(null),
                  child: const Text('OK'),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}
