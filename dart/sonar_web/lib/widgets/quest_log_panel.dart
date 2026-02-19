import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../providers/discoveries_provider.dart';
import '../providers/quest_log_provider.dart';
import '../providers/tags_provider.dart';

const _placeholderImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';

/// Bottom-sheet content for Quest Log.
/// [onFocusPoI] when user taps a POI in a quest: close sheet, fly to POI, open POI panel.
typedef OnFocusPoI = void Function(PointOfInterest poi);
typedef OnFocusTurnInQuest = void Function(Quest quest);

class QuestLogPanel extends StatefulWidget {
  const QuestLogPanel({
    super.key,
    required this.onClose,
    required this.onFocusPoI,
    required this.onFocusTurnInQuest,
  });

  final VoidCallback onClose;
  final OnFocusPoI onFocusPoI;
  final OnFocusTurnInQuest onFocusTurnInQuest;

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

  void _focusTurnInQuest(Quest quest) {
    widget.onClose();
    widget.onFocusTurnInQuest(quest);
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
        final readyToTurnIn = ql.quests
            .where((q) =>
                q.turnedInAt == null &&
                (q.readyToTurnIn || (q.currentNode == null && q.isAccepted)))
            .toList();
        final readyIds = readyToTurnIn.map((q) => q.id).toSet();
        final tracked = ql.quests
            .where((q) =>
                ql.trackedQuestIds.contains(q.id) && !readyIds.contains(q.id))
            .toList();
        final tagBuckets = <String, List<Quest>>{};
        for (final g in tags.tagGroups) {
          tagBuckets[g.id] = [];
        }
        final untagged = <Quest>[];

        for (final q in ql.quests) {
          if (readyIds.contains(q.id)) continue;
          if (ql.trackedQuestIds.contains(q.id)) continue;
          final tagNames = _questTags(q);
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

        final completed = ql.completedQuests;

        final hasQuestListItems = readyToTurnIn.isNotEmpty ||
            tracked.isNotEmpty ||
            tagBuckets.values.any((list) => list.isNotEmpty) ||
            untagged.isNotEmpty;

        return DefaultTabController(
          length: 2,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const TabBar(
                tabs: [
                  Tab(text: 'Quests'),
                  Tab(text: 'Completed'),
                ],
              ),
              Expanded(
                child: TabBarView(
                  children: [
                    SingleChildScrollView(
                      padding: const EdgeInsets.all(16),
                      child: hasQuestListItems
                          ? Column(
                              crossAxisAlignment: CrossAxisAlignment.stretch,
                              children: [
                                if (readyToTurnIn.isNotEmpty)
                                  _QuestAccordion(
                                    title: 'Ready to Turn In',
                                    quests: readyToTurnIn,
                                    expanded: _expanded['ready'] ?? true,
                                    onToggle: () {
                                      setState(() {
                                        _expanded['ready'] =
                                            !(_expanded['ready'] ?? true);
                                      });
                                    },
                                    onQuestTap: (q) =>
                                        setState(() => _selectedQuest = q),
                                    onReadyQuestTap: _focusTurnInQuest,
                                  ),
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
                                    onReadyQuestTap: _focusTurnInQuest,
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
                                    onReadyQuestTap: _focusTurnInQuest,
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
                                    onReadyQuestTap: _focusTurnInQuest,
                                  ),
                              ],
                            )
                          : const Padding(
                              padding: EdgeInsets.symmetric(vertical: 48),
                              child: Center(
                                child: Text(
                                  'No quests yet. Explore to discover new adventures.',
                                ),
                              ),
                            ),
                    ),
                    if (completed.isEmpty)
                      const Center(
                        child: Padding(
                          padding: EdgeInsets.all(24),
                          child: Text('No completed quests yet.'),
                        ),
                      )
                    else
                      SingleChildScrollView(
                        padding: const EdgeInsets.all(16),
                        child: _QuestAccordion(
                          title: 'Completed Quests',
                          quests: completed,
                          expanded: _expanded['completed'] ?? true,
                          onToggle: () {
                            setState(() {
                              _expanded['completed'] =
                                  !(_expanded['completed'] ?? true);
                            });
                          },
                          onQuestTap: (q) =>
                              setState(() => _selectedQuest = q),
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

  List<String> _questTags(Quest quest) {
    final poi = quest.currentNode?.pointOfInterest;
    if (poi == null) return [];
    return poi.tags.map((t) => t.name).toList();
  }

  Widget _buildQuestDetail(BuildContext context, Quest quest) {
    return Consumer2<QuestLogProvider, DiscoveriesProvider>(
      builder: (context, ql, discoveries, _) {
        final isTracked = ql.trackedQuestIds.contains(quest.id);
        final node = quest.currentNode;
        final poi = node?.pointOfInterest;
        final discoveredIds = <String>{
          for (final d in discoveries.discoveries) d.pointOfInterestId
        };

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
              const SizedBox(height: 8),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    quest.turnedInAt != null
                        ? 'Completed'
                        : quest.readyToTurnIn
                            ? 'Ready to turn in'
                            : quest.isAccepted
                                ? 'In progress'
                                : 'Not accepted',
                    style: Theme.of(context).textTheme.titleSmall,
                  ),
                  if (quest.turnedInAt == null)
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
              Text(
                quest.description,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 16),
              Text(
                'Rewards',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
              ),
              const SizedBox(height: 8),
              if (quest.gold > 0)
                Padding(
                  padding: const EdgeInsets.only(bottom: 6),
                  child: Text(
                    '+${quest.gold} Gold',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                ),
              if (quest.itemRewards.isNotEmpty)
                ...quest.itemRewards.map((reward) {
                  final itemName = reward.inventoryItem?.name ?? 'Item';
                  final qty = reward.quantity > 0 ? reward.quantity : 1;
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Text(
                      '+$qty $itemName',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  );
                }),
              if (quest.gold <= 0 && quest.itemRewards.isEmpty)
                Text(
                  'No rewards listed.',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(color: Colors.grey.shade600),
                ),
              const SizedBox(height: 16),
              if (node == null)
                Text(
                  quest.turnedInAt != null
                      ? 'Quest turned in. Well done!'
                      : 'Quest completed! Turn it in for rewards.',
                  style: Theme.of(context).textTheme.titleMedium,
                )
              else ...[
                Text(
                  'Current Objective',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
                const SizedBox(height: 8),
                if (poi != null)
                  _QuestPoiCard(
                    poi: poi,
                    discovered: discoveredIds.contains(poi.id),
                    onTap: () => _focusPoI(poi),
                  )
                else
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.amber.shade50,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: Colors.amber.shade200),
                    ),
                    child: const Text(
                      'Reach the highlighted quest area to submit your answer.',
                    ),
                  ),
                const SizedBox(height: 12),
                Text(
                  'Challenges',
                  style: Theme.of(context).textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
                const SizedBox(height: 8),
                ...node.challenges.map(
                  (c) => Padding(
                    padding: const EdgeInsets.only(bottom: 6),
                    child: Text(
                      '• ${c.question}',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  ),
                ),
              ],
            ],
          ),
        );
      },
    );
  }
}

class _QuestPoiCard extends StatelessWidget {
  const _QuestPoiCard({
    required this.poi,
    required this.discovered,
    required this.onTap,
  });

  final PointOfInterest poi;
  final bool discovered;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final imageUrl = discovered && poi.imageURL != null && poi.imageURL!.isNotEmpty
        ? poi.imageURL!
        : _placeholderImageUrl;
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.all(10),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: Colors.grey.shade300),
        ),
        child: Row(
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(6),
              child: Image.network(
                imageUrl,
                width: 48,
                height: 48,
                fit: BoxFit.cover,
                errorBuilder: (_, __, ___) => Container(
                  width: 48,
                  height: 48,
                  color: Colors.grey.shade300,
                  child: const Icon(Icons.place, size: 20),
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                poi.name,
                style: Theme.of(context).textTheme.titleSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
              ),
            ),
            const Icon(Icons.chevron_right),
          ],
        ),
      ),
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
    this.onReadyQuestTap,
  });

  final String title;
  final List<Quest> quests;
  final bool expanded;
  final VoidCallback onToggle;
  final void Function(Quest) onQuestTap;
  final void Function(Quest)? onReadyQuestTap;

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
                          onTap: () {
                            if (q.readyToTurnIn && onReadyQuestTap != null) {
                              onReadyQuestTap!(q);
                              return;
                            }
                            onQuestTap(q);
                          },
                          borderRadius: BorderRadius.circular(8),
                          child: Padding(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 10,
                            ),
                            child: Row(
                              children: [
                                q.turnedInAt != null || q.readyToTurnIn
                                    ? Container(
                                        width: 22,
                                        height: 22,
                                        decoration: const BoxDecoration(
                                          color: Color(0xFF3BB54A),
                                          shape: BoxShape.circle,
                                        ),
                                        child: const Icon(
                                          Icons.check,
                                          size: 14,
                                          color: Colors.white,
                                        ),
                                      )
                                    : Icon(
                                        q.isAccepted
                                            ? Icons.play_circle_fill
                                            : Icons.radio_button_unchecked,
                                        size: 22,
                                        color: q.isAccepted
                                            ? Colors.orange
                                            : Colors.grey.shade400,
                                      ),
                                const SizedBox(width: 12),
                                ClipRRect(
                                  borderRadius: BorderRadius.circular(6),
                                  child: Image.network(
                                    (q.currentNode?.pointOfInterest?.imageURL ?? '').isNotEmpty
                                        ? q.currentNode!.pointOfInterest!.imageURL!
                                        : _placeholderImageUrl,
                                    width: 36,
                                    height: 36,
                                    fit: BoxFit.cover,
                                    errorBuilder: (_, __, ___) => Container(
                                      width: 36,
                                      height: 36,
                                      color: Colors.grey.shade300,
                                      child: const Icon(Icons.place, size: 18),
                                    ),
                                  ),
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
