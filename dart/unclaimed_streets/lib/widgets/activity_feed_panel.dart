import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../models/activity_feed.dart';
import '../providers/activity_feed_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/party_provider.dart';

class ActivityFeedPanel extends StatefulWidget {
  const ActivityFeedPanel({super.key});

  @override
  State<ActivityFeedPanel> createState() => _ActivityFeedPanelState();
}

class _ActivityFeedPanelState extends State<ActivityFeedPanel> {
  final Set<String> _markedIds = {};
  final Set<String> _respondingInviteIds = {};
  bool _marking = false;

  Map<String, dynamic> _entitiesFor(ActivityFeed a) {
    final entities = a.data['entities'];
    if (entities is Map<String, dynamic>) return entities;
    if (entities is Map) return Map<String, dynamic>.from(entities);
    return const {};
  }

  Map<String, dynamic> _mapField(Map<String, dynamic> source, String key) {
    final value = source[key];
    if (value is Map<String, dynamic>) return value;
    if (value is Map) return Map<String, dynamic>.from(value);
    return const {};
  }

  List<Map<String, dynamic>> _mapListField(
    Map<String, dynamic> source,
    String key,
  ) {
    final value = source[key];
    if (value is! List) return const [];
    return value
        .whereType<Map>()
        .map((entry) => Map<String, dynamic>.from(entry))
        .toList();
  }

  String _stringField(Map<String, dynamic> source, String key) {
    final value = source[key];
    if (value == null) return '';
    return value.toString().trim();
  }

  int _intField(Map<String, dynamic> source, String key) {
    final value = source[key];
    if (value is int) return value;
    if (value is num) return value.toInt();
    return int.tryParse(value?.toString() ?? '') ?? 0;
  }

  bool _boolField(Map<String, dynamic> source, String key) {
    final value = source[key];
    if (value is bool) return value;
    final raw = value?.toString().toLowerCase().trim();
    return raw == 'true' || raw == '1' || raw == 'yes';
  }

  String _firstNonEmpty(Iterable<String> values) {
    for (final value in values) {
      if (value.trim().isNotEmpty) return value.trim();
    }
    return '';
  }

  int _itemRewardCount(List<Map<String, dynamic>> items) {
    var count = 0;
    for (final item in items) {
      final quantity = _intField(item, 'quantity');
      count += quantity > 0 ? quantity : 1;
    }
    return count;
  }

  String _countLabel(int count, String singular, String plural) {
    if (count <= 0) return '';
    return '$count ${count == 1 ? singular : plural}';
  }

  String _naturalJoin(List<String> parts) {
    final cleaned = parts.where((part) => part.trim().isNotEmpty).toList();
    if (cleaned.isEmpty) return '';
    if (cleaned.length == 1) return cleaned.first;
    if (cleaned.length == 2) return '${cleaned.first} and ${cleaned.last}';
    return '${cleaned.sublist(0, cleaned.length - 1).join(', ')}, and ${cleaned.last}';
  }

  _ActivityVisualStyle _styleFor(ActivityFeed a) {
    switch (a.activityType) {
      case 'level_up':
        return const _ActivityVisualStyle(
          accent: Color(0xFFC17A12),
          background: Color(0xFFFFF5DD),
        );
      case 'challenge_completed':
        return const _ActivityVisualStyle(
          accent: Color(0xFF157A6E),
          background: Color(0xFFEAF7F3),
        );
      case 'quest_completed':
        return const _ActivityVisualStyle(
          accent: Color(0xFF8B3C2C),
          background: Color(0xFFF9EBDD),
        );
      case 'item_received':
        return const _ActivityVisualStyle(
          accent: Color(0xFF356C9B),
          background: Color(0xFFEAF3FD),
        );
      case 'reputation_up':
        return const _ActivityVisualStyle(
          accent: Color(0xFF6E4E9D),
          background: Color(0xFFF1ECFB),
        );
      case 'monster_battle_invite':
        return const _ActivityVisualStyle(
          accent: Color(0xFF9B3D3D),
          background: Color(0xFFFBECEC),
        );
      default:
        return const _ActivityVisualStyle(
          accent: Color(0xFF5C6B73),
          background: Color(0xFFF3F6F8),
        );
    }
  }

