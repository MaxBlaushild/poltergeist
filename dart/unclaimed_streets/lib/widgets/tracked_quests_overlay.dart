import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/character.dart';
import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../models/quest_node.dart';
import '../providers/discoveries_provider.dart';
import '../providers/quest_log_provider.dart';
import 'quest_objective_display.dart';
import 'quest_turn_in_target.dart';

class TrackedQuestsOverlayController extends ChangeNotifier {
  void open() {
    notifyListeners();
  }
}

/// Expandable "Quests" overlay on the map. Lists tracked quests; tap a POI to
/// fly to it and open the POI panel.
class TrackedQuestsOverlay extends StatefulWidget {
  const TrackedQuestsOverlay({
    super.key,
    required this.onFocusPoI,
    required this.onFocusNode,
    this.onFocusTurnInQuest,
    this.onOpenQuestDetails,
    this.resolveQuestReceiverCharacter,
    this.resolveQuestReceiverPoi,
    this.controller,
    this.featuredMainStoryPoi,
    this.featuredMainStoryQuestGiverName,
    this.onFocusFeaturedMainStoryLead,
  });

  /// When user taps a POI: fly to location then open POI panel.
  final void Function(PointOfInterest poi) onFocusPoI;
  final void Function(QuestNode node) onFocusNode;
  final void Function(Quest quest)? onFocusTurnInQuest;
  final void Function(Quest quest)? onOpenQuestDetails;
  final Character? Function(Quest quest)? resolveQuestReceiverCharacter;
  final PointOfInterest? Function(Quest quest)? resolveQuestReceiverPoi;
  final TrackedQuestsOverlayController? controller;
  final PointOfInterest? featuredMainStoryPoi;
  final String? featuredMainStoryQuestGiverName;
  final VoidCallback? onFocusFeaturedMainStoryLead;

  @override
  State<TrackedQuestsOverlay> createState() => _TrackedQuestsOverlayState();
}

class _TrackedQuestsOverlayState extends State<TrackedQuestsOverlay> {
  bool _expanded = false;
  bool _showContent = false;
  TrackedQuestsOverlayController? _controller;
  List<Quest> _cachedTracked = const [];

  @override
  void initState() {
    super.initState();
    _controller = widget.controller;
    _controller?.addListener(_handleController);
  }

