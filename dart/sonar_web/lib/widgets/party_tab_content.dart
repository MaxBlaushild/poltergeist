import 'package:flutter/material.dart';
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
  ) {
    final fp = context.read<FriendProvider>();
    final memberIds = members.map((m) => m.id).toSet();
    final searchController = TextEditingController();
    fp.clearSearch();

    showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Invite to party'),
        content: SizedBox(
          width: 320,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextField(
                controller: searchController,
                decoration: const InputDecoration(
                  hintText: 'Search by username...',
                  border: OutlineInputBorder(),
                  isDense: true,
                ),
                onChanged: (q) => fp.searchForFriends(q),
              ),
              const SizedBox(height: 12),
              SizedBox(
                height: 240,
                child: Consumer<FriendProvider>(
                    builder: (context, fp2, _) {
                      final results = fp2.searchResults
                          .where((u) => !memberIds.contains(u.id))
                          .toList();
                      return ListView.builder(
                        shrinkWrap: true,
                        itemCount: results.length,
                        itemBuilder: (_, i) {
                          final u = results[i];
                          return ListTile(
                            leading: CircleAvatar(
                              radius: 16,
                              backgroundColor: Colors.grey.shade300,
                              backgroundImage:
                                  u.profilePictureUrl.isNotEmpty
                                      ? NetworkImage(u.profilePictureUrl)
                                      : null,
                              child: u.profilePictureUrl.isEmpty
                                  ? const Icon(Icons.person, size: 18)
                                  : null,
                            ),
                            title: Text(
                                u.username.isNotEmpty ? u.username : u.name),
                            trailing: TextButton(
                              onPressed: () async {
                                await party.inviteToParty(u);
                                if (ctx.mounted) Navigator.of(ctx).pop();
                              },
                              child: const Text('Invite'),
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
      ),
    ).then((_) {
      searchController.dispose();
      fp.clearSearch();
    });
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
            if (receivedInvites.isNotEmpty) _partyInvitesSection(context, party, receivedInvites, true),
            if (party.party != null) _currentPartySection(context, party, user, isLeader),
            if (party.party == null && receivedInvites.isEmpty)
              _noPartyPlaceholder(context),
            if (sentInvites.isNotEmpty) _sentInvitesSection(context, party, sentInvites),
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
    final p = party.party!;
    return Padding(
      padding: const EdgeInsets.only(top: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.purple.shade900.withValues(alpha: 0.3),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: Colors.purple.shade700),
            ),
            child: Row(
              children: [
                const Icon(Icons.group, color: Colors.white70, size: 24),
                const SizedBox(width: 8),
                Text(
                  'Party (${p.members.length}/5)',
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
                if (isLeader) ...[
                  const Spacer(),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: Colors.amber.shade700,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Text('LEADER', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Colors.white)),
                  ),
                ],
              ],
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
              onPressed: () => _showInviteToPartyDialog(context, party, p.members),
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
                  builder: (ctx) => AlertDialog(
                    title: const Text('Leave party?'),
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
                  ),
                );
                if (confirm == true && context.mounted) {
                  await party.leaveParty();
                }
              },
              icon: const Icon(Icons.logout, size: 20),
              label: const Text('Leave Party'),
              style: FilledButton.styleFrom(
                backgroundColor: Colors.red.shade700,
                foregroundColor: Colors.white,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _noPartyPlaceholder(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 24),
      child: Column(
        children: [
          Icon(Icons.group_off, size: 64, color: Colors.grey.shade600),
          const SizedBox(height: 12),
          Text(
            'No Active Party',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: Colors.grey.shade400,
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            'Accept an invite or get invited by a friend!',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Colors.grey.shade500,
                ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  Widget _sentInvitesSection(
    BuildContext context,
    PartyProvider party,
    List<PartyInvite> invites,
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
    return Container(
      margin: const EdgeInsets.only(top: 16),
      decoration: BoxDecoration(
        color: accent
            ? Colors.amber.shade900.withValues(alpha: 0.2)
            : Colors.grey.shade900.withValues(alpha: 0.2),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: accent ? Colors.amber.shade700 : Colors.grey.shade700),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InkWell(
            onTap: onToggle,
            borderRadius: BorderRadius.circular(12),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              child: Row(
                children: [
                  Icon(
                    expanded ? Icons.expand_more : Icons.chevron_right,
                    color: accent ? Colors.amber.shade300 : Colors.grey.shade400,
                    size: 24,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    title,
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      color: accent ? Colors.amber.shade100 : Colors.grey.shade300,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                    decoration: BoxDecoration(
                      color: accent ? Colors.amber.shade700 : Colors.grey.shade600,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      '$badge',
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          if (expanded) Padding(padding: const EdgeInsets.all(12), child: child),
        ],
      ),
    );
  }
}

class _PartyInviteTile extends StatelessWidget {
  const _PartyInviteTile({
    required this.invite,
    required this.received,
    required this.onAccept,
    required this.onReject,
  });

  final PartyInvite invite;
  final bool received;
  final VoidCallback onAccept;
  final VoidCallback onReject;

  @override
  Widget build(BuildContext context) {
    final user = received ? invite.inviter : invite.invitee;
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.black26,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        children: [
          _UserAvatar(user: user, size: 40),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  user.username.isNotEmpty ? user.username : user.name,
                  style: const TextStyle(
                    fontWeight: FontWeight.w600,
                    color: Colors.white,
                  ),
                ),
                Text(
                  received ? 'invited you to their party' : 'Invite pending...',
                  style: TextStyle(fontSize: 12, color: Colors.grey.shade400),
                ),
              ],
            ),
          ),
          if (received) ...[
            IconButton(
              onPressed: onAccept,
              icon: const Icon(Icons.check, color: Colors.green),
              tooltip: 'Accept',
            ),
            IconButton(
              onPressed: onReject,
              icon: const Icon(Icons.close, color: Colors.red),
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
    required this.onPromote,
  });

  final User member;
  final bool isLeader;
  final bool isCurrentUser;
  final bool canPromote;
  final bool promoting;
  final VoidCallback onPromote;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: isLeader
            ? Colors.amber.shade900.withValues(alpha: 0.3)
            : Colors.black26,
        borderRadius: BorderRadius.circular(8),
        border: isLeader
            ? Border.all(color: Colors.amber.shade700.withValues(alpha: 0.5))
            : Border.all(color: Colors.grey.shade700),
      ),
      child: Row(
        children: [
          _UserAvatar(user: member, size: 48),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(
                      member.username.isNotEmpty ? member.username : member.name,
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: isLeader ? Colors.amber.shade300 : Colors.white,
                      ),
                    ),
                    if (isCurrentUser)
                      Text(
                        ' (You)',
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.grey.shade400,
                        ),
                      ),
                  ],
                ),
                if (member.name.isNotEmpty && member.name != member.username)
                  Text(
                    member.name,
                    style: TextStyle(fontSize: 12, color: Colors.grey.shade400),
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
    return CircleAvatar(
      radius: size / 2,
      backgroundColor: Colors.grey.shade700,
      backgroundImage: user.profilePictureUrl.isNotEmpty
          ? NetworkImage(user.profilePictureUrl)
          : null,
      child: user.profilePictureUrl.isEmpty
          ? Icon(Icons.person, size: size * 0.5, color: Colors.grey.shade400)
          : null,
    );
  }
}
