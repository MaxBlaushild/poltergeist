import 'dart:async';

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';

import '../models/user.dart';
import '../providers/auth_provider.dart';
import '../providers/activity_feed_provider.dart';
import '../providers/friend_provider.dart';
import '../providers/party_provider.dart';
import '../providers/quest_log_provider.dart';
import '../providers/map_focus_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/completed_task_provider.dart';
import '../services/push_notification_service.dart';
import '../widgets/abilities_tab_content.dart';
import '../widgets/character_tab_content.dart';
import '../widgets/friends_tab_content.dart';
import '../widgets/inventory_panel.dart';
import '../widgets/party_tab_content.dart';
import '../widgets/quest_log_panel.dart';
import '../widgets/reputation_tab_content.dart';
import '../widgets/settings_tab_content.dart';
import 'user_character_screen.dart';

class LayoutShell extends StatefulWidget {
  const LayoutShell({super.key, required this.child});

  final Widget child;

  @override
  State<LayoutShell> createState() => _LayoutShellState();
}

class _LayoutShellState extends State<LayoutShell> {
  final _scaffoldKey = GlobalKey<ScaffoldState>();
  final _sideDrawerKey = GlobalKey<_SideDrawerState>();
  StreamSubscription<InAppPushEvent>? _pushEventSubscription;
  bool _drainingPartySubmissionResults = false;
  int? _pendingDrawerTabIndex;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      unawaited(_initializeForegroundInviteToasts());
    });
  }

  @override
  void dispose() {
    _pushEventSubscription?.cancel();
    _pushEventSubscription = null;
    super.dispose();
  }

  Future<void> _initializeForegroundInviteToasts() async {
    final pushService = context.read<PushNotificationService>();
    _pushEventSubscription ??= pushService.inAppEvents.listen(
      _handleInAppPushEvent,
    );
    await pushService.initializeForegroundMessageHandling();
    if (!mounted) return;
    await _drainPendingPartySubmissionResults();
  }

  void _handleInAppPushEvent(InAppPushEvent event) {
    if (!mounted) return;
    final type = event.type.trim();
    if (type != 'monster_battle_invite' &&
        type != 'party_invite' &&
        type != 'friend_invite' &&
        type != 'friend_invite_accepted' &&
        type != 'party_submission_result' &&
        type != 'push_test') {
      return;
    }

    final title = _toastTitleFor(event);
    final body = _toastBodyFor(event);
    final actionLabel = _toastActionLabelFor(type);

    if (type == 'monster_battle_invite') {
      unawaited(context.read<ActivityFeedProvider>().refresh());
    }
    if (type == 'party_invite') {
      if (event.openedFromNotification) {
        unawaited(_openPartyInviteDestination());
        return;
      }
      unawaited(context.read<PartyProvider>().fetchPartyInvites());
    }
    if (type == 'friend_invite' || type == 'friend_invite_accepted') {
      if (event.openedFromNotification) {
        unawaited(_openFriendsDestination());
        return;
      }
    }
    if (type == 'party_submission_result') {
      unawaited(_drainPendingPartySubmissionResults());
    }

    final messenger = ScaffoldMessenger.of(context);
    messenger.hideCurrentSnackBar();
    messenger.showSnackBar(
      SnackBar(
        content: Text('$title: $body'),
        duration: const Duration(seconds: 8),
        action: actionLabel == null
            ? null
            : SnackBarAction(
                label: actionLabel,
                onPressed: () => _openDrawerForPushType(type),
              ),
      ),
    );
  }

  Future<void> _openPartyInviteDestination() async {
    try {
      await context.read<PartyProvider>().fetchPartyInvites();
    } catch (_) {
      // Continue to party tab even if invite refresh fails.
    }
    if (!mounted) return;
    _openDrawerForPushType('party_invite');
  }

  Future<void> _openFriendsDestination() async {
    try {
      await context.read<FriendProvider>().refresh();
    } catch (_) {
      // Continue to friends tab even if refresh fails.
    }
    if (!mounted) return;
    _openDrawerForPushType('friend_invite');
  }

  String _toastTitleFor(InAppPushEvent event) {
    if (event.title.trim().isNotEmpty) return event.title.trim();
    switch (event.type.trim()) {
      case 'monster_battle_invite':
        return 'Party Combat Invite';
      case 'party_invite':
        return 'Party Invite';
      case 'friend_invite':
        return 'Friend Invite';
      case 'friend_invite_accepted':
        return 'Friend Invite Accepted';
      case 'party_submission_result':
        return 'Party Result Ready';
      case 'push_test':
        return 'Push Test';
      default:
        return 'Notification';
    }
  }

  String _toastBodyFor(InAppPushEvent event) {
    if (event.body.trim().isNotEmpty) return event.body.trim();
    switch (event.type.trim()) {
      case 'monster_battle_invite':
        return 'A party member invited you to join combat.';
      case 'party_invite':
        return 'You received a party invite.';
      case 'friend_invite':
        return 'You received a friend invite.';
      case 'friend_invite_accepted':
        return 'A friend invite was accepted.';
      case 'party_submission_result':
        return 'A party challenge or scenario result is ready.';
      case 'push_test':
        return 'Foreground push received.';
      default:
        return 'You received a notification.';
    }
  }

  String? _toastActionLabelFor(String type) {
    switch (type) {
      case 'monster_battle_invite':
      case 'party_invite':
      case 'friend_invite':
      case 'friend_invite_accepted':
      case 'party_submission_result':
        return 'Open';
      default:
        return null;
    }
  }

  void _openDrawerForPushType(String type) {
    _pendingDrawerTabIndex = _tabForPushType(type);
    _scaffoldKey.currentState?.openEndDrawer();
  }

  int _tabForPushType(String type) {
    switch (type) {
      case 'party_invite':
        return _SideDrawerState._partyTab;
      case 'friend_invite':
      case 'friend_invite_accepted':
        return _SideDrawerState._friendsTab;
      case 'party_submission_result':
      case 'monster_battle_invite':
      default:
        return _SideDrawerState._characterTab;
    }
  }

  Future<void> _drainPendingPartySubmissionResults() async {
    if (!mounted || _drainingPartySubmissionResults) return;
    final userId = context.read<AuthProvider>().user?.id;
    if (userId == null || userId.trim().isEmpty) return;
    _drainingPartySubmissionResults = true;
    try {
      final results = await context
          .read<PushNotificationService>()
          .getPendingPartySubmissionResults();
      if (!mounted || results.isEmpty) return;
      final provider = context.read<CompletedTaskProvider>();
      for (final result in results) {
        final type = result.type.trim();
        if (type.isEmpty) continue;
        provider.showModal(type, data: result.data);
      }
    } catch (_) {
    } finally {
      _drainingPartySubmissionResults = false;
    }
  }

  @override
  Widget build(BuildContext context) {
    return _LogoutCleaner(
      scaffoldKey: _scaffoldKey,
      onUserLoggedIn: () {
        unawaited(_drainPendingPartySubmissionResults());
      },
      child: Scaffold(
        key: _scaffoldKey,
        onEndDrawerChanged: (isOpened) {
          if (!isOpened) return;
          final pendingTab = _pendingDrawerTabIndex;
          if (pendingTab != null) {
            _sideDrawerKey.currentState?._selectTab(pendingTab);
            _pendingDrawerTabIndex = null;
          }
          _sideDrawerKey.currentState?.handleDrawerOpened();
        },
        endDrawer: _SideDrawer(key: _sideDrawerKey),
        body: SafeArea(
          top: false,
          bottom: false,
          left: false,
          right: false,
          child: Column(
            children: [
              _LayoutHeader(),
              Expanded(child: widget.child),
            ],
          ),
        ),
      ),
    );
  }
}

