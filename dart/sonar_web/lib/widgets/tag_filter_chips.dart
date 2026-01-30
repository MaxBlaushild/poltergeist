import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../providers/tags_provider.dart';

class TagFilterChips extends StatelessWidget {
  const TagFilterChips({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<TagsProvider>(
      builder: (context, tags, _) {
        if (tags.loading && tags.tags.isEmpty) {
          return const Center(child: SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)));
        }
        if (tags.tags.isEmpty) {
          return const Padding(
            padding: EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Text('No tags', style: TextStyle(fontSize: 12)),
          );
        }
        return Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            ...tags.tags.map((t) {
              final selected = tags.selectedTagIds.contains(t.id);
              return FilterChip(
                label: Text(t.name),
                selected: selected,
                onSelected: (_) => tags.toggleTag(t.id),
              );
            }),
            if (tags.selectedTagIds.isNotEmpty)
              ActionChip(
                label: const Text('Clear'),
                onPressed: tags.clearFilters,
              ),
          ],
        );
      },
    );
  }
}
