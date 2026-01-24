import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/providers/post_provider.dart';
import 'package:skunkworks/widgets/emoji_picker.dart';
import 'package:skunkworks/screens/post_detail_screen.dart';
import 'package:skunkworks/screens/profile_screen.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/widgets/video_player_widget.dart';

class PostCard extends StatefulWidget {
  final Post post;
  final Function(NavTab)? onNavigate;

  const PostCard({
    super.key,
    required this.post,
    this.onNavigate,
  });

  @override
  State<PostCard> createState() => _PostCardState();
}

class _PostCardState extends State<PostCard> {
  late Post _currentPost;
  bool _isReacting = false;
  bool _isCommenting = false;
  bool _showAllComments = false;
  final TextEditingController _commentController = TextEditingController();
  bool _commentsLoaded = false;

  @override
  void initState() {
    super.initState();
    _currentPost = widget.post;
    _commentsLoaded = widget.post.comments != null;
  }

  @override
  void dispose() {
    _commentController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(PostCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    // Always update to get latest reactions and comments from provider
    _currentPost = widget.post;
    // Update comments loaded state if comments are now present
    if (widget.post.comments != null && !_commentsLoaded) {
      _commentsLoaded = true;
    }
  }

  String _formatTimestamp(DateTime? dateTime) {
    if (dateTime == null) return '';
    
    final now = DateTime.now();
    final difference = now.difference(dateTime);

    if (difference.inDays > 7) {
      return '${dateTime.day}/${dateTime.month}/${dateTime.year}';
    } else if (difference.inDays > 0) {
      return '${difference.inDays}d';
    } else if (difference.inHours > 0) {
      return '${difference.inHours}h';
    } else if (difference.inMinutes > 0) {
      return '${difference.inMinutes}m';
    } else {
      return 'now';
    }
  }

  Future<void> _handleReaction(String emoji) async {
    if (_isReacting || _currentPost.id == null) return;
    
    setState(() {
      _isReacting = true;
    });

    try {
      final postProvider = context.read<PostProvider>();
      ReactionSummary? currentReaction;
      if (_currentPost.reactions != null) {
        try {
          currentReaction = _currentPost.reactions!.firstWhere(
            (r) => r.userReacted,
          );
        } catch (e) {
          // No reaction found
          currentReaction = null;
        }
      }

      // If user already reacted with the same emoji, remove it
      if (currentReaction != null && currentReaction.emoji == emoji) {
        await postProvider.removeReaction(_currentPost.id!);
      } else {
        await postProvider.reactToPost(_currentPost.id!, emoji);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to react: ${e.toString()}')),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isReacting = false;
        });
      }
    }
  }

  String? _getUserReactionEmoji() {
    if (_currentPost.reactions == null) return null;
    try {
      final userReaction = _currentPost.reactions!.firstWhere(
        (r) => r.userReacted,
      );
      return userReaction.emoji;
    } catch (e) {
      return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    final username = _currentPost.user?.username ?? _currentPost.user?.phoneNumber ?? 'Unknown';
    final profilePictureUrl = _currentPost.user?.profilePictureUrl;
    final userReactionEmoji = _getUserReactionEmoji();
    final authProvider = context.watch<AuthProvider>();
    final currentUserId = authProvider.user?.id;
    final isPostOwner = currentUserId != null && _currentPost.userId == currentUserId;

    return Container(
      margin: const EdgeInsets.only(bottom: 24),
      decoration: BoxDecoration(
        color: AppColors.warmWhite,
        border: Border(
          top: BorderSide(color: Colors.grey.shade300, width: 0.5),
          bottom: BorderSide(color: Colors.grey.shade300, width: 0.5),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // User header
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
            child: Row(
              children: [
                GestureDetector(
                  onTap: () {
                    if (_currentPost.userId != null) {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => ProfileScreen(
                            userId: _currentPost.userId!,
                            user: _currentPost.user,
                            onNavigate: widget.onNavigate ?? (_) {},
                          ),
                        ),
                      );
                    }
                  },
                  child: CircleAvatar(
                    radius: 18,
                    backgroundColor: Colors.grey.shade300,
                    backgroundImage: profilePictureUrl != null
                        ? NetworkImage(profilePictureUrl)
                        : null,
                    child: profilePictureUrl == null
                        ? Text(
                            username.isNotEmpty ? username[0].toUpperCase() : 'U',
                            style: const TextStyle(color: Colors.grey),
                          )
                        : null,
                  ),
                ),
                const SizedBox(width: 12),
                GestureDetector(
                  onTap: () {
                    if (_currentPost.userId != null) {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => ProfileScreen(
                            userId: _currentPost.userId!,
                            user: _currentPost.user,
                            onNavigate: widget.onNavigate ?? (_) {},
                          ),
                        ),
                      );
                    }
                  },
                  child: Text(
                    username,
                    style: const TextStyle(
                      fontWeight: FontWeight.w600,
                      fontSize: 14,
                    ),
                  ),
                ),
                const Spacer(),
                if (_currentPost.createdAt != null)
                  Text(
                    _formatTimestamp(_currentPost.createdAt),
                    style: TextStyle(
                      color: AppColors.graphiteInk.withOpacity(0.6),
                      fontSize: 12,
                    ),
                  ),
                if (isPostOwner) ...[
                  const SizedBox(width: 8),
                  PopupMenuButton<String>(
                    icon: const Icon(Icons.more_vert, size: 20, color: Colors.grey),
                    onSelected: (value) {
                      if (value == 'delete') {
                        _handleDeletePost();
                      }
                    },
                    itemBuilder: (context) => [
                      PopupMenuItem<String>(
                        value: 'delete',
                        child: Row(
                          children: [
                            Icon(Icons.delete, color: AppColors.coralPop, size: 20),
                            const SizedBox(width: 8),
                            Text(
                              'Delete',
                              style: TextStyle(color: AppColors.coralPop),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ],
              ],
            ),
          ),
          // Media - tappable to view detail
          if (_currentPost.imageUrl != null)
            GestureDetector(
              onTap: () {
                if (_currentPost.id != null) {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => PostDetailScreen(
                        postId: _currentPost.id!,
                        onNavigate: widget.onNavigate ?? (_) {},
                      ),
                    ),
                  );
                }
              },
              child: _currentPost.isVideo
                  ? SizedBox(
                      height: 300,
                      child: VideoPlayerWidget(
                        videoUrl: _currentPost.imageUrl!,
                        autoPlay: false,
                        showControls: true,
                        fit: BoxFit.cover,
                      ),
                    )
                  : Image.network(
                      _currentPost.imageUrl!,
                      width: double.infinity,
                      fit: BoxFit.cover,
                      loadingBuilder: (context, child, loadingProgress) {
                        if (loadingProgress == null) return child;
                        return Container(
                          height: 300,
                          color: Colors.grey.shade200,
                          child: const Center(
                            child: CircularProgressIndicator(),
                          ),
                        );
                      },
                      errorBuilder: (context, error, stackTrace) {
                        return Container(
                          height: 300,
                          color: Colors.grey.shade200,
                          child: const Icon(Icons.error),
                        );
                      },
                    ),
            ),
          // Caption
          if (_currentPost.caption != null && _currentPost.caption!.isNotEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
              child: Text(
                _currentPost.caption!,
                style: const TextStyle(fontSize: 14),
              ),
            ),
          // Reactions section
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Display aggregated reactions
                if (_currentPost.reactions != null && _currentPost.reactions!.isNotEmpty)
                  Wrap(
                    spacing: 8,
                    runSpacing: 4,
                    children: _currentPost.reactions!.map((reaction) {
                      return Chip(
                        label: Text(
                          '${reaction.emoji} ${reaction.count}',
                          style: TextStyle(
                            fontSize: 12,
                            fontWeight: reaction.userReacted ? FontWeight.w600 : FontWeight.normal,
                            color: reaction.userReacted ? AppColors.softRealBlue : AppColors.graphiteInk,
                          ),
                        ),
                        backgroundColor: reaction.userReacted 
                            ? AppColors.softRealBlue.withOpacity(0.1)
                            : Colors.grey.shade100,
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                        materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                      );
                    }).toList(),
                  ),
                const SizedBox(height: 8),
                // Reaction button
                Row(
                  children: [
                    GestureDetector(
                      onTap: _isReacting ? null : () {
                        EmojiPicker.show(context, _handleReaction);
                      },
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                        decoration: BoxDecoration(
                          color: userReactionEmoji != null 
                              ? AppColors.softRealBlue.withOpacity(0.1)
                              : Colors.grey.shade100,
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              userReactionEmoji ?? '👍',
                              style: const TextStyle(fontSize: 18),
                            ),
                            const SizedBox(width: 4),
                            Text(
                              userReactionEmoji != null ? 'Reacted' : 'React',
                              style: TextStyle(
                                fontSize: 14,
                                color: userReactionEmoji != null 
                                    ? AppColors.softRealBlue 
                                    : AppColors.graphiteInk,
                                fontWeight: userReactionEmoji != null 
                                    ? FontWeight.w600 
                                    : FontWeight.normal,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          // Comments section
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Comment count or "View all comments" link
                if ((_currentPost.commentCount ?? 0) > 0 || 
                    (_currentPost.comments != null && _currentPost.comments!.isNotEmpty))
                  GestureDetector(
                    onTap: () async {
                      if (!_commentsLoaded && _currentPost.id != null) {
                        try {
                          final postProvider = context.read<PostProvider>();
                          await postProvider.loadComments(_currentPost.id!);
                          setState(() {
                            _commentsLoaded = true;
                            _showAllComments = true;
                          });
                        } catch (e) {
                          if (mounted) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(content: Text('Failed to load comments: ${e.toString()}')),
                            );
                          }
                        }
                      } else {
                        setState(() {
                          _showAllComments = !_showAllComments;
                        });
                      }
                    },
                    child: Padding(
                      padding: const EdgeInsets.only(bottom: 8),
                      child: Text(
                        _showAllComments 
                            ? 'Hide comments'
                            : 'View all ${_currentPost.commentCount ?? _currentPost.comments?.length ?? 0} comments',
                        style: TextStyle(
                          color: AppColors.graphiteInk.withOpacity(0.6),
                          fontSize: 14,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                  ),
                // Display comments
                if (_showAllComments && _currentPost.comments != null && _currentPost.comments!.isNotEmpty)
                  ..._currentPost.comments!.map((comment) => _buildComment(comment)),
                // Comment input
                if (_currentPost.id != null)
                  _buildCommentInput(),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildComment(Comment comment) {
    final authProvider = context.watch<AuthProvider>();
    final currentUserId = authProvider.user?.id;
    final canDelete = currentUserId != null && 
        (comment.userId == currentUserId || _currentPost.userId == currentUserId);
    
    final username = comment.user?.username ?? comment.user?.phoneNumber ?? 'Unknown';
    final profilePictureUrl = comment.user?.profilePictureUrl;

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          CircleAvatar(
            radius: 14,
            backgroundColor: Colors.grey.shade300,
            backgroundImage: profilePictureUrl != null
                ? NetworkImage(profilePictureUrl)
                : null,
            child: profilePictureUrl == null
                ? Text(
                    username.isNotEmpty ? username[0].toUpperCase() : 'U',
                    style: TextStyle(
                      color: Colors.grey,
                      fontSize: 12,
                    ),
                  )
                : null,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: Colors.grey.shade100,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        username,
                        style: const TextStyle(
                          fontWeight: FontWeight.w600,
                          fontSize: 13,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        comment.text ?? '',
                        style: const TextStyle(fontSize: 14),
                      ),
                    ],
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.only(top: 4, left: 4),
                  child: Row(
                    children: [
                      if (comment.createdAt != null)
                        Text(
                          _formatTimestamp(comment.createdAt),
                          style: TextStyle(
                            color: Colors.grey.shade600,
                            fontSize: 12,
                          ),
                        ),
                      if (canDelete) ...[
                        const SizedBox(width: 12),
                        GestureDetector(
                          onTap: () => _handleDeleteComment(comment),
                          child: Text(
                            'Delete',
                            style: TextStyle(
                              color: AppColors.coralPop,
                              fontSize: 12,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCommentInput() {
    return Padding(
      padding: const EdgeInsets.only(top: 8),
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _commentController,
              decoration: InputDecoration(
                hintText: 'Add a comment...',
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(20),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 10,
                ),
                filled: true,
                fillColor: Colors.grey.shade100,
              ),
              maxLines: null,
              textInputAction: TextInputAction.send,
              onSubmitted: (_) => _handleSubmitComment(),
            ),
          ),
          const SizedBox(width: 8),
          IconButton(
            icon: _isCommenting
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.send),
            onPressed: _isCommenting ? null : _handleSubmitComment,
            color: Colors.blue,
          ),
        ],
      ),
    );
  }

  Future<void> _handleSubmitComment() async {
    if (_commentController.text.trim().isEmpty || _currentPost.id == null) return;

    final text = _commentController.text.trim();
    _commentController.clear();

    setState(() {
      _isCommenting = true;
    });

    try {
      final postProvider = context.read<PostProvider>();
      await postProvider.createComment(_currentPost.id!, text);
      
      // Ensure comments are visible after creating
      if (!_showAllComments) {
        setState(() {
          _showAllComments = true;
          _commentsLoaded = true;
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to post comment: ${e.toString()}')),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isCommenting = false;
        });
      }
    }
  }

  Future<void> _handleDeleteComment(Comment comment) async {
    if (_currentPost.id == null || comment.id == null) return;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Comment'),
        content: const Text('Are you sure you want to delete this comment?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    try {
      final postProvider = context.read<PostProvider>();
      await postProvider.deleteComment(_currentPost.id!, comment.id!);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to delete comment: ${e.toString()}')),
        );
      }
    }
  }

  Future<void> _handleDeletePost() async {
    if (_currentPost.id == null) return;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Post'),
        content: const Text('Are you sure you want to delete this post? This action cannot be undone.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    try {
      final postProvider = context.read<PostProvider>();
      await postProvider.deletePost(_currentPost.id!);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to delete post: ${e.toString()}')),
        );
      }
    }
  }
}