class _LogoutCleaner extends StatefulWidget {
  const _LogoutCleaner({
    required this.child,
    required this.scaffoldKey,
    this.onUserLoggedIn,
  });

  final Widget child;
  final GlobalKey<ScaffoldState> scaffoldKey;
  final VoidCallback? onUserLoggedIn;

  @override
  State<_LogoutCleaner> createState() => _LogoutCleanerState();
}

class _LogoutCleanerState extends State<_LogoutCleaner> {
  String? _lastUserId;

  @override
  Widget build(BuildContext context) {
    final user = context.watch<AuthProvider>().user;
    final uid = user?.id;
    if (uid == null && _lastUserId != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        context.read<PartyProvider>().clear();
        context.read<FriendProvider>().clear();
        widget.scaffoldKey.currentState?.closeEndDrawer();
      });
      _lastUserId = null;
    } else if (uid != null) {
      final isNewUser = _lastUserId != uid;
      if (isNewUser) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          widget.onUserLoggedIn?.call();
        });
      }
      _lastUserId = uid;
    }
    return widget.child;
  }
}

class _LayoutHeader extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Consumer<AuthProvider>(
      builder: (context, auth, _) {
        final theme = Theme.of(context);
        final surfaceColor = theme.colorScheme.surface.withValues(alpha: 0.95);
        final topPadding = MediaQuery.of(context).padding.top;
        return Container(
          height: 52 + topPadding,
          padding: EdgeInsets.fromLTRB(16, topPadding + 4, 16, 4),
          decoration: BoxDecoration(
            color: surfaceColor,
            border: Border(
              bottom: BorderSide(
                color: theme.colorScheme.outlineVariant,
                width: 1.5,
              ),
            ),
            boxShadow: [
              BoxShadow(
                color: const Color(0x332D2416),
                blurRadius: 12,
                offset: const Offset(0, 4),
              ),
            ],
          ),
          child: Row(
            children: [
              GestureDetector(
                onTap: () => context.go('/'),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      'unclaimed streets',
                      style: GoogleFonts.cinzelDecorative(
                        fontWeight: FontWeight.w700,
                        fontSize: 24,
                        letterSpacing: 0.6,
                        color: theme.colorScheme.onSurface,
                      ),
                    ),
                  ],
                ),
              ),
              const Spacer(),
              if (auth.user != null)
                _UserAvatar(auth: auth)
              else if (!auth.loading)
                TextButton(
                  onPressed: () => context.go(
                    '/?from=${Uri.encodeComponent('/single-player')}',
                  ),
                  child: const Text('Log in'),
                ),
            ],
          ),
        );
      },
    );
  }
}

