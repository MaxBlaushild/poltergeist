import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:travel_angels/constants/api_constants.dart';
import 'package:travel_angels/models/friend_invite.dart';
import 'package:travel_angels/models/user.dart';
import 'package:travel_angels/providers/auth_provider.dart';
import 'package:travel_angels/services/api_client.dart';
import 'package:travel_angels/services/friend_service.dart';
import 'package:intl/intl.dart';

/// My Network screen for managing friends and friend invites
class MyNetworkScreen extends StatefulWidget {
  const MyNetworkScreen({super.key});

  @override
  State<MyNetworkScreen> createState() => _MyNetworkScreenState();
}

class _MyNetworkScreenState extends State<MyNetworkScreen> {
  final FriendService _friendService = FriendService(
    APIClient(ApiConstants.baseUrl),
  );

  List<User> _friends = [];
  List<FriendInvite> _friendInvites = [];
  List<User> _searchResults = [];
  bool _isLoadingFriends = true;
  bool _isLoadingInvites = true;
  bool _isSearching = false;
  String? _errorMessage;

  // Section expansion states
  bool _isFriendsExpanded = true;
  bool _isSearchExpanded = false;
  bool _isReceivedInvitesExpanded = true;
  bool _isSentInvitesExpanded = true;

  // Loading states for actions
  Set<String> _acceptingInvites = {};
  Set<String> _rejectingInvites = {};
  Set<String> _sendingInvites = {};