  IconData _iconFor(ActivityFeed a) {
    switch (a.activityType) {
      case 'level_up':
        return Icons.workspace_premium_rounded;
      case 'challenge_completed':
        return Icons.bolt_rounded;
      case 'quest_completed':
        return Icons.flag_rounded;
      case 'item_received':
        return Icons.inventory_2_rounded;
      case 'reputation_up':
        return Icons.military_tech_rounded;
      case 'monster_battle_invite':
        return Icons.shield_rounded;
      default:
        return Icons.notifications_rounded;
    }
  }

  String _summaryFor(ActivityFeed a) {
    final entities = _entitiesFor(a);
    switch (a.activityType) {
      case 'level_up':
        final newLevel = _stringField(_mapField(entities, 'level'), 'newLevel');
        return newLevel.isNotEmpty
            ? 'You reached level $newLevel.'
            : 'You leveled up.';
      case 'challenge_completed':
        final question = _stringField(
          _mapField(entities, 'challenge'),
          'question',
        );
        final poi = _stringField(_mapField(entities, 'currentPoi'), 'name');
        final zone = _stringField(_mapField(entities, 'zone'), 'name');
        final quest = _stringField(_mapField(entities, 'quest'), 'name');
        final xp = _intField(a.data, 'experienceAwarded');
        final gold = _intField(a.data, 'goldAwarded');
        final itemCount = _itemRewardCount(
          _mapListField(a.data, 'itemsAwarded'),
        );
        final rewards = <String>[
          if (xp > 0) '$xp XP',
          if (gold > 0) '$gold gold',
          if (itemCount > 0)
            _countLabel(itemCount, 'item reward', 'item rewards'),
        ];
        final subject = question.isNotEmpty
            ? 'You solved "$question"'
            : 'You completed a challenge';
        final location = _firstNonEmpty([poi, zone]);
        if (location.isNotEmpty && rewards.isNotEmpty) {
          return '$subject at $location and earned ${_naturalJoin(rewards)}.';
        }
        if (location.isNotEmpty) {
          return '$subject at $location.';
        }
        if (_boolField(a.data, 'questCompleted') && quest.isNotEmpty) {
          return '$subject and wrapped up $quest.';
        }
        if (rewards.isNotEmpty) {
          return '$subject and earned ${_naturalJoin(rewards)}.';
        }
        return '$subject.';
      case 'quest_completed':
        final quest = _stringField(_mapField(entities, 'quest'), 'name');
        final gold = _intField(a.data, 'goldAwarded');
        final itemCount = _itemRewardCount(
          _mapListField(a.data, 'itemsAwarded'),
        );
        final spellCount = _mapListField(a.data, 'spellsAwarded').length;
        final rewards = <String>[
          if (gold > 0) '$gold gold',
          if (itemCount > 0)
            _countLabel(itemCount, 'item reward', 'item rewards'),
          if (spellCount > 0) _countLabel(spellCount, 'spell', 'spells'),
        ];
        final subject = quest.isNotEmpty
            ? 'You completed $quest'
            : 'You completed a quest';
        if (rewards.isNotEmpty) {
          return '$subject and collected ${_naturalJoin(rewards)}.';
        }
        return '$subject.';
      case 'item_received':
        final item = _stringField(_mapField(entities, 'item'), 'name');
        return item.isNotEmpty
            ? 'You received $item.'
            : 'You received a new item.';
      case 'reputation_up':
        final zone = _stringField(_mapField(entities, 'zone'), 'name');
        final newLevel = _intField(a.data, 'newLevel');
        if (zone.isNotEmpty && newLevel > 0) {
          return 'Your reputation in $zone reached level $newLevel.';
        }
        if (zone.isNotEmpty) {
          return 'Your reputation improved in $zone.';
        }
        return 'Your reputation increased.';
      case 'monster_battle_invite':
        final inviter = _stringField(a.data, 'inviterName');
        final monster = _stringField(a.data, 'monsterName');
        if (inviter.isNotEmpty && monster.isNotEmpty) {
          return '$inviter invited you to join the fight against $monster.';
        }
        if (monster.isNotEmpty) {
          return 'You were invited to join the fight against $monster.';
        }
        return 'You were invited to join a party battle.';
      default:
        return a.activityType;
    }
  }

