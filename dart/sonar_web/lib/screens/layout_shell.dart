import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../providers/auth_provider.dart';
import '../providers/friend_provider.dart';
import '../providers/party_provider.dart';
import '../widgets/character_tab_content.dart';
import '../widgets/friends_tab_content.dart';
import '../widgets/party_tab_content.dart';

class LayoutShell extends StatelessWidget {
  const LayoutShell({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return _LogoutCleaner(
      child: Scaffold(
        endDrawer: const _SideDrawer(),
        body: SafeArea(
          top: false,
          bottom: false,
          left: false,
          right: false,
          child: Column(
            children: [
              _LayoutHeader(),
              Expanded(child: child),
            ],
          ),
        ),
      ),
    );
  }
}

class _LogoutCleaner extends StatefulWidget {
  const _LogoutCleaner({required this.child});

  final Widget child;

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
          height: 61 + topPadding,
          padding: EdgeInsets.fromLTRB(16, topPadding + 8, 16, 8),
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
                      style: TextStyle(
                        fontFamily: 'Cinzel',
                        fontWeight: FontWeight.bold,
                        fontSize: 24,
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
    return GestureDetector(
      onTap: () {
        Scaffold.of(context).openEndDrawer();
      },
      child: CircleAvatar(
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
    );
  }
}

class _SideDrawer extends StatefulWidget {
  const _SideDrawer();

  @override
  State<_SideDrawer> createState() => _SideDrawerState();
}

class _SideDrawerState extends State<_SideDrawer> {
  int _tabIndex = 0;

  @override
  Widget build(BuildContext context) {
    return Drawer(
      child: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            const Text(
              'Profile',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 16),
            Consumer<AuthProvider>(
              builder: (context, auth, _) {
                final u = auth.user;
                if (u == null) return const SizedBox.shrink();
                return Row(
                  children: [
                    CircleAvatar(
                      radius: 24,
                      backgroundColor: Colors.grey.shade300,
                      backgroundImage:
                          u.profilePictureUrl.isNotEmpty
                              ? NetworkImage(u.profilePictureUrl)
                              : null,
                      child: u.profilePictureUrl.isEmpty
                          ? const Icon(Icons.person)
                          : null,
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            u.username.isNotEmpty ? u.username : u.name,
                            style: const TextStyle(
                              fontWeight: FontWeight.w600,
                              fontSize: 16,
                            ),
                          ),
                          if (u.username.isNotEmpty && u.name != u.username)
                            Text(
                              u.name,
                              style: TextStyle(
                                fontSize: 12,
                                color: Colors.grey.shade600,
                              ),
                            ),
                        ],
                      ),
                    ),
                  ],
                );
              },
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                _TabButton(
                  label: 'Character',
                  selected: _tabIndex == 0,
                  onTap: () => setState(() => _tabIndex = 0),
                ),
                _TabButton(
                  label: 'Party',
                  selected: _tabIndex == 1,
                  onTap: () => setState(() => _tabIndex = 1),
                ),
                _TabButton(
                  label: 'Friends',
                  selected: _tabIndex == 2,
                  onTap: () => setState(() => _tabIndex = 2),
                ),
              ],
            ),
            const SizedBox(height: 8),
            AnimatedSwitcher(
              duration: const Duration(milliseconds: 200),
              child: _tabIndex == 0
                  ? const CharacterTabContent(key: ValueKey('char'))
                  : _tabIndex == 1
                      ? const PartyTabContent(key: ValueKey('party'))
                      : const FriendsTabContent(key: ValueKey('friends')),
            ),
            const Divider(),
            TextButton(
              onPressed: () {
                Navigator.of(context).pop();
                context.go('/create-point-of-interest');
              },
              child: const Text('Create POI'),
            ),
            TextButton(
              onPressed: () {
                Navigator.of(context).pop();
                context.go('/adminfuckoff');
              },
              child: const Text('Admin'),
            ),
            TextButton(
              onPressed: () {
                Navigator.of(context).pop();
                context.go('/logout');
              },
              child: const Text('Log out'),
            ),
          ],
        ),
      ),
    );
  }
}

class _TabButton extends StatelessWidget {
  const _TabButton({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: TextButton(
        onPressed: onTap,
        style: TextButton.styleFrom(
          foregroundColor: selected
              ? Theme.of(context).colorScheme.primary
              : Colors.grey.shade600,
        ),
        child: Text(
          label,
          style: TextStyle(
            fontWeight: selected ? FontWeight.bold : FontWeight.normal,
          ),
        ),
      ),
    );
  }
}
