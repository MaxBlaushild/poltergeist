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
  String? _pendingOpenItemId;

  void open({String? itemId}) {
    final trimmed = itemId?.trim() ?? '';
    _pendingOpenItemId = trimmed.isEmpty ? null : trimmed;
    notifyListeners();
  }

  String? consumePendingOpenItemId() {
    final pendingItemId = _pendingOpenItemId;
    _pendingOpenItemId = null;
    return pendingItemId;
  }
}

/// Expandable "Quests" overlay on the map. Lists tracked quests and lets the
/// screen decide whether a tap should focus in place or switch zones first.
class TrackedQuestsOverlay extends StatefulWidget {
  const TrackedQuestsOverlay({
    super.key,
    required this.onFocusPoI,
    required this.onFocusNode,
    this.onFocusTurnInQuest,
    this.onPreviewPoI,
    this.onPreviewNode,
    this.onPreviewTurnInQuest,
    this.resolveQuestReceiverCharacter,
    this.resolveQuestReceiverPoi,
    this.controller,
    this.onCloseOverlay,
    this.tutorialScenarioTitle,
    this.tutorialScenarioDetail,
    this.tutorialScenarioObjectiveCopy,
    this.onFocusTutorialScenario,
    this.onPreviewTutorialScenario,
    this.tutorialMonsterTitle,
    this.tutorialMonsterDetail,
    this.tutorialMonsterObjectiveCopy,
    this.onFocusTutorialMonster,
    this.onPreviewTutorialMonster,
    this.expandUpwards = false,
    this.collapsedHeight = 48,
    this.maxExpandedHeight = 384,
  });

  /// When user taps a POI: focus that quest target on the map.
  final void Function(PointOfInterest poi) onFocusPoI;
  final void Function(QuestNode node) onFocusNode;
  final void Function(Quest quest)? onFocusTurnInQuest;
  final void Function(PointOfInterest poi)? onPreviewPoI;
  final void Function(QuestNode node)? onPreviewNode;
  final void Function(Quest quest)? onPreviewTurnInQuest;
  final Character? Function(Quest quest)? resolveQuestReceiverCharacter;
  final PointOfInterest? Function(Quest quest)? resolveQuestReceiverPoi;
  final TrackedQuestsOverlayController? controller;
  final VoidCallback? onCloseOverlay;
  final String? tutorialScenarioTitle;
  final String? tutorialScenarioDetail;
  final String? tutorialScenarioObjectiveCopy;
  final VoidCallback? onFocusTutorialScenario;
  final VoidCallback? onPreviewTutorialScenario;
  final String? tutorialMonsterTitle;
  final String? tutorialMonsterDetail;
  final String? tutorialMonsterObjectiveCopy;
  final VoidCallback? onFocusTutorialMonster;
  final VoidCallback? onPreviewTutorialMonster;
  final bool expandUpwards;
  final double collapsedHeight;
  final double maxExpandedHeight;

  @override
  State<TrackedQuestsOverlay> createState() => _TrackedQuestsOverlayState();
}

class _TrackedQuestsOverlayState extends State<TrackedQuestsOverlay> {
  static const double _carouselSwipeThreshold = 40;
  static const double _carouselSwipeVelocityThreshold = 300;