  @override
  void didUpdateWidget(covariant TrackedQuestsOverlay oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.controller != widget.controller) {
      oldWidget.controller?.removeListener(_handleController);
      _controller = widget.controller;
      _controller?.addListener(_handleController);
    }
  }

  @override
  void dispose() {
    _controller?.removeListener(_handleController);
    super.dispose();
  }

  void _handleController() {
    if (!mounted) return;
    _expand();
  }

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

  void _expand() {
    if (_expanded) return;
    setState(() {
      _expanded = true;
      _showContent = false;
    });
    Future.delayed(const Duration(milliseconds: 300), () {
      if (mounted) setState(() => _showContent = true);
    });
  }

  void _onPoITap(PointOfInterest poi) {
    setState(() {
      _expanded = false;
      _showContent = false;
    });
    widget.onFocusPoI(poi);
  }

  void _onFeaturedMainStoryLeadTap() {
    final poi = widget.featuredMainStoryPoi;
    if (poi == null) return;
    setState(() {
      _expanded = false;
      _showContent = false;
    });
    final onFocusLead = widget.onFocusFeaturedMainStoryLead;
    if (onFocusLead != null) {
      onFocusLead();
      return;
    }
    widget.onFocusPoI(poi);
  }

  @override
  Widget build(BuildContext context) {
    return Consumer2<QuestLogProvider, DiscoveriesProvider>(
      builder: (context, ql, discoveries, _) {
        final tracked =
            ql.quests.where((q) => ql.trackedQuestIds.contains(q.id)).toList()
              ..sort((a, b) {
                if (a.isTutorial != b.isTutorial) {
                  return a.isTutorial ? -1 : 1;
                }
                if (a.isMainStory != b.isMainStory) {
                  return a.isMainStory ? -1 : 1;
                }
                final aAwaitingTurnIn = questIsAwaitingTurnIn(a);
                final bAwaitingTurnIn = questIsAwaitingTurnIn(b);
                if (aAwaitingTurnIn != bAwaitingTurnIn) {
                  return aAwaitingTurnIn ? -1 : 1;
                }
                return a.name.toLowerCase().compareTo(b.name.toLowerCase());
              });
        if (tracked.isNotEmpty) {
          _cachedTracked = List<Quest>.from(tracked);
        } else if (ql.trackedQuestIds.isEmpty) {
          _cachedTracked = const [];
        } else if (_cachedTracked.isNotEmpty) {
          _cachedTracked = _cachedTracked
              .where((q) => ql.trackedQuestIds.contains(q.id))
              .toList();
        }
        final visibleTracked = tracked.isNotEmpty ? tracked : _cachedTracked;
        final featuredMainStoryPoi = widget.featuredMainStoryPoi;
        final discoveredIds = <String>{
          for (final d in discoveries.discoveries) d.pointOfInterestId,
        };

        if (visibleTracked.isEmpty && featuredMainStoryPoi == null) {
          return const SizedBox.shrink();
        }

        final screenWidth = MediaQuery.sizeOf(context).width;
        const rightMargin = 16.0;
        const minSideMargin = 16.0;
        const collapsedWidth = 96.0;
        final maxAllowedWidth = (screenWidth - rightMargin - minSideMargin)
            .clamp(collapsedWidth, 288.0);
        final width = _expanded ? maxAllowedWidth : collapsedWidth;

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
                    padding: EdgeInsets.fromLTRB(
                      12,
                      12,
                      _expanded ? 12 : 24,
                      12,
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Quests',
                          style: Theme.of(
                            context,
                          ).textTheme.titleSmall?.copyWith(color: Colors.white),
                        ),
                        Icon(
                          _expanded ? Icons.expand_less : Icons.expand_more,
                          color: Colors.white,
                          size: 24,
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
                                children: [
                                  if (featuredMainStoryPoi != null)
                                    Padding(
                                      padding: const EdgeInsets.only(bottom: 8),
                                      child: _ImportantQuestLeadCard(
                                        poi: featuredMainStoryPoi,
                                        questGiverName: widget
                                            .featuredMainStoryQuestGiverName,
                                        onTap: _onFeaturedMainStoryLeadTap,
                                      ),
                                    ),
                                  ...visibleTracked.map(
                                    (quest) => Padding(
                                      padding: const EdgeInsets.only(bottom: 8),
                                      child: _TrackedQuestCard(
                                        quest: quest,
                                        discoveredIds: discoveredIds,
                                        onPoITap: _onPoITap,
                                        onNodeTap: widget.onFocusNode,
                                        onTurnInTap: widget.onFocusTurnInQuest,
                                        onOpenQuestDetails:
                                            widget.onOpenQuestDetails,
                                        resolveQuestReceiverCharacter: widget
                                            .resolveQuestReceiverCharacter,
                                        resolveQuestReceiverPoi:
                                            widget.resolveQuestReceiverPoi,
                                      ),
                                    ),
                                  ),
                                ],
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

class _ImportantQuestLeadCard extends StatelessWidget {
  const _ImportantQuestLeadCard({
    required this.poi,
    required this.onTap,
    this.questGiverName,
  });

  final PointOfInterest poi;
  final VoidCallback onTap;
  final String? questGiverName;

  @override
  Widget build(BuildContext context) {
    final questGiver = questGiverName?.trim() ?? '';
    final locationName = poi.name.trim().isNotEmpty
        ? poi.name.trim()
        : 'nearby';
    final headline = questGiver.isNotEmpty
        ? 'Talk to $questGiver.'
        : 'A main story thread is waiting nearby.';

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        padding: const EdgeInsets.all(10),
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            colors: [Color(0xFF7A1823), Color(0xFF4E0F17)],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: const Color(0xFFE7C36A)),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: const Color(0xFFE7C36A),
                borderRadius: BorderRadius.circular(999),
              ),
              child: Text(
                'Important Quest Available',
                style: Theme.of(context).textTheme.labelSmall?.copyWith(
                  color: const Color(0xFF3A1A11),
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
            const SizedBox(height: 10),
            Text(
              headline,
              style: Theme.of(context).textTheme.titleSmall?.copyWith(
                color: Colors.white,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              'Head to $locationName.',
              style: Theme.of(
                context,
              ).textTheme.bodySmall?.copyWith(color: const Color(0xFFF8EBD0)),
            ),
          ],
        ),
      ),
    );
  }
}

class _TrackedQuestCard extends StatelessWidget {
  const _TrackedQuestCard({
    required this.quest,
    required this.discoveredIds,
    required this.onPoITap,
    required this.onNodeTap,
    this.onTurnInTap,
    this.onOpenQuestDetails,
    this.resolveQuestReceiverCharacter,
    this.resolveQuestReceiverPoi,
  });

  final Quest quest;
  final Set<String> discoveredIds;
  final void Function(PointOfInterest) onPoITap;
  final void Function(QuestNode) onNodeTap;
  final void Function(Quest quest)? onTurnInTap;
  final void Function(Quest quest)? onOpenQuestDetails;
  final Character? Function(Quest quest)? resolveQuestReceiverCharacter;
  final PointOfInterest? Function(Quest quest)? resolveQuestReceiverPoi;

  @override
  Widget build(BuildContext context) {
    final node = quest.currentNode;
    final poi = node?.pointOfInterest;
    final awaitingTurnIn = questIsAwaitingTurnIn(quest);
    final questReceiver = resolveQuestReceiverCharacter?.call(quest);
    final questReceiverPoi = resolveQuestReceiverPoi?.call(quest);
    final objectiveLines = questObjectiveLines(node);
    final hasDirectFocusTarget = questNodeHasDirectFocusTarget(node);
    final detailFallbackTap = onOpenQuestDetails == null
        ? null
        : () => onOpenQuestDetails!(quest);
    final VoidCallback? nodeTap = hasDirectFocusTarget
        ? () => onNodeTap(node!)
        : detailFallbackTap;

    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.black26,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  quest.name,
                  style: Theme.of(context).textTheme.titleSmall?.copyWith(
                    color: Colors.white,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          if (awaitingTurnIn)
            _TrackedQuestTurnInTile(
              questReceiver: questReceiver,
              pointOfInterest: questReceiverPoi,
              onTap:
                  onTurnInTap == null ||
                      (questReceiver == null && questReceiverPoi == null)
                  ? detailFallbackTap
                  : () => onTurnInTap!(quest),
            )
          else if (poi != null)
            _QuestPoiTile(
              node: node!,
              poi: poi,
              discoveredIds: discoveredIds,
              onTap: () => onPoITap(poi),
              onChallengeTap: () => onNodeTap(node),
              onChevronTap: onOpenQuestDetails == null
                  ? null
                  : () => onOpenQuestDetails!(quest),
              objectiveLines: objectiveLines,
            )
          else
            InkWell(
              onTap: nodeTap,
              borderRadius: BorderRadius.circular(6),
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: 2),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    QuestObjectiveIcon(
                      node: node,
                      discoveredPoiIds: discoveredIds,
                      size: 28,
                      borderRadius: 4,
                      iconColor: Colors.white70,
                      backgroundColor: Colors.grey.shade700,
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          QuestObjectiveChallengeBadge(node: node),
                          if (questObjectiveChallengeLabel(node) != null)
                            const SizedBox(height: 6),
                          ...objectiveLines.map(
                            (line) => Padding(
                              padding: const EdgeInsets.only(bottom: 4),
                              child: Text(
                                line,
                                style: Theme.of(context).textTheme.bodySmall
                                    ?.copyWith(color: Colors.white70),
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
        ],
      ),
    );
  }
}

class _TrackedQuestTurnInTile extends StatelessWidget {
  const _TrackedQuestTurnInTile({
    required this.questReceiver,
    required this.pointOfInterest,
    this.onTap,
  });

  final Character? questReceiver;
  final PointOfInterest? pointOfInterest;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final receiverName = questTurnInReceiverLabel(questReceiver);
    final locationName = questTurnInLocationLabel(
      character: questReceiver,
      pointOfInterest: pointOfInterest,
    );
    final headline = questReceiver != null
        ? 'Return to $receiverName'
        : 'Turn this quest in';
    final subhead = pointOfInterest != null
        ? 'Collect your rewards at $locationName.'
        : 'Collect your rewards from $receiverName.';

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.all(10),
        decoration: BoxDecoration(
          color: const Color(0x1FFFFFFF),
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: const Color(0x33F1D597)),
        ),
        child: Row(
          children: [
            QuestTurnInPortrait(
              character: questReceiver,
              size: 40,
              backgroundColor: const Color(0x33F1D597),
              foregroundColor: const Color(0xFFF1D597),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF1D597),
                      borderRadius: BorderRadius.circular(999),
                    ),
                    child: Text(
                      'Ready to Turn In',
                      style: Theme.of(context).textTheme.labelSmall?.copyWith(
                        color: const Color(0xFF3A1A11),
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    headline,
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Colors.white,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subhead,
                    style: Theme.of(
                      context,
                    ).textTheme.bodySmall?.copyWith(color: Colors.white70),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            const Icon(Icons.chevron_right, size: 18, color: Colors.white70),
          ],
        ),
      ),
    );
  }
}