class _UserAvatar extends StatelessWidget {
  const _UserAvatar({required this.auth});

  final AuthProvider auth;

  @override
  Widget build(BuildContext context) {
    final hasUnspentPoints = context
        .watch<CharacterStatsProvider>()
        .hasUnspentPoints;
    return GestureDetector(
      onTap: () {
        Scaffold.of(context).openEndDrawer();
      },
      child: Stack(
        clipBehavior: Clip.none,
        children: [
          CircleAvatar(
            radius: 20,
            backgroundColor: Colors.grey.shade300,
            backgroundImage:
                auth.user?.profilePictureUrl != null &&
                    auth.user!.profilePictureUrl.isNotEmpty
                ? NetworkImage(auth.user!.profilePictureUrl)
                : null,
            child:
                auth.user?.profilePictureUrl == null ||
                    auth.user!.profilePictureUrl.isEmpty
                ? const Icon(Icons.person)
                : null,
          ),
          if (hasUnspentPoints)
            Positioned(
              right: -2,
              top: -2,
              child: Container(
                padding: const EdgeInsets.all(3),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFD54F),
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: Theme.of(context).colorScheme.surface,
                    width: 1.2,
                  ),
                ),
                child: const Icon(
                  Icons.arrow_upward,
                  size: 10,
                  color: Colors.white,
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _SideDrawer extends StatefulWidget {
  const _SideDrawer({super.key});

  @override
  State<_SideDrawer> createState() => _SideDrawerState();
}

class _SideDrawerState extends State<_SideDrawer> {
  static const int _characterTab = 0;
  static const int _inventoryTab = 1;
  static const int _questLogTab = 2;
  static const int _partyTab = 3;
  static const int _friendsTab = 4;
  static const int _reputationTab = 5;
  static const int _abilitiesTab = 6;
  static const int _settingsTab = 7;

  int _tabIndex = 0;
  User? _profileUser;

  void handleDrawerOpened() {
    _refreshCharacterStatsIfVisible();
  }

  void _refreshCharacterStatsIfVisible() {
    final showsCharacterOrAbilities =
        _tabIndex == _characterTab || _tabIndex == _abilitiesTab;
    if (_profileUser != null || !showsCharacterOrAbilities) return;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      context.read<CharacterStatsProvider>().refresh(silent: true);
    });
  }

  void _selectTab(int index) {
    final shouldRefreshQuestLog = index == _questLogTab;
    final shouldRefreshActivityFeed = index == _characterTab;
    final shouldRefreshCharacterStats =
        index == _characterTab || index == _abilitiesTab;
    if (_tabIndex == index) {
      if (shouldRefreshQuestLog) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (!mounted) return;
          context.read<QuestLogProvider>().refresh();
        });
      }
      if (shouldRefreshActivityFeed) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (!mounted) return;
          context.read<ActivityFeedProvider>().refresh();
        });
      }
      if (shouldRefreshCharacterStats) {
        _refreshCharacterStatsIfVisible();
      }
      return;
    }
    setState(() {
      _tabIndex = index;
      _profileUser = null;
    });
    if (shouldRefreshQuestLog) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        context.read<QuestLogProvider>().refresh();
      });
    }
    if (shouldRefreshActivityFeed) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        context.read<ActivityFeedProvider>().refresh();
      });
    }
    if (shouldRefreshCharacterStats) {
      _refreshCharacterStatsIfVisible();
    }
  }

  void _openProfile(User user) {
    setState(() => _profileUser = user);
  }

  void _closeProfile() {
    setState(() => _profileUser = null);
    _refreshCharacterStatsIfVisible();
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final drawerWidth = (screenWidth * 0.9).clamp(320.0, 520.0);
    final theme = Theme.of(context);
    final profileUser = _profileUser;
    return Drawer(
      width: drawerWidth,
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.surfaceContainerHighest,
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                          color: theme.colorScheme.outlineVariant,
                          width: 1.2,
                        ),
                      ),
                      child: Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          _DrawerMenuButton(
                            label: 'Character',
                            icon: Icons.person,
                            selected: _tabIndex == _characterTab,
                            onTap: () => _selectTab(_characterTab),
                          ),
                          _DrawerMenuButton(
                            label: 'Abilities',
                            icon: Icons.auto_fix_high,
                            selected: _tabIndex == _abilitiesTab,
                            onTap: () => _selectTab(_abilitiesTab),
                          ),
                          _DrawerMenuButton(
                            label: 'Inventory',
                            icon: Icons.inventory_2,
                            selected: _tabIndex == _inventoryTab,
                            onTap: () => _selectTab(_inventoryTab),
                          ),
                          _DrawerMenuButton(
                            label: 'Quest Log',
                            icon: Icons.menu_book,
                            selected: _tabIndex == _questLogTab,
                            onTap: () => _selectTab(_questLogTab),
                          ),
                          _DrawerMenuButton(
                            label: 'Party',
                            icon: Icons.groups,
                            selected: _tabIndex == _partyTab,
                            onTap: () => _selectTab(_partyTab),
                          ),
                          _DrawerMenuButton(
                            label: 'Friends',
                            icon: Icons.people,
                            selected: _tabIndex == _friendsTab,
                            onTap: () => _selectTab(_friendsTab),
                          ),
                          _DrawerMenuButton(
                            label: 'Reputation',
                            icon: Icons.stars,
                            selected: _tabIndex == _reputationTab,
                            onTap: () => _selectTab(_reputationTab),
                          ),
                          _DrawerMenuButton(
                            label: 'Settings',
                            icon: Icons.settings,
                            selected: _tabIndex == _settingsTab,
                            onTap: () => _selectTab(_settingsTab),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 12),
                    Expanded(
                      child: SizedBox.expand(
                        child: AnimatedSwitcher(
                          duration: const Duration(milliseconds: 200),
                          child: profileUser != null
                              ? _DrawerCharacterProfile(
                                  key: ValueKey('profile-${profileUser.id}'),
                                  user: profileUser,
                                  onBack: _closeProfile,
                                )
                              : _tabIndex == _characterTab
                              ? const CharacterTabContent(
                                  key: ValueKey('character'),
                                )
                              : _tabIndex == _inventoryTab
                              ? InventoryPanel(
                                  key: const ValueKey('inventory'),
                                  onClose: () => Navigator.of(context).pop(),
                                )
                              : _tabIndex == _questLogTab
                              ? QuestLogPanel(
                                  key: const ValueKey('quest-log'),
                                  onClose: () => Navigator.of(context).pop(),
                                  onFocusPoI: (poi) {
                                    context.read<MapFocusProvider>().focusPoi(
                                      poi,
                                    );
                                    context.go('/single-player');
                                  },
                                  onFocusTurnInQuest: (quest) {
                                    context
                                        .read<MapFocusProvider>()
                                        .focusTurnInQuest(quest);
                                    context.go('/single-player');
                                  },
                                )
                              : _tabIndex == _partyTab
                              ? PartyTabContent(
                                  key: const ValueKey('party'),
                                  onViewProfile: _openProfile,
                                )
                              : _tabIndex == _friendsTab
                              ? FriendsTabContent(
                                  key: const ValueKey('friends'),
                                  onViewProfile: _openProfile,
                                )
                              : _tabIndex == _abilitiesTab
                              ? const AbilitiesTabContent(
                                  key: ValueKey('abilities'),
                                )
                              : _tabIndex == _reputationTab
                              ? const ReputationTabContent(
                                  key: ValueKey('reputation'),
                                )
                              : const Align(
                                  alignment: Alignment.topCenter,
                                  child: SettingsTabContent(
                                    key: ValueKey('settings'),
                                  ),
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
    );
  }
}

class _DrawerCharacterProfile extends StatelessWidget {
  const _DrawerCharacterProfile({
    super.key,
    required this.user,
    required this.onBack,
  });

  final User user;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final scheme = theme.colorScheme;
    final displayName = user.username.isNotEmpty ? user.username : user.name;
    final secondaryName = user.username.isNotEmpty && user.name != user.username
        ? user.name
        : null;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Material(
          color: scheme.surface,
          elevation: 1,
          shadowColor: const Color(0x332D2416),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
            side: BorderSide(color: scheme.outlineVariant),
          ),
          child: Padding(
            padding: const EdgeInsets.fromLTRB(8, 8, 12, 8),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                IconButton(
                  onPressed: onBack,
                  icon: const Icon(Icons.arrow_back),
                  tooltip: 'Back',
                ),
                const SizedBox(width: 4),
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.only(top: 6, bottom: 2),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisSize: MainAxisSize.min,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          'Character',
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                        if (displayName.isNotEmpty)
                          Text(
                            displayName,
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: scheme.onSurfaceVariant,
                            ),
                          ),
                        if (secondaryName != null)
                          Text(
                            secondaryName,
                            style: theme.textTheme.bodySmall?.copyWith(
                              color: scheme.onSurfaceVariant,
                            ),
                          ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 12),
        Expanded(
          child: UserCharacterScreen(
            key: ValueKey('drawer-character-${user.id}'),
            userId: user.id,
          ),
        ),
      ],
    );
  }
}

class _DrawerMenuButton extends StatelessWidget {
  const _DrawerMenuButton({
    required this.label,
    required this.icon,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final accent = theme.colorScheme.primary;
    final textColor = selected
        ? theme.colorScheme.onPrimaryContainer
        : theme.colorScheme.onSurface;
    final bgColor = selected
        ? theme.colorScheme.primaryContainer
        : theme.colorScheme.surface;
    final borderColor = selected ? accent : theme.colorScheme.outlineVariant;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(999),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 160),
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          decoration: BoxDecoration(
            color: bgColor,
            borderRadius: BorderRadius.circular(999),
            border: Border.all(color: borderColor, width: selected ? 1.4 : 1.0),
            boxShadow: selected
                ? const [
                    BoxShadow(
                      color: Color(0x1F2D2416),
                      blurRadius: 8,
                      offset: Offset(0, 3),
                    ),
                  ]
                : const [],
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                icon,
                size: 18,
                color: selected
                    ? theme.colorScheme.onPrimaryContainer
                    : theme.colorScheme.onSurfaceVariant,
              ),
              const SizedBox(width: 8),
              Text(
                label,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: textColor,
                  fontWeight: selected ? FontWeight.w700 : FontWeight.w500,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
