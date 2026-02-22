import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../models/party_invite.dart';
import '../models/user.dart';
import '../providers/auth_provider.dart';
import '../providers/friend_provider.dart';
import '../providers/party_provider.dart';

class PartyTabContent extends StatefulWidget {
  const PartyTabContent({super.key});

  @override
  State<PartyTabContent> createState() => _PartyTabContentState();
}

class _PartyTabContentState extends State<PartyTabContent> {
  bool _invitesExpanded = true;
  bool _sentInvitesExpanded = true;
  String? _promotingMemberId;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _fetchIfLoggedIn();
      _startPolling();
    });
  }

  void _fetchIfLoggedIn() {
    final auth = context.read<AuthProvider>();
    final party = context.read<PartyProvider>();
    if (auth.user != null) {
      party.refresh();
    }
  }

  void _startPolling() {
    Future<void> poll() async {
      await Future.delayed(const Duration(seconds: 5));
      if (!mounted) return;
      final auth = context.read<AuthProvider>();
      final party = context.read<PartyProvider>();
      if (auth.user != null) party.refresh();
      if (mounted) _startPolling();
    }

    poll();
  }

  void _showInviteToPartyDialog(
    BuildContext context,
    PartyProvider party,
    List<User> members,
    List<PartyInvite> invites,
    String currentUserId,
  ) {
    final fp = context.read<FriendProvider>();
    fp.fetchFriends();
    final memberIds = members.map((m) => m.id).toSet();
    final pendingIds = invites
        .map((invite) =>
            invite.inviterId == currentUserId ? invite.inviteeId : invite.inviterId)
        .toSet();
    final blockedIds = <String>{...memberIds, ...pendingIds, currentUserId};
    final searchController = TextEditingController();
    final inviting = <String>{};
    String query = '';

    showDialog<void>(
      context: context,
      builder: (ctx) {
        final theme = Theme.of(ctx);
        final scheme = theme.colorScheme;
        return StatefulBuilder(
          builder: (context, setState) {
            return AlertDialog(
              backgroundColor: scheme.surface,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
                side: BorderSide(color: scheme.outlineVariant),
              ),
              title: Text(
                'Invite friends',
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              content: SizedBox(
                width: 320,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    TextField(
                      controller: searchController,
                      decoration: const InputDecoration(
                        hintText: 'Search your friends...',
                        prefixIcon: Icon(Icons.search),
                        isDense: true,
                      ),
                      onChanged: (q) => setState(() => query = q),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Send invites to start a party or add members.',
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: scheme.onSurfaceVariant,
                      ),
                    ),
                    const SizedBox(height: 12),
                    SizedBox(
                      height: 240,
                      child: Consumer<FriendProvider>(
                        builder: (context, fp2, _) {
                          final friends = fp2.friends;
                          if (friends.isEmpty) {
                            return Center(
                              child: Text(
                                'No friends available yet.',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  color: scheme.onSurfaceVariant,
                                ),
                              ),
                            );
                          }
                          final q = query.trim().toLowerCase();
                          final results = friends
                              .where((u) => !blockedIds.contains(u.id))
                              .where((u) {
                                if (q.isEmpty) return true;
                                final haystack =
                                    '${u.username} ${u.name}'.toLowerCase();
                                return haystack.contains(q);
                              })
                              .toList();

                          if (results.isEmpty) {
                            return Center(
                              child: Text(
                                q.isEmpty
                                    ? 'No friends to invite.'
                                    : 'No matching friends found.',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  color: scheme.onSurfaceVariant,
                                ),
                              ),
                            );
                          }

                          return ListView.builder(
                            shrinkWrap: true,
                            itemCount: results.length,
                            itemBuilder: (_, i) {
                              final u = results[i];
                              final isInviting = inviting.contains(u.id);
                              return Padding(
                                padding: const EdgeInsets.only(bottom: 8),
                                child: Material(
                                  color: scheme.surfaceContainerHighest,
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12),
                                    side:
                                        BorderSide(color: scheme.outlineVariant),
                                  ),
                                  child: ListTile(
                                    leading: CircleAvatar(
                                      radius: 16,
                                      backgroundColor: scheme.surfaceVariant,
                                      backgroundImage:
                                          u.profilePictureUrl.isNotEmpty
                                              ? NetworkImage(u.profilePictureUrl)
                                              : null,
                                      child: u.profilePictureUrl.isEmpty
                                          ? Icon(
                                              Icons.person,
                                              size: 18,
                                              color: scheme.onSurfaceVariant,
                                            )
                                          : null,
                                    ),
                                    title: Text(
                                      u.username.isNotEmpty
                                          ? u.username
                                          : u.name,
                                    ),
                                    trailing: FilledButton.tonal(
                                      onPressed: isInviting
                                          ? null
                                          : () async {
                                              setState(
                                                  () => inviting.add(u.id));
                                              try {
                                                await party.inviteToParty(u);
                                                if (!ctx.mounted) return;
                                                setState(() {
                                                  pendingIds.add(u.id);
                                                  blockedIds.add(u.id);
                                                });
                                              } finally {
                                                if (ctx.mounted) {
                                                  setState(() =>
                                                      inviting.remove(u.id));
                                                }
                                              }
                                            },
                                      child: Text(isInviting ? '…' : 'Invite'),
                                    ),
                                  ),
                                ),
                              );
                            },
                          );
                        },
                      ),
                    ),
                  ],
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(ctx).pop(),
                  child: const Text('Close'),
                ),
              ],
            );
          },
        );
      },
    ).then((_) => searchController.dispose());
  }

  void _openCharacterProfile(
    BuildContext context,
    User target,
    String currentUserId,
  ) {
    if (target.id.isEmpty || target.id == currentUserId) return;
    Scaffold.maybeOf(context)?.closeEndDrawer();
    context.go('/character/${target.id}');
  }

  @override
  Widget build(BuildContext context) {
    return Consumer2<AuthProvider, PartyProvider>(
      builder: (context, auth, party, _) {
        final user = auth.user;
        if (user == null) {
          return const Padding(
            padding: EdgeInsets.all(24),
            child: Center(child: Text('Log in to see party')),
          );
        }

        if (party.loading && party.party == null && party.partyInvites.isEmpty) {
          return const Padding(
            padding: EdgeInsets.all(24),
            child: Center(child: CircularProgressIndicator()),
          );
        }

        final isLeader = party.party?.leaderId == user.id;
        final receivedInvites = party.partyInvites
            .where((i) => i.inviteeId == user.id)
            .toList();
        final sentInvites = party.partyInvites
            .where((i) => i.inviterId == user.id)
            .toList();

        return Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (receivedInvites.isNotEmpty)
              _partyInvitesSection(context, party, receivedInvites, true, user.id),
            if (party.party != null)
              _currentPartySection(context, party, user, isLeader),
            if (party.party == null)
              _noPartyPlaceholder(
                context,
                onCreate: () => _showInviteToPartyDialog(
                  context,
                  party,
              const [],
              party.partyInvites,
              user.id,
            ),
          ),
            if (sentInvites.isNotEmpty)
              _sentInvitesSection(context, party, sentInvites, user.id),
            const SizedBox(height: 24),
          ],
        );
      },
    );
  }

  Widget _partyInvitesSection(
    BuildContext context,
    PartyProvider party,
    List<PartyInvite> invites,
    bool received,
    String currentUserId,
  ) {
    return _AccordionSection(
      title: 'Party Invites',
      badge: invites.length,
      expanded: _invitesExpanded,
      onToggle: () => setState(() => _invitesExpanded = !_invitesExpanded),
      accent: true,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: invites
            .map((invite) => _PartyInviteTile(
                  invite: invite,
                  received: received,
                  onViewProfile: () => _openCharacterProfile(
                    context,
                    received ? invite.inviter : invite.invitee,
                    currentUserId,
                  ),
                  onAccept: () => party.acceptPartyInvite(invite.id),
                  onReject: () => party.rejectPartyInvite(invite.id),
                ))
            .toList(),
      ),
    );
  }

  Widget _currentPartySection(
    BuildContext context,
    PartyProvider party,
    User user,
    bool isLeader,
  ) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final p = party.party!;
    return Padding(
      padding: const EdgeInsets.only(top: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Material(
            color: scheme.surfaceVariant,
            elevation: 1,
            shadowColor: const Color(0x332D2416),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(14),
              side: BorderSide(color: scheme.outlineVariant),
            ),
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Row(
                children: [
                  Icon(Icons.group, color: scheme.primary, size: 24),
                  const SizedBox(width: 8),
                  Text(
                    'Party (${p.members.length}/5)',
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                      color: scheme.onSurface,
                    ),
                  ),
                  if (isLeader) ...[
                    const Spacer(),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: scheme.tertiary,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        'LEADER',
                        style: theme.textTheme.labelSmall?.copyWith(
                          fontWeight: FontWeight.w700,
                          color: scheme.onTertiary,
                          letterSpacing: 0.6,
                        ),
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
          const SizedBox(height: 8),
          ...p.members.map((member) {
            final isMemberLeader = member.id == p.leaderId;
            final isCurrentUser = member.id == user.id;
            return _PartyMemberTile(
              member: member,
              isLeader: isMemberLeader,
              isCurrentUser: isCurrentUser,
              canPromote: isLeader && !isCurrentUser && !isMemberLeader,
              promoting: _promotingMemberId == member.id,
              onViewProfile: isCurrentUser
                  ? null
                  : () => _openCharacterProfile(context, member, user.id),
              onPromote: () async {
                setState(() => _promotingMemberId = member.id);
                try {
                  await party.setLeader(member);
                } finally {
                  if (mounted) setState(() => _promotingMemberId = null);
                }
              },
            );
          }),
          if (isLeader && p.members.length < 5) ...[
            const SizedBox(height: 8),
            OutlinedButton.icon(
              onPressed: () => _showInviteToPartyDialog(
                context,
                party,
                p.members,
                party.partyInvites,
                user.id,
              ),
              icon: const Icon(Icons.person_add, size: 18),
              label: const Text('Invite to party'),
            ),
          ],
          const SizedBox(height: 8),
          SizedBox(
            width: double.infinity,
            child: FilledButton.icon(
              onPressed: () async {
                final confirm = await showDialog<bool>(
                  context: context,
                  builder: (ctx) {
                    final theme = Theme.of(ctx);
                    final scheme = theme.colorScheme;
                    return AlertDialog(
                      backgroundColor: scheme.surface,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                        side: BorderSide(color: scheme.outlineVariant),
                      ),
                      title: Text(
                        'Leave party?',
                        style: theme.textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      content: Text(
                        isLeader
                            ? 'You are the leader. Are you sure you want to leave?'
                            : 'Are you sure you want to leave the party?',
                      ),
                      actions: [
                        TextButton(
                          onPressed: () => Navigator.of(ctx).pop(false),
                          child: const Text('Cancel'),
                        ),
                        FilledButton(
                          onPressed: () => Navigator.of(ctx).pop(true),
                          child: const Text('Leave'),
                        ),
                      ],
                    );
                  },
                );
                if (confirm == true && context.mounted) {
                  await party.leaveParty();
                }
              },
              icon: const Icon(Icons.logout, size: 20),
              label: const Text('Leave Party'),
              style: FilledButton.styleFrom(
                backgroundColor: scheme.error,
                foregroundColor: scheme.onError,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _noPartyPlaceholder(BuildContext context, {VoidCallback? onCreate}) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 24),
      child: Column(
        children: [
          Icon(Icons.group_off, size: 64, color: scheme.onSurfaceVariant),
          const SizedBox(height: 12),
          Text(
            'No Active Party',
            style: theme.textTheme.titleMedium?.copyWith(
              color: scheme.onSurface,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Start a party by inviting friends.',
            style: theme.textTheme.bodySmall?.copyWith(
              color: scheme.onSurfaceVariant,
            ),
            textAlign: TextAlign.center,
          ),
          if (onCreate != null) ...[
            const SizedBox(height: 12),
            OutlinedButton.icon(
              onPressed: onCreate,
              icon: const Icon(Icons.group_add, size: 18),
              label: const Text('Create Party'),
            ),
          ],
        ],
      ),
    );
  }

  Widget _sentInvitesSection(
    BuildContext context,
    PartyProvider party,
    List<PartyInvite> invites,
    String currentUserId,
  ) {
    return _AccordionSection(
      title: 'Pending Invites',
      badge: invites.length,
      expanded: _sentInvitesExpanded,
      onToggle: () => setState(() => _sentInvitesExpanded = !_sentInvitesExpanded),
      accent: false,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: invites
            .map((invite) => _PartyInviteTile(
                  invite: invite,
                  received: false,
                  onViewProfile: () => _openCharacterProfile(
                    context,
                    invite.invitee,
                    currentUserId,
                  ),
                  onAccept: () {},
                  onReject: () => party.rejectPartyInvite(invite.id),
                ))
            .toList(),
      ),
    );
  }
}

class _AccordionSection extends StatelessWidget {
  const _AccordionSection({
    required this.title,
    required this.badge,
    required this.expanded,
    required this.onToggle,
    required this.accent,
    required this.child,
  });

  final String title;
  final int badge;
  final bool expanded;
  final VoidCallback onToggle;
  final bool accent;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final accentColor = accent ? scheme.tertiary : scheme.primary;
    final accentText = accent ? scheme.onTertiary : scheme.onPrimary;

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
                    Icon(
                      expanded ? Icons.expand_more : Icons.chevron_right,
                      color: scheme.onSurfaceVariant,
                      size: 24,
                    ),
                    const SizedBox(width: 8),
                    Text(
                      title,
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                        color: scheme.onSurface,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                      decoration: BoxDecoration(
                        color: accentColor,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        '$badge',
                        style: theme.textTheme.labelSmall?.copyWith(
                          fontWeight: FontWeight.w700,
                          color: accentText,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            if (expanded) Divider(height: 1, color: scheme.outlineVariant),
            if (expanded) Padding(padding: const EdgeInsets.all(12), child: child),
          ],
        ),
      ),
    );
  }
}

class _PartyInviteTile extends StatelessWidget {
  const _PartyInviteTile({
    required this.invite,
    required this.received,
    required this.onViewProfile,
    required this.onAccept,
    required this.onReject,
  });

  final PartyInvite invite;
  final bool received;
  final VoidCallback onViewProfile;
  final VoidCallback onAccept;
  final VoidCallback onReject;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final user = received ? invite.inviter : invite.invitee;
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
          GestureDetector(
            onTap: onViewProfile,
            child: _UserAvatar(user: user, size: 40),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                GestureDetector(
                  onTap: onViewProfile,
                  child: Text(
                    user.username.isNotEmpty ? user.username : user.name,
                    style: theme.textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                      color: scheme.onSurface,
                    ),
                  ),
                ),
                Text(
                  received ? 'invited you to their party' : 'Invite pending...',
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: scheme.onSurfaceVariant,
                  ),
                ),
              ],
            ),
          ),
          if (received) ...[
            IconButton(
              onPressed: onAccept,
              icon: Icon(Icons.check, color: scheme.secondary),
              tooltip: 'Accept',
            ),
            IconButton(
              onPressed: onReject,
              icon: Icon(Icons.close, color: scheme.error),
              tooltip: 'Decline',
            ),
          ] else
            TextButton(
              onPressed: onReject,
              child: const Text('Cancel'),
            ),
        ],
      ),
    );
  }
}