class _QuestPoiTile extends StatelessWidget {
  const _QuestPoiTile({
    required this.node,
    required this.poi,
    required this.discoveredIds,
    required this.onTap,
    required this.onChallengeTap,
    required this.objectiveLines,
    this.onChevronTap,
  });

  final QuestNode node;
  final PointOfInterest poi;
  final Set<String> discoveredIds;
  final VoidCallback onTap;
  final VoidCallback onChallengeTap;
  final List<String> objectiveLines;
  final VoidCallback? onChevronTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 4, top: 4, bottom: 4, right: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: InkWell(
              onTap: onTap,
              borderRadius: BorderRadius.circular(6),
              child: Padding(
                padding: const EdgeInsets.only(right: 6),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    QuestObjectiveIcon(
                      node: node,
                      discoveredPoiIds: discoveredIds,
                      size: 28,
                      borderRadius: 4,
                      iconColor: Colors.white70,
                      backgroundColor: Colors.grey.shade700,
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            poi.name,
                            style: Theme.of(context).textTheme.bodySmall
                                ?.copyWith(
                                  color: Colors.white,
                                  fontWeight: FontWeight.w600,
                                ),
                          ),
                          const SizedBox(height: 4),
                          QuestObjectiveChallengeBadge(node: node),
                          if (questObjectiveChallengeLabel(node) != null)
                            const SizedBox(height: 4),
                          ...objectiveLines.map(
                            (line) => GestureDetector(
                              behavior: HitTestBehavior.opaque,
                              onTap: onChallengeTap,
                              child: Padding(
                                padding: const EdgeInsets.symmetric(
                                  vertical: 2,
                                ),
                                child: Text(
                                  line,
                                  style: Theme.of(context).textTheme.bodySmall
                                      ?.copyWith(color: Colors.white70),
                                ),
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
          ),
          InkWell(
            onTap: onChevronTap ?? onTap,
            borderRadius: BorderRadius.circular(6),
            child: const Padding(
              padding: EdgeInsets.all(2),
              child: Icon(Icons.chevron_right, size: 16, color: Colors.white70),
            ),
          ),
        ],
      ),
    );
  }
}
