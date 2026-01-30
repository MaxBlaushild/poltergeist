import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/activity_feed.dart';
import '../providers/activity_feed_provider.dart';

class ActivityFeedPanel extends StatelessWidget {
  const ActivityFeedPanel({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<ActivityFeedProvider>(
      builder: (context, feed, _) {
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
            ListView.builder(
              shrinkWrap: true,
              itemCount: list.length,
              itemBuilder: (_, i) {
                final a = list[i];
                return ListTile(
                  title: Text(_titleFor(a)),
                  subtitle: Text(a.activityType),
                  trailing: a.seen ? null : Icon(Icons.circle, size: 8, color: Theme.of(context).colorScheme.primary),
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
