import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../providers/quest_filter_provider.dart';
import '../providers/tags_provider.dart';

class QuestFilterPanel extends StatefulWidget {
  const QuestFilterPanel({super.key});

  @override
  State<QuestFilterPanel> createState() => _QuestFilterPanelState();
}

class _QuestFilterPanelState extends State<QuestFilterPanel> {
  final TextEditingController _searchController = TextEditingController();
  String _search = '';

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Consumer2<QuestFilterProvider, TagsProvider>(
      builder: (context, filters, tags, _) {
        final allTags = tags.tags;
        final search = _search.trim().toLowerCase();
        final filteredTags = search.isEmpty
            ? allTags
            : allTags.where((t) => t.name.toLowerCase().contains(search)).toList();
        final tagsEnabled = filters.enableTagFilter;
        final hasSelectedTags = tags.selectedTagIds.isNotEmpty;

        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              'Show',
              style: Theme.of(context).textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w600,
                  ),
            ),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                FilterChip(
                  label: const Text('Current Quest Points'),
                  selected: filters.showCurrentQuestPoints,
                  onSelected: (_) => filters.toggleCurrentQuestPoints(),
                ),
                FilterChip(
                  label: const Text('Quest Available'),
                  selected: filters.showQuestAvailablePoints,
                  onSelected: (_) => filters.toggleQuestAvailablePoints(),
                ),
                FilterChip(
                  label: const Text('Treasure Chests'),
                  selected: filters.showTreasureChests,
                  onSelected: (_) => filters.toggleTreasureChests(),
                ),
                FilterChip(
                  label: const Text('Tags'),
                  selected: filters.enableTagFilter,
                  onSelected: (_) => filters.toggleTagFilter(),
                ),
                if (filters.showCurrentQuestPoints ||
                    filters.showQuestAvailablePoints ||
                    filters.enableTagFilter ||
                    tags.selectedTagIds.isNotEmpty)
                  ActionChip(
                    label: const Text('Reset'),
                    onPressed: () {
                      filters.clearAll();
                      tags.clearFilters();
                    },
                  ),
              ],
            ),
            const SizedBox(height: 16),
            Text(
              'Tags',
              style: Theme.of(context).textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w600,
                  ),
            ),
            const SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                TextButton(
                  onPressed: tagsEnabled ? tags.selectAll : null,
                  child: const Text('Select all'),
                ),
                const SizedBox(width: 8),
                TextButton(
                  onPressed: tagsEnabled ? tags.deselectAll : null,
                  child: const Text('Deselect all'),
                ),
              ],
            ),
            TextField(
              controller: _searchController,
              decoration: InputDecoration(
                prefixIcon: const Icon(Icons.search),
                hintText: 'Search tags',
                suffixIcon: _search.isEmpty
                    ? null
                    : IconButton(
                        icon: const Icon(Icons.close),
                        onPressed: () => setState(() {
                          _search = '';
                          _searchController.clear();
                        }),
                      ),
                border: const OutlineInputBorder(),
              ),
              onChanged: (value) => setState(() => _search = value),
            ),
            const SizedBox(height: 12),
            if (tags.loading && tags.tags.isEmpty)
              const Center(
                child: SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              )
            else if (filteredTags.isEmpty)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 8),
                child: Text(
                  search.isEmpty ? 'No tags' : 'No matching tags',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              )
            else
              Opacity(
                opacity: tagsEnabled ? 1.0 : 0.4,
                child: IgnorePointer(
                  ignoring: !tagsEnabled,
                  child: Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      ...filteredTags.map((t) {
                        final selected = tags.selectedTagIds.contains(t.id);
                        return FilterChip(
                          label: Text(t.name),
                          selected: selected,
                          onSelected: (_) => tags.toggleTag(t.id),
                        );
                      }),
                      if (hasSelectedTags)
                        ActionChip(
                          label: const Text('Clear Tags'),
                          onPressed: tags.clearFilters,
                        ),
                    ],
                  ),
                ),
              ),
            if (!tagsEnabled)
              Padding(
                padding: const EdgeInsets.only(top: 8),
                child: Text(
                  'Enable the Tags filter to apply tag selection.',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Theme.of(context).colorScheme.onSurface.withValues(alpha: 0.6),
                      ),
                ),
              ),
          ],
        );
      },
    );
  }
}
