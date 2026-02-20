import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:provider/provider.dart';

import '../providers/auth_provider.dart';
import '../providers/activity_feed_provider.dart';
import '../providers/friend_provider.dart';
import '../providers/party_provider.dart';
import '../providers/quest_log_provider.dart';
import '../providers/map_focus_provider.dart';
import '../providers/character_stats_provider.dart';
import '../providers/user_level_provider.dart';
import '../widgets/friends_tab_content.dart';
import '../widgets/inventory_panel.dart';
import '../widgets/party_tab_content.dart';
import '../widgets/quest_log_panel.dart';
import '../widgets/reputation_tab_content.dart';

class LayoutShell extends StatefulWidget {
  const LayoutShell({super.key, required this.child});

  final Widget child;

  @override
  State<LayoutShell> createState() => _LayoutShellState();
}

class _LayoutShellState extends State<LayoutShell> {
  final _scaffoldKey = GlobalKey<ScaffoldState>();

  @override
  Widget build(BuildContext context) {
    return _LogoutCleaner(
      scaffoldKey: _scaffoldKey,
      child: Scaffold(
        key: _scaffoldKey,
        endDrawer: const _SideDrawer(),
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
  });

  final Widget child;
  final GlobalKey<ScaffoldState> scaffoldKey;

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
                  onPressed: () => context
                      .go('/?from=${Uri.encodeComponent('/single-player')}'),
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
    final hasUnspentPoints = context.watch<CharacterStatsProvider>().hasUnspentPoints;
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
            backgroundImage: auth.user?.profilePictureUrl != null &&
                    auth.user!.profilePictureUrl.isNotEmpty
                ? NetworkImage(auth.user!.profilePictureUrl)
                : null,
            child: auth.user?.profilePictureUrl == null ||
                    auth.user!.profilePictureUrl.isEmpty
                ? const Icon(Icons.person)
                : null,
          ),
          if (hasUnspentPoints)
            Positioned(
              right: -2,
              top: -2,
              child: Container(
                padding: const EdgeInsets.all(4),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.primary,
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: Theme.of(context).colorScheme.surface,
                    width: 1.2,
                  ),
                ),
                child: Icon(
                  Icons.arrow_upward,
                  size: 12,
                  color: Theme.of(context).colorScheme.onPrimary,
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _SideDrawer extends StatefulWidget {
  const _SideDrawer();

  @override
  State<_SideDrawer> createState() => _SideDrawerState();
}

class _SideDrawerState extends State<_SideDrawer> {
  int _tabIndex = 1;

  void _selectTab(int index) {
    final shouldRefreshQuestLog = index == 2;
    final shouldRefreshActivityFeed = index == 0;
    final shouldRefreshCharacterStats = index == 0;
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
        WidgetsBinding.instance.addPostFrameCallback((_) {
          if (!mounted) return;
          context.read<CharacterStatsProvider>().refresh();
        });
      }
      return;
    }
    setState(() => _tabIndex = index);
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
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        context.read<CharacterStatsProvider>().refresh();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final drawerWidth = (screenWidth * 0.9).clamp(320.0, 520.0);
    final theme = Theme.of(context);
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
                            selected: _tabIndex == 0,
                            onTap: () => _selectTab(0),
                          ),
                          _DrawerMenuButton(
                            label: 'Inventory',
                            icon: Icons.inventory_2,
                            selected: _tabIndex == 1,
                            onTap: () => _selectTab(1),
                          ),
                          _DrawerMenuButton(
                            label: 'Quest Log',
                            icon: Icons.menu_book,
                            selected: _tabIndex == 2,
                            onTap: () => _selectTab(2),
                          ),
                          _DrawerMenuButton(
                            label: 'Party',
                            icon: Icons.groups,
                            selected: _tabIndex == 3,
                            onTap: () => _selectTab(3),
                          ),
                          _DrawerMenuButton(
                            label: 'Friends',
                            icon: Icons.people,
                            selected: _tabIndex == 4,
                            onTap: () => _selectTab(4),
                          ),
                          _DrawerMenuButton(
                            label: 'Reputation',
                            icon: Icons.stars,
                            selected: _tabIndex == 5,
                            onTap: () => _selectTab(5),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 12),
                    Expanded(
                      child: SizedBox.expand(
                        child: AnimatedSwitcher(
                          duration: const Duration(milliseconds: 200),
                          child: _tabIndex == 0
                              ? const _CharacterTabContent(
                                  key: ValueKey('character'),
                                )
                              : _tabIndex == 1
                                  ? InventoryPanel(
                                      key: const ValueKey('inventory'),
                                      onClose: () => Navigator.of(context).pop(),
                                    )
                                  : _tabIndex == 2
                                  ? QuestLogPanel(
                                      key: const ValueKey('quest-log'),
                                      onClose: () => Navigator.of(context).pop(),
                                      onFocusPoI: (poi) {
                                        context.read<MapFocusProvider>().focusPoi(poi);
                                        context.go('/single-player');
                                      },
                                      onFocusTurnInQuest: (quest) {
                                        context
                                            .read<MapFocusProvider>()
                                            .focusTurnInQuest(quest);
                                        context.go('/single-player');
                                      },
                                    )
                                  : _tabIndex == 3
                                      ? const PartyTabContent(
                                          key: ValueKey('party'),
                                        )
                                      : _tabIndex == 4
                                          ? const FriendsTabContent(
                                              key: ValueKey('friends'),
                                            )
                                          : const ReputationTabContent(
                                              key: ValueKey('reputation'),
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
    final textColor =
        selected ? theme.colorScheme.onPrimaryContainer : theme.colorScheme.onSurface;
    final bgColor =
        selected ? theme.colorScheme.primaryContainer : theme.colorScheme.surface;
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
                color:
                    selected ? theme.colorScheme.onPrimaryContainer : theme.colorScheme.onSurfaceVariant,
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

class _CharacterTabContent extends StatefulWidget {
  const _CharacterTabContent({super.key});

  @override
  State<_CharacterTabContent> createState() => _CharacterTabContentState();
}

class _CharacterTabContentState extends State<_CharacterTabContent> {
  static const Map<String, String> _labels = {
    'strength': 'Strength',
    'dexterity': 'Dexterity',
    'constitution': 'Constitution',
    'intelligence': 'Intelligence',
    'wisdom': 'Wisdom',
    'charisma': 'Charisma',
  };

  String? _lastUserId;
  Map<String, int> _pending = {};
  final ScrollController _scrollController = ScrollController();
  bool _showTopFade = false;
  bool _showBottomFade = false;

  int get _pendingTotal =>
      _pending.values.where((value) => value > 0).fold(0, (a, b) => a + b);

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_updateFades);
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final uid = context.watch<AuthProvider>().user?.id;
    if (uid != _lastUserId) {
      _lastUserId = uid;
      _pending = {};
    }
    final unspent = context.watch<CharacterStatsProvider>().unspentPoints;
    if (unspent == 0 && _pending.isNotEmpty) {
      _pending = {};
    }
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
  }

  @override
  void dispose() {
    _scrollController.removeListener(_updateFades);
    _scrollController.dispose();
    super.dispose();
  }

  void _bumpStat(String key, int delta, int remaining) {
    if (delta > 0 && remaining <= 0) return;
    final current = _pending[key] ?? 0;
    final next = (current + delta).clamp(0, 999);
    setState(() {
      if (next == 0) {
        _pending.remove(key);
      } else {
        _pending[key] = next;
      }
    });
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
  }

  Future<void> _confirmAllocations(CharacterStatsProvider stats) async {
    if (_pendingTotal == 0) return;
    final success = await stats.applyAllocations(_pending);
    if (!mounted) return;
    if (success) {
      setState(() => _pending = {});
      WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
      return;
    }
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Unable to apply stat points.')),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateFades());
    final auth = context.watch<AuthProvider>();
    final levels = context.watch<UserLevelProvider>();
    final statsProvider = context.watch<CharacterStatsProvider>();
    final u = auth.user;
    if (u == null) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: Text('Log in to see your character.'),
        ),
      );
    }
    final userLevel = levels.userLevel;
    final levelLoading = levels.loading;
    final displayLevel = userLevel?.level ?? statsProvider.level;

    void showProfileImage() {
      if (u.profilePictureUrl.isEmpty) return;
      showDialog<void>(
        context: context,
        barrierColor: Colors.black54,
        builder: (context) {
          final theme = Theme.of(context);
          return Dialog(
            backgroundColor: Colors.transparent,
            insetPadding: const EdgeInsets.all(24),
            child: Stack(
              alignment: Alignment.topRight,
              children: [
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.surface,
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(
                      color: theme.colorScheme.outlineVariant,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.2),
                        blurRadius: 18,
                        offset: const Offset(0, 10),
                      ),
                    ],
                  ),
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(16),
                    child: Image.network(
                      u.profilePictureUrl,
                      width: 320,
                      height: 320,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => Container(
                        width: 320,
                        height: 320,
                        color: theme.colorScheme.surfaceVariant,
                        child: const Icon(Icons.person, size: 96),
                      ),
                    ),
                  ),
                ),
                IconButton(
                  onPressed: () => Navigator.of(context).pop(),
                  icon: const Icon(Icons.close),
                  style: IconButton.styleFrom(
                    backgroundColor: theme.colorScheme.surfaceContainerHighest,
                    shape: const CircleBorder(),
                  ),
                ),
              ],
            ),
          );
        },
      );
    }

    return Stack(
      children: [
        SingleChildScrollView(
          controller: _scrollController,
          padding: const EdgeInsets.only(bottom: 24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              GestureDetector(
                onTap: u.profilePictureUrl.isNotEmpty ? showProfileImage : null,
                child: CircleAvatar(
                  radius: 28,
                  backgroundColor: Colors.grey.shade300,
                  backgroundImage: u.profilePictureUrl.isNotEmpty
                      ? NetworkImage(u.profilePictureUrl)
                      : null,
                  child: u.profilePictureUrl.isEmpty ? const Icon(Icons.person) : null,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      u.username.isNotEmpty ? u.username : u.name,
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    if (u.username.isNotEmpty && u.name != u.username)
                      Text(
                        u.name,
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: theme.colorScheme.surface,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: theme.colorScheme.outlineVariant),
            ),
            child: levelLoading
                ? Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      Text(
                        'Level',
                        style: theme.textTheme.titleSmall?.copyWith(
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                      const SizedBox(height: 10),
                      LinearProgressIndicator(
                        minHeight: 8,
                        color: theme.colorScheme.primary,
                        backgroundColor: theme.colorScheme.surfaceContainerHighest,
                      ),
                    ],
                  )
                : userLevel == null
                    ? Text(
                        'Level data unavailable right now.',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      )
                    : Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Row(
                            children: [
                              Text(
                                'Level ${userLevel.level}',
                                style: theme.textTheme.titleSmall?.copyWith(
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                              const Spacer(),
                              Text(
                                '${userLevel.experiencePointsOnLevel} / ${userLevel.experienceToNextLevel} XP',
                                style: theme.textTheme.bodySmall?.copyWith(
                                  color: theme.colorScheme.onSurfaceVariant,
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 8),
                          ClipRRect(
                            borderRadius: BorderRadius.circular(999),
                            child: LinearProgressIndicator(
                              value: userLevel.experienceToNextLevel > 0
                                  ? (userLevel.experiencePointsOnLevel /
                                          userLevel.experienceToNextLevel)
                                      .clamp(0.0, 1.0)
                                  : 0.0,
                              minHeight: 8,
                              color: theme.colorScheme.primary,
                              backgroundColor: theme.colorScheme.surfaceContainerHighest,
                            ),
                          ),
                        ],
                      ),
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: theme.colorScheme.surface,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: theme.colorScheme.outlineVariant),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  'Character stats',
                  style: theme.textTheme.titleSmall?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Text(
                      'Level $displayLevel',
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    const Spacer(),
                    Text(
                      'Unspent: ${statsProvider.unspentPoints}',
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: statsProvider.hasUnspentPoints
                            ? theme.colorScheme.primary
                            : theme.colorScheme.onSurfaceVariant,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                ),
                if (_pendingTotal > 0) ...[
                  const SizedBox(height: 6),
                  Text(
                    'Remaining after pending: ${statsProvider.unspentPoints - _pendingTotal}',
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
                const SizedBox(height: 12),
                Column(
                  children: CharacterStatsProvider.statKeys.map((key) {
                    final label = _labels[key] ?? key;
                    final baseValue = statsProvider.stats[key] ??
                        CharacterStatsProvider.baseStatValue;
                    final pendingValue = _pending[key] ?? 0;
                    final displayValue = baseValue + pendingValue;
                    final remaining = statsProvider.unspentPoints - _pendingTotal;
                    final canAdd = remaining > 0;
                    final canRemove = pendingValue > 0;
                    return Container(
                      margin: const EdgeInsets.only(bottom: 8),
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 8,
                      ),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.surfaceContainerHighest,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: theme.colorScheme.outlineVariant),
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  label,
                                  style: theme.textTheme.bodyMedium?.copyWith(
                                    fontWeight: FontWeight.w700,
                                  ),
                                ),
                                if (pendingValue > 0)
                                  Text(
                                    '+$pendingValue pending',
                                    style: theme.textTheme.bodySmall?.copyWith(
                                      color: theme.colorScheme.primary,
                                    ),
                                  ),
                              ],
                            ),
                          ),
                          Text(
                            '$displayValue',
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                          const SizedBox(width: 8),
                          IconButton(
                            visualDensity: VisualDensity.compact,
                            onPressed: canRemove
                                ? () => _bumpStat(key, -1, remaining)
                                : null,
                            icon: const Icon(Icons.remove_circle_outline),
                          ),
                          IconButton(
                            visualDensity: VisualDensity.compact,
                            onPressed: canAdd
                                ? () => _bumpStat(key, 1, remaining)
                                : null,
                            icon: const Icon(Icons.add_circle_outline),
                          ),
                        ],
                      ),
                    );
                  }).toList(),
                ),
                if (_pendingTotal > 0) ...[
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Expanded(
                        child: OutlinedButton(
                          onPressed: () => setState(() => _pending = {}),
                          child: const Text('Cancel'),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: FilledButton(
                          onPressed: () => _confirmAllocations(statsProvider),
                          child: const Text('Confirm'),
                        ),
                      ),
                    ],
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(height: 16),
          OutlinedButton(
            onPressed: () {
              Navigator.of(context).pop();
              context.go('/logout');
            },
            child: const Text('Log out'),
          ),
            ],
          ),
        ),
        if (_showTopFade)
          Positioned(
            left: 0,
            right: 0,
            top: 0,
            child: IgnorePointer(
              child: Container(
                height: 18,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    colors: [
                      theme.colorScheme.surface,
                      theme.colorScheme.surface.withOpacity(0.0),
                    ],
                  ),
                ),
              ),
            ),
          ),
        if (_showBottomFade)
          Positioned(
            left: 0,
            right: 0,
            bottom: 0,
            child: IgnorePointer(
              child: Container(
                height: 20,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.bottomCenter,
                    end: Alignment.topCenter,
                    colors: [
                      theme.colorScheme.surface,
                      theme.colorScheme.surface.withOpacity(0.0),
                    ],
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }

  void _updateFades() {
    if (!mounted || !_scrollController.hasClients) return;
    final position = _scrollController.position;
    final showTop = position.pixels > 0.5;
    final showBottom = position.pixels < position.maxScrollExtent - 0.5;
    if (showTop == _showTopFade && showBottom == _showBottomFade) return;
    setState(() {
      _showTopFade = showTop;
      _showBottomFade = showBottom;
    });
  }
}