  bool _expanded = false;
  bool _showContent = false;
  TrackedQuestsOverlayController? _controller;
  List<Quest> _cachedTracked = const [];
  List<_TrackedQuestCarouselItem> _carouselItems = const [];
  int _currentItemIndex = 0;
  int _carouselTransitionDirection = 1;
  double _dragDeltaX = 0;
  bool _previewCurrentItemAfterBuild = false;
  String? _lastViewedCarouselItemId;
  String? _preferredOpenCarouselItemId;
  String? _pendingOpenCarouselItemId;
  Set<String> _lastAcceptedTrackedQuestIds = <String>{};
  bool _hasAcceptedTrackedQuestSnapshot = false;

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
    final pendingOpenItemId = _controller?.consumePendingOpenItemId();
    if (pendingOpenItemId != null) {
      _preferredOpenCarouselItemId = pendingOpenItemId;
      _pendingOpenCarouselItemId = pendingOpenItemId;
      if (_expanded) {
        final currentIndex = _resolvedItemIndex(_carouselItems.length);
        final pendingIndex = _indexOfCarouselItemId(pendingOpenItemId);
        setState(() {
          if (pendingIndex != -1) {
            _carouselTransitionDirection = pendingIndex >= currentIndex
                ? 1
                : -1;
            _currentItemIndex = pendingIndex;
          }
          _previewCurrentItemAfterBuild = true;
        });
        return;
      }
    }
    _expand();
  }

  void _toggle() {
    if (_expanded) {
      _collapseAndRestorePlayer();
      return;
    }
    _expand();
  }

  void _expand() {
    if (_expanded) return;
    setState(() {
      _expanded = true;
      _showContent = false;
      _previewCurrentItemAfterBuild = true;
      _pendingOpenCarouselItemId =
          _preferredOpenCarouselItemId ?? _lastViewedCarouselItemId;
      _preferredOpenCarouselItemId = null;
    });
    Future.delayed(const Duration(milliseconds: 300), () {
      if (mounted) setState(() => _showContent = true);
    });
  }

  void _collapseAndRestorePlayer() {
    _collapse();
    widget.onCloseOverlay?.call();
  }

  void _collapse() {
    setState(() {
      _expanded = false;
      _showContent = false;
    });
  }

  void _onPoITap(PointOfInterest poi) {
    _collapse();
    widget.onFocusPoI(poi);
  }

  int _resolvedItemIndex(int itemCount) {
    if (itemCount <= 0) return 0;
    if (_currentItemIndex < 0) return 0;
    if (_currentItemIndex >= itemCount) return itemCount - 1;
    return _currentItemIndex;
  }

  void _syncCurrentItemIndexIfNeeded(int desiredIndex) {
    if (desiredIndex == _currentItemIndex || !mounted) return;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || _currentItemIndex == desiredIndex) return;
      setState(() => _currentItemIndex = desiredIndex);
    });
  }

  int _indexOfCarouselItemId(String itemId) {
    return _carouselItems.indexWhere((item) => item.id == itemId);
  }

  int _resolvedDisplayItemIndex() {
    if (_carouselItems.isEmpty) return 0;

    final pendingOpenItemId = _pendingOpenCarouselItemId;
    if (pendingOpenItemId != null) {
      final pendingIndex = _indexOfCarouselItemId(pendingOpenItemId);
      if (pendingIndex != -1) {
        return pendingIndex;
      }
    }

    final lastViewedCarouselItemId = _lastViewedCarouselItemId;
    if (lastViewedCarouselItemId != null) {
      final lastViewedIndex = _indexOfCarouselItemId(lastViewedCarouselItemId);
      if (lastViewedIndex != -1) {
        return lastViewedIndex;
      }
    }

    return _resolvedItemIndex(_carouselItems.length);
  }

  void _rememberCarouselItemAt(int index) {
    if (index < 0 || index >= _carouselItems.length) return;
    _lastViewedCarouselItemId = _carouselItems[index].id;
    _preferredOpenCarouselItemId = null;
  }

  void _previewCarouselItem(int index) {
    if (index < 0 || index >= _carouselItems.length) return;
    _carouselItems[index].onPreview?.call();
  }

  VoidCallback? _buildQuestPreviewCallback(Quest quest) {
    final node = quest.currentNode;
    final poi = node?.pointOfInterest;
    final awaitingTurnIn = questIsAwaitingTurnIn(quest);
    final questReceiver = widget.resolveQuestReceiverCharacter?.call(quest);
    final questReceiverPoi = widget.resolveQuestReceiverPoi?.call(quest);
    final hasDirectFocusTarget = questNodeHasDirectFocusTarget(node);

    if (awaitingTurnIn) {
      if (widget.onPreviewTurnInQuest == null ||
          (questReceiver == null && questReceiverPoi == null)) {
        return null;
      }
      return () => widget.onPreviewTurnInQuest!(quest);
    }
    if (poi != null) {
      if (widget.onPreviewPoI == null) return null;
      return () => widget.onPreviewPoI!(poi);
    }
    if (hasDirectFocusTarget && node != null) {
      if (widget.onPreviewNode == null) return null;
      return () => widget.onPreviewNode!(node);
    }
    return null;
  }

  void _goToItem(int nextIndex, {required int itemCount}) {
    final currentIndex = _resolvedItemIndex(itemCount);
    if (nextIndex < 0 || nextIndex >= itemCount || nextIndex == currentIndex) {
      return;
    }
    setState(() {
      _carouselTransitionDirection = nextIndex > currentIndex ? 1 : -1;
      _currentItemIndex = nextIndex;
    });
    _rememberCarouselItemAt(nextIndex);
    _previewCarouselItem(nextIndex);
  }

  void _showPreviousItem({required int itemCount}) {
    _goToItem(_resolvedItemIndex(itemCount) - 1, itemCount: itemCount);
  }

  void _showNextItem({required int itemCount}) {
    _goToItem(_resolvedItemIndex(itemCount) + 1, itemCount: itemCount);
  }

  void _handleHorizontalDragStart(DragStartDetails details) {
    _dragDeltaX = 0;
  }

  void _handleHorizontalDragUpdate(DragUpdateDetails details) {
    _dragDeltaX += details.delta.dx;
  }

  void _handleHorizontalDragCancel() {
    _dragDeltaX = 0;
  }

  void _handleHorizontalDragEnd(
    DragEndDetails details, {
    required int itemCount,
  }) {
    final velocity = details.primaryVelocity ?? 0;
    final dragDelta = _dragDeltaX;
    _dragDeltaX = 0;
    if (itemCount <= 1) return;

    final swipedLeft =
        velocity <= -_carouselSwipeVelocityThreshold ||
        dragDelta <= -_carouselSwipeThreshold;
    final swipedRight =
        velocity >= _carouselSwipeVelocityThreshold ||
        dragDelta >= _carouselSwipeThreshold;

    if (swipedLeft) {
      _showNextItem(itemCount: itemCount);
    } else if (swipedRight) {
      _showPreviousItem(itemCount: itemCount);
    }
  }

  void _onNodeTap(QuestNode node) {
    _collapse();
    widget.onFocusNode(node);
  }

  void _onTurnInTap(Quest quest) {
    _collapse();
    widget.onFocusTurnInQuest?.call(quest);
  }

  void _onTutorialScenarioTap() {
    _collapse();
    widget.onFocusTutorialScenario?.call();
  }

  void _onTutorialMonsterTap() {
    _collapse();
    widget.onFocusTutorialMonster?.call();
  }

  @override
  Widget build(BuildContext context) {
    return Consumer2<QuestLogProvider, DiscoveriesProvider>(
      builder: (context, ql, discoveries, _) {
        final tutorialScenarioObjectiveCopy =
            widget.tutorialScenarioObjectiveCopy?.trim() ?? '';
        final tutorialMonsterObjectiveCopy =
            widget.tutorialMonsterObjectiveCopy?.trim() ?? '';
        final tutorialItems = <_TrackedQuestCarouselItem>[
          if (tutorialScenarioObjectiveCopy.isNotEmpty)
            _TrackedQuestCarouselItem(
              id: 'tutorial:scenario',
              child: KeyedSubtree(
                key: const ValueKey('tutorial-tracked-scenario'),
                child: _TutorialTrackedObjectiveCard(
                  title: widget.tutorialScenarioTitle ?? 'Tutorial Scenario',
                  detail: widget.tutorialScenarioDetail ?? '',
                  objectiveCopy: tutorialScenarioObjectiveCopy,
                  icon: Icons.auto_awesome_rounded,
                  onTap: widget.onFocusTutorialScenario == null
                      ? null
                      : _onTutorialScenarioTap,
                ),
              ),
              onPreview: widget.onPreviewTutorialScenario,
            ),
          if (tutorialMonsterObjectiveCopy.isNotEmpty)
            _TrackedQuestCarouselItem(
              id: 'tutorial:monster',
              child: KeyedSubtree(
                key: const ValueKey('tutorial-tracked-monster'),
                child: _TutorialTrackedObjectiveCard(
                  title: widget.tutorialMonsterTitle ?? 'Tutorial Monster',
                  detail: widget.tutorialMonsterDetail ?? '',
                  objectiveCopy: tutorialMonsterObjectiveCopy,
                  icon: Icons.gps_fixed_rounded,
                  onTap: widget.onFocusTutorialMonster == null
                      ? null
                      : _onTutorialMonsterTap,
                ),
              ),
              onPreview: widget.onPreviewTutorialMonster,
            ),
        ];
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
        final discoveredIds = <String>{
          for (final d in discoveries.discoveries) d.pointOfInterestId,
        };
        final acceptedTrackedQuestIds = <String>{
          for (final quest in visibleTracked)
            if (quest.isAccepted && quest.id.trim().isNotEmpty) quest.id.trim(),
        };
        if (_hasAcceptedTrackedQuestSnapshot) {
          final newlyAcceptedTrackedQuestIds = acceptedTrackedQuestIds
              .difference(_lastAcceptedTrackedQuestIds);
          if (newlyAcceptedTrackedQuestIds.isNotEmpty) {
            for (final quest in visibleTracked) {
              final questId = quest.id.trim();
              if (!newlyAcceptedTrackedQuestIds.contains(questId)) continue;
              _preferredOpenCarouselItemId = 'quest:$questId';
              break;
            }
          }
        }
        _lastAcceptedTrackedQuestIds = acceptedTrackedQuestIds;
        _hasAcceptedTrackedQuestSnapshot = true;
        final contentItems = <_TrackedQuestCarouselItem>[
          ...tutorialItems,
          ...visibleTracked.map(
            (quest) => _TrackedQuestCarouselItem(
              id: 'quest:${quest.id}',
              child: KeyedSubtree(
                key: ValueKey('tracked-quest-${quest.id}'),
                child: _TrackedQuestCard(
                  quest: quest,
                  discoveredIds: discoveredIds,
                  tutorialScenarioObjectiveCopy: tutorialScenarioObjectiveCopy,
                  tutorialMonsterObjectiveCopy: tutorialMonsterObjectiveCopy,
                  onPoITap: _onPoITap,
                  onNodeTap: _onNodeTap,
                  onTurnInTap: widget.onFocusTurnInQuest == null
                      ? null
                      : _onTurnInTap,
                  resolveQuestReceiverCharacter:
                      widget.resolveQuestReceiverCharacter,
                  resolveQuestReceiverPoi: widget.resolveQuestReceiverPoi,
                ),
              ),
              onPreview: _buildQuestPreviewCallback(quest),
            ),
          ),
        ];
        _carouselItems = contentItems;
        final itemCount = contentItems.length;
        final currentItemIndex = _resolvedDisplayItemIndex();
        final hasMultipleItems = itemCount > 1;
        _syncCurrentItemIndexIfNeeded(currentItemIndex);
        if (_previewCurrentItemAfterBuild && itemCount > 0) {
          _previewCurrentItemAfterBuild = false;
          _pendingOpenCarouselItemId = null;
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            _rememberCarouselItemAt(currentItemIndex);
            _previewCarouselItem(currentItemIndex);
          });
        }

        if (contentItems.isEmpty) {
          return const SizedBox.shrink();
        }

        final screenWidth = MediaQuery.sizeOf(context).width;
        const collapsedWidth = 96.0;
        const sideMargin = 16.0;
        final expandUpwards = widget.expandUpwards;
        final collapsedHeight = widget.collapsedHeight;
        final arrowIcon = _expanded
            ? (expandUpwards ? Icons.expand_more : Icons.expand_less)
            : (expandUpwards ? Icons.expand_less : Icons.expand_more);
        final maxAllowedWidth = (screenWidth - (sideMargin * 2)).clamp(
          collapsedWidth,
          288.0,
        );
        final width = _expanded ? maxAllowedWidth : collapsedWidth;
        final header = InkWell(
          onTap: _toggle,
          splashFactory: NoSplash.splashFactory,
          overlayColor: WidgetStateProperty.all(Colors.transparent),
          borderRadius: BorderRadius.circular(12),
          child: Padding(
            padding: EdgeInsets.fromLTRB(12, 12, _expanded ? 12 : 24, 12),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Quests',
                  style: Theme.of(
                    context,
                  ).textTheme.titleSmall?.copyWith(color: Colors.white),
                ),
                Icon(arrowIcon, color: Colors.white, size: 24),
              ],
            ),
          ),
        );
        final content = AnimatedSize(
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOutCubic,
          alignment: expandUpwards
              ? Alignment.bottomCenter
              : Alignment.topCenter,
          clipBehavior: Clip.hardEdge,
          child: _expanded && _showContent
              ? Padding(
                  padding: EdgeInsets.fromLTRB(
                    8,
                    expandUpwards ? 8 : 0,
                    8,
                    expandUpwards ? 0 : 8,
                  ),
                  child: GestureDetector(
                    behavior: HitTestBehavior.translucent,
                    onHorizontalDragStart: hasMultipleItems
                        ? _handleHorizontalDragStart
                        : null,
                    onHorizontalDragUpdate: hasMultipleItems
                        ? _handleHorizontalDragUpdate
                        : null,
                    onHorizontalDragCancel: hasMultipleItems
                        ? _handleHorizontalDragCancel
                        : null,
                    onHorizontalDragEnd: hasMultipleItems
                        ? (details) => _handleHorizontalDragEnd(
                            details,
                            itemCount: itemCount,
                          )
                        : null,
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.center,
                      children: [
                        if (hasMultipleItems)
                          _TrackedQuestCarouselChevron(
                            icon: Icons.chevron_left,
                            enabled: currentItemIndex > 0,
                            onTap: currentItemIndex > 0
                                ? () => _showPreviousItem(itemCount: itemCount)
                                : null,
                          ),
                        Expanded(
                          child: Padding(
                            padding: const EdgeInsets.symmetric(vertical: 4),
                            child: ClipRect(
                              child: AnimatedSwitcher(
                                duration: const Duration(milliseconds: 220),
                                switchInCurve: Curves.easeOutCubic,
                                switchOutCurve: Curves.easeOutCubic,
                                layoutBuilder:
                                    (currentChild, previousChildren) => Stack(
                                      alignment: Alignment.topCenter,
                                      children: [
                                        ...previousChildren,
                                        if (currentChild != null) currentChild,
                                      ],
                                    ),
                                transitionBuilder: (child, animation) {
                                  final curved = CurvedAnimation(
                                    parent: animation,
                                    curve: Curves.easeOutCubic,
                                  );
                                  final offsetTween = Tween<Offset>(
                                    begin: Offset(
                                      _carouselTransitionDirection * 0.18,
                                      0,
                                    ),
                                    end: Offset.zero,
                                  );
                                  return FadeTransition(
                                    opacity: animation,
                                    child: SlideTransition(
                                      position: offsetTween.animate(curved),
                                      child: child,
                                    ),
                                  );
                                },
                                child: contentItems[currentItemIndex].child,
                              ),
                            ),
                          ),
                        ),
                        if (hasMultipleItems)
                          _TrackedQuestCarouselChevron(
                            icon: Icons.chevron_right,
                            enabled: currentItemIndex < itemCount - 1,
                            onTap: currentItemIndex < itemCount - 1
                                ? () => _showNextItem(itemCount: itemCount)
                                : null,
                          ),
                      ],
                    ),
                  ),
                )
              : const SizedBox.shrink(),
        );

        return AnimatedContainer(
          duration: const Duration(milliseconds: 300),
          width: width,
          height: _expanded ? null : collapsedHeight,
          constraints: _expanded
              ? BoxConstraints(minHeight: collapsedHeight)
              : BoxConstraints.tightFor(height: collapsedHeight),
          child: Material(
            color: Colors.black54,
            borderRadius: BorderRadius.circular(12),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (expandUpwards) ...[
                  content,
                  header,
                ] else ...[
                  header,
                  content,
                ],
              ],
            ),
          ),
        );
      },
    );
  }
}