  String _imageUrlFor(ActivityFeed a, {String currentUserProfileUrl = ''}) {
    final entities = _entitiesFor(a);
    switch (a.activityType) {
      case 'level_up':
        return currentUserProfileUrl.trim();
      case 'challenge_completed':
        return _firstNonEmpty([
          _stringField(_mapField(entities, 'currentPoi'), 'imageUrl'),
          _stringField(_mapField(entities, 'currentPoi'), 'imageURL'),
          _stringField(_mapField(entities, 'quest'), 'imageUrl'),
          _stringField(_mapField(entities, 'quest'), 'imageURL'),
          _stringField(_mapField(a.data, 'currentPOI'), 'imageUrl'),
          _stringField(_mapField(a.data, 'currentPOI'), 'imageURL'),
          _stringField(_mapField(entities, 'nextPoi'), 'imageUrl'),
          _stringField(_mapField(entities, 'nextPoi'), 'imageURL'),
          currentUserProfileUrl,
        ]);
      case 'quest_completed':
        return _firstNonEmpty([
          _stringField(_mapField(entities, 'quest'), 'imageUrl'),
          _stringField(_mapField(entities, 'quest'), 'imageURL'),
          currentUserProfileUrl,
        ]);
      case 'item_received':
        return _firstNonEmpty([
          _stringField(_mapField(entities, 'item'), 'imageUrl'),
          _stringField(_mapField(entities, 'item'), 'imageURL'),
          currentUserProfileUrl,
        ]);
      case 'reputation_up':
        return currentUserProfileUrl.trim();
      case 'monster_battle_invite':
        return _firstNonEmpty([
          _stringField(_mapField(entities, 'inviter'), 'profilePictureUrl'),
          _stringField(a.data, 'inviterProfilePictureUrl'),
          _stringField(_mapField(entities, 'monster'), 'imageUrl'),
          _stringField(a.data, 'monsterImageUrl'),
        ]);
      default:
        return '';
    }
  }

  String _subtitleFor(ActivityFeed a) {
    final parts = <String>[];
    final timestamp = _timestampLabel(a.createdAt);
    if (timestamp.isNotEmpty) {
      parts.add(timestamp);
    }
    if (a.activityType == 'monster_battle_invite') {
      final expires = _formatRemainingTime(_stringField(a.data, 'expiresAt'));
      if (expires.isNotEmpty) {
        parts.add(expires);
      }
    }
    return parts.join(' · ');
  }

