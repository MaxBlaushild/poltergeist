import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/friend_invite.dart';
import '../models/user.dart';
import '../providers/auth_provider.dart';
import '../providers/friend_provider.dart';

class FriendsTabContent extends StatefulWidget {
  const FriendsTabContent({super.key});

  @override
  State<FriendsTabContent> createState() => _FriendsTabContentState();
}

class _FriendsTabContentState extends State<FriendsTabContent> {
  final _searchController = TextEditingController();
  bool _friendsExpanded = true;
  bool _receivedExpanded = true;
  bool _sentExpanded = true;
  bool _searchExpanded = true;
  final Set<String> _accepting = {};
  final Set<String> _rejecting = {};
  final Set<String> _sending = {};

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _fetchIfLoggedIn();
      _startPolling();
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _fetchIfLoggedIn() {
    final auth = context.read<AuthProvider>();
    final fp = context.read<FriendProvider>();
    if (auth.user != null) {
      fp.refresh();
    }
  }

  void _startPolling() {
    Future<void> poll() async {
      await Future.delayed(const Duration(seconds: 5));
      if (!mounted) return;
      final auth = context.read<AuthProvider>();
      final fp = context.read<FriendProvider>();
      if (auth.user != null) fp.refresh();
      if (mounted) _startPolling();
    }

    poll();
  }

