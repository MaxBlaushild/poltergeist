import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/models/album.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/services/album_service.dart';
import 'package:skunkworks/services/post_service.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/widgets/post_card.dart';

class AlbumDetailScreen extends StatefulWidget {
  final String albumId;
  final String albumName;
  final Function(NavTab) onNavigate;

  const AlbumDetailScreen({
    super.key,
    required this.albumId,
    required this.albumName,
    required this.onNavigate,
  });

  @override
  State<AlbumDetailScreen> createState() => _AlbumDetailScreenState();
}

class _AlbumDetailScreenState extends State<AlbumDetailScreen> {
  Album? _album;
  List<Post> _posts = [];
  String? _role;
  List<dynamic> _members = [];
  List<dynamic> _pendingInvites = [];
  bool _loading = true;
  String? _error;

  bool get _canAdmin => _role == 'owner' || _role == 'admin';
  bool get _canAddRemovePosts => _role == 'owner' || _role == 'admin' || _role == 'poster';

  @override
  void initState() {
    super.initState();
    _loadAlbum();
  }

  Future<void> _loadAlbum() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final albumService = AlbumService(apiClient);
      final data = await albumService.getAlbum(widget.albumId);
      if (mounted) {
        setState(() {
          _album = data['album'] as Album;
          _posts = data['posts'] as List<Post>;
          _role = data['role'] as String?;
          _members = data['members'] as List<dynamic>? ?? [];
          _pendingInvites = data['pendingInvites'] as List<dynamic>? ?? [];
          _loading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _loading = false;
        });
      }
    }
  }

  Future<void> _addTag() async {
    if (!_canAdmin || _album?.id == null) return;
    final controller = TextEditingController();
    final result = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Add Tag'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            labelText: 'Tag',
            border: OutlineInputBorder(),
          ),
          autofocus: true,
          onSubmitted: (_) => Navigator.pop(ctx, true),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Add'),
          ),
        ],
      ),
    );
    if (result != true || !mounted) return;
    final tag = controller.text.trim();
    if (tag.isEmpty) return;
    try {
      await AlbumService(APIClient(ApiConstants.baseUrl)).addAlbumTag(widget.albumId, tag);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Tag added')));
        _loadAlbum();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
      }
    }
  }

  Future<void> _removeTag(String tag) async {
    if (!_canAdmin || _album?.id == null) return;
    try {
      await AlbumService(APIClient(ApiConstants.baseUrl)).removeAlbumTag(widget.albumId, tag);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Tag removed')));
        _loadAlbum();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
      }
    }
  }

  Future<void> _inviteUser() async {
    if (!_canAdmin) return;
    final controller = TextEditingController();
    String role = 'poster';
    final result = await showDialog<bool>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setState) => AlertDialog(
          title: const Text('Invite to Album'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text('Enter username to search and invite:'),
              const SizedBox(height: 8),
              TextField(
                controller: controller,
                decoration: const InputDecoration(
                  hintText: 'Username',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              const Text('Role:'),
              SegmentedButton<String>(
                segments: const [
                  ButtonSegment(value: 'poster', label: Text('Poster')),
                  ButtonSegment(value: 'admin', label: Text('Admin')),
                ],
                selected: {role},
                onSelectionChanged: (v) {
                  setState(() => role = v.first);
                },
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(ctx, true),
              child: const Text('Invite'),
            ),
          ],
        ),
      ),
    );
    if (result != true || !mounted) return;
    final query = controller.text.trim();
    if (query.isEmpty) return;
    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final users = await apiClient.get<List<dynamic>>(ApiConstants.searchUsersEndpoint(query));
      if (users.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('No users found')));
        }
        return;
      }
      final userMap = users[0] as Map<String, dynamic>;
      final userId = userMap['id']?.toString();
      if (userId == null) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('User not found')));
        }
        return;
      }
      await AlbumService(apiClient).inviteToAlbum(widget.albumId, userId, role);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Invite sent')));
        _loadAlbum();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
      }
    }
  }

  Future<void> _addPost() async {
    if (!_canAddRemovePosts) return;
    final albumTags = _album?.tags ?? [];
    if (albumTags.isEmpty) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Add tags to the album first')),
        );
      }
      return;
    }
    final currentUser = context.read<AuthProvider>().user;
    if (currentUser?.id == null) return;
    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final postService = PostService(apiClient);
      final myPosts = await postService.getUserPosts(currentUser!.id!);
      if (myPosts.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('No posts to add')));
        }
        return;
      }
      final selected = await showModalBottomSheet<Post>(
        context: context,
        isScrollControlled: true,
        builder: (ctx) => DraggableScrollableSheet(
          initialChildSize: 0.6,
          maxChildSize: 0.9,
          expand: false,
          builder: (_, scrollController) => Column(
            children: [
              const Padding(
                padding: EdgeInsets.all(16),
                child: Text('Select a post', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600)),
              ),
              Expanded(
                child: ListView.builder(
                  controller: scrollController,
                  itemCount: myPosts.length,
                  itemBuilder: (_, i) {
                    final post = myPosts[i];
                    return ListTile(
                      leading: post.imageUrl != null && !(post.isVideo)
                          ? Image.network(post.imageUrl!, width: 48, height: 48, fit: BoxFit.cover)
                          : const Icon(Icons.image),
                      title: Text(post.caption ?? 'No caption', maxLines: 2, overflow: TextOverflow.ellipsis),
                      onTap: () => Navigator.pop(ctx, post),
                    );
                  },
                ),
              ),
            ],
          ),
        ),
      );
      if (selected?.id == null || !mounted) return;
      final post = selected!;
      final existingTags = post.tags ?? [];
      final tagsToAdd = albumTags.where((t) => !existingTags.contains(t)).toList();
      if (tagsToAdd.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Post already has all album tags')),
          );
        }
        return;
      }
      final selectedTags = <String>{};
      final chosenTags = await showModalBottomSheet<List<String>>(
        context: context,
        isScrollControlled: true,
        builder: (ctx) => StatefulBuilder(
          builder: (ctx2, setModalState) {
            return DraggableScrollableSheet(
              initialChildSize: 0.4,
              maxChildSize: 0.6,
              expand: false,
              builder: (_, scrollController) => Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    const Text(
                      'Select album tag(s) to add to this post',
                      style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
                    ),
                    const SizedBox(height: 12),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: tagsToAdd.map((tag) {
                        final isSelected = selectedTags.contains(tag);
                        return FilterChip(
                          label: Text(tag),
                          selected: isSelected,
                          onSelected: (v) {
                            setModalState(() {
                              if (v) {
                                selectedTags.add(tag);
                              } else {
                                selectedTags.remove(tag);
                              }
                            });
                          },
                        );
                      }).toList(),
                    ),
                    const SizedBox(height: 16),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        TextButton(
                          onPressed: () => Navigator.pop(ctx2, <String>[]),
                          child: const Text('Cancel'),
                        ),
                        FilledButton(
                          onPressed: () => Navigator.pop(ctx2, selectedTags.toList()),
                          child: const Text('Add tags'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            );
          },
        ),
      );
      if (chosenTags == null || chosenTags.isEmpty || !mounted) return;
      await postService.addTagsToPost(post.id!, chosenTags);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Tags added to post')));
        _loadAlbum();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
      }
    }
  }

  Future<void> _removeMember(String userId) async {
    if (!_canAdmin) return;
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Remove Member'),
        content: const Text('Remove this member from the album?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancel')),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('Remove'),
          ),
        ],
      ),
    );
    if (confirm != true || !mounted) return;
    try {
      await AlbumService(APIClient(ApiConstants.baseUrl)).removeAlbumMember(widget.albumId, userId);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Member removed')));
        _loadAlbum();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
      }
    }
  }

  Future<void> _removePost(Post post) async {
    if (!_canAddRemovePosts || post.id == null) return;
    final currentUser = context.read<AuthProvider>().user;
    if (_role == 'poster' && post.userId != currentUser?.id) return;
    final albumTags = _album?.tags ?? [];
    final postTags = post.tags ?? [];
    final albumTagsOnPost = albumTags.where((t) => postTags.contains(t)).toList();
    if (albumTagsOnPost.isEmpty) return;
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Remove from album'),
        content: Text(
          albumTagsOnPost.length == 1
              ? 'Remove tag "${albumTagsOnPost.first}" from this post? It will no longer appear in this album.'
              : 'Remove ${albumTagsOnPost.length} album tags from this post? It will no longer appear in this album.',
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancel')),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('Remove'),
          ),
        ],
      ),
    );
    if (confirm != true || !mounted) return;
    try {
      final postService = PostService(APIClient(ApiConstants.baseUrl));
      final currentUserId = currentUser?.id;
      final isOwnPost = post.userId == currentUserId;
      final albumIdForAdmin = (!isOwnPost && _canAdmin) ? widget.albumId : null;
      for (final tag in albumTagsOnPost) {
        await postService.removeTagFromPost(post.id!, tag, albumId: albumIdForAdmin);
      }
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Post removed from album')));
        _loadAlbum();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Failed: $e')));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.warmWhite,
      appBar: AppBar(
        backgroundColor: AppColors.warmWhite,
        elevation: 0,
        title: Row(
          children: [
            Expanded(
              child: Text(
                widget.albumName,
                style: const TextStyle(
                  color: AppColors.graphiteInk,
                  fontWeight: FontWeight.w600,
                  fontSize: 18,
                ),
              ),
            ),
            if (_role != null)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: AppColors.softRealBlue.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(_role!, style: TextStyle(fontSize: 12, color: AppColors.softRealBlue, fontWeight: FontWeight.w600)),
              ),
          ],
        ),
        actions: [
          if (_canAdmin)
            IconButton(
              icon: const Icon(Icons.person_add),
              onPressed: _inviteUser,
              tooltip: 'Invite',
            ),
          if (_canAddRemovePosts)
            IconButton(
              icon: const Icon(Icons.add_photo_alternate),
              onPressed: _addPost,
              tooltip: 'Add tags to post',
            ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_error!, textAlign: TextAlign.center, style: TextStyle(color: Colors.grey.shade700)),
                      const SizedBox(height: 16),
                      TextButton(onPressed: _loadAlbum, child: const Text('Retry')),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _loadAlbum,
                  child: SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: const EdgeInsets.only(bottom: 80),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (_canAdmin && (_members.isNotEmpty || _pendingInvites.isNotEmpty || (_album?.tags ?? []).isNotEmpty)) ...[
                          _buildManageSection(),
                          const Divider(height: 1),
                        ],
                        if (_posts.isEmpty)
                          Padding(
                            padding: const EdgeInsets.all(24),
                            child: Center(
                              child: Text(
                                'No posts in this album yet.\n${_canAddRemovePosts ? "Tap + to add posts." : "Posts with the album\'s tags appear here."}',
                                textAlign: TextAlign.center,
                                style: TextStyle(fontSize: 16, color: Colors.grey.shade600),
                              ),
                            ),
                          )
                        else
                          ..._posts.map((post) => _buildPostTile(post)),
                      ],
                    ),
                  ),
                ),
      bottomNavigationBar: BottomNav(
        currentTab: NavTab.profile,
        onTabChanged: widget.onNavigate,
      ),
    );
  }

  Widget _buildManageSection() {
    return ExpansionTile(
      title: const Text('Manage', style: TextStyle(fontWeight: FontWeight.w600)),
      initiallyExpanded: false,
      children: [
        if ((_album?.tags ?? []).isNotEmpty) ...[
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 8, 16, 4),
            child: Text('Tags', style: TextStyle(fontSize: 12, color: Colors.grey)),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Wrap(
              spacing: 6,
              runSpacing: 6,
              children: (_album!.tags).map((tag) => Chip(
                label: Text(tag),
                deleteIcon: const Icon(Icons.close, size: 18),
                onDeleted: _canAdmin ? () => _removeTag(tag) : null,
              )).toList(),
            ),
          ),
          if (_canAdmin)
            Padding(
              padding: const EdgeInsets.all(16),
              child: OutlinedButton.icon(
                onPressed: _addTag,
                icon: const Icon(Icons.add, size: 18),
                label: const Text('Add tag'),
              ),
            ),
        ] else if (_canAdmin)
          Padding(
            padding: const EdgeInsets.all(16),
            child: OutlinedButton.icon(
              onPressed: _addTag,
              icon: const Icon(Icons.add, size: 18),
              label: const Text('Add tag'),
            ),
          ),
        if (_members.isNotEmpty) ...[
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 16, 16, 4),
            child: Text('Members', style: TextStyle(fontSize: 12, color: Colors.grey)),
          ),
          ..._members.map((m) {
            final u = m['user'] as Map<String, dynamic>?;
            final role = m['role'] as String? ?? '';
            final userId = m['userId']?.toString() ?? u?['id']?.toString();
            final username = u?['username'] ?? u?['phoneNumber'] ?? 'Unknown';
            final isOwner = role == 'owner';
            return ListTile(
              dense: true,
              leading: CircleAvatar(
                backgroundColor: Colors.grey.shade300,
                backgroundImage: u?['profilePictureUrl'] != null
                    ? NetworkImage(u!['profilePictureUrl'] as String)
                    : null,
                child: u?['profilePictureUrl'] == null ? Text('${username[0]}'.toUpperCase()) : null,
              ),
              title: Text(username),
              trailing: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(role, style: TextStyle(fontSize: 12, color: Colors.grey.shade600)),
                  if (_canAdmin && !isOwner && userId != null)
                    IconButton(
                      icon: const Icon(Icons.remove_circle_outline, size: 20),
                      onPressed: () => _removeMember(userId),
                      tooltip: 'Remove',
                    ),
                ],
              ),
            );
          }),
        ],
        if (_pendingInvites.isNotEmpty) ...[
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 16, 16, 4),
            child: Text('Pending Invites', style: TextStyle(fontSize: 12, color: Colors.grey)),
          ),
          ..._pendingInvites.map((inv) {
            final u = inv['invitedUser'] as Map<String, dynamic>?;
            final username = u?['username'] ?? u?['phoneNumber'] ?? 'Unknown';
            return ListTile(dense: true, title: Text(username), subtitle: const Text('Pending'));
          }),
        ],
      ],
    );
  }

  bool _canRemovePost(Post post) {
    if (!_canAddRemovePosts) return false;
    if (_role == 'owner' || _role == 'admin') return true;
    if (_role == 'poster') {
      final currentUser = context.read<AuthProvider>().user;
      return post.userId == currentUser?.id;
    }
    return false;
  }

  Widget _buildPostTile(Post post) {
    return Stack(
      children: [
        PostCard(
          post: post,
          onNavigate: widget.onNavigate,
        ),
        if (_canRemovePost(post))
          Positioned(
            top: 8,
            right: 8,
            child: Material(
              color: Colors.black54,
              borderRadius: BorderRadius.circular(20),
              child: InkWell(
                onTap: () => _removePost(post),
                borderRadius: BorderRadius.circular(20),
                child: const Padding(
                  padding: EdgeInsets.all(6),
                  child: Icon(Icons.remove_circle_outline, color: Colors.white, size: 24),
                ),
              ),
            ),
          ),
      ],
    );
  }
}
