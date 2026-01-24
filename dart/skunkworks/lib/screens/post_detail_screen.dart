import 'dart:io';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:gal/gal.dart';
import 'package:http/http.dart' as http;
import 'package:path_provider/path_provider.dart';
import 'package:skunkworks/constants/app_colors.dart';
import 'package:skunkworks/models/post.dart';
import 'package:skunkworks/providers/auth_provider.dart';
import 'package:skunkworks/providers/post_provider.dart';
import 'package:skunkworks/services/api_client.dart';
import 'package:skunkworks/services/post_service.dart';
import 'package:skunkworks/constants/api_constants.dart';
import 'package:skunkworks/widgets/bottom_nav.dart';
import 'package:skunkworks/screens/profile_screen.dart';
import 'package:skunkworks/widgets/emoji_picker.dart';
import 'package:skunkworks/widgets/video_player_widget.dart';

class PostDetailScreen extends StatefulWidget {
  final String postId;
  final Function(NavTab) onNavigate;

  const PostDetailScreen({
    super.key,
    required this.postId,
    required this.onNavigate,
  });

  @override
  State<PostDetailScreen> createState() => _PostDetailScreenState();
}

class _PostDetailScreenState extends State<PostDetailScreen> {
  Post? _post;
  bool _loading = true;
  String? _error;
  bool _isReacting = false;
  bool _isCommenting = false;
  final TextEditingController _commentController = TextEditingController();
  Map<String, dynamic>? _blockchainTransaction;

  @override
  void initState() {
    super.initState();
    _loadPostData();
  }

  @override
  void dispose() {
    _commentController.dispose();
    super.dispose();
  }