  Widget _buildActivityCard(
    BuildContext context,
    ActivityFeed a, {
    required String currentUserProfileUrl,
  }) {
    final style = _styleFor(a);
    final summary = _summaryFor(a);
    final subtitle = _subtitleFor(a);
    final imageUrl = _imageUrlFor(
      a,
      currentUserProfileUrl: currentUserProfileUrl,
    );

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: a.seen ? Colors.white : style.background.withValues(alpha: 0.6),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: a.seen
              ? Colors.black.withValues(alpha: 0.06)
              : style.accent.withValues(alpha: 0.18),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _ActivityThumbnail(
            imageUrl: imageUrl,
            icon: _iconFor(a),
            style: style,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  summary,
                  maxLines: 3,
                  overflow: TextOverflow.ellipsis,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey.shade900,
                    fontWeight: FontWeight.w600,
                    height: 1.35,
                  ),
                ),
                if (subtitle.isNotEmpty) ...[
                  const SizedBox(height: 6),
                  Row(
                    children: [
                      if (!a.seen) ...[
                        Container(
                          width: 7,
                          height: 7,
                          decoration: BoxDecoration(
                            color: style.accent,
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 6),
                      ],
                      Expanded(
                        child: Text(
                          subtitle,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: Theme.of(context).textTheme.bodySmall
                              ?.copyWith(color: Colors.grey.shade600),
                        ),
                      ),
                    ],
                  ),
                ],
                if (a.activityType == 'monster_battle_invite') ...[
                  const SizedBox(height: 10),
                  _monsterBattleInviteActions(context, a),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _monsterBattleInviteActions(BuildContext context, ActivityFeed a) {
    final inviteId = _stringField(a.data, 'inviteId').trim();
    final monsterId = _stringField(a.data, 'monsterId').trim();
    final battleId = _stringField(a.data, 'battleId').trim();
    if (inviteId.isEmpty) return const SizedBox.shrink();

    final expiresAtRaw = _stringField(a.data, 'expiresAt').trim();
    final expiresAt = DateTime.tryParse(expiresAtRaw);
    final expired =
        expiresAt != null && DateTime.now().isAfter(expiresAt.toLocal());
    final busy = _respondingInviteIds.contains(inviteId);

    if (expired) {
      return Text(
        'This invite has expired.',
        style: Theme.of(
          context,
        ).textTheme.bodySmall?.copyWith(color: Colors.grey.shade600),
      );
    }

    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        OutlinedButton(
          onPressed: busy
              ? null
              : () => _respondToMonsterBattleInvite(
                  inviteId,
                  monsterId: monsterId,
                  battleId: battleId,
                  accept: false,
                ),
          child: const Text('Decline'),
        ),
        FilledButton(
          onPressed: busy
              ? null
              : () => _respondToMonsterBattleInvite(
                  inviteId,
                  monsterId: monsterId,
                  battleId: battleId,
                  accept: true,
                ),
          child: const Text('Join fight'),
        ),
        if (busy)
          const SizedBox(
            width: 18,
            height: 18,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
      ],
    );
  }

  Future<void> _respondToMonsterBattleInvite(
    String inviteId, {
    required String monsterId,
    required String battleId,
    required bool accept,
  }) async {
    setState(() {
      _respondingInviteIds.add(inviteId);
    });
    try {
      final partyProvider = context.read<PartyProvider>();
      if (accept) {
        final response = await partyProvider.acceptMonsterBattleInvite(
          inviteId,
        );
        final resolvedBattleId = _stringFromBattleDetail(response, 'id').trim();
        final resolvedMonsterId = _stringFromBattleDetail(
          response,
          'monsterId',
        ).trim();
        final trimmedMonsterId = resolvedMonsterId.isNotEmpty
            ? resolvedMonsterId
            : monsterId.trim();
        final trimmedBattleId = resolvedBattleId.isNotEmpty
            ? resolvedBattleId
            : battleId.trim();
        if (trimmedMonsterId.isNotEmpty && mounted) {
          final targetUri = Uri(
            path: '/single-player',
            queryParameters: {
              'joinMonsterId': trimmedMonsterId,
              'partyBattle': '1',
              'inviteId': inviteId,
              if (trimmedBattleId.isNotEmpty) 'battleId': trimmedBattleId,
            },
          );
          context.go(targetUri.toString());
        }
      } else {
        await partyProvider.rejectMonsterBattleInvite(inviteId);
      }
      if (!mounted) return;
      await context.read<ActivityFeedProvider>().refresh();
    } catch (_) {
      // Keep failures non-blocking in feed UI.
    } finally {
      if (mounted) {
        setState(() {
          _respondingInviteIds.remove(inviteId);
        });
      }
    }
  }

  String _timestampLabel(String raw) {
    if (raw.isEmpty) return '';
    final parsed = DateTime.tryParse(raw);
    if (parsed == null) return '';
    final relative = _formatRelativeTime(parsed);
    if (relative.isNotEmpty) return relative;
    return _formatTimestamp(parsed);
  }

  String _formatTimestamp(DateTime parsed) {
    final local = parsed.toLocal();
    final month = _monthLabel(local.month);
    final hour = local.hour % 12 == 0 ? 12 : local.hour % 12;
    final minute = local.minute.toString().padLeft(2, '0');
    final suffix = local.hour >= 12 ? 'PM' : 'AM';
    return '$month ${local.day}, ${local.year} $hour:$minute $suffix';
  }

  String _formatRelativeTime(DateTime parsed) {
    final local = parsed.toLocal();
    final diff = DateTime.now().difference(local);
    if (diff.inSeconds < 45) return 'Just now';
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    if (diff.inDays < 7) return '${diff.inDays}d ago';
    return '';
  }

  String _formatRemainingTime(String raw) {
    if (raw.isEmpty) return '';
    final parsed = DateTime.tryParse(raw);
    if (parsed == null) return '';
    final remaining = parsed.toLocal().difference(DateTime.now());
    if (remaining.inSeconds <= 0) return 'Expired';
    if (remaining.inDays >= 1) return '${remaining.inDays}d left';
    if (remaining.inHours >= 1) return '${remaining.inHours}h left';
    if (remaining.inMinutes >= 1) return '${remaining.inMinutes}m left';
    return '${remaining.inSeconds}s left';
  }

  String _monthLabel(int month) {
    const labels = [
      'Jan',
      'Feb',
      'Mar',
      'Apr',
      'May',
      'Jun',
      'Jul',
      'Aug',
      'Sep',
      'Oct',
      'Nov',
      'Dec',
    ];
    if (month < 1 || month > labels.length) return '';
    return labels[month - 1];
  }

  String _stringFromBattleDetail(Map<String, dynamic> detail, String key) {
    final battleRaw = detail['battle'];
    if (battleRaw is Map<String, dynamic>) {
      return battleRaw[key]?.toString() ?? '';
    }
    if (battleRaw is Map) {
      return Map<String, dynamic>.from(battleRaw)[key]?.toString() ?? '';
    }
    return '';
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
    final currentUserProfileUrl = context.select<AuthProvider, String>(
      (auth) => auth.user?.profilePictureUrl ?? '',
    );
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
                    onPressed: () =>
                        feed.markAsSeen(unseen.map((a) => a.id).toList()),
                    child: const Text('Mark all as seen'),
                  ),
                ),
              ),
            ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: list.length,
              separatorBuilder: (_, index) => const SizedBox(height: 10),
              itemBuilder: (_, i) => _buildActivityCard(
                context,
                list[i],
                currentUserProfileUrl: currentUserProfileUrl,
              ),
            ),
            if (feed.loadingMore) ...[
              const SizedBox(height: 16),
              const Center(
                child: Padding(
                  padding: EdgeInsets.symmetric(vertical: 8),
                  child: CircularProgressIndicator(),
                ),
              ),
            ] else if (feed.hasMore) ...[
              const SizedBox(height: 16),
              Center(
                child: OutlinedButton(
                  onPressed: feed.loadMore,
                  child: const Text('Load more'),
                ),
              ),
            ],
          ],
        );
      },
    );
  }
}

class _ActivityVisualStyle {
  const _ActivityVisualStyle({required this.accent, required this.background});

  final Color accent;
  final Color background;
}

class _ActivityThumbnail extends StatelessWidget {
  const _ActivityThumbnail({
    required this.imageUrl,
    required this.icon,
    required this.style,
  });

  final String imageUrl;
  final IconData icon;
  final _ActivityVisualStyle style;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 60,
      height: 60,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(14),
        color: style.background,
        border: Border.all(color: Colors.black.withValues(alpha: 0.06)),
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(13),
        child: imageUrl.isNotEmpty
            ? Image.network(
                imageUrl,
                fit: BoxFit.cover,
                errorBuilder: (_, error, stackTrace) =>
                    _ThumbnailFallback(icon: icon, style: style),
              )
            : _ThumbnailFallback(icon: icon, style: style),
      ),
    );
  }
}

class _ThumbnailFallback extends StatelessWidget {
  const _ThumbnailFallback({required this.icon, required this.style});

  final IconData icon;
  final _ActivityVisualStyle style;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: style.background,
      child: Center(child: Icon(icon, color: style.accent, size: 28)),
    );
  }
}