class _PartyMemberTile extends StatelessWidget {
  const _PartyMemberTile({
    required this.member,
    required this.isLeader,
    required this.isCurrentUser,
    required this.canPromote,
    required this.promoting,
    required this.onViewProfile,
    required this.onPromote,
  });

  final User member;
  final bool isLeader;
  final bool isCurrentUser;
  final bool canPromote;
  final bool promoting;
  final VoidCallback? onViewProfile;
  final VoidCallback onPromote;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final isInactive = member.isActive != true;
    final nameColor = isInactive
        ? scheme.onSurfaceVariant
        : (isLeader ? scheme.tertiary : scheme.onSurface);
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: isLeader
            ? scheme.tertiary.withValues(alpha: 0.12)
            : scheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isLeader
              ? scheme.tertiary.withValues(alpha: 0.45)
              : scheme.outlineVariant,
        ),
      ),
      child: Row(
        children: [
          GestureDetector(
            onTap: onViewProfile,
            child: _UserAvatar(user: member, size: 48),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                GestureDetector(
                  onTap: onViewProfile,
                  child: Row(
                    children: [
                      Text(
                        member.username.isNotEmpty
                            ? member.username
                            : member.name,
                        style: TextStyle(
                          fontWeight: FontWeight.bold,
                          color: nameColor,
                        ),
                      ),
                      if (isCurrentUser)
                        Text(
                          ' (You)',
                          style: TextStyle(
                            fontSize: 14,
                            color: scheme.onSurfaceVariant,
                          ),
                        ),
                      if (isInactive)
                        Text(
                          ' (inactive)',
                          style: TextStyle(
                            fontSize: 14,
                            color: scheme.onSurfaceVariant,
                          ),
                        ),
                    ],
                  ),
                ),
                if (member.name.isNotEmpty && member.name != member.username)
                  Text(
                    member.name,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: scheme.onSurfaceVariant,
                    ),
                  ),
              ],
            ),
          ),
          if (canPromote)
            FilledButton(
              onPressed: promoting ? null : onPromote,
              child: Text(promoting ? '…' : 'Promote'),
            ),
        ],
      ),
    );
  }
}

class _UserAvatar extends StatelessWidget {
  const _UserAvatar({required this.user, this.size = 40});

  final User user;
  final double size;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return CircleAvatar(
      radius: size / 2,
      backgroundColor: scheme.surfaceVariant,
      backgroundImage: user.profilePictureUrl.isNotEmpty
          ? NetworkImage(user.profilePictureUrl)
          : null,
      child: user.profilePictureUrl.isEmpty
          ? Icon(
              Icons.person,
              size: size * 0.5,
              color: scheme.onSurfaceVariant,
            )
          : null,
    );
  }
}