  Future<void> _loadPostData() async {
    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final apiClient = APIClient(ApiConstants.baseUrl);
      final postService = PostService(apiClient);

      // Load post
      final post = await postService.getPost(widget.postId);
      
      // Load comments
      final comments = await postService.getComments(widget.postId);
      
      // Create post with comments
      final postWithComments = Post(
        id: post.id,
        createdAt: post.createdAt,
        updatedAt: post.updatedAt,
        userId: post.userId,
        imageUrl: post.imageUrl,
        caption: post.caption,
        manifestUri: post.manifestUri,
        manifestHash: post.manifestHash,
        certFingerprint: post.certFingerprint,
        assetId: post.assetId,
        user: post.user,
        reactions: post.reactions,
        commentCount: post.commentCount,
        comments: comments,
      );

      // Load blockchain transaction if manifest hash exists
      Map<String, dynamic>? tx;
      if (post.manifestHash != null && post.manifestHash!.isNotEmpty) {
        try {
          tx = await postService.getBlockchainTransaction(widget.postId);
        } catch (e) {
          // Ignore errors loading transaction - it's optional
          print('Failed to load blockchain transaction: $e');
        }
      }

      if (mounted) {
        setState(() {
          _post = postWithComments;
          _blockchainTransaction = tx;
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
    if (_isReacting || _post?.id == null) return;
    
    setState(() {
      _isReacting = true;
    });

    try {
      final postProvider = context.read<PostProvider>();
      ReactionSummary? currentReaction;
      if (_post?.reactions != null) {
        try {
          currentReaction = _post!.reactions!.firstWhere(
            (r) => r.userReacted,
          );
        } catch (e) {
          currentReaction = null;
        }
      }

      if (currentReaction != null && currentReaction.emoji == emoji) {
        await postProvider.removeReaction(_post!.id!);
      } else {
        await postProvider.reactToPost(_post!.id!, emoji);
      }

      // Reload post to get updated reactions
      await _loadPostData();
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
    if (_post?.reactions == null) return null;
    try {
      final userReaction = _post!.reactions!.firstWhere(
        (r) => r.userReacted,
      );
      return userReaction.emoji;
    } catch (e) {
      return null;
    }
  }

  Future<void> _handleSubmitComment() async {
    if (_commentController.text.trim().isEmpty || _post?.id == null) return;

    final text = _commentController.text.trim();
    _commentController.clear();

    setState(() {
      _isCommenting = true;
    });

    try {
      final postProvider = context.read<PostProvider>();
      await postProvider.createComment(_post!.id!, text);
      
      // Reload post to get updated comments
      await _loadPostData();
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
    if (_post?.id == null || comment.id == null) return;

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
      await postProvider.deleteComment(_post!.id!, comment.id!);
      
      // Reload post to get updated comments
      await _loadPostData();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to delete comment: ${e.toString()}')),
        );
      }
    }
  }

  Future<void> _handleSaveImage() async {
    if (_post?.imageUrl == null || _post!.isVideo) return;

    try {
      // Show loading indicator
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Row(
              children: [
                SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
                SizedBox(width: 16),
                Text('Saving image...'),
              ],
            ),
            duration: Duration(seconds: 2),
          ),
        );
      }

      // Download the image
      final response = await http.get(Uri.parse(_post!.imageUrl!));
      if (response.statusCode != 200) {
        throw Exception('Failed to download image: ${response.statusCode}');
      }

      // Get temporary directory
      final tempDir = await getTemporaryDirectory();
      final file = File('${tempDir.path}/image_${DateTime.now().millisecondsSinceEpoch}.jpg');
      await file.writeAsBytes(response.bodyBytes);

      // Save to gallery
      await Gal.putImage(file.path);

      // Show success message
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Image saved to Photos!'),
            backgroundColor: Colors.green,
            duration: Duration(seconds: 2),
          ),
        );
      }

      // Clean up temporary file
      await file.delete();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to save image: ${e.toString()}'),
            backgroundColor: AppColors.coralPop,
            duration: const Duration(seconds: 3),
          ),
        );
      }
    }
  }

  Future<void> _handleDeletePost() async {
    if (_post?.id == null) return;

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
      await postProvider.deletePost(_post!.id!);
      
      // Navigate back after successful deletion
      if (mounted) {
        Navigator.pop(context);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to delete post: ${e.toString()}')),
        );
      }
    }
  }

  String _getBlockExplorerUrl(String txHash, int? chainId) {
    switch (chainId) {
      case 84532: // Base Sepolia
        return 'https://sepolia.basescan.org/tx/$txHash';
      case 1: // Ethereum Mainnet
        return 'https://etherscan.io/tx/$txHash';
      case 11155111: // Sepolia
        return 'https://sepolia.etherscan.io/tx/$txHash';
      default:
        return 'https://sepolia.basescan.org/tx/$txHash';
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return Scaffold(
        backgroundColor: AppColors.warmWhite,
        appBar: AppBar(
          backgroundColor: AppColors.warmWhite,
          elevation: 0,
          leading: IconButton(
            icon: Icon(Icons.arrow_back, color: AppColors.graphiteInk),
            onPressed: () => Navigator.pop(context),
          ),
        ),
        body: const Center(
          child: CircularProgressIndicator(),
        ),
      );
    }

    if (_error != null || _post == null) {
      return Scaffold(
        backgroundColor: AppColors.warmWhite,
        appBar: AppBar(
          backgroundColor: AppColors.warmWhite,
          elevation: 0,
          leading: IconButton(
            icon: Icon(Icons.arrow_back, color: AppColors.graphiteInk),
            onPressed: () => Navigator.pop(context),
          ),
        ),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(
                'Error: ${_error ?? "Post not found"}',
                style: TextStyle(color: AppColors.coralPop),
              ),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: _loadPostData,
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      );
    }

    final username = _post!.user?.username ?? _post!.user?.phoneNumber ?? 'Unknown';
    final userReactionEmoji = _getUserReactionEmoji();
    final authProvider = context.watch<AuthProvider>();
    final currentUserId = authProvider.user?.id;
    final isPostOwner = currentUserId != null && _post!.userId == currentUserId;

    return Scaffold(
      backgroundColor: AppColors.warmWhite,
      appBar: AppBar(
        backgroundColor: AppColors.warmWhite,
        elevation: 0,
        leading: IconButton(
          icon: Icon(Icons.arrow_back, color: AppColors.graphiteInk),
          onPressed: () => Navigator.pop(context),
        ),
        title: GestureDetector(
          onTap: () {
            if (_post!.userId != null) {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (context) => ProfileScreen(
                    userId: _post!.userId!,
                    user: _post!.user,
                    onNavigate: widget.onNavigate,
                  ),
                ),
              );
            }
          },
          child: Text(
            username,
            style: TextStyle(
              color: AppColors.graphiteInk,
              fontWeight: FontWeight.w600,
              fontSize: 18,
            ),
          ),
        ),
        actions: isPostOwner
            ? [
                if (_post!.imageUrl != null && !_post!.isVideo)
                  IconButton(
                    icon: const Icon(Icons.download, color: AppColors.softRealBlue),
                    onPressed: _handleSaveImage,
                    tooltip: 'Save to Photos',
                  ),
                IconButton(
                  icon: Icon(Icons.delete, color: AppColors.coralPop),
                  onPressed: _handleDeletePost,
                ),
              ]
            : null,
      ),
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Media - larger display
            if (_post!.imageUrl != null)
              _post!.isVideo
                  ? SizedBox(
                      height: 400,
                      child: VideoPlayerWidget(
                        videoUrl: _post!.imageUrl!,
                        autoPlay: false,
                        showControls: true,
                        fit: BoxFit.contain,
                      ),
                    )
                  : Image.network(
                      _post!.imageUrl!,
                      width: double.infinity,
                      fit: BoxFit.contain,
                      loadingBuilder: (context, child, loadingProgress) {
                        if (loadingProgress == null) return child;
                        return Container(
                          height: 400,
                          color: Colors.grey.shade200,
                          child: const Center(
                            child: CircularProgressIndicator(),
                          ),
                        );
                      },
                      errorBuilder: (context, error, stackTrace) {
                        return Container(
                          height: 400,
                          color: Colors.grey.shade200,
                          child: const Icon(Icons.error),
                        );
                      },
                    ),
            // Caption
            if (_post!.caption != null && _post!.caption!.isNotEmpty)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                child: Text(
                  _post!.caption!,
                  style: const TextStyle(fontSize: 16),
                ),
              ),
            // Reactions section
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Display all reactions
                  if (_post!.reactions != null && _post!.reactions!.isNotEmpty)
                    Wrap(
                      spacing: 8,
                      runSpacing: 4,
                      children: _post!.reactions!.map((reaction) {
                        return Chip(
                          label: Text(
                            '${reaction.emoji} ${reaction.count}',
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: reaction.userReacted ? FontWeight.w600 : FontWeight.normal,
                              color: reaction.userReacted ? AppColors.softRealBlue : AppColors.graphiteInk,
                            ),
                          ),
                          backgroundColor: reaction.userReacted 
                              ? AppColors.softRealBlue.withOpacity(0.1)
                              : Colors.grey.shade100,
                          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                          materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                        );
                      }).toList(),
                    ),
                  const SizedBox(height: 12),
                  // Reaction button
                  GestureDetector(
                    onTap: _isReacting ? null : () {
                      EmojiPicker.show(context, _handleReaction);
                    },
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                      decoration: BoxDecoration(
                        color: userReactionEmoji != null 
                            ? AppColors.softRealBlue.withOpacity(0.1)
                            : Colors.grey.shade100,
                        borderRadius: BorderRadius.circular(24),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            userReactionEmoji ?? '👍',
                            style: const TextStyle(fontSize: 20),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            userReactionEmoji != null ? 'Reacted' : 'React',
                            style: TextStyle(
                              fontSize: 16,
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
            ),
            const Divider(),
            // Comments section
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Comments (${_post!.comments?.length ?? 0})',
                    style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 12),
                  // Display all comments
                  if (_post!.comments != null && _post!.comments!.isNotEmpty)
                    ..._post!.comments!.map((comment) => _buildComment(comment)),
                  const SizedBox(height: 12),
                  // Comment input
                  _buildCommentInput(),
                ],
              ),
            ),
            const Divider(),
            // C2PA Manifest accordion
            if (_post!.manifestUri != null || _post!.manifestHash != null)
              _buildManifestAccordion(),
            // Blockchain Transaction accordion
            if (_blockchainTransaction != null)
              _buildBlockchainTransactionAccordion(),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }

  Widget _buildComment(Comment comment) {
    final authProvider = context.watch<AuthProvider>();
    final currentUserId = authProvider.user?.id;
    final canDelete = currentUserId != null && 
        (comment.userId == currentUserId || _post!.userId == currentUserId);
    
    final username = comment.user?.username ?? comment.user?.phoneNumber ?? 'Unknown';
    final profilePictureUrl = comment.user?.profilePictureUrl;

    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          CircleAvatar(
            radius: 16,
            backgroundColor: Colors.grey.shade300,
            backgroundImage: profilePictureUrl != null
                ? NetworkImage(profilePictureUrl)
                : null,
            child: profilePictureUrl == null
                ? Text(
                    username.isNotEmpty ? username[0].toUpperCase() : 'U',
                    style: TextStyle(
                      color: Colors.grey,
                      fontSize: 14,
                    ),
                  )
                : null,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.grey.shade100,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        username,
                        style: const TextStyle(
                          fontWeight: FontWeight.w600,
                          fontSize: 14,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        comment.text ?? '',
                        style: const TextStyle(fontSize: 15),
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
    return Row(
      children: [
        Expanded(
          child: TextField(
            controller: _commentController,
            decoration: InputDecoration(
              hintText: 'Add a comment...',
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(24),
                borderSide: BorderSide(color: Colors.grey.shade300),
              ),
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 20,
                vertical: 12,
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
                  width: 24,
                  height: 24,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Icon(Icons.send),
          onPressed: _isCommenting ? null : _handleSubmitComment,
          color: Colors.blue,
        ),
      ],
    );
  }

  Widget _buildManifestAccordion() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Card(
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side: BorderSide(color: Colors.grey.shade300),
        ),
        child: ExpansionTile(
          initiallyExpanded: false,
          leading: Icon(Icons.description, color: AppColors.softRealBlue),
          title: const Text(
            'C2PA Manifest',
            style: TextStyle(
              fontWeight: FontWeight.w600,
              fontSize: 16,
            ),
          ),
          children: [
            const Divider(height: 1),
            if (_post!.manifestHash != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 12.0),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(Icons.fingerprint, size: 18, color: Colors.grey.shade600),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Manifest Hash',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey.shade600,
                            ),
                          ),
                          const SizedBox(height: 4),
                          SelectableText(
                            _post!.manifestHash!,
                            style: const TextStyle(
                              fontSize: 12,
                              fontFamily: 'monospace',
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            if (_post!.manifestUri != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(Icons.link, size: 18, color: Colors.grey.shade600),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Manifest URI',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey.shade600,
                            ),
                          ),
                          const SizedBox(height: 4),
                          SelectableText(
                            _post!.manifestUri!,
                            style: const TextStyle(
                              fontSize: 12,
                              fontFamily: 'monospace',
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            if (_post!.assetId != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(Icons.tag, size: 18, color: Colors.grey.shade600),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Asset ID',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey.shade600,
                            ),
                          ),
                          const SizedBox(height: 4),
                          SelectableText(
                            _post!.assetId!,
                            style: const TextStyle(
                              fontSize: 14,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }

  Widget _buildBlockchainTransactionAccordion() {
    final tx = _blockchainTransaction!;
    final txHash = tx['txHash'] as String?;
    final status = tx['status'] as String?;
    final blockNumber = tx['blockNumber'] as int?;
    final chainId = tx['chainId'] as int?;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Card(
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side: BorderSide(color: Colors.grey.shade300),
        ),
        child: ExpansionTile(
          initiallyExpanded: false,
          leading: Icon(Icons.account_balance_wallet, color: Colors.blue.shade700),
          title: const Text(
            'Blockchain Transaction',
            style: TextStyle(
              fontWeight: FontWeight.w600,
              fontSize: 16,
            ),
          ),
          children: [
            const Divider(height: 1),
            // Status
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 12.0),
              child: Row(
                children: [
                  const Text(
                    'Status: ',
                    style: TextStyle(
                      fontSize: 14,
                      color: Colors.grey,
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: (status == 'confirmed')
                          ? Colors.green.shade100
                          : (status == 'pending')
                              ? Colors.orange.shade100
                              : Colors.grey.shade200,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      status ?? 'Unknown',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                        color: (status == 'confirmed')
                            ? Colors.green.shade800
                            : (status == 'pending')
                                ? Colors.orange.shade800
                                : Colors.grey.shade700,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            // Transaction Hash
            if (txHash != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(Icons.receipt, size: 18, color: Colors.grey.shade600),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Transaction Hash',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey.shade600,
                            ),
                          ),
                          const SizedBox(height: 4),
                          SelectableText(
                            txHash,
                            style: const TextStyle(
                              fontSize: 12,
                              fontFamily: 'monospace',
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            // Block Number
            if (blockNumber != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
                child: Row(
                  children: [
                    Icon(Icons.numbers, size: 18, color: Colors.grey.shade600),
                    const SizedBox(width: 8),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Block Number',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade600,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          blockNumber.toString(),
                          style: const TextStyle(
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            // Block Explorer Link
            if (txHash != null && chainId != null)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
                child: InkWell(
                  onTap: () async {
                    final url = _getBlockExplorerUrl(txHash, chainId);
                    final uri = Uri.parse(url);
                    if (await canLaunchUrl(uri)) {
                      await launchUrl(uri, mode: LaunchMode.externalApplication);
                    } else {
                      if (mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(
                            content: Text('Could not open block explorer: $url'),
                            backgroundColor: Theme.of(context).colorScheme.error,
                          ),
                        );
                      }
                    }
                  },
                  child: Row(
                    children: [
                      Icon(Icons.open_in_new, size: 18, color: Colors.blue.shade700),
                      const SizedBox(width: 8),
                      Text(
                        'View on Block Explorer',
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.blue.shade700,
                          decoration: TextDecoration.underline,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }
}
