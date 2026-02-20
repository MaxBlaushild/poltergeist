import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/point_of_interest.dart';
import '../models/quest.dart';
import '../models/quest_node.dart';
import '../providers/discoveries_provider.dart';
import '../providers/quest_log_provider.dart';

const _placeholderImageUrl =
    'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp';

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
    this.controller,
  });

  /// When user taps a POI: fly to location then open POI panel.
  final void Function(PointOfInterest poi) onFocusPoI;
  final void Function(QuestNode node) onFocusNode;
  final TrackedQuestsOverlayController? controller;

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

  @override
  Widget build(BuildContext context) {
    return Consumer2<QuestLogProvider, DiscoveriesProvider>(
      builder: (context, ql, discoveries, _) {
        final tracked = ql.quests
            .where((q) => ql.trackedQuestIds.contains(q.id))
            .toList();
        if (tracked.isNotEmpty) {
          _cachedTracked = List<Quest>.from(tracked);
        } else if (ql.trackedQuestIds.isEmpty) {
          _cachedTracked = const [];
        } else if (_cachedTracked.isNotEmpty) {
          _cachedTracked = _cachedTracked
              .where((q) => ql.trackedQuestIds.contains(q.id))
              .toList();
        }
        final visibleTracked =
            tracked.isNotEmpty ? tracked : _cachedTracked;
        final discoveredIds = <String>{
          for (final d in discoveries.discoveries) d.pointOfInterestId
        };

        if (visibleTracked.isEmpty) return const SizedBox.shrink();

        final screenWidth = MediaQuery.sizeOf(context).width;
        const rightMargin = 16.0;
        const minSideMargin = 16.0;
        final maxAllowedWidth = (screenWidth - rightMargin - minSideMargin)
            .clamp(80.0, 288.0);
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
                          style: Theme.of(context)
                              .textTheme
                              .titleSmall
                              ?.copyWith(color: Colors.white),
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
                                children: visibleTracked
                                    .map(
                                      (quest) => Padding(
                                        padding: const EdgeInsets.only(bottom: 8),
                                        child: _TrackedQuestCard(
                                          quest: quest,
                                          discoveredIds: discoveredIds,
                                          onPoITap: _onPoITap,
                                          onNodeTap: widget.onFocusNode,
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
    required this.onNodeTap,
  });

  final Quest quest;
  final Set<String> discoveredIds;
  final void Function(PointOfInterest) onPoITap;
  final void Function(QuestNode) onNodeTap;

  @override
  Widget build(BuildContext context) {
    final node = quest.currentNode;
    final poi = node?.pointOfInterest;

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
          if (node == null)
            const Text(
              'Quest completed! Turn it in for rewards.',
              style: TextStyle(color: Colors.white70),
            )
          else if (poi != null)
            _QuestPoiTile(
              poi: poi,
              discovered: discoveredIds.contains(poi.id),
              onTap: () => onPoITap(poi),
              onChallengeTap: () => onNodeTap(node),
              challenges: node.challenges.map((c) => c.question).toList(),
            )
          else
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Padding(
                  padding: const EdgeInsets.only(bottom: 4),
                  child: Text(
                    'Reach the highlighted area to submit your answer.',
                    style: Theme.of(context)
                        .textTheme
                        .bodySmall
                        ?.copyWith(color: Colors.white70),
                  ),
                ),
                ...node.challenges.map(
                  (q) => GestureDetector(
                    behavior: HitTestBehavior.opaque,
                    onTap: () => onNodeTap(node),
                    child: Padding(
                      padding: const EdgeInsets.symmetric(vertical: 2),
                      child: Text(
                        q.question,
                        style: Theme.of(context)
                            .textTheme
                            .bodySmall
                            ?.copyWith(color: Colors.white70),
                      ),
                    ),
                  ),
                ),
              ],
            ),
        ],
      ),
    );
  }
}

class _QuestPoiTile extends StatelessWidget {
  const _QuestPoiTile({
    required this.poi,
    required this.discovered,
    required this.onTap,
    required this.onChallengeTap,
    required this.challenges,
  });

  final PointOfInterest poi;
  final bool discovered;
  final VoidCallback onTap;
  final VoidCallback onChallengeTap;
  final List<String> challenges;

  @override
  Widget build(BuildContext context) {
    final imageUrl = discovered &&
            (poi.imageURL != null && poi.imageURL!.isNotEmpty)
        ? poi.imageURL!
        : _placeholderImageUrl;

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(6),
      child: Padding(
        padding: const EdgeInsets.only(left: 4, top: 4, bottom: 4, right: 8),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: Image.network(
                imageUrl,
                width: 28,
                height: 28,
                fit: BoxFit.cover,
                errorBuilder: (_, __, ___) => Container(
                  width: 28,
                  height: 28,
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
                          fontWeight: FontWeight.w600,
                        ),
                  ),
                  ...challenges.map(
                    (q) => GestureDetector(
                      behavior: HitTestBehavior.opaque,
                      onTap: onChallengeTap,
                      child: Padding(
                        padding: const EdgeInsets.symmetric(vertical: 2),
                        child: Text(
                          q,
                          style: Theme.of(context)
                              .textTheme
                              .bodySmall
                              ?.copyWith(color: Colors.white70),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const Icon(Icons.chevron_right, size: 16, color: Colors.white70),
          ],
        ),
      ),
    );
  }
}
