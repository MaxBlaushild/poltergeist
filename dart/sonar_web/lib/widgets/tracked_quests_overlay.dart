import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../providers/discoveries_provider.dart';
import '../providers/quest_log_provider.dart';

const _placeholderImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';

/// Expandable "Quests" overlay on the map. Lists tracked quests; tap a POI to
/// fly to it and open the POI panel.
class TrackedQuestsOverlay extends StatefulWidget {
  const TrackedQuestsOverlay({
    super.key,
    required this.onFocusPoI,
  });

  /// When user taps a POI: fly to location then open POI panel.
  final void Function(PointOfInterest poi) onFocusPoI;

  @override
  State<TrackedQuestsOverlay> createState() => _TrackedQuestsOverlayState();
}

class _TrackedQuestsOverlayState extends State<TrackedQuestsOverlay> {
  bool _expanded = false;
  bool _showContent = false;

  void _toggle() {
    setState(() {
      _expanded = !_expanded;
      if (_expanded) {
        Future.delayed(const Duration(milliseconds: 300), () {
          if (mounted) setState(() => _showContent = true);
        });
      } else {
        _showContent = false;
      }
    });
  }

  void _onPoITap(PointOfInterest poi) {
    setState(() {
      _expanded = false;
      _showContent = false;
    });
    widget.onFocusPoI(poi);
  }

  @override
  Widget build(BuildContext context) {
    return Consumer2<QuestLogProvider, DiscoveriesProvider>(
      builder: (context, ql, discoveries, _) {
        final tracked = ql.quests
            .where((q) => ql.trackedQuestIds.contains(q.id))
            .toList();
        final discoveredIds = <String>{};
        for (final d in discoveries.discoveries) {
          discoveredIds.add(d.pointOfInterestId);
        }

        if (tracked.isEmpty) return const SizedBox.shrink();

        final screenWidth = MediaQuery.sizeOf(context).width;
        const rightMargin = 16.0;
        const minSideMargin = 16.0;
        final maxAllowedWidth = (screenWidth - rightMargin - minSideMargin).clamp(80.0, 288.0);
        final width = _expanded ? maxAllowedWidth : 80.0;

        return AnimatedContainer(
          duration: const Duration(milliseconds: 300),
          width: width,
          child: Material(
            color: Colors.black54,
            borderRadius: BorderRadius.circular(12),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                InkWell(
                  onTap: _toggle,
                  borderRadius: BorderRadius.circular(12),
                  child: Padding(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 10,
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Quests',
                          style: Theme.of(context).textTheme.titleSmall?.copyWith(
                                color: Colors.white,
                              ),
                        ),
                        Icon(
                          _expanded ? Icons.expand_less : Icons.expand_more,
                          color: Colors.white,
                          size: 20,
                        ),
                      ],
                    ),
                  ),
                ),
                AnimatedCrossFade(
                  firstChild: const SizedBox.shrink(),
                  secondChild: _showContent
                      ? Padding(
                          padding: const EdgeInsets.fromLTRB(8, 0, 8, 8),
                          child: ConstrainedBox(
                            constraints: const BoxConstraints(maxHeight: 384),
                            child: SingleChildScrollView(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.stretch,
                                children: tracked
                                    .map(
                                      (quest) => Padding(
                                        padding: const EdgeInsets.only(bottom: 8),
                                        child: _TrackedQuestCard(
                                          quest: quest,
                                          discoveredIds: discoveredIds,
                                          onPoITap: _onPoITap,
                                        ),
                                      ),
                                    )
                                    .toList(),
                              ),
                            ),
                          ),
                        )
                      : const SizedBox.shrink(),
                  crossFadeState: _expanded && _showContent
                      ? CrossFadeState.showSecond
                      : CrossFadeState.showFirst,
                  duration: const Duration(milliseconds: 200),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

class _TrackedQuestCard extends StatelessWidget {
  const _TrackedQuestCard({
    required this.quest,
    required this.discoveredIds,
    required this.onPoITap,
  });

  final Quest quest;
  final Set<String> discoveredIds;
  final void Function(PointOfInterest) onPoITap;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.black26,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            quest.name,
            style: Theme.of(context).textTheme.titleSmall?.copyWith(
                  color: Colors.white,
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: 8),
          _QuestNodeTile(
            node: quest.rootNode,
            discoveredIds: discoveredIds,
            onPoITap: onPoITap,
            indent: 0,
          ),
        ],
      ),
    );
  }
}

class _QuestNodeTile extends StatelessWidget {
  const _QuestNodeTile({
    required this.node,
    required this.discoveredIds,
    required this.onPoITap,
    this.indent = 0,
  });

  final QuestNode node;
  final Set<String> discoveredIds;
  final void Function(PointOfInterest) onPoITap;
  final int indent;

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
          borderRadius: BorderRadius.circular(6),
          child: Padding(
            padding: EdgeInsets.only(
              left: 8.0 + indent * 16,
              top: 4,
              bottom: 4,
              right: 8,
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: Image.network(
                    imageUrl,
                    width: 24,
                    height: 24,
                    fit: BoxFit.cover,
                    errorBuilder: (_, __, ___) => Container(
                      width: 24,
                      height: 24,
                      color: Colors.grey.shade700,
                      child: const Icon(Icons.place, size: 14, color: Colors.white70),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        poi.name,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Colors.white,
                              fontWeight: FontWeight.w500,
                            ),
                      ),
                      ...node.objectives.map(
                        (o) => Text(
                          '${o.challenge.question}${o.isCompleted ? ' ✅' : ''}',
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: Colors.white70,
                                fontSize: 11,
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
              (o) => _QuestNodeTile(
                node: o.nextNode!,
                discoveredIds: discoveredIds,
                onPoITap: onPoITap,
                indent: indent + 1,
              ),
            ),
      ],
    );
  }
}
