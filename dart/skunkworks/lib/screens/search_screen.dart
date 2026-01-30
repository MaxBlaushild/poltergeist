import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/models/friend.dart';
import 'package:skunkworks/models/user.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/providers/friend_provider.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/screens/profile_screen.dart';

class SearchScreen extends StatefulWidget {
  final Function(NavTab) onNavigate;

  const SearchScreen({
    super.key,
    required this.onNavigate,
  });

  @override
  State<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen> {
  final _searchController = TextEditingController();
  String _currentUserId = '';

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final authProvider = context.read<AuthProvider>();
      _currentUserId = authProvider.user?.id ?? '';
      final friendProvider = context.read<FriendProvider>();
      friendProvider.loadInvites();
      friendProvider.loadFriends();
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _onSearchChanged(String query) {
    if (query.trim().isEmpty) {
      context.read<FriendProvider>().searchUsers('');
    } else {
      context.read<FriendProvider>().searchUsers(query);
    }
  }

  bool _isFriend(List<User> friends, String userId) {
    return friends.any((friend) => friend.id == userId);
  }

  bool _hasPendingInvite(List<FriendInvite> invites, String userId) {
    return invites.any((invite) =>
        (invite.inviterID == _currentUserId && invite.inviteeID == userId) ||
        (invite.inviteeID == _currentUserId && invite.inviterID == userId));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: const Text(
          'Search',
          style: TextStyle(
            color: Colors.black,
            fontWeight: FontWeight.w600,
            fontSize: 18,
          ),
        ),
      ),
      body: Consumer<FriendProvider>(
        builder: (context, friendProvider, child) {
          // Get incoming invites (for current user)
          final incomingInvites = friendProvider.friendInvites
              .where((invite) => invite.inviteeID == _currentUserId)
              .toList();

          // Get outgoing invites (sent by current user)
          final outgoingInvites = friendProvider.friendInvites
              .where((invite) => invite.inviterID == _currentUserId)
              .toList();

          return Column(
            children: [
              // Search bar
              Padding(
                padding: const EdgeInsets.all(16.0),
                child: TextField(
                  controller: _searchController,
                  decoration: InputDecoration(
                    hintText: 'Search users...',
                    prefixIcon: const Icon(Icons.search),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                    filled: true,
                    fillColor: Colors.grey.shade100,
                  ),
                  onChanged: _onSearchChanged,
                ),
              ),
              
              // Friend invites section
              if (incomingInvites.isNotEmpty || outgoingInvites.isNotEmpty) ...[
                const Padding(
                  padding: EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      'Friend Requests',
                      style: TextStyle(
                        fontWeight: FontWeight.w600,
                        fontSize: 16,
                      ),
                    ),
                  ),
                ),
                // Incoming invites
                ...incomingInvites.map((invite) => _buildInviteItem(
                  invite,
                  friendProvider,
                  true,
                )),
                // Outgoing invites
                ...outgoingInvites.map((invite) => _buildInviteItem(
                  invite,
                  friendProvider,
                  false,
                )),
                const Divider(),
              ],

              // Search results or friends
              Expanded(
                child: _searchController.text.trim().isEmpty
                    ? _buildFriendsList(friendProvider)
                    : _buildSearchResults(friendProvider),
              ),
            ],
          );
        },
      ),
      bottomNavigationBar: BottomNav(
        currentTab: NavTab.search,
        onTabChanged: widget.onNavigate,
      ),
    );
  }

  Widget _buildFriendsList(FriendProvider friendProvider) {
    if (friendProvider.loading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (friendProvider.friends.isEmpty) {
      return const Center(
        child: Text(
          'No friends yet.\nSearch for users to add friends!',
          textAlign: TextAlign.center,
          style: TextStyle(color: Colors.grey),
        ),
      );
    }

    return ListView.builder(
      itemCount: friendProvider.friends.length,
      itemBuilder: (context, index) {
        final friend = friendProvider.friends[index];
        return _buildUserItem(friend, friendProvider, true);
      },
    );
  }

  Widget _buildSearchResults(FriendProvider friendProvider) {
    if (friendProvider.searchResults.isEmpty) {
      return const Center(
        child: Text(
          'No users found',
          style: TextStyle(color: Colors.grey),
        ),
      );
    }

    return ListView.builder(
      itemCount: friendProvider.searchResults.length,
      itemBuilder: (context, index) {
        final user = friendProvider.searchResults[index];
        // Don't show current user in search results
        if (user.id == _currentUserId) return const SizedBox.shrink();
        return _buildUserItem(
          user,
          friendProvider,
          _isFriend(friendProvider.friends, user.id ?? ''),
        );
      },
    );
  }

  Widget _buildUserItem(User user, FriendProvider friendProvider, bool isFriend) {
    final hasPendingInvite = _hasPendingInvite(
      friendProvider.friendInvites,
      user.id ?? '',
    );
    final username = user.username ?? user.phoneNumber ?? 'Unknown';

    return ListTile(
      leading: CircleAvatar(
        radius: 20,
        backgroundColor: Colors.grey.shade300,
        backgroundImage: user.profilePictureUrl != null
            ? NetworkImage(user.profilePictureUrl!)
            : null,
        child: user.profilePictureUrl == null
            ? Text(
                username.isNotEmpty ? username[0].toUpperCase() : 'U',
                style: const TextStyle(color: Colors.grey),
              )
            : null,
      ),
      title: Text(
        username,
        style: const TextStyle(fontWeight: FontWeight.w500),
      ),
      trailing: isFriend
          ? const Text(
              'Friends',
              style: TextStyle(color: Colors.grey),
            )
          : hasPendingInvite
              ? const Text(
                  'Pending',
                  style: TextStyle(color: Colors.grey),
                )
              : TextButton(
                  onPressed: () async {
                    try {
                      await friendProvider.sendInvite(user.id!);
                      await friendProvider.loadInvites();
                      if (mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(content: Text('Friend request sent')),
                        );
                      }
                    } catch (e) {
                      if (mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text('Failed to send request: $e')),
                        );
                      }
                    }
                  },
                  child: const Text('Add'),
                ),
      onTap: () {
        if (user.id != null) {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => ProfileScreen(
                userId: user.id!,
                user: user,
                onNavigate: widget.onNavigate,
              ),
            ),
          );
        }
      },
    );
  }

  Widget _buildInviteItem(
    FriendInvite invite,
    FriendProvider friendProvider,
    bool isIncoming,
  ) {
    // For now, we'll show a simple invite item
    // In a real app, you'd fetch the inviter/invitee user details
    return ListTile(
      title: Text(isIncoming ? 'Incoming request' : 'Outgoing request'),
      trailing: isIncoming
          ? Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextButton(
                  onPressed: () async {
                    try {
                      await friendProvider.acceptInvite(invite.id!);
                      if (mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(content: Text('Friend request accepted')),
                        );
                      }
                    } catch (e) {
                      if (mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text('Failed to accept: $e')),
                        );
                      }
                    }
                  },
                  child: const Text('Accept'),
                ),
                TextButton(
                  onPressed: () async {
                    try {
                      await friendProvider.deleteInvite(invite.id!);
                      if (mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(content: Text('Request declined')),
                        );
                      }
                    } catch (e) {
                      if (mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text('Failed to decline: $e')),
                        );
                      }
                    }
                  },
                  child: const Text('Decline'),
                ),
              ],
            )
          : TextButton(
              onPressed: () async {
                try {
                  await friendProvider.deleteInvite(invite.id!);
                  if (mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Request cancelled')),
                    );
                  }
                } catch (e) {
                  if (mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('Failed to cancel: $e')),
                    );
                  }
                }
              },
              child: const Text('Cancel'),
            ),
    );
  }
}

