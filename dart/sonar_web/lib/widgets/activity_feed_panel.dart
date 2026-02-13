import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/activity_feed.dart';
import '../providers/activity_feed_provider.dart';

class ActivityFeedPanel extends StatefulWidget {
  const ActivityFeedPanel({super.key});

  @override
  State<ActivityFeedPanel> createState() => _ActivityFeedPanelState();
}

class _ActivityFeedPanelState extends State<ActivityFeedPanel> {
  final Set<String> _markedIds = {};
  bool _marking = false;

  Map<String, dynamic> _entitiesFor(ActivityFeed a) {
    final data = a.data;
    if (data is Map && data['entities'] is Map<String, dynamic>) {
      return data['entities'] as Map<String, dynamic>;
    }
    if (data is Map && data['entities'] is Map) {
      return Map<String, dynamic>.from(data['entities'] as Map);
    }
    return const {};
  }

  String _stringField(Map<String, dynamic> source, String key) {
    final value = source[key];
    if (value == null) return '';
    return value.toString();
  }

  List<Widget> _detailLines(BuildContext context, ActivityFeed a) {
    final entities = _entitiesFor(a);
    final style = Theme.of(context).textTheme.bodySmall?.copyWith(
          color: Colors.grey.shade700,
        );

    final lines = <String>[];

    switch (a.activityType) {
      case 'quest_completed':
        final quest = entities['quest'];
        if (quest is Map) {
          final name = _stringField(Map<String, dynamic>.from(quest), 'name');
          if (name.isNotEmpty) {
            lines.add('Quest: $name');
          }
        }
        break;
      case 'challenge_completed':
        final quest = entities['quest'];
        if (quest is Map) {
          final name = _stringField(Map<String, dynamic>.from(quest), 'name');
          if (name.isNotEmpty) {
            lines.add('Quest: $name');
          }
        }
        final challenge = entities['challenge'];
        if (challenge is Map) {
          final question = _stringField(Map<String, dynamic>.from(challenge), 'question');
          if (question.isNotEmpty) {
            lines.add('Challenge: $question');
          }
        }
        final zone = entities['zone'];
        if (zone is Map) {
          final name = _stringField(Map<String, dynamic>.from(zone), 'name');
          if (name.isNotEmpty) {
            lines.add('Zone: $name');
          }
        }
        final currentPoi = entities['currentPoi'];
        if (currentPoi is Map) {
          final name = _stringField(Map<String, dynamic>.from(currentPoi), 'name');
          if (name.isNotEmpty) {
            lines.add('Current POI: $name');
          }
        }
        final nextPoi = entities['nextPoi'];
        if (nextPoi is Map) {
          final name = _stringField(Map<String, dynamic>.from(nextPoi), 'name');
          if (name.isNotEmpty) {
            lines.add('Next POI: $name');
          }
        }
        break;
      case 'item_received':
        final item = entities['item'];
        if (item is Map) {
          final name = _stringField(Map<String, dynamic>.from(item), 'name');
          if (name.isNotEmpty) {
            lines.add('Item: $name');
          }
        }
        break;
      case 'reputation_up':
        final zone = entities['zone'];
        if (zone is Map) {
          final name = _stringField(Map<String, dynamic>.from(zone), 'name');
          if (name.isNotEmpty) {
            lines.add('Zone: $name');
          }
        }
        break;
      case 'level_up':
        final level = entities['level'];
        if (level is Map) {
          final newLevel = _stringField(Map<String, dynamic>.from(level), 'newLevel');
          if (newLevel.isNotEmpty) {
            lines.add('New Level: $newLevel');
          }
        }
        break;
    }

    return lines
        .map(
          (line) => Padding(
            padding: const EdgeInsets.only(top: 4),
            child: Text(line, style: style),
          ),
        )
        .toList();
  }

  String _formatTimestamp(String raw) {
    if (raw.isEmpty) return '';
    final parsed = DateTime.tryParse(raw);
    if (parsed == null) return '';
    final local = parsed.toLocal();
    final hh = local.hour % 12 == 0 ? 12 : local.hour % 12;
    final mm = local.minute.toString().padLeft(2, '0');
    final suffix = local.hour >= 12 ? 'PM' : 'AM';
    return '${local.month}/${local.day}/${local.year} $hh:$mm $suffix';
  }

  void _markVisibleUnseen(ActivityFeedProvider feed) {
    if (_marking) return;
    final unseenIds = feed.unseenActivities
        .map((a) => a.id)
        .where((id) => !_markedIds.contains(id))
        .toList();
    if (unseenIds.isEmpty) return;
    _marking = true;
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      await feed.markAsSeen(unseenIds);
      if (!mounted) return;
      setState(() {
        _markedIds.addAll(unseenIds);
        _marking = false;
      });
    });
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<ActivityFeedProvider>(
      builder: (context, feed, _) {
        _markVisibleUnseen(feed);
        if (feed.loading && feed.activities.isEmpty) {
          return const Center(child: CircularProgressIndicator());
        }
        final list = feed.activities;
        if (list.isEmpty) {
          return const Padding(
            padding: EdgeInsets.all(24),
            child: Center(child: Text('No activities yet')),
          );
        }
        final unseen = feed.unseenActivities;
        return Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (unseen.isNotEmpty)
              Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Align(
                  alignment: Alignment.centerRight,
                  child: TextButton(
                    onPressed: () => feed.markAsSeen(unseen.map((a) => a.id).toList()),
                    child: const Text('Mark all as seen'),
                  ),
                ),
              ),
            ListView.separated(
              shrinkWrap: true,
              itemCount: list.length,
              separatorBuilder: (_, __) => const SizedBox(height: 10),
              itemBuilder: (_, i) {
                final a = list[i];
                final timestamp = _formatTimestamp(a.createdAt);
                return Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  decoration: BoxDecoration(
                    color: a.seen ? Colors.grey.shade50 : Colors.white,
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: a.seen ? Colors.grey.shade200 : Colors.red.shade100,
                    ),
                    boxShadow: const [
                      BoxShadow(
                        color: Color(0x0D000000),
                        blurRadius: 6,
                        offset: Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Container(
                        margin: const EdgeInsets.only(top: 2),
                        width: 8,
                        height: 8,
                        decoration: BoxDecoration(
                          color: a.seen ? Colors.transparent : Colors.red.shade600,
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _titleFor(a),
                              style: Theme.of(context).textTheme.titleSmall?.copyWith(
                                    fontWeight: FontWeight.w600,
                                  ),
                            ),
                            ..._detailLines(context, a),
                            if (timestamp.isNotEmpty) ...[
                              const SizedBox(height: 6),
                              Text(
                                timestamp,
                                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                      color: Colors.grey.shade600,
                                    ),
                              ),
                            ],
                          ],
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ],
        );
      },
    );
  }

  String _titleFor(ActivityFeed a) {
    switch (a.activityType) {
      case 'level_up':
        return 'Level up!';
      case 'challenge_completed':
        return 'Challenge completed';
      case 'quest_completed':
        return 'Quest completed';
      case 'item_received':
        return 'Item received';
      case 'reputation_up':
        return 'Reputation up!';
      default:
        return a.activityType;
    }
  }
}