  @override
  Widget build(BuildContext context) {
    return Consumer2<AuthProvider, FriendProvider>(
      builder: (context, auth, fp, _) {
        final user = auth.user;
        if (user == null) {
          return const Padding(
            padding: EdgeInsets.all(24),
            child: Center(child: Text('Log in to see friends')),
          );
        }

        final received = fp.friendInvites
            .where((i) => i.inviteeId == user.id)
            .toList();
        final sent = fp.friendInvites
            .where((i) => i.inviterId == user.id)
            .toList();

        return SingleChildScrollView(
          primary: false,
          padding: const EdgeInsets.only(bottom: 12),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _section(
                context,
                title: 'Friends',
                subtitle: '(${fp.friends.length})',
                expanded: _friendsExpanded,
                onToggle: () => setState(() => _friendsExpanded = !_friendsExpanded),
                child: fp.friends.isEmpty
                    ? const _Empty(message: 'No friends yet')
                    : Column(
                        children: fp.friends
                            .map((f) => _FriendTile(friend: f))
                            .toList(),
                      ),
              ),
              _section(
                context,
                title: 'Find Friends',
                expanded: _searchExpanded,
                onToggle: () => setState(() => _searchExpanded = !_searchExpanded),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    TextField(
                      controller: _searchController,
                      decoration: const InputDecoration(
                        hintText: 'Search by username...',
                        prefixIcon: Icon(Icons.search),
                        isDense: true,
                      ),
                      onChanged: (q) => fp.searchForFriends(q),
                    ),
                    if (_searchController.text.trim().isNotEmpty) ...[
                      const SizedBox(height: 12),
                      fp.searchResults.isEmpty
                          ? const _Empty(message: 'No users found')
                          : Column(
                              children: fp.searchResults
                                  .map((u) => _SearchResultTile(
                                        user: u,
                                        currentUserId: user.id,
                                        friends: fp.friends,
                                        sending: _sending.contains(u.id),
                                        onInvite: () async {
                                          setState(
                                              () => _sending.add(u.id));
                                          try {
                                            await fp.createFriendInvite(u.id);
                                          } finally {
                                            if (mounted) {
                                              setState(
                                                  () => _sending.remove(u.id));
                                            }
                                          }
                                        },
                                      ))
                                  .toList(),
                            ),
                    ],
                  ],
                ),
              ),
              _section(
                context,
                title: 'Received Invites',
                badge: received.length,
                expanded: _receivedExpanded,
                onToggle: () =>
                    setState(() => _receivedExpanded = !_receivedExpanded),
                child: received.isEmpty
                    ? const _Empty(message: 'No pending invites')
                    : Column(
                        children: received
                            .map((i) => _FriendInviteTile(
                                  invite: i,
                                  received: true,
                                  accepting: _accepting.contains(i.id),
                                  rejecting: _rejecting.contains(i.id),
                                  onAccept: () async {
                                    setState(() => _accepting.add(i.id));
                                    try {
                                      await fp.acceptFriendInvite(i.id);
                                    } finally {
                                      if (mounted) {
                                        setState(() => _accepting.remove(i.id));
                                      }
                                    }
                                  },
                                  onReject: () async {
                                    setState(() => _rejecting.add(i.id));
                                    try {
                                      await fp.deleteFriendInvite(i.id);
                                    } finally {
                                      if (mounted) {
                                        setState(() => _rejecting.remove(i.id));
                                      }
                                    }
                                  },
                                ))
                            .toList(),
                      ),
              ),
              _section(
                context,
                title: 'Sent Invites',
                subtitle: '(${sent.length})',
                expanded: _sentExpanded,
                onToggle: () => setState(() => _sentExpanded = !_sentExpanded),
                child: sent.isEmpty
                    ? const _Empty(message: 'No sent invites')
                    : Column(
                        children: sent
                            .map((i) => _FriendInviteTile(
                                  invite: i,
                                  received: false,
                                  accepting: false,
                                  rejecting: _rejecting.contains(i.id),
                                  onAccept: () {},
                                  onReject: () async {
                                    setState(() => _rejecting.add(i.id));
                                    try {
                                      await fp.deleteFriendInvite(i.id);
                                    } finally {
                                      if (mounted) {
                                        setState(() => _rejecting.remove(i.id));
                                      }
                                    }
                                  },
                                ))
                            .toList(),
                      ),
              ),
              const SizedBox(height: 24),
            ],
          ),
        );
      },
    );
  }

  Widget _section(
    BuildContext context, {
    required String title,
    String? subtitle,
    int? badge,
    required bool expanded,
    required VoidCallback onToggle,
    required Widget child,
  }) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final isSearch = title == 'Find Friends';
    final badgeColor = isSearch ? scheme.secondary : scheme.primary;
    final badgeTextColor = isSearch ? scheme.onSecondary : scheme.onPrimary;

    return Container(
      margin: const EdgeInsets.only(top: 16),
      child: Material(
        color: scheme.surface,
        elevation: 1,
        shadowColor: const Color(0x332D2416),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: BorderSide(color: scheme.outlineVariant),
        ),
        clipBehavior: Clip.antiAlias,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            InkWell(
              onTap: onToggle,
              borderRadius: BorderRadius.circular(16),
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                child: Row(
                  children: [
                    Expanded(
                      child: Row(
                        children: [
                          if (isSearch)
                            Icon(Icons.search, size: 20, color: scheme.primary),
                          if (isSearch) const SizedBox(width: 8),
                          Text(
                            title,
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.w700,
                              color: scheme.onSurface,
                            ),
                          ),
                          if (subtitle != null)
                            Text(
                              ' $subtitle',
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: scheme.onSurfaceVariant,
                              ),
                            ),
                          if (badge != null && badge > 0) ...[
                            const SizedBox(width: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 8, vertical: 2),
                              decoration: BoxDecoration(
                                color: badgeColor,
                                borderRadius: BorderRadius.circular(12),
                              ),
                              child: Text(
                                '$badge',
                                style: theme.textTheme.labelSmall?.copyWith(
                                  fontWeight: FontWeight.w700,
                                  color: badgeTextColor,
                                ),
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                    Icon(
                      expanded ? Icons.expand_more : Icons.chevron_right,
                      color: scheme.onSurfaceVariant,
                    ),
                  ],
                ),
              ),
            ),
            if (expanded) Divider(height: 1, color: scheme.outlineVariant),
            if (expanded)
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
                child: child,
              ),
          ],
        ),
      ),
    );
  }
}

class _Empty extends StatelessWidget {
  const _Empty({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 16),
      child: Center(
        child: Text(
          message,
          style: theme.textTheme.bodySmall?.copyWith(
            color: scheme.onSurfaceVariant,
          ),
        ),
      ),
    );
  }
}

class _FriendTile extends StatelessWidget {
  const _FriendTile({required this.friend});

