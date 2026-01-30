import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../providers/discoveries_provider.dart';
import '../providers/quest_log_provider.dart';
import '../providers/tags_provider.dart';
import 'tag_filter_chips.dart';

const _placeholderImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';

/// Bottom-sheet content for Quest Log.
/// [onFocusPoI] when user taps a POI in a quest: close sheet, fly to POI, open POI panel.
typedef OnFocusPoI = void Function(PointOfInterest poi);

class QuestLogPanel extends StatefulWidget {
  const QuestLogPanel({
    super.key,
    required this.onClose,
    required this.onFocusPoI,
  });

  final VoidCallback onClose;
  final OnFocusPoI onFocusPoI;

  @override
  State<QuestLogPanel> createState() => _QuestLogPanelState();
}

class _QuestLogPanelState extends State<QuestLogPanel> {
  Quest? _selectedQuest;
  final Map<String, bool> _expanded = {};

  void _focusPoI(PointOfInterest poi) {
    widget.onClose();
    widget.onFocusPoI(poi);
  }

  @override
  Widget build(BuildContext context) {
    if (_selectedQuest != null) {
      return _buildQuestDetail(context, _selectedQuest!);
    }
    return _buildQuestList(context);
  }

  Widget _buildQuestList(BuildContext context) {
    return Consumer3<QuestLogProvider, TagsProvider, DiscoveriesProvider>(
      builder: (context, ql, tags, discoveries, _) {
        if (ql.loading && ql.quests.isEmpty) {
          return const Center(
            child: Padding(
              padding: EdgeInsets.all(24),
              child: CircularProgressIndicator(),
            ),
          );
        }

        final tracked = ql.quests
            .where((q) => ql.trackedQuestIds.contains(q.id))
            .toList();
        final tagBuckets = <String, List<Quest>>{};
        for (final g in tags.tagGroups) {
          tagBuckets[g.id] = [];
        }
        final untagged = <Quest>[];

        for (final q in ql.quests) {
          if (ql.trackedQuestIds.contains(q.id)) continue;
          final tagNames = getQuestTags(q);
          var added = false;
          for (final g in tags.tagGroups) {
            final hasMatch = tagNames.any((tName) => tags.tags
                .any((x) => x.tagGroupId == g.id && x.name == tName));
            if (hasMatch) {
              tagBuckets[g.id]!.add(q);
              added = true;
            }
          }
          if (!added) untagged.add(q);
        }

        return DefaultTabController(
          length: 2,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TabBar(
                tabs: const [
                  Tab(text: 'Quests'),
                  Tab(text: 'Filters'),
                ],
              ),
              Expanded(
                child: TabBarView(
                  children: [
                    SingleChildScrollView(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          if (tracked.isNotEmpty)
                            _QuestAccordion(
                              title: 'Tracked Quests',
                              quests: tracked,
                              expanded: _expanded['tracked'] ?? false,
                              onToggle: () {
                                setState(() {
                                  _expanded['tracked'] =
                                      !(_expanded['tracked'] ?? false);
                                });
                              },
                              onQuestTap: (q) =>
                                  setState(() => _selectedQuest = q),
                            ),
                          ...tags.tagGroups.map((g) {
                            final list = tagBuckets[g.id] ?? [];
                            if (list.isEmpty) return const SizedBox.shrink();
                            return _QuestAccordion(
                              key: ValueKey(g.id),
                              title: g.name,
                              quests: list,
                              expanded: _expanded[g.id] ?? false,
                              onToggle: () {
                                setState(() {
                                  _expanded[g.id] =
                                      !(_expanded[g.id] ?? false);
                                });
                              },
                              onQuestTap: (q) =>
                                  setState(() => _selectedQuest = q),
                            );
                          }),
                          if (untagged.isNotEmpty)
                            _QuestAccordion(
                              title: 'The Rest',
                              quests: untagged,
                              expanded: _expanded['untagged'] ?? false,
                              onToggle: () {
                                setState(() {
                                  _expanded['untagged'] =
                                      !(_expanded['untagged'] ?? false);
                                });
                              },
                              onQuestTap: (q) =>
                                  setState(() => _selectedQuest = q),
                            ),
                        ],
                      ),
                    ),
                    SingleChildScrollView(
                      padding: const EdgeInsets.all(16),
                      child: const TagFilterChips(),
                    ),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildQuestDetail(BuildContext context, Quest quest) {
    return Consumer2<QuestLogProvider, DiscoveriesProvider>(
      builder: (context, ql, discoveries, _) {
        final isTracked = ql.trackedQuestIds.contains(quest.id);
        final discoveredIds = <String>{};
        for (final d in discoveries.discoveries) {
          discoveredIds.add(d.pointOfInterestId);
        }

        var completed = 0;
        var total = 0;
        void count(QuestNode node) {
          total++;
          if (node.objectives.every((o) => o.isCompleted)) completed++;
          for (final o in node.objectives) {
            if (o.nextNode != null) count(o.nextNode!);
          }
        }
        count(quest.rootNode);

        return SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextButton.icon(
                onPressed: () => setState(() => _selectedQuest = null),
                icon: const Icon(Icons.arrow_back),
                label: const Text('Back to Quests'),
              ),
              const SizedBox(height: 8),
              Text(
                quest.name,
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
              ),
              const SizedBox(height: 12),
              ClipRRect(
                borderRadius: BorderRadius.circular(12),
                child: Image.network(
                  quest.imageUrl.isNotEmpty
                      ? quest.imageUrl
                      : _placeholderImageUrl,
                  height: 180,
                  width: double.infinity,
                  fit: BoxFit.cover,
                  errorBuilder: (_, __, ___) => Container(
                    height: 180,
                    color: Colors.grey.shade300,
                    child: const Icon(Icons.help_outline, size: 48),
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Tasks completed: $completed/$total',
                    style: Theme.of(context).textTheme.titleSmall,
                  ),
                  FilledButton(
                    onPressed: () async {
                      if (isTracked) {
                        await ql.untrackQuest(quest.id);
                      } else {
                        await ql.trackQuest(quest.id);
                      }
                    },
                    child: Text(isTracked ? 'Untrack Quest' : 'Track Quest'),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              DefaultTabController(
                length: 2,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    const TabBar(
                      tabs: [
                        Tab(text: 'Description'),
                        Tab(text: 'Tasks'),
                      ],
                    ),
                    SizedBox(
                      height: 320,
                      child: TabBarView(
                        children: [
                          SingleChildScrollView(
                            padding: const EdgeInsets.only(top: 12),
                            child: Text(
                              quest.description,
                              style: Theme.of(context).textTheme.bodyMedium,
                            ),
                          ),
                          SingleChildScrollView(
                            padding: const EdgeInsets.only(top: 12),
                            child: _QuestNodeContent(
                              node: quest.rootNode,
                              discoveredIds: discoveredIds,
                              onPoITap: _focusPoI,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _QuestAccordion extends StatelessWidget {
  const _QuestAccordion({
    super.key,
    required this.title,
    required this.quests,
    required this.expanded,
    required this.onToggle,
    required this.onQuestTap,
  });

  final String title;
  final List<Quest> quests;
  final bool expanded;
  final VoidCallback onToggle;
  final void Function(Quest) onQuestTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Material(
        elevation: 1,
        borderRadius: BorderRadius.circular(12),
        child: Column(
          children: [
            InkWell(
              onTap: onToggle,
              borderRadius: BorderRadius.circular(12),
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                child: Row(
                  children: [
                    if (title == 'Tracked Quests')
                      const Padding(
                        padding: EdgeInsets.only(right: 8),
                        child: Text('⭐', style: TextStyle(fontSize: 20)),
                      ),
                    Expanded(
                      child: Text(
                        title,
                        style: Theme.of(context).textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.w600,
                            ),
                      ),
                    ),
                    Text(
                      '(${quests.length})',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: Theme.of(context)
                                .colorScheme
                                .onSurface
                                .withValues(alpha: 0.7),
                          ),
                    ),
                    Icon(
                      expanded ? Icons.expand_less : Icons.expand_more,
                      color: Theme.of(context)
                          .colorScheme
                          .onSurface
                          .withValues(alpha: 0.7),
                    ),
                  ],
                ),
              ),
            ),
            if (expanded)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
                child: Column(
                  children: quests
                      .map(
                        (q) => InkWell(
                          onTap: () => onQuestTap(q),
                          borderRadius: BorderRadius.circular(8),
                          child: Padding(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 10,
                            ),
                            child: Row(
                              children: [
                                Icon(
                                  q.isCompleted
                                      ? Icons.check_circle
                                      : Icons.radio_button_unchecked,
                                  size: 22,
                                  color: q.isCompleted
                                      ? Colors.green
                                      : Colors.grey.shade400,
                                ),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: Text(
                                    q.name,
                                    style: Theme.of(context).textTheme.bodyLarge,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      )
                      .toList(),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _QuestNodeContent extends StatelessWidget {
  const _QuestNodeContent({
    required this.node,
    required this.discoveredIds,
    required this.onPoITap,
  });

  final QuestNode node;
  final Set<String> discoveredIds;
  final void Function(PointOfInterest) onPoITap;

  @override
  Widget build(BuildContext context) {
    final poi = node.pointOfInterest;
    final discovered = discoveredIds.contains(poi.id);
    final imageUrl = discovered &&
            (poi.imageURL != null && poi.imageURL!.isNotEmpty)
        ? poi.imageURL!
        : _placeholderImageUrl;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        InkWell(
          onTap: () => onPoITap(poi),
          borderRadius: BorderRadius.circular(8),
          child: Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(6),
                  child: Image.network(
                    imageUrl,
                    width: 40,
                    height: 40,
                    fit: BoxFit.cover,
                    errorBuilder: (_, __, ___) => Container(
                      width: 40,
                      height: 40,
                      color: Colors.grey.shade300,
                      child: const Icon(Icons.place, size: 20),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        poi.name,
                        style: Theme.of(context).textTheme.titleSmall?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                      ),
                      ...node.objectives.map(
                        (o) => Padding(
                          padding: const EdgeInsets.only(top: 2),
                          child: Text(
                            '${o.challenge.question}${o.isCompleted ? ' ✅' : ''}',
                            style: Theme.of(context).textTheme.bodySmall,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
        ...node.objectives
            .where((o) => o.isCompleted && o.nextNode != null)
            .map(
              (o) => Padding(
                padding: const EdgeInsets.only(left: 24),
                child: _QuestNodeContent(
                  node: o.nextNode!,
                  discoveredIds: discoveredIds,
                  onPoITap: onPoITap,
                ),
              ),
            ),
      ],
    );
  }
}
