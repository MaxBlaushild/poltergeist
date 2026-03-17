import 'package:flutter/material.dart';

import '../models/base.dart';
import 'paper_texture.dart';

class BasePanel extends StatelessWidget {
  const BasePanel({super.key, required this.base, required this.onClose});

  final BasePin base;
  final VoidCallback onClose;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return DraggableScrollableSheet(
      initialChildSize: 0.48,
      minChildSize: 0.32,
      maxChildSize: 0.7,
      builder: (_, scrollController) => PaperSheet(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Base',
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  IconButton(onPressed: onClose, icon: const Icon(Icons.close)),
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
                    ClipRRect(
                      borderRadius: BorderRadius.circular(14),
                      child: AspectRatio(
                        aspectRatio: 1,
                        child: Image.network(
                          base.thumbnailUrl,
                          fit: BoxFit.cover,
                          errorBuilder: (_, _, _) => Container(
                            color: theme.colorScheme.surfaceVariant,
                            child: const Icon(
                              Icons.home_work_outlined,
                              size: 48,
                            ),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 16),
                    Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.surfaceVariant.withValues(
                          alpha: 0.42,
                        ),
                        borderRadius: BorderRadius.circular(14),
                        border: Border.all(
                          color: theme.colorScheme.outlineVariant,
                        ),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            '${base.owner.displayName}\'s Base',
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                          if (base.owner.secondaryName.isNotEmpty) ...[
                            const SizedBox(height: 4),
                            Text(
                              base.owner.secondaryName,
                              style: theme.textTheme.bodyMedium,
                            ),
                          ],
                          const SizedBox(height: 10),
                          Text(
                            'A marked home base shared on the map with trusted allies.',
                            style: theme.textTheme.bodyMedium?.copyWith(
                              height: 1.35,
                            ),
                          ),
                        ],
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