  final User friend;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: scheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: scheme.outlineVariant),
      ),
      child: Row(
        children: [
          _Avatar(user: friend),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  friend.username.isNotEmpty ? friend.username : friend.name,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: scheme.onSurface,
                  ),
                ),
                if (friend.name.isNotEmpty && friend.name != friend.username)
                  Text(
                    friend.name,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: scheme.onSurfaceVariant,
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _SearchResultTile extends StatelessWidget {
  const _SearchResultTile({
    required this.user,
    required this.currentUserId,
    required this.friends,
    required this.sending,
    required this.onInvite,
  });

  final User user;
  final String currentUserId;
  final List<User> friends;
  final bool sending;
  final VoidCallback onInvite;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final isYou = user.id == currentUserId;
    final isFriend = friends.any((f) => f.id == user.id);

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: scheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: scheme.outlineVariant),
      ),
      child: Row(
        children: [
          _Avatar(user: user),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  user.username.isNotEmpty ? user.username : user.name,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: scheme.onSurface,
                  ),
                ),
                if (user.name.isNotEmpty && user.name != user.username)
                  Text(
                    user.name,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: scheme.onSurfaceVariant,
                    ),
                  ),
              ],
            ),
          ),
          if (isYou)
            Text(
              'You',
              style: theme.textTheme.bodySmall?.copyWith(
                color: scheme.onSurfaceVariant,
              ),
            )
          else if (isFriend)
            Text(
              'Friends',
              style: theme.textTheme.bodySmall?.copyWith(
                color: scheme.secondary,
                fontWeight: FontWeight.w700,
              ),
            )
          else
            FilledButton(
              onPressed: sending ? null : onInvite,
              child: Text(sending ? 'Sending...' : 'Invite Friend'),
            ),
        ],
      ),
    );
  }
}

class _FriendInviteTile extends StatelessWidget {
  const _FriendInviteTile({
    required this.invite,
    required this.received,
    required this.accepting,
    required this.rejecting,
    required this.onAccept,
    required this.onReject,
  });

  final FriendInvite invite;
  final bool received;
  final bool accepting;
  final bool rejecting;
  final VoidCallback onAccept;
  final VoidCallback onReject;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final user = received ? invite.inviter : invite.invitee;
    final date = invite.createdAt;
    String dateStr = date;
    try {
      dateStr = DateTime.parse(date).toString().split(' ').first;
    } catch (_) {}

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: scheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: scheme.outlineVariant),
      ),
      child: Row(
        children: [
          _Avatar(user: user),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  user.username.isNotEmpty ? user.username : user.name,
                  style: theme.textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w600,
                    color: scheme.onSurface,
                  ),
                ),
                Text(
                  dateStr,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: scheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
          ),
          if (received) ...[
            FilledButton(
              onPressed: (accepting || rejecting) ? null : onAccept,
              child: Text(accepting ? 'Accepting...' : 'Accept'),
            ),
            const SizedBox(width: 8),
            OutlinedButton(
              onPressed: (accepting || rejecting) ? null : onReject,
              style: OutlinedButton.styleFrom(
                foregroundColor: scheme.error,
                side: BorderSide(color: scheme.error),
              ),
              child: Text(rejecting ? 'Rejecting...' : 'Reject'),
            ),
          ] else
            OutlinedButton(
              onPressed: rejecting ? null : onReject,
              child: Text(rejecting ? 'Canceling...' : 'Cancel'),
            ),
        ],
      ),
    );
  }
}

class _Avatar extends StatelessWidget {
  const _Avatar({required this.user});

  final User user;
  static const double _size = 40;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return SizedBox(
      width: _size,
      height: _size,
      child: Stack(
        fit: StackFit.expand,
        children: [
          CircleAvatar(
            backgroundColor: scheme.surfaceVariant,
            backgroundImage: user.profilePictureUrl.isNotEmpty
                ? NetworkImage(user.profilePictureUrl)
                : null,
            child: user.profilePictureUrl.isEmpty
                ? Icon(
                    Icons.person,
                    size: _size * 0.5,
                    color: scheme.onSurfaceVariant,
                  )
                : null,
          ),
          Positioned(
            right: 0,
            bottom: 0,
            child: Container(
              width: 12,
              height: 12,
              decoration: BoxDecoration(
                color:
                    (user.isActive == true) ? scheme.secondary : scheme.outlineVariant,
                shape: BoxShape.circle,
                border: Border.all(color: scheme.surface, width: 2),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