  final TextEditingController _searchController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadData();
    _searchController.addListener(_onSearchChanged);
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    await Future.wait([
      _loadFriends(),
      _loadFriendInvites(),
    ]);
  }

  Future<void> _loadFriends() async {
    setState(() {
      _isLoadingFriends = true;
      _errorMessage = null;
    });

    try {
      final friends = await _friendService.getFriends();
      setState(() {
        _friends = friends;
        _isLoadingFriends = false;
      });
    } catch (e) {
      setState(() {
        _isLoadingFriends = false;
        _errorMessage = 'Failed to load friends: ${e.toString()}';
      });
    }
  }

  Future<void> _loadFriendInvites() async {
    setState(() {
      _isLoadingInvites = true;
    });

    try {
      final invites = await _friendService.getFriendInvites();
      setState(() {
        _friendInvites = invites;
        _isLoadingInvites = false;
      });
    } catch (e) {
      setState(() {
        _isLoadingInvites = false;
        _errorMessage = 'Failed to load friend invites: ${e.toString()}';
      });
    }
  }

  void _onSearchChanged() {
    final query = _searchController.text.trim();
    if (query.isEmpty) {
      setState(() {
        _searchResults = [];
      });
      return;
    }

    _performSearch(query);
  }

  Future<void> _performSearch(String query) async {
    setState(() {
      _isSearching = true;
    });

    try {
      final results = await _friendService.searchUsers(query);
      setState(() {
        _searchResults = results;
        _isSearching = false;
      });
    } catch (e) {
      setState(() {
        _isSearching = false;
        _errorMessage = 'Failed to search users: ${e.toString()}';
      });
    }
  }

  Future<void> _handleAcceptInvite(String inviteId) async {
    setState(() {
      _acceptingInvites.add(inviteId);
    });

    try {
      await _friendService.acceptFriendInvite(inviteId);
      await _loadData(); // Refresh friends and invites
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Friend invite accepted!'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to accept invite: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _acceptingInvites.remove(inviteId);
        });
      }
    }
  }

  Future<void> _handleRejectInvite(String inviteId) async {
    setState(() {
      _rejectingInvites.add(inviteId);
    });

    try {
      await _friendService.deleteFriendInvite(inviteId);
      await _loadFriendInvites(); // Refresh invites
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Friend invite rejected'),
            backgroundColor: Colors.orange,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to reject invite: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _rejectingInvites.remove(inviteId);
        });
      }
    }
  }

  Future<void> _handleSendInvite(String userId) async {
    setState(() {
      _sendingInvites.add(userId);
    });

    try {
      await _friendService.createFriendInvite(userId);
      await _loadFriendInvites(); // Refresh invites
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Friend invite sent!'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      String errorMsg = 'Failed to send invite';
      if (e is DioException) {
        if (e.response != null && e.response?.data != null) {
          final errorData = e.response?.data as Map<String, dynamic>?;
          errorMsg = errorData?['error']?.toString() ?? errorMsg;
        }
      }
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(errorMsg),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _sendingInvites.remove(userId);
        });
      }
    }
  }

  bool _isAlreadyFriend(String userId) {
    return _friends.any((friend) => friend.id == userId);
  }

  bool _isCurrentUser(String userId) {
    final authProvider = context.read<AuthProvider>();
    return authProvider.user?.id == userId;
  }

  bool _hasPendingInvite(String userId) {
    return _friendInvites.any((invite) =>
        (invite.inviterId == userId || invite.inviteeId == userId));
  }

  List<FriendInvite> get _receivedInvites {
    final authProvider = context.read<AuthProvider>();
    final currentUserId = authProvider.user?.id;
    if (currentUserId == null) return [];
    return _friendInvites
        .where((invite) => invite.inviteeId == currentUserId)
        .toList();
  }

  List<FriendInvite> get _sentInvites {
    final authProvider = context.read<AuthProvider>();
    final currentUserId = authProvider.user?.id;
    if (currentUserId == null) return [];
    return _friendInvites
        .where((invite) => invite.inviterId == currentUserId)
        .toList();
  }

  Widget _buildUserAvatar(User? user, {double radius = 20}) {
    final theme = Theme.of(context);
    final profilePictureUrl = user?.profilePictureUrl;

    if (profilePictureUrl != null && profilePictureUrl.isNotEmpty) {
      return CircleAvatar(
        radius: radius,
        backgroundImage: NetworkImage(profilePictureUrl),
        onBackgroundImageError: (exception, stackTrace) {
          // Handle error silently
        },
        child: profilePictureUrl.isEmpty ? _buildFallbackAvatar(user, theme, radius) : null,
      );
    }

    return _buildFallbackAvatar(user, theme, radius);
  }

  Widget _buildFallbackAvatar(User? user, ThemeData theme, double radius) {
    final initials = _getInitials(user);
    return CircleAvatar(
      radius: radius,
      backgroundColor: theme.colorScheme.primaryContainer,
      child: Text(
        initials,
        style: TextStyle(
          fontSize: radius * 0.6,
          fontWeight: FontWeight.bold,
          color: theme.colorScheme.onPrimaryContainer,
        ),
      ),
    );
  }

  String _getInitials(User? user) {
    if (user == null) return '?';
    final name = user.name ?? user.username ?? '';
    if (name.isEmpty) return '?';
    final parts = name.trim().split(' ');
    if (parts.length >= 2) {
      return '${parts[0][0]}${parts[1][0]}'.toUpperCase();
    }
    return name[0].toUpperCase();
  }

  Widget _buildFriendsSection() {
    return ExpansionTile(
      title: Row(
        children: [
          const Text('Friends'),
          const SizedBox(width: 8),
          Text(
            '(${_friends.length})',
            style: TextStyle(
              color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
              fontSize: 14,
            ),
          ),
        ],
      ),
      initiallyExpanded: _isFriendsExpanded,
      onExpansionChanged: (expanded) {
        setState(() {
          _isFriendsExpanded = expanded;
        });
      },
      children: [
        if (_isLoadingFriends)
          const Padding(
            padding: EdgeInsets.all(16.0),
            child: Center(child: CircularProgressIndicator()),
          )
        else if (_friends.isEmpty)
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: Text(
              'No friends yet',
              style: TextStyle(
                color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
              ),
              textAlign: TextAlign.center,
            ),
          )
        else
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _friends.length,
            itemBuilder: (context, index) {
              final friend = _friends[index];
              return Card(
                margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                child: ListTile(
                  leading: _buildUserAvatar(friend),
                  title: Text(friend.username ?? 'Unknown'),
                  subtitle: friend.name != null ? Text(friend.name!) : null,
                ),
              );
            },
          ),
      ],
    );
  }

  Widget _buildSearchSection() {
    return ExpansionTile(
      title: const Row(
        children: [
          Icon(Icons.search, size: 20),
          SizedBox(width: 8),
          Text('Find Friends'),
        ],
      ),
      initiallyExpanded: _isSearchExpanded,
      onExpansionChanged: (expanded) {
        setState(() {
          _isSearchExpanded = expanded;
        });
      },
      children: [
        Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            children: [
              TextField(
                controller: _searchController,
                decoration: const InputDecoration(
                  labelText: 'Search by username',
                  hintText: 'Enter username...',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.search),
                ),
              ),
              const SizedBox(height: 16),
              if (_isSearching)
                const Center(child: CircularProgressIndicator())
              else if (_searchController.text.trim().isNotEmpty)
                if (_searchResults.isEmpty)
                  Text(
                    'No users found',
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
                    ),
                  )
                else
                  ListView.builder(
                    shrinkWrap: true,
                    physics: const NeverScrollableScrollPhysics(),
                    itemCount: _searchResults.length,
                    itemBuilder: (context, index) {
                      final user = _searchResults[index];
                      final isCurrent = _isCurrentUser(user.id ?? '');
                      final isFriend = _isAlreadyFriend(user.id ?? '');
                      final hasInvite = _hasPendingInvite(user.id ?? '');

                      return Card(
                        margin: const EdgeInsets.symmetric(vertical: 4),
                        child: ListTile(
                          leading: _buildUserAvatar(user),
                          title: Text(user.username ?? 'Unknown'),
                          subtitle: user.name != null ? Text(user.name!) : null,
                          trailing: isCurrent
                              ? Text(
                                  'You',
                                  style: TextStyle(
                                    color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
                                  ),
                                )
                              : isFriend
                                  ? Chip(
                                      label: const Text('Friends'),
                                      backgroundColor: Colors.green.withOpacity(0.2),
                                    )
                                  : ElevatedButton(
                                      onPressed: _sendingInvites.contains(user.id) || hasInvite
                                          ? null
                                          : () => _handleSendInvite(user.id ?? ''),
                                      child: _sendingInvites.contains(user.id)
                                          ? const SizedBox(
                                              width: 16,
                                              height: 16,
                                              child: CircularProgressIndicator(strokeWidth: 2),
                                            )
                                          : const Text('Invite'),
                                    ),
                        ),
                      );
                    },
                  ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildReceivedInvitesSection() {
    final receivedInvites = _receivedInvites;
    return ExpansionTile(
      title: Row(
        children: [
          const Text('Received Invites'),
          if (receivedInvites.isNotEmpty) ...[
            const SizedBox(width: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.primary,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                '${receivedInvites.length}',
                style: TextStyle(
                  color: Theme.of(context).colorScheme.onPrimary,
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ],
        ],
      ),
      initiallyExpanded: _isReceivedInvitesExpanded,
      onExpansionChanged: (expanded) {
        setState(() {
          _isReceivedInvitesExpanded = expanded;
        });
      },
      children: [
        if (_isLoadingInvites)
          const Padding(
            padding: EdgeInsets.all(16.0),
            child: Center(child: CircularProgressIndicator()),
          )
        else if (receivedInvites.isEmpty)
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: Text(
              'No pending invites',
              style: TextStyle(
                color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
              ),
              textAlign: TextAlign.center,
            ),
          )
        else
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: receivedInvites.length,
            itemBuilder: (context, index) {
              final invite = receivedInvites[index];
              final inviter = invite.inviter;
              final isAccepting = _acceptingInvites.contains(invite.id);
              final isRejecting = _rejectingInvites.contains(invite.id);
              final isProcessing = isAccepting || isRejecting;

              return Card(
                margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                child: ListTile(
                  leading: _buildUserAvatar(inviter),
                  title: Text(inviter?.username ?? 'Unknown'),
                  subtitle: Text(
                    invite.createdAt != null
                        ? DateFormat.yMMMd().format(invite.createdAt!)
                        : 'Unknown date',
                  ),
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      ElevatedButton(
                        onPressed: isProcessing
                            ? null
                            : () => _handleAcceptInvite(invite.id),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.green,
                          foregroundColor: Colors.white,
                        ),
                        child: isAccepting
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                                ),
                              )
                            : const Text('Accept'),
                      ),
                      const SizedBox(width: 8),
                      OutlinedButton(
                        onPressed: isProcessing
                            ? null
                            : () => _handleRejectInvite(invite.id),
                        child: isRejecting
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              )
                            : const Text('Reject'),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
      ],
    );
  }

  Widget _buildSentInvitesSection() {
    final sentInvites = _sentInvites;
    return ExpansionTile(
      title: Row(
        children: [
          const Text('Sent Invites'),
          const SizedBox(width: 8),
          Text(
            '(${sentInvites.length})',
            style: TextStyle(
              color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
              fontSize: 14,
            ),
          ),
        ],
      ),
      initiallyExpanded: _isSentInvitesExpanded,
      onExpansionChanged: (expanded) {
        setState(() {
          _isSentInvitesExpanded = expanded;
        });
      },
      children: [
        if (_isLoadingInvites)
          const Padding(
            padding: EdgeInsets.all(16.0),
            child: Center(child: CircularProgressIndicator()),
          )
        else if (sentInvites.isEmpty)
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: Text(
              'No sent invites',
              style: TextStyle(
                color: Theme.of(context).colorScheme.onSurface.withOpacity(0.6),
              ),
              textAlign: TextAlign.center,
            ),
          )
        else
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: sentInvites.length,
            itemBuilder: (context, index) {
              final invite = sentInvites[index];
              final invitee = invite.invitee;
              final isRejecting = _rejectingInvites.contains(invite.id);

              return Card(
                margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                child: ListTile(
                  leading: _buildUserAvatar(invitee),
                  title: Text(invitee?.username ?? 'Unknown'),
                  subtitle: Text(
                    invite.createdAt != null
                        ? DateFormat.yMMMd().format(invite.createdAt!)
                        : 'Unknown date',
                  ),
                  trailing: OutlinedButton(
                    onPressed: isRejecting
                        ? null
                        : () => _handleRejectInvite(invite.id),
                    child: isRejecting
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('Cancel'),
                  ),
                ),
              );
            },
          ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('My Circle'),
      ),
      body: RefreshIndicator(
        onRefresh: _loadData,
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          padding: const EdgeInsets.all(16.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (_errorMessage != null) ...[
                Container(
                  padding: const EdgeInsets.all(12.0),
                  margin: const EdgeInsets.only(bottom: 16.0),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.errorContainer,
                    borderRadius: BorderRadius.circular(8.0),
                  ),
                  child: Row(
                    children: [
                      Icon(
                        Icons.error_outline,
                        color: theme.colorScheme.onErrorContainer,
                      ),
                      const SizedBox(width: 12.0),
                      Expanded(
                        child: Text(
                          _errorMessage!,
                          style: TextStyle(
                            color: theme.colorScheme.onErrorContainer,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
              Card(
                child: _buildFriendsSection(),
              ),
              const SizedBox(height: 16),
              Card(
                child: _buildSearchSection(),
              ),
              const SizedBox(height: 16),
              Card(
                child: _buildReceivedInvitesSection(),
              ),
              const SizedBox(height: 16),
              Card(
                child: _buildSentInvitesSection(),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