class _TrackedQuestCarouselItem {
  const _TrackedQuestCarouselItem({
    required this.id,
    required this.child,
    this.onPreview,
  });

  final String id;
  final Widget child;
  final VoidCallback? onPreview;
}

class _TrackedQuestCarouselChevron extends StatelessWidget {
  const _TrackedQuestCarouselChevron({
    required this.icon,
    required this.enabled,
    this.onTap,
  });

  final IconData icon;
  final bool enabled;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 28,
      child: AnimatedOpacity(
        duration: const Duration(milliseconds: 150),
        opacity: enabled ? 0.9 : 0.3,
        child: IconButton(
          onPressed: onTap,
          icon: Icon(icon, color: Colors.white),
          iconSize: 22,
          visualDensity: VisualDensity.compact,
          padding: EdgeInsets.zero,
          constraints: const BoxConstraints.tightFor(width: 24, height: 36),
        ),
      ),
    );
  }
}

class _TutorialTrackedObjectiveCard extends StatelessWidget {
  const _TutorialTrackedObjectiveCard({
    required this.title,
    required this.objectiveCopy,
    required this.icon,
    this.detail = '',
    this.onTap,
  });

  final String title;
  final String objectiveCopy;
  final String detail;
  final IconData icon;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final normalizedTitle = 'Current Objective';
    final normalizedDetail = detail.trim();

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        padding: const EdgeInsets.all(10),
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            colors: [Color(0xFF5D3E12), Color(0xFF2E1E0A)],
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
                'Tutorial Objective',
                style: Theme.of(context).textTheme.labelSmall?.copyWith(
                  color: const Color(0xFF3A1A11),
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
            const SizedBox(height: 10),
            _TrackedQuestTileFrame(
              leading: Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  color: const Color(0x33F1D597),
                  borderRadius: BorderRadius.circular(10),
                ),
                alignment: Alignment.center,
                child: Icon(icon, color: const Color(0xFFF1D597), size: 22),
              ),
              crossAxisAlignment: CrossAxisAlignment.start,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    normalizedTitle,
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Colors.white,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    objectiveCopy.trim(),
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: const Color(0xFFF8EBD0),
                      height: 1.35,
                    ),
                  ),
                  if (normalizedDetail.isNotEmpty) ...[
                    const SizedBox(height: 4),
                    Text(
                      normalizedDetail,
                      style: Theme.of(
                        context,
                      ).textTheme.bodySmall?.copyWith(color: Colors.white70),
                    ),
                  ],
                ],
              ),
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
    required this.tutorialScenarioObjectiveCopy,
    required this.tutorialMonsterObjectiveCopy,
    required this.onPoITap,
    required this.onNodeTap,
    this.onTurnInTap,
    this.resolveQuestReceiverCharacter,
    this.resolveQuestReceiverPoi,
  });

  final Quest quest;
  final Set<String> discoveredIds;
  final String tutorialScenarioObjectiveCopy;
  final String tutorialMonsterObjectiveCopy;
  final void Function(PointOfInterest) onPoITap;
  final void Function(QuestNode) onNodeTap;
  final void Function(Quest quest)? onTurnInTap;
  final Character? Function(Quest quest)? resolveQuestReceiverCharacter;
  final PointOfInterest? Function(Quest quest)? resolveQuestReceiverPoi;

  @override
  Widget build(BuildContext context) {
    final node = quest.currentNode;
    final poi = node?.pointOfInterest;
    final awaitingTurnIn = questIsAwaitingTurnIn(quest);
    final questReceiver = resolveQuestReceiverCharacter?.call(quest);
    final questReceiverPoi = resolveQuestReceiverPoi?.call(quest);
    final hasTutorialMonsterObjective =
        quest.isTutorial &&
        ((node?.monsterEncounterId?.trim().isNotEmpty ?? false) ||
            (node?.monsterId?.trim().isNotEmpty ?? false)) &&
        tutorialMonsterObjectiveCopy.isNotEmpty;
    final hasTutorialScenarioObjective =
        quest.isTutorial &&
        (node?.scenarioId?.trim().isNotEmpty ?? false) &&
        tutorialScenarioObjectiveCopy.isNotEmpty;
    final hasTutorialObjectiveOverride =
        hasTutorialMonsterObjective || hasTutorialScenarioObjective;
    final tutorialObjectiveCopy = hasTutorialMonsterObjective
        ? tutorialMonsterObjectiveCopy
        : hasTutorialScenarioObjective
        ? tutorialScenarioObjectiveCopy
        : '';
    final objectiveLines = hasTutorialObjectiveOverride
        ? <String>[tutorialObjectiveCopy]
        : questObjectiveLines(node);
    final objectiveTitle = hasTutorialObjectiveOverride
        ? 'Current Objective'
        : poi?.name ?? _trackedQuestObjectiveTitle(node);
    final hasDirectFocusTarget = questNodeHasDirectFocusTarget(node);
    final VoidCallback? cardTap;
    if (awaitingTurnIn) {
      cardTap =
          onTurnInTap == null ||
              (questReceiver == null && questReceiverPoi == null)
          ? null
          : () => onTurnInTap!(quest);
    } else if (poi != null) {
      cardTap = () => onPoITap(poi);
    } else if (hasDirectFocusTarget) {
      cardTap = () => onNodeTap(node!);
    } else {
      cardTap = null;
    }

    return Ink(
      decoration: BoxDecoration(
        color: Colors.black26,
        borderRadius: BorderRadius.circular(8),
      ),
      child: InkWell(
        onTap: cardTap,
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      quest.isTutorial ? 'Current Objective' : quest.name,
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
                )
              else if (poi != null)
                _TrackedQuestObjectiveTile(
                  node: node!,
                  title: objectiveTitle,
                  discoveredIds: discoveredIds,
                  objectiveLines: objectiveLines,
                )
              else
                _TrackedQuestObjectiveTile(
                  node: node,
                  title: objectiveTitle,
                  discoveredIds: discoveredIds,
                  objectiveLines: objectiveLines,
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _TrackedQuestTurnInTile extends StatelessWidget {
  const _TrackedQuestTurnInTile({
    required this.questReceiver,
    required this.pointOfInterest,
  });

  final Character? questReceiver;
  final PointOfInterest? pointOfInterest;

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

    return _TrackedQuestTileFrame(
      leading: QuestTurnInPortrait(
        character: questReceiver,
        size: 40,
        backgroundColor: const Color(0x33F1D597),
        foregroundColor: const Color(0xFFF1D597),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
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
    );
  }
}

class _TrackedQuestObjectiveTile extends StatelessWidget {
  const _TrackedQuestObjectiveTile({
    required this.node,
    required this.title,
    required this.discoveredIds,
    required this.objectiveLines,
  });

  final QuestNode? node;
  final String title;
  final Set<String> discoveredIds;
  final List<String> objectiveLines;

  @override
  Widget build(BuildContext context) {
    final challengeLabel = questObjectiveChallengeLabel(node);
    final normalizedTitle = title.trim();
    final detailLines = objectiveLines
        .where((line) => line.trim().isNotEmpty)
        .where((line) => line.trim() != normalizedTitle)
        .toList();

    return _TrackedQuestObjectivePulseFrame(
      child: _TrackedQuestTileFrame(
        crossAxisAlignment: CrossAxisAlignment.start,
        leading: QuestObjectiveIcon(
          node: node,
          discoveredPoiIds: discoveredIds,
          size: 40,
          borderRadius: 10,
          iconColor: const Color(0xFFF1D597),
          backgroundColor: const Color(0x33F1D597),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              normalizedTitle,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: Colors.white,
                fontWeight: FontWeight.w700,
              ),
            ),
            if (challengeLabel != null) ...[
              const SizedBox(height: 6),
              QuestObjectiveChallengeBadge(node: node),
            ],
            if (detailLines.isNotEmpty) ...[
              const SizedBox(height: 4),
              ...detailLines.map(
                (line) => Padding(
                  padding: const EdgeInsets.only(top: 2, bottom: 2),
                  child: Text(
                    line,
                    style: Theme.of(
                      context,
                    ).textTheme.bodySmall?.copyWith(color: Colors.white70),
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _TrackedQuestTileFrame extends StatelessWidget {
  const _TrackedQuestTileFrame({
    required this.leading,
    required this.child,
    this.crossAxisAlignment = CrossAxisAlignment.center,
  });

  final Widget leading;
  final Widget child;
  final CrossAxisAlignment crossAxisAlignment;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: const Color(0x1FFFFFFF),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: const Color(0x33F1D597)),
      ),
      child: Row(
        crossAxisAlignment: crossAxisAlignment,
        children: [
          leading,
          const SizedBox(width: 10),
          Expanded(child: child),
        ],
      ),
    );
  }
}

class _TrackedQuestObjectivePulseFrame extends StatefulWidget {
  const _TrackedQuestObjectivePulseFrame({required this.child});

  final Widget child;

  @override
  State<_TrackedQuestObjectivePulseFrame> createState() =>
      _TrackedQuestObjectivePulseFrameState();
}

class _TrackedQuestObjectivePulseFrameState
    extends State<_TrackedQuestObjectivePulseFrame>
    with SingleTickerProviderStateMixin {
  static const _pulseCoreColor = Color(0xFFB53A4B);
  static const _pulseMistColor = Color(0xFFF3CBD2);
  static const _pulseRingColor = Color(0xFF7A1823);

  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1800),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Widget _buildPulseLayer(double progress) {
    final curved = Curves.easeOutCubic.transform(progress.clamp(0.0, 1.0));
    final scale = 1.0 + (0.045 * curved);
    final opacity = (1.0 - curved).clamp(0.0, 1.0);
    return Transform.scale(
      scale: scale,
      child: DecoratedBox(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(10),
          color: _pulseCoreColor.withValues(alpha: 0.05 * opacity),
          border: Border.all(
            color: _pulseRingColor.withValues(alpha: 0.68 * opacity),
            width: 1.4,
          ),
          boxShadow: [
            BoxShadow(
              color: _pulseCoreColor.withValues(alpha: 0.2 * opacity),
              blurRadius: 18 + (curved * 14),
              spreadRadius: 1.2 + (curved * 2.6),
            ),
            BoxShadow(
              color: _pulseMistColor.withValues(alpha: 0.14 * opacity),
              blurRadius: 22 + (curved * 16),
              spreadRadius: 0.8 + (curved * 1.8),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      child: widget.child,
      builder: (context, child) {
        final primaryProgress = _controller.value.clamp(0.0, 1.0);
        final secondaryProgress = (primaryProgress + 0.5) % 1.0;
        return Stack(
          clipBehavior: Clip.none,
          children: [
            Positioned.fill(
              child: IgnorePointer(child: _buildPulseLayer(secondaryProgress)),
            ),
            Positioned.fill(
              child: IgnorePointer(child: _buildPulseLayer(primaryProgress)),
            ),
            if (child != null) child,
          ],
        );
      },
    );
  }
}

String _trackedQuestObjectiveTitle(QuestNode? node) {
  final fetchCharacterName = node?.fetchCharacter?.name.trim() ?? '';
  if (fetchCharacterName.isNotEmpty) {
    return 'Deliver to $fetchCharacterName';
  }

  final expositionTitle = node?.exposition?.title.trim() ?? '';
  if (expositionTitle.isNotEmpty) {
    return expositionTitle;
  }

  if (node?.challengeId?.trim().isNotEmpty ?? false) {
    return 'Challenge objective';
  }

  final hasMonsterObjective =
      (node?.monsterEncounterId?.trim().isNotEmpty ?? false) ||
      (node?.monsterId?.trim().isNotEmpty ?? false);
  if (hasMonsterObjective) {
    return 'Monster encounter';
  }

  if (node?.scenarioId?.trim().isNotEmpty ?? false) {
    return 'Scenario objective';
  }

  if (node?.polygon.isNotEmpty ?? false) {
    return 'Quest area';
  }

  return 'Current objective';
}
