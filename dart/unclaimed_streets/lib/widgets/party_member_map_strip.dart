import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../models/party.dart';
import '../models/user.dart';
import '../models/user_character_profile.dart';
import '../providers/auth_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/party_provider.dart';
import '../screens/layout_shell.dart';
import '../services/user_character_service.dart';

class PartyMemberMapStrip extends StatefulWidget {
  const PartyMemberMapStrip({super.key});

  @override
  State<PartyMemberMapStrip> createState() => _PartyMemberMapStripState();
}

class _PartyMemberMapStripState extends State<PartyMemberMapStrip> {
  static const Duration _pollInterval = Duration(seconds: 10);

  final Map<String, UserCharacterProfile?> _profilesByUserId =
      <String, UserCharacterProfile?>{};
  final Set<String> _loadingProfileIds = <String>{};
  Set<String> _lastMemberIds = <String>{};
  Timer? _pollTimer;
  String? _lastUserId;
  bool _refreshInFlight = false;
  bool _expanded = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      _handleAuthChange();
      unawaited(_refreshPartyAndProfiles(force: false));
    });
    _pollTimer = Timer.periodic(_pollInterval, (_) {
      if (!mounted) return;
      unawaited(_refreshPartyAndProfiles(force: true));
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _handleAuthChange();
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    super.dispose();
  }

  void _handleAuthChange() {
    final userId = context.read<AuthProvider>().user?.id;
    if (userId == _lastUserId) return;
    _lastUserId = userId;
    _lastMemberIds = <String>{};
    _profilesByUserId.clear();
    _loadingProfileIds.clear();
    if (mounted) {
      setState(() => _expanded = false);
    }
  }

  void _toggleExpanded() {
    if (!mounted) return;
    setState(() => _expanded = !_expanded);
  }

  Future<void> _refreshPartyAndProfiles({required bool force}) async {
    if (_refreshInFlight) return;
    final authUser = context.read<AuthProvider>().user;
    if (authUser == null || authUser.id.isEmpty) return;

    _refreshInFlight = true;
    try {
      final partyProvider = context.read<PartyProvider>();
      await partyProvider.fetchParty();
      if (!mounted) return;
      final members = _combinedMembers(authUser, partyProvider.party);
      await _refreshProfilesForMembers(
        members,
        currentUserId: authUser.id,
        force: force,
      );
    } finally {
      _refreshInFlight = false;
    }
  }

  List<User> _combinedMembers(User currentUser, Party? party) {
    if (party == null) return <User>[currentUser];
    final users = <User>[];
    final seenIds = <String>{};

    if (currentUser.id.isNotEmpty) {
      seenIds.add(currentUser.id);
      users.add(currentUser);
    }

    for (final member in party.members) {
      if (member.id.isEmpty || !seenIds.add(member.id)) continue;
      users.add(member);
    }

    return users;
  }

  void _scheduleProfileSync(List<User> members, String currentUserId) {
    final memberIds = members
        .where((member) => member.id.isNotEmpty)
        .map((member) => member.id)
        .toSet();
    if (setEquals(memberIds, _lastMemberIds)) return;
    _lastMemberIds = memberIds;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      unawaited(
        _refreshProfilesForMembers(
          members,
          currentUserId: currentUserId,
          force: false,
        ),
      );
    });
  }

  Future<void> _refreshProfilesForMembers(
    List<User> members, {
    required String currentUserId,
    required bool force,
  }) async {
    final remoteMembers = members
        .where((member) => member.id.isNotEmpty && member.id != currentUserId)
        .toList(growable: false);
    final remoteIds = remoteMembers.map((member) => member.id).toSet();

    if (mounted) {
      setState(() {
        _profilesByUserId.removeWhere(
          (userId, _) => !remoteIds.contains(userId),
        );
        _loadingProfileIds.removeWhere((userId) => !remoteIds.contains(userId));
      });
    }

    final membersToLoad = remoteMembers
        .where((member) {
          if (_loadingProfileIds.contains(member.id)) return false;
          if (!force && _profilesByUserId.containsKey(member.id)) return false;
          return true;
        })
        .toList(growable: false);
    if (membersToLoad.isEmpty) return;

    if (mounted) {
      setState(() {
        _loadingProfileIds.addAll(membersToLoad.map((member) => member.id));
      });
    }

    final service = context.read<UserCharacterService>();
    final results = await Future.wait(
      membersToLoad.map((member) async {
        final profile = await service.getProfile(member.id);
        return MapEntry(member.id, profile);
      }),
    );
    if (!mounted) return;

    setState(() {
      for (final entry in results) {
        _profilesByUserId[entry.key] = entry.value;
        _loadingProfileIds.remove(entry.key);
      }
    });
  }

  void _openMemberProfile(
    BuildContext context,
    User member,
    String currentUserId,
  ) {
    final drawerController = LayoutShellDrawerController.maybeOf(context);
    if (drawerController != null) {
      if (member.id == currentUserId) {
        drawerController.openCharacter();
      } else {
        drawerController.openProfile(member);
      }
      return;
    }

    Scaffold.maybeOf(context)?.openEndDrawer();
    if (member.id != currentUserId) {
      context.go('/character/${member.id}');
    }
  }

  _MemberResources? _resourcesForMember(
    User member,
    CharacterStatsProvider statsProvider,
    String currentUserId,
  ) {
    if (member.id == currentUserId) {
      return _MemberResources(
        health: statsProvider.health,
        maxHealth: statsProvider.maxHealth,
        mana: statsProvider.mana,
        maxMana: statsProvider.maxMana,
      );
    }

    final profile = _profilesByUserId[member.id];
    final stats = profile?.stats;
    if (stats == null) return null;

    return _MemberResources(
      health: stats.health,
      maxHealth: stats.maxHealth,
      mana: stats.mana,
      maxMana: stats.maxMana,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Consumer3<AuthProvider, PartyProvider, CharacterStatsProvider>(
      builder: (context, auth, partyProvider, statsProvider, _) {
        final currentUser = auth.user;
        if (currentUser == null) {
          return const SizedBox.shrink();
        }

        final party = partyProvider.party;
        final members = party == null
            ? <User>[currentUser]
            : _combinedMembers(currentUser, party);

        _scheduleProfileSync(members, currentUser.id);

        final theme = Theme.of(context);
        final currentMember = members.first;
        final currentMemberResources = _resourcesForMember(
          currentMember,
          statsProvider,
          currentUser.id,
        );
        final extraMembers = members.skip(1).toList(growable: false);
        final hasExpandableParty = extraMembers.isNotEmpty;
        final screenWidth = MediaQuery.sizeOf(context).width;
        const tileWidth = 58.0;
        const tileGap = 10.0;
        const toggleGap = 2.0;
        const toggleWidth = 26.0;
        const horizontalPadding = 12.0;
        final closedWidth =
            tileWidth +
            (hasExpandableParty ? toggleGap + toggleWidth : 0.0) +
            horizontalPadding;
        final expandedWidth =
            tileWidth * members.length +
            tileGap * extraMembers.length +
            (hasExpandableParty ? toggleGap + toggleWidth : 0.0) +
            horizontalPadding;
        final maxAllowedWidth = (screenWidth - 32.0).clamp(
          closedWidth,
          expandedWidth,
        );
        final containerWidth = _expanded && hasExpandableParty
            ? maxAllowedWidth
            : closedWidth;
        return ConstrainedBox(
          constraints: BoxConstraints(maxWidth: maxAllowedWidth),
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 240),
            curve: Curves.easeInOut,
            width: containerWidth,
            decoration: BoxDecoration(
              color: theme.colorScheme.surface.withValues(alpha: 0.94),
              borderRadius: BorderRadius.circular(18),
              border: Border.all(color: theme.colorScheme.outlineVariant),
              boxShadow: const [
                BoxShadow(
                  color: Color(0x332D2416),
                  blurRadius: 16,
                  offset: Offset(0, 6),
                ),
              ],
            ),
            child: Padding(
              padding: const EdgeInsets.fromLTRB(6, 10, 6, 8),
              child: SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                physics: _expanded
                    ? const BouncingScrollPhysics()
                    : const NeverScrollableScrollPhysics(),
                child: Row(
                  children: [
                    _PartyMemberThumbnail(
                      member: currentMember,
                      isLeader: currentMember.id == party?.leaderId,
                      isCurrentUser: true,
                      isLoading: false,
                      resources: currentMemberResources,
                      onTap: () => _openMemberProfile(
                        context,
                        currentMember,
                        currentUser.id,
                      ),
                    ),
                    AnimatedCrossFade(
                      firstChild: const SizedBox.shrink(),
                      secondChild: Row(
                        children: [
                          const SizedBox(width: 10),
                          for (
                            var index = 0;
                            index < extraMembers.length;
                            index++
                          ) ...[
                            _PartyMemberThumbnail(
                              member: extraMembers[index],
                              isLeader:
                                  extraMembers[index].id == party?.leaderId,
                              isCurrentUser: false,
                              isLoading: _loadingProfileIds.contains(
                                extraMembers[index].id,
                              ),
                              resources: _resourcesForMember(
                                extraMembers[index],
                                statsProvider,
                                currentUser.id,
                              ),
                              onTap: () => _openMemberProfile(
                                context,
                                extraMembers[index],
                                currentUser.id,
                              ),
                            ),
                            if (index < extraMembers.length - 1)
                              const SizedBox(width: 10),
                          ],
                        ],
                      ),
                      crossFadeState: _expanded && hasExpandableParty
                          ? CrossFadeState.showSecond
                          : CrossFadeState.showFirst,
                      duration: const Duration(milliseconds: 180),
                      sizeCurve: Curves.easeInOut,
                    ),
                    if (hasExpandableParty) ...[
                      const SizedBox(width: 2),
                      _AccordionToggleButton(
                        expanded: _expanded,
                        onTap: _toggleExpanded,
                      ),
                    ],
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}

class _AccordionToggleButton extends StatelessWidget {
  const _AccordionToggleButton({required this.expanded, required this.onTap});

  final bool expanded;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Material(
      color: theme.colorScheme.surfaceContainerHighest,
      borderRadius: BorderRadius.circular(999),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(999),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 18),
          child: Icon(
            expanded ? Icons.chevron_right : Icons.chevron_left,
            size: 18,
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
      ),
    );
  }
}

class _PartyMemberThumbnail extends StatelessWidget {
  const _PartyMemberThumbnail({
    required this.member,
    required this.isLeader,
    required this.isCurrentUser,
    required this.isLoading,
    required this.resources,
    required this.onTap,
  });

  final User member;
  final bool isLeader;
  final bool isCurrentUser;
  final bool isLoading;
  final _MemberResources? resources;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final label = member.username.isNotEmpty ? member.username : member.name;
    final borderColor = isCurrentUser
        ? theme.colorScheme.primary
        : theme.colorScheme.outlineVariant;

    return Tooltip(
      message: label.isEmpty ? 'Party member' : label,
      child: SizedBox(
        width: 58,
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: onTap,
            borderRadius: BorderRadius.circular(14),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 1, vertical: 2),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Stack(
                    clipBehavior: Clip.none,
                    children: [
                      Container(
                        width: 54,
                        height: 54,
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(14),
                          border: Border.all(color: borderColor, width: 1.5),
                          color: theme.colorScheme.surfaceContainerHighest,
                          image: member.profilePictureUrl.isNotEmpty
                              ? DecorationImage(
                                  image: NetworkImage(member.profilePictureUrl),
                                  fit: BoxFit.cover,
                                )
                              : null,
                        ),
                        child: member.profilePictureUrl.isNotEmpty
                            ? null
                            : Icon(
                                Icons.person,
                                color: theme.colorScheme.onSurfaceVariant,
                              ),
                      ),
                      if (isLeader)
                        Positioned(
                          right: -4,
                          top: -4,
                          child: Container(
                            width: 18,
                            height: 18,
                            decoration: BoxDecoration(
                              color: const Color(0xFFD4A64A),
                              shape: BoxShape.circle,
                              border: Border.all(
                                color: theme.colorScheme.surface,
                                width: 1.2,
                              ),
                            ),
                            child: const Icon(
                              Icons.star,
                              size: 11,
                              color: Colors.white,
                            ),
                          ),
                        ),
                      if (isCurrentUser)
                        Positioned(
                          left: -2,
                          bottom: -2,
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 4,
                              vertical: 2,
                            ),
                            decoration: BoxDecoration(
                              color: theme.colorScheme.primaryContainer,
                              borderRadius: BorderRadius.circular(999),
                              border: Border.all(
                                color: theme.colorScheme.surface,
                                width: 1,
                              ),
                            ),
                            child: Text(
                              'You',
                              style: theme.textTheme.labelSmall?.copyWith(
                                color: theme.colorScheme.onPrimaryContainer,
                                fontWeight: FontWeight.w700,
                                fontSize: 9,
                              ),
                            ),
                          ),
                        ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  _MiniResourceBar(
                    value: resources?.health ?? 0,
                    maxValue: resources?.maxHealth ?? 0,
                    fillColor: const Color(0xFFB33939),
                    loading: isLoading && resources == null,
                  ),
                  const SizedBox(height: 4),
                  _MiniResourceBar(
                    value: resources?.mana ?? 0,
                    maxValue: resources?.maxMana ?? 0,
                    fillColor: const Color(0xFF1E6FA7),
                    loading: isLoading && resources == null,
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _MiniResourceBar extends StatelessWidget {
  const _MiniResourceBar({
    required this.value,
    required this.maxValue,
    required this.fillColor,
    required this.loading,
  });

  final int value;
  final int maxValue;
  final Color fillColor;
  final bool loading;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final normalizedMax = maxValue <= 0 ? 1 : maxValue;
    final progress = loading ? null : (value / normalizedMax).clamp(0.0, 1.0);

    return SizedBox(
      width: 54,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(999),
        child: LinearProgressIndicator(
          value: progress,
          minHeight: 5,
          color: fillColor,
          backgroundColor: theme.colorScheme.surfaceContainerHighest,
        ),
      ),
    );
  }
}

class _MemberResources {
  const _MemberResources({
    required this.health,
    required this.maxHealth,
    required this.mana,
    required this.maxMana,
  });

  final int health;
  final int maxHealth;
  final int mana;
  final int maxMana;
}
